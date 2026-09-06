package docs

import (
	"sort"
	"strings"
)

// WithLocaleFallback keeps a partially translated site usable. Without it, a
// page tagged `locale: en` simply disappears from a build for another locale,
// so the French site is missing pages rather than showing English ones.
//
// With it, a route family that has no variant in the build locale falls back to
// the variant in defaultLocale. A family that does have a translation is
// unaffected, so a fallback never shadows real translated content. An empty
// defaultLocale falls back to routes that carry no locale at all.
func WithLocaleFallback(defaultLocale string) Option {
	return func(r *Router) {
		r.localeFallback = true
		r.fallbackLocale = strings.TrimSpace(defaultLocale)
	}
}

// effectiveLocale is the locale a route counts as for pairing purposes.
//
// A page declares its locale in front matter, but a site rarely marks its
// original pages: writing `locale: en` on thirty English files to get a
// language selector is busywork, and forgetting it fails silently, because a
// route in no locale pairs with nothing. So when the project has named a
// default locale, an unmarked route is treated as being in it.
//
// This is deliberately only about pairing. Which routes publish is decided by
// localeAllows, and that is left alone: a route with no declared locale is
// still served in every locale build.
func (r *Router) effectiveLocale(route *Route) string {
	if route == nil {
		return ""
	}
	if locale := normalizeLocale(route.Metadata.Locale); locale != "" {
		return locale
	}
	return normalizeLocale(r.fallbackLocale)
}

// LanguageFor returns the language of the document served at path.
//
// Hand it to app.WithLangFunc, and every page carries its own <html lang>
// instead of the host-wide one, which is what Pagefind reads to choose a
// language index and what a screen reader reads to choose pronunciation
// rules. A path with no route, such as a 404, takes the language of the
// deepest route above it, so a missed URL under /es still answers in Spanish.
// A path in no language at all is the host language.
func (r *Router) LanguageFor(path string) string {
	if r == nil {
		return "en"
	}
	path = normalizePath(path)
	var best *Route
	for _, route := range r.routes {
		if !pathActive(route.Path, path) {
			continue
		}
		if best == nil || len(route.Path) > len(best.Path) {
			best = route
		}
	}
	if locale := r.effectiveLocale(best); locale != "" {
		return locale
	}
	return r.Language()
}

// familyOf is the family a route pairs within: the routes that are the same
// page in other languages or versions.
//
// Two ways to declare it, and a site can use both. A path that mirrors the
// original's with a locale segment added, /es/docs/guide for /docs/guide,
// pairs by shape alone, which suits a tree translated folder for folder. A
// page whose path does not mirror, because its slug is translated or it lives
// somewhere else, names the original in TranslationOf and joins that family.
// The reference is followed through the target, so pointing at another
// translation lands in the same family as pointing at the original.
//
// A reference to a path nothing serves is kept as the family, so the page
// still pairs with nothing rather than with the wrong thing, and validation
// can name it.
func (r *Router) familyOf(route *Route) string {
	if route == nil {
		return ""
	}
	seen := map[*Route]bool{}
	for route.Metadata.TranslationOf != "" && !seen[route] {
		seen[route] = true
		ref := normalizePath(route.Metadata.TranslationOf)
		target := r.routes[ref]
		if target == nil {
			return strings.Trim(ref, "/")
		}
		route = target
	}
	return variantFamily(route)
}

// alternatesFor is the page's hreflang map: every declared alternate, plus one
// entry per published variant of its family in another language. A pair
// declared either way emits both directions, so nobody hand-maintains the
// alternates map for a site that already knows its translations.
func (r *Router) alternatesFor(route *Route) map[string]string {
	if route == nil {
		return nil
	}
	out := cloneStringMap(route.Metadata.Alternates)
	family := r.familyOf(route)
	self := r.effectiveLocale(route)
	for _, candidate := range r.Routes() {
		if candidate == route || !r.variantPublished(candidate) || r.familyOf(candidate) != family {
			continue
		}
		if candidate.Metadata.Version != route.Metadata.Version {
			continue
		}
		locale := r.effectiveLocale(candidate)
		if locale == "" || locale == self {
			continue
		}
		if out == nil {
			out = map[string]string{}
		}
		if _, declared := out[locale]; !declared {
			out[locale] = candidate.Path
		}
	}
	return out
}

// localeFamilyLocales maps each variant family to the locales present in it.
// It is rebuilt whenever routes change, and deliberately ignores the locale
// filter it exists to inform.
func (r *Router) localeFamilyLocales() map[string]map[string]bool {
	if r.localeFamilies != nil {
		return r.localeFamilies
	}
	families := make(map[string]map[string]bool)
	for _, route := range r.routes {
		if route.Kind == KindGroup || route.Hidden {
			continue
		}
		if route.Metadata.Draft && !r.includeDrafts && !route.includeDrafts {
			continue
		}
		if r.version != "" && route.Metadata.Version != "" && route.Metadata.Version != r.version {
			continue
		}
		locale := route.Metadata.Locale
		if locale == "" {
			continue
		}
		family := r.familyOf(route)
		if families[family] == nil {
			families[family] = make(map[string]bool)
		}
		families[family][locale] = true
	}
	r.localeFamilies = families
	return families
}

// localeAllows decides whether a route survives the build's locale filter.
func (r *Router) localeAllows(route *Route) bool {
	if r.locale == "" || route.Metadata.Locale == "" {
		return true
	}
	if route.Metadata.Locale == r.locale {
		return true
	}
	if !r.localeFallback {
		return false
	}
	// A translation exists, so the untranslated variant stays hidden.
	if r.localeFamilyLocales()[r.familyOf(route)][r.locale] {
		return false
	}
	if r.fallbackLocale == "" {
		// No default declared, so nothing stands in for the missing page.
		return false
	}
	return route.Metadata.Locale == r.fallbackLocale
}

// UntranslatedFamilies reports the route families that have no variant in the
// given locale, so a project can see its translation coverage rather than
// discovering gaps in production. Families are returned sorted.
func (r *Router) UntranslatedFamilies(locale string) []string {
	locale = strings.TrimSpace(locale)
	if r == nil || locale == "" {
		return nil
	}
	var missing []string
	for family, locales := range r.localeFamilyLocales() {
		if !locales[locale] {
			missing = append(missing, family)
		}
	}
	sort.Strings(missing)
	return missing
}

// LocaleCoverage reports, for every locale present in the route tree, which
// families are missing a translation.
func (r *Router) LocaleCoverage() map[string][]string {
	if r == nil {
		return nil
	}
	coverage := make(map[string][]string)
	for _, locale := range r.Locales() {
		if missing := r.UntranslatedFamilies(locale); len(missing) > 0 {
			coverage[locale] = missing
		}
	}
	return coverage
}
