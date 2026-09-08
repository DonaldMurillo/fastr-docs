package docs

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"
)

// UIStrings contains the framework-owned labels rendered around a document.
// Content remains project-owned, but every piece of surrounding chrome must be
// translatable for sites that publish more than one language.
//
// Fields holding a %s or %d are format strings. Translations may reorder the
// surrounding words but must keep the same verbs.
type UIStrings struct {
	// Contents titles the sidebar rail above the route tree. One word;
	// the English default is "Contents".
	Contents string
	// Home labels the sidebar link to the site root, rendered in the
	// language of the page being read. One word.
	Home string
	// SkipToContent labels the skip link that jumps the keyboard past the
	// navigation chrome. GoFastr renders the element with English text; the
	// runtime rewrites it from this label per page language.
	SkipToContent string
	// SectionHelp is the helper line under the drawer's section select,
	// wired as the select's accessible description.
	SectionHelp string
	// AnchorLabel names the copy-anchor button beside each heading, so a
	// screen reader announces an action rather than a glyph.
	AnchorLabel string
	// Sections labels the section select at the top of the mobile drawer,
	// where the whole site's navigation lives because the header tabs are
	// hidden below md.
	Sections string
	// OnThisPage heads the in-page heading rail on wide viewports and the
	// section select that replaces it below md.
	OnThisPage string
	// Search is the visible label of the header search trigger. One word.
	Search string
	// SearchPlaceholder is the command palette input placeholder. An
	// ellipsis is conventional; keep it if your language does.
	SearchPlaceholder string
	// OpenSearch is the aria-label of the search trigger. Imperative, not
	// rendered visually.
	OpenSearch string
	// CloseSearch is the aria-label of the palette close button.
	// Imperative, not rendered visually.
	CloseSearch string
	// OpenNavigation is the aria-label of the mobile navigation trigger,
	// the hamburger in the phone header. Imperative, not rendered visually.
	OpenNavigation string
	// EditPage is the text of the edit link under a document. Imperative;
	EditPage string
	// LastUpdated prefixes the modified date under a document. One word
	// or short phrase, capitalized as your language prefixes metadata.
	LastUpdated string
	// Published prefixes the date on a blog post. One word.
	Published string
	// By joins a post and its authors in the byline. Lowercase in English
	// ("By ..."); follow your language byline convention.
	By string
	// Language is the aria-label of the language selector. Not rendered
	// visually.
	Language string
	// Version is the aria-label of the version selector. Not rendered
	// visually.
	Version string
	// Previous and Next label the pager under a document. They carry their own
	// arrows because a translation may want them on the other side of the word.
	Previous string
	Next     string

	// DateFormat is a Go time layout used for published and updated dates.
	// The standard library has no CLDR data, so the layout is the project's
	// choice rather than something derived from the locale.
	DateFormat string
	// Months and ShortMonths name the months in the language, January first,
	// for a DateFormat that spells them out. Go prints month names in English
	// only, so "2 January 2006" gave "30 August 2026" on a Spanish page; with
	// twelve names here the same layout prints "30 agosto 2026". A numeric
	// layout needs neither. Twelve entries, or none.
	Months      []string
	ShortMonths []string
	// VersionDescription describes a generated version route. Takes the
	// version name.
	VersionDescription string

	// Blog nests the publication labels, resolved per collection through the
	// collection's own locale.
	Blog BlogStrings
	// NotFound nests the branded 404 labels.
	NotFound NotFoundStrings
}

// BlogStrings are the labels for the publication surface: route titles for the
// generated views, and the chrome around posts.
type BlogStrings struct {
	// Title names the collection in its sidebar rail. One word; the English
	// default is "Blog".
	Title string
	// Navigation is the aria-label of the collection sidebar. Not rendered
	// visually.
	Navigation string
	// Breadcrumb is the root crumb of the collection trail. One word.
	Breadcrumb string
	// AllPosts labels the link to the collection root.
	AllPosts string
	// LatestPosts heads the collection root when it lists the newest posts.
	LatestPosts string
	// RecentPosts titles the sidebar group of recent posts.
	RecentPosts string
	// KeepReading is the continue-reading link after a truncated post.
	// Imperative.
	KeepReading string
	// NoPostsYet is the line an empty collection renders. Full sentence.
	NoPostsYet string
	// NothingClassified is the line a tag or author page renders when it
	// has no posts. Full sentence.
	NothingClassified string
	// CollectionDescription is the meta description of generated views.
	// Format string taking the collection title as %s.
	CollectionDescription string
	// Search titles the collection search view. One word.
	Search string
	// SearchDescription is the meta description of the search view.
	// Format string taking the collection title as %s.
	SearchDescription string
	// SearchPosts is the visible heading over search results.
	SearchPosts string
	// SearchThePublication is the search input placeholder. Ellipsis is
	// conventional.
	SearchThePublication string
	// ResultsFor heads results with the query. Format string taking the query
	// as %s.
	ResultsFor string
	// NoPostsMatched is the empty-results line. Full sentence; the English
	// default names no query.
	NoPostsMatched string
	// MatchesSummary counts results for screen readers and the live filter.
	// Format string taking the count as %d; it also rides a data attribute
	// the client-side filter re-renders, so the %d must survive.
	MatchesSummary string
	// Archive titles the archive view. One word.
	Archive string
	// ArchiveDescription is the meta description of the archive view.
	// Format string taking the collection title as %s.
	ArchiveDescription string
	// ArchiveYear heads one year in the archive. Format string taking the
	// year as %s.
	ArchiveYear string
	// ArchiveYearDescription is the meta description of one year. Format
	// string taking the year as %d.
	ArchiveYearDescription string
	// Tag heads one tag page. One word.
	Tag string
	// Tags titles the tag index. One word.
	Tags string
	// TagsDescription is the meta description of the tag index. Format
	// string taking the collection title as %s.
	TagsDescription string
	// ExploreTerms is the lede of the tag index. Format string taking a
	// summary of the terms as %s.
	ExploreTerms string
	// PostCount counts a term's or author's posts. Format string taking the
	// count as %d.
	PostCount string
	// Publication names the publication in generated copy where "blog"
	// reads wrong for the project.
	Publication string
	// ReadingTime estimates a post's length. Format string taking the minute
	// count as %s.
	ReadingTime string
	// Featured labels the featured posts group. One word.
	Featured string
	// Latest labels the newest-posts group. One word.
	Latest string
	// Feed labels the link to the collection's RSS feed. One word.
	Feed string
	// TopicsPrefix introduces a post's topics for screen readers. Not
	// rendered visually; sentence fragment.
	TopicsPrefix string
	// Lede is the intro line of the collection's post list.
	Lede string
	// TagDescription is the meta description of one tag page. Format string
	// taking the term as %s.
	TagDescription string
	// Authors titles the author index. One word.
	Authors string
	// AuthorsDescription is the meta description of the author index.
	// Format string taking the collection title as %s.
	AuthorsDescription string
	// Author heads one author page. One word.
	Author string
	// AuthorDescription is the meta description of one author page. Format
	// string taking the name as %s.
	AuthorDescription string
	// Page is the word "Page" in a numbered page title. One word.
	Page string
	// PagedTitle titles a numbered page of the collection. Format string
	// taking the page number as %d.
	PagedTitle string
	// PageDescription is the meta description of a numbered page. Format
	// string taking the collection title as %s.
	PageDescription string
	// OlderPosts labels the link to older pages of the feed. Comparative.
	OlderPosts string
	// NewerPosts labels the link to newer pages of the feed. Comparative.
	NewerPosts string
	// Pagination is the aria-label of the pager between pages. Not
	// rendered visually.
	Pagination string
	// PostActions heads the share actions under a post.
	PostActions string
	// Share is the short share heading. One word.
	Share string
	// ShareThisPost is the full share heading. Imperative.
	ShareThisPost string
	// CopyLink is the copy-link button. Imperative.
	CopyLink string
	// LinkCopied is the transient confirmation after copying; announced to
	// screen readers. Past participle.
	LinkCopied string
}

// NotFoundStrings are the labels on the branded 404 page.
type NotFoundStrings struct {
	// Heading is the 404 page title. Short and capitalized.
	Heading string
	// Message is the static body line of the 404. Full sentence.
	Message string
	// MessageForURL is the body line naming the missed address. Format
	// string taking the requested path as %s.
	MessageForURL string
	// BackTo labels the way back into the site. Format string taking the
	// site name as %s.
	BackTo string
	// SiteFallback is the link text of the fallback to the site root.
	SiteFallback string
	// Search labels the way back into search from a dead end. One word.
	Search string
}

var defaultUIStrings = UIStrings{
	Contents:           "Contents",
	Previous:           "← Previous",
	Next:               "Next →",
	Home:               "Home",
	SkipToContent:      "Skip to main content",
	AnchorLabel:        "Copy link to this section",
	SectionHelp:        "Jump to a top-level section.",
	Sections:           "Sections",
	OnThisPage:         "On this page",
	Search:             "Search",
	SearchPlaceholder:  "Search documentation…",
	OpenSearch:         "Open documentation search",
	CloseSearch:        "Close search",
	OpenNavigation:     "Open navigation",
	EditPage:           "Edit this page ↗",
	LastUpdated:        "Updated",
	Published:          "Published",
	By:                 "By",
	Language:           "Language",
	Version:            "Version",
	DateFormat:         "Jan 2, 2006",
	VersionDescription: "Documentation for version %s.",
	Blog: BlogStrings{
		Title:                  "Blog",
		Navigation:             "Blog navigation",
		Breadcrumb:             "Breadcrumb",
		AllPosts:               "All posts",
		LatestPosts:            "Latest posts",
		RecentPosts:            "Recent posts",
		KeepReading:            "Keep reading",
		NoPostsYet:             "No posts are published here yet.",
		NothingClassified:      "Nothing has been classified yet.",
		CollectionDescription:  "Published posts in this collection.",
		Search:                 "Search",
		SearchDescription:      "Search the published posts.",
		SearchPosts:            "Search posts",
		SearchThePublication:   "Search the publication",
		ResultsFor:             "Results for “%s”",
		NoPostsMatched:         "No posts matched that search.",
		MatchesSummary:         "1 match|%d matches",
		Archive:                "Archive",
		ArchiveDescription:     "Browse every published post by year.",
		ArchiveYear:            "%s archive",
		ArchiveYearDescription: "Posts published in %s.",
		Tag:                    "Tag",
		Tags:                   "Tags",
		TagsDescription:        "Browse posts by topic.",
		ExploreTerms:           "Explore %s across the publication.",
		PostCount:              "1 post|%d posts",
		Publication:            "Publication",
		ReadingTime:            "%s min read",
		Featured:               "Featured",
		Latest:                 "Latest",
		Feed:                   "RSS",
		TopicsPrefix:           "Topics: ",
		Lede:                   "Find release notes, essays, and implementation updates.",
		TagDescription:         "Posts tagged %s.",
		Authors:                "Authors",
		AuthorsDescription:     "Browse posts by author.",
		Author:                 "Author",
		AuthorDescription:      "Posts by %s.",
		Page:                   "Page %d",
		PagedTitle:             "Blog · Page %d",
		PageDescription:        "More posts from %s.",
		OlderPosts:             "Older posts →",
		NewerPosts:             "← Newer posts",
		Pagination:             "Blog pagination",
		PostActions:            "Post actions",
		Share:                  "Share",
		ShareThisPost:          "Share this post",
		CopyLink:               "Copy link",
		LinkCopied:             "Link copied",
	},
	NotFound: NotFoundStrings{
		Heading:       "Page not found",
		Message:       "This page does not exist.",
		MessageForURL: "No page matches %s.",
		BackTo:        "Back to %s",
		SiteFallback:  "Documentation",
		Search:        "Search",
	},
}

// WithUIStrings overrides the framework-owned labels. Empty fields retain the
// English defaults, so a project can translate one label at a time.
func WithUIStrings(strings UIStrings) Option {
	return func(r *Router) {
		r.ui = mergeUIStrings(r.ui, strings)
	}
}

// UIStrings returns the effective labels used by this Router.
func (r *Router) UIStrings() UIStrings {
	if r == nil {
		return defaultUIStrings
	}
	return r.ui
}

// mergeUIStrings copies every non-empty field of override onto base, walking
// nested groups. Reflection keeps this from becoming one if-statement per
// label; there are more than seventy of them.
func mergeUIStrings(base, override UIStrings) UIStrings {
	if strings.TrimSpace(base.Contents) == "" {
		base = defaultUIStrings
	}
	merged := base
	overlayStrings(reflect.ValueOf(&merged).Elem(), reflect.ValueOf(override))
	return merged
}

// overlayUIStrings copies the non-empty fields of override onto base and
// nothing else.
//
// mergeUIStrings cannot be used for a locale. It substitutes the English
// defaults whenever its base looks uninitialized, which is right for a Router
// built without options and wrong here: it would fill every field of a partial
// translation with English and then overwrite the project's own Router-wide
// labels with them.
func overlayUIStrings(base, override UIStrings) UIStrings {
	merged := base
	overlayStrings(reflect.ValueOf(&merged).Elem(), reflect.ValueOf(override))
	return merged
}

func overlayStrings(target, override reflect.Value) {
	for i := range override.NumField() {
		switch field := override.Field(i); field.Kind() {
		case reflect.String:
			if value := strings.TrimSpace(field.String()); value != "" {
				target.Field(i).SetString(value)
			}
		case reflect.Struct:
			overlayStrings(target.Field(i), field)
		case reflect.Slice:
			if field.Len() > 0 {
				target.Field(i).Set(field)
			}
		}
	}
}

// format applies a label's arguments, and falls back to the label itself when a
// translation dropped the placeholders.
func formatLabel(label string, args ...any) string {
	if !strings.ContainsRune(label, '%') {
		return label
	}
	// A translation that dropped or added a placeholder must not reach the
	// page as Go's %!(EXTRA ...) or %!s(MISSING) artifacts: the label is
	// wrong either way, so it renders as written and the placeholder check
	// in ContentIssues names it at build time.
	if countFormatDirectives(label) != len(args) {
		return label
	}
	return fmt.Sprintf(label, args...)
}

// countFormatDirectives counts the printf verbs a label carries, ignoring
// the %% escape.
func countFormatDirectives(label string) int {
	count := 0
	for i := 0; i < len(label); i++ {
		if label[i] != '%' || i+1 >= len(label) {
			continue
		}
		if label[i+1] == '%' {
			i++
			continue
		}
		count++
	}
	return count
}

// Keys lists the label field names a translation can set, so tooling can
// diff a locale against the inventory without importing the struct layout.
func (UIStrings) Keys() []string {
	return uiStringKeys()
}

func uiStringKeys() []string {
	var keys []string
	var walk func(reflect.Type, string)
	walk = func(t reflect.Type, prefix string) {
		for i := range t.NumField() {
			field := t.Field(i)
			name := prefix + field.Name
			switch field.Type.Kind() {
			case reflect.String, reflect.Slice:
				keys = append(keys, name)
			case reflect.Struct:
				walk(field.Type, name+".")
			}
		}
	}
	walk(reflect.TypeOf(UIStrings{}), "")
	// Sorted, so diffing two builds of a locale is stable field by field.
	sort.Strings(keys)
	return keys
}

// formatCount renders a counted label. A label may carry the CLDR pipe form
// "1 post|%d posts" so a language with different plural rules can name both;
// without a pipe the label renders as written, count and all.
func formatCount(label string, n int) string {
	parts := strings.Split(label, "|")
	switch len(parts) {
	case 3:
		// "0 posts|1 post|%d posts": a zero form, a singular form, and
		// the plural. Languages whose zero differs from their plural
		// (French "aucun article") need all three.
		if n == 0 {
			return parts[0]
		}
		if n == 1 {
			return parts[1]
		}
		return formatLabel(parts[2], n)
	case 2:
		if n == 1 {
			return parts[0]
		}
		return formatLabel(parts[1], n)
	}
	return formatLabel(label, n)
}

// labelIssues checks the locale label sets a project registered: month
// calendars that cannot render every month, and translations that dropped a
// placeholder the default carries, which would silently lose an argument at
// render time.
func (r *Router) labelIssues() []ContentIssue {
	var issues []ContentIssue
	locales := make([]string, 0, len(r.localeUI))
	for locale := range r.localeUI {
		locales = append(locales, locale)
	}
	sort.Strings(locales)
	for _, locale := range locales {
		set := r.localeUI[locale]
		if len(set.Months) != 0 && len(set.Months) != 12 {
			issues = append(issues, ContentIssue{RoutePath: localeTagPath(locale),
				Message: fmt.Sprintf("locale %q translates %d month names, want 12", locale, len(set.Months))})
		}
		if len(set.ShortMonths) != 0 && len(set.ShortMonths) != 12 {
			issues = append(issues, ContentIssue{RoutePath: localeTagPath(locale),
				Message: fmt.Sprintf("locale %q translates %d short month names, want 12", locale, len(set.ShortMonths))})
		}
		if empty := emptyStringIn(set.Months); empty {
			issues = append(issues, ContentIssue{RoutePath: localeTagPath(locale),
				Message: fmt.Sprintf("locale %q leaves a month name empty", locale)})
		}
		if dup := duplicateFoldedString(set.Months); dup != "" {
			issues = append(issues, ContentIssue{RoutePath: localeTagPath(locale),
				Message: fmt.Sprintf("locale %q names two months %q once casing and spacing are ignored", locale, dup)})
		}
		if !localeShape.MatchString(locale) {
			issues = append(issues, ContentIssue{RoutePath: localeTagPath(locale),
				Message: fmt.Sprintf("locale %q is not a BCP 47 shaped language tag", locale)})
		}
		issues = append(issues, placeholderIssues(locale, reflect.ValueOf(r.localeUI[locale]), reflect.ValueOf(defaultUIStrings))...)
	}
	return issues
}

func localeTagPath(locale string) string {
	return "labels/" + locale
}

func duplicateString(values []string) string {
	seen := map[string]bool{}
	for _, value := range values {
		if value == "" || seen[value] {
			return value
		}
		seen[value] = true
	}
	return ""
}

// placeholderIssues walks two label structs field by field and reports any
// string whose verb count no longer matches the default's.
func placeholderIssues(locale string, got, want reflect.Value) []ContentIssue {
	return placeholderIssuesAt(locale, got, want, "")
}

// placeholderIssuesAt walks two label structs field by field and reports any
// string whose verb count no longer matches the default's. The path names the
// field with its group, "Blog.PostCount", because the label text itself may
// be arbitrary prose a tool cannot look up.
func placeholderIssuesAt(locale string, got, want reflect.Value, path string) []ContentIssue {
	var issues []ContentIssue
	if got.Type() != want.Type() {
		return issues
	}
	switch got.Kind() {
	case reflect.Struct:
		for i := range got.NumField() {
			if !got.Field(i).CanInterface() {
				continue
			}
			issues = append(issues, placeholderIssuesAt(locale, got.Field(i), want.Field(i), path+got.Type().Field(i).Name+".")...)
		}
	case reflect.String:
		gotLabel, wantLabel := got.String(), want.String()
		if gotLabel == "" {
			return issues
		}
		if countFormatDirectives(gotLabel) != countFormatDirectives(wantLabel) {
			name := strings.TrimSuffix(path, ".")
			issues = append(issues, ContentIssue{RoutePath: localeTagPath(locale),
				Message: fmt.Sprintf("locale %q label %s (%q) carries %d placeholders, the default carries %d",
					locale, name, gotLabel, countFormatDirectives(gotLabel), countFormatDirectives(wantLabel))})
		}
	}
	return issues
}

func emptyStringIn(values []string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return true
		}
	}
	return false
}

func duplicateFoldedString(values []string) string {
	seen := map[string]bool{}
	for _, value := range values {
		key := foldRunes(strings.TrimSpace(value))
		if key == "" {
			continue
		}
		if seen[key] {
			return value
		}
		seen[key] = true
	}
	return ""
}

// formatDate renders a date with the configured layout.
func (u UIStrings) formatDate(value time.Time) string {
	layout := strings.TrimSpace(u.DateFormat)
	if layout == "" {
		layout = defaultUIStrings.DateFormat
	}
	out := value.Format(layout)
	// Go names months in English only. Full names first, since "January"
	// contains "Jan".
	if len(u.Months) == 12 {
		for i, name := range u.Months {
			out = strings.ReplaceAll(out, time.Month(i+1).String(), name)
		}
	}
	if len(u.ShortMonths) == 12 {
		for i, name := range u.ShortMonths {
			out = strings.ReplaceAll(out, time.Month(i + 1).String()[:3], name)
		}
	}
	return out
}

// WithLocaleUIStrings translates the chrome for one locale.
//
// WithUIStrings sets labels for the whole Router, which is all a site needs
// when one build serves one language. A site that serves several locales from a
// single build needs a set per locale, or its Spanish pages come wrapped in
// English furniture: "Contents", "On this page", "Search".
//
// Fields left empty fall back to WithUIStrings, and then to the English
// defaults, so a locale can be translated a label at a time.
func WithLocaleUIStrings(locale string, strings UIStrings) Option {
	return func(r *Router) {
		locale = normalizeLocale(locale)
		if locale == "" {
			return
		}
		if r.localeUI == nil {
			r.localeUI = make(map[string]UIStrings)
		}
		r.localeUI[locale] = overlayUIStrings(r.localeUI[locale], strings)
	}
}

// WithLocaleNames gives locales the names a reader should see in the language
// selector. Without it the selector shows raw codes, and "es" is a worse label
// than "Español" for exactly the person who needs it.
//
// Go's standard library carries no locale display names, so a project supplies
// them rather than the framework pretending to know.
func WithLocaleNames(names map[string]string) Option {
	return func(r *Router) {
		if r.localeNames == nil {
			r.localeNames = make(map[string]string)
		}
		for locale, name := range names {
			if locale = normalizeLocale(locale); locale != "" {
				r.localeNames[locale] = strings.TrimSpace(name)
			}
		}
	}
}

// UIStringsForLocale returns the labels for one locale, merged over the
// Router-wide strings and the English defaults.
func (r *Router) UIStringsForLocale(locale string) UIStrings {
	if r == nil {
		return defaultUIStrings
	}
	locale = normalizeLocale(locale)
	if locale == "" {
		return r.ui
	}
	overrides, ok := r.localeUI[locale]
	// A regional page reads the primary language's furniture rather than
	// the Router-wide English one.
	if !ok {
		if primary := localePrimary(locale); primary != locale {
			if regional, regionalOK := r.localeUI[primary]; regionalOK {
				overrides = regional
				ok = true
			}
		}
	}
	if !ok {
		return r.ui
	}
	return overlayUIStrings(r.ui, overrides)
}

// LocaleName returns the display name for a locale, falling back to the code
// itself so a selector is never empty.
func (r *Router) LocaleName(locale string) string {
	locale = normalizeLocale(locale)
	if r == nil || locale == "" {
		return locale
	}
	if name := strings.TrimSpace(r.localeNames[locale]); name != "" {
		return name
	}
	// A regional reader gets the primary language's display name rather
	// than a bare code.
	if primary := localePrimary(locale); primary != locale {
		if name := strings.TrimSpace(r.localeNames[primary]); name != "" {
			return name
		}
	}
	return locale
}

// uiForRoute resolves the labels for a route that is already in hand, which is
// the common case inside page rendering.
func (r *Router) uiForRoute(route *Route) UIStrings {
	if r == nil {
		return defaultUIStrings
	}
	if route == nil || len(r.localeUI) == 0 {
		return r.ui
	}
	return r.UIStringsForLocale(route.Metadata.Locale)
}

// uiAt resolves the labels for the page being rendered. Chrome is rendered per
// route, so the route's own locale decides, which is what lets one build serve
// several languages with the right furniture around each.
func (r *Router) uiAt(currentPath string) UIStrings {
	if r == nil {
		return defaultUIStrings
	}
	if len(r.localeUI) == 0 {
		return r.ui
	}
	route := r.routeAtPath(currentPath)
	if route == nil {
		return r.ui
	}
	return r.UIStringsForLocale(route.Metadata.Locale)
}

func normalizeLocale(locale string) string {
	return strings.ToLower(strings.TrimSpace(locale))
}
