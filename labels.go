package docs

import "strings"

// UIStrings contains the framework-owned labels rendered around a document.
// Content remains project-owned, but the surrounding chrome must be
// translatable for sites that publish more than one language.
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
}

var defaultUIStrings = UIStrings{
	Contents:          "Contents",
	Home:              "Home",
	OnThisPage:        "On this page",
	Search:            "Search",
	SearchPlaceholder: "Search documentation…",
	OpenSearch:        "Open documentation search",
	CloseSearch:       "Close search",
	OpenNavigation:    "Open navigation",
	EditPage:          "Edit this page ↗",
	LastUpdated:       "Updated",
	Published:         "Published",
	By:                "By",
	Language:          "Language",
	Version:           "Version",
}

// WithUIStrings overrides the framework-owned labels. Empty fields retain the
// English defaults, so callers can translate one label at a time.
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

func mergeUIStrings(base, override UIStrings) UIStrings {
	if strings.TrimSpace(base.Contents) == "" {
		base = defaultUIStrings
	}
	if strings.TrimSpace(override.Contents) != "" {
		base.Contents = strings.TrimSpace(override.Contents)
	}
	if strings.TrimSpace(override.Home) != "" {
		base.Home = strings.TrimSpace(override.Home)
	}
	if strings.TrimSpace(override.OnThisPage) != "" {
		base.OnThisPage = strings.TrimSpace(override.OnThisPage)
	}
	if strings.TrimSpace(override.Search) != "" {
		base.Search = strings.TrimSpace(override.Search)
	}
	if strings.TrimSpace(override.SearchPlaceholder) != "" {
		base.SearchPlaceholder = strings.TrimSpace(override.SearchPlaceholder)
	}
	if strings.TrimSpace(override.OpenSearch) != "" {
		base.OpenSearch = strings.TrimSpace(override.OpenSearch)
	}
	if strings.TrimSpace(override.CloseSearch) != "" {
		base.CloseSearch = strings.TrimSpace(override.CloseSearch)
	}
	if strings.TrimSpace(override.OpenNavigation) != "" {
		base.OpenNavigation = strings.TrimSpace(override.OpenNavigation)
	}
	if strings.TrimSpace(override.EditPage) != "" {
		base.EditPage = strings.TrimSpace(override.EditPage)
	}
	if strings.TrimSpace(override.LastUpdated) != "" {
		base.LastUpdated = strings.TrimSpace(override.LastUpdated)
	}
	if strings.TrimSpace(override.Published) != "" {
		base.Published = strings.TrimSpace(override.Published)
	}
	if strings.TrimSpace(override.By) != "" {
		base.By = strings.TrimSpace(override.By)
	}
	if strings.TrimSpace(override.Language) != "" {
		base.Language = strings.TrimSpace(override.Language)
	}
	if strings.TrimSpace(override.Version) != "" {
		base.Version = strings.TrimSpace(override.Version)
	}
	return base
}
