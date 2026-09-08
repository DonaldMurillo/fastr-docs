package docs

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
)

// ExportManifest is the deployment-neutral route manifest emitted alongside
// static runtime assets. Hosting adapters can use it to build redirects,
// preloads, or edge metadata without parsing rendered HTML.
type ExportManifest struct {
	// Schema is the manifest schema version, for consumers that
	// check it before reading.
	Schema string `json:"schema"`
	// SiteName is the site the manifest describes.
	SiteName string `json:"siteName"`
	// BasePath is the prefix the export serves under, empty at the
	// root of a domain.
	BasePath string `json:"basePath,omitempty"`
	// SearchIndex is the URL of the search index the runtime fetches.
	SearchIndex string `json:"searchIndex"`
	// SearchBackend names the backend answering searches, json or
	// pagefind.
	SearchBackend SearchBackend `json:"searchBackend"`
	// SearchPath is the URL of the Pagefind bundle when that is the
	// backend.
	SearchPath string `json:"searchPath,omitempty"`
	// AssetsPrefix is where the runtime assets are served from.
	AssetsPrefix string `json:"assetsPrefix"`
	// Locales lists the languages the build serves.
	Locales []string `json:"locales,omitempty"`
	// Versions lists the route versions present in the tree.
	Versions []string `json:"versions,omitempty"`
	// Routes lists every published route with its language, version,
	// and fetch metadata.
	Routes []ManifestRoute `json:"routes"`
	// Redirects lists the source-to-target pairs the export writes.
	Redirects []ManifestRedirect `json:"redirects,omitempty"`
	// Drawers names the mounted navigation drawer widgets, one per
	// language plus one per blog collection, so an offline host can
	// precache exactly the chrome that exists.
	Drawers []string `json:"drawers,omitempty"`
}

// ManifestRoute is one published route in the export manifest, carrying the
// fields a static host or AI consumer needs: where it is, what it is called,
// which language and version it serves, and how it may be fetched.
type ManifestRoute struct {
	// ID is the stable route identifier.
	ID string `json:"id"`
	// Path is the route path relative to the deployment base.
	Path string `json:"path"`
	// Title is the route heading and nav label.
	Title string `json:"title"`
	// Description is the route's meta description.
	Description string `json:"description,omitempty"`
	// Kind says whether the route is a page, screen, or group.
	Kind RouteKind `json:"kind"`
	// Locale is the language the route serves, empty for unmarked
	// routes.
	Locale string `json:"locale,omitempty"`
	// Version is the route's version variant, empty when the site
	// is not versioned.
	Version string `json:"version,omitempty"`
	// Tags label the route for tag views.
	Tags []string `json:"tags,omitempty"`
	// Offline says the service worker precaches the route.
	Offline bool `json:"offline,omitempty"`
	// NoIndex asks search engines to skip the route.
	NoIndex bool `json:"noIndex,omitempty"`
	// Blog says the route belongs to a publication collection.
	Blog bool `json:"blog,omitempty"`
	// BlogIndex marks a collection's generated landing view.
	BlogIndex bool `json:"blogIndex,omitempty"`
	// DatePublished is the page's publish date as declared, for consumers
	// that sort or display recency.
	DatePublished string `json:"datePublished,omitempty"`
	// Alternates maps language tags to the route's translations, the
	// hreflang pairs derived from the route tree.
	Alternates map[string]string `json:"alternates,omitempty"`
}

// ManifestRedirect is one source-to-target pair of the export's redirects.
type ManifestRedirect struct {
	// From is the requested path; To is where it lands.
	From string `json:"from"`
	To   string `json:"to"`
}

// ExportManifestJSON builds a deterministic manifest for the current public
// route slice and deployment base path.
func (r *Router) ExportManifestJSON(basePath string) ([]byte, error) {
	if err := r.Validate(); err != nil {
		return nil, err
	}
	basePath = normalizeBasePath(basePath)
	manifest := ExportManifest{
		Schema: "fastr-docs/v1", SiteName: r.SiteName(), BasePath: basePath,
		SearchIndex:   basePath + r.SearchIndexPath(),
		SearchBackend: r.SearchBackend(),
		SearchPath:    basePath + r.PagefindPath(),
		AssetsPrefix:  basePath + "/assets/",
		Locales:       r.publishedLocales(),
		Versions:      r.Versions(),
		Drawers:       r.navigationDrawerNames(),
	}
	for _, route := range r.PublishedRoutes() {
		manifest.Routes = append(manifest.Routes, ManifestRoute{
			ID: route.ID, Path: basePath + route.Path, Title: route.Title,
			Description: route.Description, Kind: route.Kind,
			// The effective locale, not the declared one, so a consumer
			// filtering the manifest by language keeps the unmarked
			// pages of the default language the way search does.
			Locale: r.effectiveLocale(route), Version: route.Metadata.Version,
			Tags: cloneStrings(route.Tags), Offline: route.Offline,
			NoIndex: route.Metadata.NoIndex, Blog: route.Blog, BlogIndex: route.BlogIndex,
			DatePublished: route.Metadata.DatePublished, Alternates: r.alternatesFor(route),
		})
		for _, from := range route.Metadata.Redirects {
			if clean := safeRedirectPath(from); clean != "" {
				manifest.Redirects = append(manifest.Redirects, ManifestRedirect{From: basePath + clean, To: basePath + route.Path})
			}
		}
	}
	return json.MarshalIndent(manifest, "", "  ")
}

// WriteExportManifest writes ExportManifestJSON to a deployment artifact.
func (r *Router) WriteExportManifest(w io.Writer, basePath string) error {
	if w == nil {
		return errors.New("docs: export manifest writer is nil")
	}
	body, err := r.ExportManifestJSON(basePath)
	if err != nil {
		return err
	}
	_, err = w.Write(body)
	return err
}

func normalizeBasePath(base string) string {
	base = strings.TrimSpace(base)
	if base == "" || base == "/" {
		return ""
	}
	if !strings.HasPrefix(base, "/") {
		base = "/" + base
	}
	return strings.TrimRight(base, "/")
}
