package docs

import (
	"fmt"
	"reflect"
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
	Contents          string
	Home              string
	OnThisPage        string
	Search            string
	SearchPlaceholder string
	OpenSearch        string
	CloseSearch       string
	OpenNavigation    string
	EditPage          string
	LastUpdated       string
	Published         string
	By                string
	Language          string
	Version           string

	// DateFormat is a Go time layout used for published and updated dates.
	// The standard library has no CLDR data, so the layout is the project's
	// choice rather than something derived from the locale.
	DateFormat string
	// VersionDescription describes a generated version route. Takes the
	// version name.
	VersionDescription string

	Blog     BlogStrings
	NotFound NotFoundStrings
}

// BlogStrings are the labels for the publication surface: route titles for the
// generated views, and the chrome around posts.
type BlogStrings struct {
	Title                  string
	Navigation             string
	Breadcrumb             string
	AllPosts               string
	LatestPosts            string
	RecentPosts            string
	KeepReading            string
	NoPostsYet             string
	NothingClassified      string
	CollectionDescription  string
	Search                 string
	SearchDescription      string
	SearchPosts            string
	SearchThePublication   string
	ResultsFor             string
	NoPostsMatched         string
	MatchesSummary         string
	Archive                string
	ArchiveDescription     string
	ArchiveYear            string
	ArchiveYearDescription string
	Tag                    string
	Tags                   string
	TagsDescription        string
	ExploreTerms           string
	PostCount              string
	Publication            string
	ReadingTime            string
	Featured               string
	Latest                 string
	Feed                   string
	TopicsPrefix           string
	Lede                   string
	TagDescription         string
	Authors                string
	AuthorsDescription     string
	Author                 string
	AuthorDescription      string
	Page                   string
	PagedTitle             string
	PageDescription        string
	OlderPosts             string
	NewerPosts             string
	Pagination             string
	PostActions            string
	Share                  string
	ShareThisPost          string
	CopyLink               string
	LinkCopied             string
}

// NotFoundStrings are the labels on the branded 404 page.
type NotFoundStrings struct {
	Heading       string
	Message       string
	MessageForURL string
	BackTo        string
	SiteFallback  string
}

var defaultUIStrings = UIStrings{
	Contents:           "Contents",
	Home:               "Home",
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
		MatchesSummary:         "%d matches",
		Archive:                "Archive",
		ArchiveDescription:     "Browse every published post by year.",
		ArchiveYear:            "%s archive",
		ArchiveYearDescription: "Posts published in %s.",
		Tag:                    "Tag",
		Tags:                   "Tags",
		TagsDescription:        "Browse posts by topic.",
		ExploreTerms:           "Explore %s across the publication.",
		PostCount:              "%d posts",
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

func overlayStrings(target, override reflect.Value) {
	for i := range override.NumField() {
		switch field := override.Field(i); field.Kind() {
		case reflect.String:
			if value := strings.TrimSpace(field.String()); value != "" {
				target.Field(i).SetString(value)
			}
		case reflect.Struct:
			overlayStrings(target.Field(i), field)
		}
	}
}

// format applies a label's arguments, and falls back to the label itself when a
// translation dropped the placeholders.
func formatLabel(label string, args ...any) string {
	if !strings.ContainsRune(label, '%') {
		return label
	}
	return fmt.Sprintf(label, args...)
}

// formatDate renders a date with the configured layout.
func (u UIStrings) formatDate(value time.Time) string {
	layout := strings.TrimSpace(u.DateFormat)
	if layout == "" {
		layout = defaultUIStrings.DateFormat
	}
	return value.Format(layout)
}
