package docs

import (
	"errors"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

type fakeAssetPlugin struct {
	name   string
	files  map[string][]byte
	broken bool
}

func (p fakeAssetPlugin) Name() string          { return p.name }
func (p fakeAssetPlugin) Apply(_ *Router) error { return nil }
func (p fakeAssetPlugin) RuntimeAssets() (map[string][]byte, error) {
	if p.broken {
		return nil, errors.New("boom")
	}
	return p.files, nil
}

func assetRouter(t *testing.T, plugins ...Plugin) *Router {
	t.Helper()
	r := NewRouter(WithSiteName("Assets"))
	r.MustPage("/", PageConfig{Title: "Home", Description: "Landing", Order: 1, Source: "# Home\n"})
	for _, plugin := range plugins {
		if err := r.Use(plugin); err != nil {
			t.Fatalf("Use(%s) error = %v", plugin.Name(), err)
		}
	}
	return r
}

func TestRuntimeAssetsCollectsTheCoreFilesAndPluginContributions(t *testing.T) {
	r := assetRouter(t, fakeAssetPlugin{name: "charts", files: map[string][]byte{"charts.js": []byte("chart();")}})
	assets, err := r.RuntimeAssets("")
	if err != nil {
		t.Fatalf("RuntimeAssets() error = %v", err)
	}
	for _, want := range []string{"docs.js", "search.json", "manifest.json", "charts.js"} {
		if _, ok := assets[want]; !ok {
			t.Fatalf("RuntimeAssets() missing %q: %v", want, slices.Sorted(maps.Keys(assets)))
		}
	}
	if string(assets["charts.js"]) != "chart();" {
		t.Fatalf("plugin asset body = %q", assets["charts.js"])
	}
}

// A plugin must not be able to overwrite the docs runtime or escape the asset
// directory, since its file names are not the project's own.
func TestRuntimeAssetsRejectsCollidingAndEscapingPluginPaths(t *testing.T) {
	for name, files := range map[string]map[string][]byte{
		"collision": {"docs.js": []byte("hijacked")},
		"traversal": {"../../etc/passwd": []byte("x")},
		"absolute":  {"/docs.js": []byte("x")},
		"empty":     {"": []byte("x")},
	} {
		t.Run(name, func(t *testing.T) {
			r := assetRouter(t, fakeAssetPlugin{name: "bad", files: files})
			if _, err := r.RuntimeAssets(""); err == nil {
				t.Fatalf("RuntimeAssets() accepted %v", files)
			}
		})
	}
}

func TestRuntimeAssetsSurfacesPluginErrors(t *testing.T) {
	r := assetRouter(t, fakeAssetPlugin{name: "broken", broken: true})
	if _, err := r.RuntimeAssets(""); err == nil {
		t.Fatal("RuntimeAssets() hid a plugin error")
	}
}

func TestWriteRuntimeAssetsProducesTheExportLayout(t *testing.T) {
	r := assetRouter(t, fakeAssetPlugin{name: "charts", files: map[string][]byte{"charts.js": []byte("chart();")}})
	dir := t.TempDir()
	if err := r.WriteRuntimeAssets(dir, "", ""); err != nil {
		t.Fatalf("WriteRuntimeAssets() error = %v", err)
	}
	for _, want := range []string{"docs.js", "search.json", "manifest.json", "charts.js"} {
		path := filepath.Join(dir, "__fastr-docs", want)
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("export missing %s: %v", want, err)
		}
	}
}

// The asset prefix follows the configured search index path, so the two cannot
// disagree and leave the palette fetching a 404.
func TestAssetPrefixFollowsTheSearchIndexPath(t *testing.T) {
	if got := NewRouter().AssetPrefix(); got != DefaultAssetPrefix {
		t.Fatalf("AssetPrefix() = %q, want %q", got, DefaultAssetPrefix)
	}
	r := NewRouter(WithSearchIndexPath("/__manual/search.json"))
	if got := r.AssetPrefix(); got != "/__manual" {
		t.Fatalf("AssetPrefix() = %q, want /__manual", got)
	}

	dir := t.TempDir()
	r.MustPage("/", PageConfig{Title: "Home", Description: "Landing", Order: 1, Source: "# Home\n"})
	if err := r.WriteRuntimeAssets(dir, r.AssetPrefix(), ""); err != nil {
		t.Fatalf("WriteRuntimeAssets() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "__manual", "search.json")); err != nil {
		t.Fatalf("custom prefix export missing search.json: %v", err)
	}
}

func TestRuntimeAssetNamesAreSorted(t *testing.T) {
	r := assetRouter(t, fakeAssetPlugin{name: "charts", files: map[string][]byte{"charts.js": []byte("x")}})
	names, err := r.RuntimeAssetNames("")
	if err != nil {
		t.Fatalf("RuntimeAssetNames() error = %v", err)
	}
	if !slices.IsSorted(names) {
		t.Fatalf("RuntimeAssetNames() = %v, want sorted", names)
	}
}

func TestPageScriptsAreUniqueAndReachThePage(t *testing.T) {
	t.Run("duplicate page script names error", func(t *testing.T) {
		r := NewRouter(WithPageScript("poll", "a"), WithPageScript("poll", "b"))
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G"})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "poll") {
			t.Fatalf("Validate() = %v, want the duplicate script name flagged", err)
		}
	})
	t.Run("a per-page script hook exists", func(t *testing.T) {
		r := NewRouter(WithPageScript("analytics", "window.q=[];"))
		scripts := r.PageScripts()
		if len(scripts) != 1 || scripts[0].Name != "analytics" || scripts[0].JS != "window.q=[];" {
			t.Fatalf("PageScripts() = %+v", scripts)
		}
	})
}

func TestTwoPluginsMayNotClaimOneAssetName(t *testing.T) {
	r := assetRouter(t,
		fakeAssetPlugin{name: "one", files: map[string][]byte{"clash": []byte("one")}},
		fakeAssetPlugin{name: "two", files: map[string][]byte{"clash": []byte("two")}},
	)
	if _, err := r.RuntimeAssets(""); err == nil || !strings.Contains(err.Error(), "collides") {
		t.Fatal("two plugins writing different bytes under one asset name silently overwrite each other")
	}
}
