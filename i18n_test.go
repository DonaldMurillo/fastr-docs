package docs

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

// localeSite registers one family translated into English and French, and a
// second family that only exists in English.
func localeSite(t *testing.T, options ...Option) *Router {
	t.Helper()
	r := NewRouter(options...)
	r.MustPage("/en/guide", PageConfig{
		Title: "Guide", Description: "English guide", Order: 1,
		Source: "# Guide\n", Metadata: ContentMetadata{Locale: "en"},
	})
	r.MustPage("/fr/guide", PageConfig{
		Title: "Guide", Description: "Guide français", Order: 2,
		Source: "# Guide\n", Metadata: ContentMetadata{Locale: "fr"},
	})
	r.MustPage("/en/only", PageConfig{
		Title: "Only", Description: "Untranslated", Order: 3,
		Source: "# Only\n", Metadata: ContentMetadata{Locale: "en"},
	})
	r.MustPage("/shared", PageConfig{
		Title: "Shared", Description: "No locale", Order: 4, Source: "# Shared\n",
	})
	return r
}

func publishedPaths(r *Router) []string {
	var paths []string
	for _, route := range r.PublishedRoutes() {
		paths = append(paths, route.Path)
	}
	slices.Sort(paths)
	return paths
}

// Without a fallback, an untranslated page simply vanishes. This is the
// behaviour that made i18n a gap, and it stays the default.
func TestWithoutFallbackAnUntranslatedPageDisappears(t *testing.T) {
	r := localeSite(t, WithLocale("fr"))
	got := publishedPaths(r)
	if slices.Contains(got, "/en/only") {
		t.Fatalf("English page leaked into the French build: %v", got)
	}
	if !slices.Contains(got, "/fr/guide") || !slices.Contains(got, "/shared") {
		t.Fatalf("French build lost its own pages: %v", got)
	}
}

func TestLocaleFallbackServesTheDefaultForUntranslatedPages(t *testing.T) {
	r := localeSite(t, WithLocale("fr"), WithLocaleFallback("en"))
	got := publishedPaths(r)

	// The untranslated family falls back rather than disappearing.
	if !slices.Contains(got, "/en/only") {
		t.Fatalf("fallback did not serve the untranslated page: %v", got)
	}
	// The translated family is unaffected: the French page wins and the
	// English one stays hidden, so a fallback never shadows a translation.
	if !slices.Contains(got, "/fr/guide") {
		t.Fatalf("translated page missing: %v", got)
	}
	if slices.Contains(got, "/en/guide") {
		t.Fatalf("fallback shadowed an existing translation: %v", got)
	}
}

func TestLocaleFallbackWithoutADefaultLocaleChangesNothing(t *testing.T) {
	r := localeSite(t, WithLocale("fr"), WithLocaleFallback(""))
	if slices.Contains(publishedPaths(r), "/en/only") {
		t.Fatal("fallback served a page with no default locale declared")
	}
}

func TestUntranslatedFamiliesReportsCoverage(t *testing.T) {
	r := localeSite(t)
	missing := r.UntranslatedFamilies("fr")
	if len(missing) != 1 || !strings.Contains(missing[0], "only") {
		t.Fatalf("UntranslatedFamilies(fr) = %v, want the untranslated family", missing)
	}
	if got := r.UntranslatedFamilies("en"); len(got) != 0 {
		t.Fatalf("UntranslatedFamilies(en) = %v, want none", got)
	}

	coverage := r.LocaleCoverage()
	if len(coverage["fr"]) != 1 {
		t.Fatalf("LocaleCoverage() = %v, want one French gap", coverage)
	}
	// The default locale appears with an empty gap list: the key set is
	// the language inventory, and "en" being complete is information.
	if gaps, ok := coverage["en"]; !ok || len(gaps) != 0 {
		t.Fatalf("LocaleCoverage()[en] = %v (present=%v), want an empty gap list", gaps, ok)
	}
}

// Coverage is advisory. A partially translated site is a normal state, and
// wiring it into Validate would fail every build that has not finished
// translating.
func TestTranslationGapsDoNotFailValidation(t *testing.T) {
	r := localeSite(t, WithLocale("fr"), WithLocaleFallback("en"))
	if err := r.Validate(); err != nil {
		t.Fatalf("Validate() failed on a partially translated site: %v", err)
	}
}

func TestUIStringsMergeOverridesNestedGroupsAndKeepsDefaults(t *testing.T) {
	r := NewRouter(WithUIStrings(UIStrings{
		Search: "Rechercher",
		Blog: BlogStrings{
			Archive:    "Archives",
			ResultsFor: "Résultats pour « %s »",
		},
		NotFound: NotFoundStrings{Heading: "Page introuvable"},
	}))
	labels := r.UIStrings()

	if labels.Search != "Rechercher" {
		t.Fatalf("Search = %q", labels.Search)
	}
	if labels.Blog.Archive != "Archives" || labels.NotFound.Heading != "Page introuvable" {
		t.Fatalf("nested override did not apply: %+v", labels)
	}
	// Untouched fields keep their defaults, so a project can translate one
	// label at a time.
	if labels.Contents != "Contents" || labels.Blog.Tags != "Tags" || labels.NotFound.Message != "This page does not exist." {
		t.Fatalf("merge clobbered untranslated defaults: %+v", labels)
	}
	if got := formatLabel(labels.Blog.ResultsFor, "go"); got != "Résultats pour « go »" {
		t.Fatalf("formatLabel() = %q", got)
	}
}

func TestDateFormatIsConfigurable(t *testing.T) {
	when := time.Date(2026, 3, 9, 12, 0, 0, 0, time.UTC)
	if got := defaultUIStrings.formatDate(when); got != "Mar 9, 2026" {
		t.Fatalf("default date = %q", got)
	}
	custom := mergeUIStrings(defaultUIStrings, UIStrings{DateFormat: "2006-01-02"})
	if got := custom.formatDate(when); got != "2026-03-09" {
		t.Fatalf("custom date = %q", got)
	}
}

// A translation that drops a placeholder must not produce Go's %!(EXTRA ...)
// noise in the middle of a page.
func TestFormatLabelToleratesTranslationsWithoutPlaceholders(t *testing.T) {
	// The labels come from a table so vet cannot treat this as a printf call
	// site; the whole point is passing an argument a translation dropped.
	for _, testCase := range []struct{ label, want string }{
		{"Archives", "Archives"},
		{"%s archive", "2026 archive"},
	} {
		if got := formatLabel(testCase.label, "2026"); got != testCase.want {
			t.Fatalf("formatLabel(%q) = %q, want %q", testCase.label, got, testCase.want)
		}
	}
}

func TestBlogChromeUsesTranslatedLabels(t *testing.T) {
	r := NewRouter(WithUIStrings(UIStrings{Blog: BlogStrings{
		Archive: "Archives", Tags: "Étiquettes", Authors: "Auteurs",
		AllPosts: "Tous les articles", NoPostsYet: "Aucun article pour le moment.",
	}}))
	dir := t.TempDir()
	for name, body := range map[string]string{
		"index.md": "# Blog\n\nUpdates.",
		"post.md":  "---\ntitle: A post\ndescription: One post.\ndate: 2026-02-01\nauthors: [Ada]\ntags: [release]\n---\n# A post\n\nBody.",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := r.MarkdownBlog("/blog", dir, BlogConfig{Title: "Blog", Description: "Posts", Order: 1}); err != nil {
		t.Fatalf("MarkdownBlog() error = %v", err)
	}
	if err := r.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	var titles []string
	for _, route := range r.Routes() {
		titles = append(titles, route.Title)
	}
	for _, want := range []string{"Archives", "Étiquettes", "Auteurs"} {
		if !slices.Contains(titles, want) {
			t.Fatalf("blog view titles %v missing %q", titles, want)
		}
	}
	for _, unwanted := range []string{"Archive", "Tags", "Authors"} {
		if slices.Contains(titles, unwanted) {
			t.Fatalf("blog kept the English title %q: %v", unwanted, titles)
		}
	}
}
