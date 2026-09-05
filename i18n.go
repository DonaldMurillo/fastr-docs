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
		family := variantFamily(route)
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
	if r.localeFamilyLocales()[variantFamily(route)][r.locale] {
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
