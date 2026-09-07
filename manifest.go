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
	Schema        string             `json:"schema"`
	SiteName      string             `json:"siteName"`
	BasePath      string             `json:"basePath,omitempty"`
	SearchIndex   string             `json:"searchIndex"`
	SearchBackend SearchBackend      `json:"searchBackend"`
	SearchPath    string             `json:"searchPath,omitempty"`
	AssetsPrefix  string             `json:"assetsPrefix"`
	Locales       []string           `json:"locales,omitempty"`
	Versions      []string           `json:"versions,omitempty"`
	Routes        []ManifestRoute    `json:"routes"`
	Redirects     []ManifestRedirect `json:"redirects,omitempty"`
}

// ManifestRoute is one published route in the export manifest, carrying the
// fields a static host or AI consumer needs: where it is, what it is called,
// which language and version it serves, and how it may be fetched.
type ManifestRoute struct {
	ID          string    `json:"id"`
	Path        string    `json:"path"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Kind        RouteKind `json:"kind"`
	Locale      string    `json:"locale,omitempty"`
	Version     string    `json:"version,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	Offline     bool      `json:"offline,omitempty"`
	NoIndex     bool      `json:"noIndex,omitempty"`
	Blog        bool      `json:"blog,omitempty"`
	BlogIndex   bool      `json:"blogIndex,omitempty"`
}

// ManifestRedirect is one source-to-target pair of the export's redirects.
type ManifestRedirect struct {
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
		SearchIndex:   basePath + "/__fastr-docs/search.json",
		SearchBackend: r.SearchBackend(),
		SearchPath:    basePath + r.PagefindPath(),
		AssetsPrefix:  basePath + "/assets/",
		Locales:       r.Locales(),
		Versions:      r.Versions(),
	}
	for _, route := range r.PublishedRoutes() {
		manifest.Routes = append(manifest.Routes, ManifestRoute{
			ID: route.ID, Path: basePath + route.Path, Title: route.Title,
			Description: route.Description, Kind: route.Kind,
			Locale: route.Metadata.Locale, Version: route.Metadata.Version,
			Tags: cloneStrings(route.Tags), Offline: route.Offline,
			NoIndex: route.Metadata.NoIndex, Blog: route.Blog, BlogIndex: route.BlogIndex,
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
