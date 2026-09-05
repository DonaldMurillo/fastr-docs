package docs

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"testing/fstest"

	gofastrRouter "github.com/DonaldMurillo/gofastr/core/router"
	"github.com/DonaldMurillo/gofastr/core/static"
)

// DefaultAssetPrefix is where a generated project serves the docs runtime,
// search index, and plugin assets.
const DefaultAssetPrefix = "/__fastr-docs"

// RuntimeAssetPlugin contributes files to the docs asset prefix. Implement it
// so a project never has to hand-wire a plugin's browser runtime into its own
// main.go, for both serving and static export.
type RuntimeAssetPlugin interface {
	Plugin
	// RuntimeAssets returns file names relative to the asset prefix.
	RuntimeAssets() (map[string][]byte, error)
}

// RuntimeAssets collects every file that belongs under the docs asset prefix:
// the browser runtime, the search index, the export manifest, and whatever each
// registered plugin contributes.
//
// Before this existed, each of those was written out by hand in the generated
// main.go, once for serving and again for export, so adding any asset-bearing
// plugin meant editing every project that wanted it.
func (r *Router) RuntimeAssets(exportBase string) (map[string][]byte, error) {
	if r == nil {
		return nil, errors.New("docs: RuntimeAssets requires a Router")
	}
	searchIndex, err := r.SearchIndexJSON()
	if err != nil {
		return nil, fmt.Errorf("docs: build search index: %w", err)
	}
	manifest, err := r.ExportManifestJSON(exportBase)
	if err != nil {
		return nil, fmt.Errorf("docs: build export manifest: %w", err)
	}
	assets := map[string][]byte{
		"docs.js":       []byte(RuntimeJS()),
		"search.json":   searchIndex,
		"manifest.json": manifest,
	}
	for _, plugin := range r.plugins {
		contributor, ok := plugin.(RuntimeAssetPlugin)
		if !ok {
			continue
		}
		files, err := contributor.RuntimeAssets()
		if err != nil {
			return nil, fmt.Errorf("docs: plugin %q assets: %w", plugin.Name(), err)
		}
		for name, body := range files {
			clean := path.Clean(strings.TrimPrefix(strings.TrimSpace(name), "/"))
			if clean == "" || clean == "." || strings.HasPrefix(clean, "..") {
				return nil, fmt.Errorf("docs: plugin %q contributed an unsafe asset path %q", plugin.Name(), name)
			}
			if _, taken := assets[clean]; taken {
				return nil, fmt.Errorf("docs: plugin %q asset %q collides with an existing file", plugin.Name(), clean)
			}
			assets[clean] = body
		}
	}
	return assets, nil
}

// RuntimeAssetNames lists the collected asset names, sorted. Useful for
// precache lists and for asserting an export is complete.
func (r *Router) RuntimeAssetNames(exportBase string) ([]string, error) {
	assets, err := r.RuntimeAssets(exportBase)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(assets))
	for name := range assets {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

// MountRuntimeAssets serves the collected assets under prefix. An empty prefix
// uses DefaultAssetPrefix.
func (r *Router) MountRuntimeAssets(httpRouter *gofastrRouter.Router, prefix, exportBase string) error {
	if httpRouter == nil {
		return errors.New("docs: MountRuntimeAssets requires a GoFastr router")
	}
	prefix = normalizeAssetPrefix(prefix)
	assets, err := r.RuntimeAssets(exportBase)
	if err != nil {
		return err
	}
	files := make(fstest.MapFS, len(assets))
	for name, body := range assets {
		files[name] = &fstest.MapFile{Data: body}
	}
	httpRouter.Get(prefix+"/{path...}", static.Handler(static.Config{FS: files, Prefix: prefix}))
	return nil
}

// WriteRuntimeAssets writes the collected assets into a static export.
func (r *Router) WriteRuntimeAssets(dir, prefix, exportBase string) error {
	if strings.TrimSpace(dir) == "" {
		return errors.New("docs: WriteRuntimeAssets requires an output directory")
	}
	assets, err := r.RuntimeAssets(exportBase)
	if err != nil {
		return err
	}
	root := filepath.Join(dir, filepath.FromSlash(strings.TrimPrefix(normalizeAssetPrefix(prefix), "/")))
	for name, body := range assets {
		target := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return fmt.Errorf("docs: create %s: %w", filepath.Dir(target), err)
		}
		if err := os.WriteFile(target, body, 0o644); err != nil {
			return fmt.Errorf("docs: write %s: %w", target, err)
		}
	}
	return nil
}

// AssetPrefix returns the prefix the runtime assets are served from, matching
// the configured search index path so the two cannot disagree.
func (r *Router) AssetPrefix() string {
	if r == nil {
		return DefaultAssetPrefix
	}
	dir := path.Dir(r.SearchIndexPath())
	if dir == "" || dir == "." || dir == "/" {
		return DefaultAssetPrefix
	}
	return dir
}

func normalizeAssetPrefix(prefix string) string {
	prefix = strings.TrimRight(strings.TrimSpace(prefix), "/")
	if prefix == "" {
		return DefaultAssetPrefix
	}
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	return prefix
}

// PageScriptPlugin narrows which of a plugin's runtime assets belong in a
// <script> tag on every page.
//
// Without this, a project picks page scripts by file extension, and any .js a
// plugin contributes is loaded on every page. That is wrong for a plugin whose
// bundle is meant for somewhere else: the Mermaid frame bundle is several
// megabytes and is loaded by the frame document, never by the page.
type PageScriptPlugin interface {
	RuntimeAssetPlugin
	// PageScripts returns the subset of RuntimeAssets names to load on every
	// page. An empty slice means the plugin has no page script at all.
	PageScripts() []string
}

// RuntimeScriptNames lists the assets that belong in a page <script> tag,
// sorted, relative to the asset prefix.
//
// A plugin that does not implement PageScriptPlugin contributes every .js file
// it owns, which is what a single-file browser runtime wants.
func (r *Router) RuntimeScriptNames(exportBase string) ([]string, error) {
	if r == nil {
		return nil, errors.New("docs: RuntimeScriptNames requires a Router")
	}
	scripts := []string{"docs.js"}
	for _, plugin := range r.plugins {
		contributor, ok := plugin.(RuntimeAssetPlugin)
		if !ok {
			continue
		}
		if narrowed, ok := plugin.(PageScriptPlugin); ok {
			for _, name := range narrowed.PageScripts() {
				scripts = append(scripts, path.Clean(strings.TrimPrefix(strings.TrimSpace(name), "/")))
			}
			continue
		}
		files, err := contributor.RuntimeAssets()
		if err != nil {
			return nil, fmt.Errorf("docs: plugin %q assets: %w", plugin.Name(), err)
		}
		for name := range files {
			if strings.HasSuffix(name, ".js") {
				scripts = append(scripts, path.Clean(strings.TrimPrefix(strings.TrimSpace(name), "/")))
			}
		}
	}
	sort.Strings(scripts)
	return scripts, nil
}
