package docs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestMarkdownVersionedCollectionMountsCurrentAndArchivedPaths(t *testing.T) {
	dir := t.TempDir()
	write := func(version, name, body string) {
		t.Helper()
		path := filepath.Join(dir, version, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, version := range []string{"v1", "v2"} {
		write(version, "index.md", "# "+version+" docs\n\nOverview")
		write(version, "guide.md", "# Guide "+version+"\n\nRead this.")
	}

	r := NewRouter()
	if err := r.MarkdownVersionedCollection("/docs", dir, VersionedCollectionConfig{Current: "v2", Versions: []string{"v2", "v1"}, Collection: CollectionConfig{Offline: true}}); err != nil {
		t.Fatalf("MarkdownVersionedCollection() error = %v", err)
	}
	if err := r.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	paths := make(map[string]string)
	for _, route := range r.Routes() {
		paths[route.Path] = route.Metadata.Version
	}
	for path, wantVersion := range map[string]string{
		"/docs": "v2", "/docs/guide": "v2", "/docs/v1": "v1", "/docs/v1/guide": "v1",
	} {
		if got := paths[path]; got != wantVersion {
			t.Fatalf("route %s version = %q, want %q; routes = %#v", path, got, wantVersion, r.Routes())
		}
	}
	if got := strings.Join(r.Versions(), ","); got != "v1,v2" {
		t.Fatalf("Versions() = %q", got)
	}
	filtered := NewRouter(WithVersion("v1"))
	if err := filtered.MarkdownVersionedCollection("/docs", dir, VersionedCollectionConfig{Current: "v2"}); err != nil {
		t.Fatalf("filtered collection error = %v", err)
	}
	if got := len(filtered.PublishedRoutes()); got != 2 || filtered.PublishedRoutes()[0].Path != "/docs/v1" {
		t.Fatalf("filtered routes = %#v", filtered.PublishedRoutes())
	}
}

func TestMarkdownVersionedCollectionCreatesGroupsForSnapshotsWithoutIndex(t *testing.T) {
	dir := t.TempDir()
	for _, version := range []string{"v1", "v2"} {
		versionDir := filepath.Join(dir, version)
		if err := os.MkdirAll(versionDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if version == "v2" {
			if err := os.WriteFile(filepath.Join(versionDir, "index.md"), []byte("# Current\n\nCurrent docs."), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.WriteFile(filepath.Join(versionDir, "guide.md"), []byte("# "+version+" guide\n\nGuide."), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	r := NewRouter()
	if err := r.MarkdownVersionedCollection("/docs", dir, VersionedCollectionConfig{Current: "v2", Versions: []string{"v2", "v1"}}); err != nil {
		t.Fatal(err)
	}
	if err := r.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if route := r.routes["/docs/v1"]; route == nil || route.Kind != KindGroup || len(route.Children) != 1 {
		t.Fatalf("missing version group for index-less snapshot: %#v", route)
	}
}

func TestMarkdownVersionedCollectionFSSupportsEmbeddedSnapshots(t *testing.T) {
	content := fstest.MapFS{
		"versions/v1/index.md": &fstest.MapFile{Data: []byte("# Old docs\n\nOld")},
		"versions/v1/guide.md": &fstest.MapFile{Data: []byte("# Old guide\n\nOld")},
		"versions/v2/index.md": &fstest.MapFile{Data: []byte("# Current docs\n\nCurrent")},
	}
	r := NewRouter()
	if err := r.MarkdownVersionedCollectionFS("/docs", content, "versions", VersionedCollectionConfig{Current: "v2"}); err != nil {
		t.Fatalf("MarkdownVersionedCollectionFS() error = %v", err)
	}
	if got := len(r.Routes()); got != 3 {
		t.Fatalf("versioned FS routes = %d, want 3", got)
	}
	paths := map[string]bool{}
	for _, route := range r.Routes() {
		paths[route.Path] = true
	}
	for _, path := range []string{"/docs", "/docs/v1", "/docs/v1/guide"} {
		if !paths[path] {
			t.Fatalf("versioned FS routes missing %s: %#v", path, r.Routes())
		}
	}
}

func TestMarkdownVersionedCollectionRejectsAmbiguousConfiguration(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "v1"), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, cfg := range map[string]VersionedCollectionConfig{
		"missing current": {Versions: []string{"v1"}},
		"current absent":  {Current: "v2", Versions: []string{"v1"}},
		"duplicate":       {Current: "v1", Versions: []string{"v1", "v1"}},
		"nested":          {Current: "v1/x", Versions: []string{"v1/x"}},
	} {
		t.Run(name, func(t *testing.T) {
			r := NewRouter()
			if err := r.MarkdownVersionedCollection("/docs", dir, cfg); err == nil {
				t.Fatal("expected version configuration error")
			}
		})
	}
}

func TestVersionedCollectionRejectsStrayRootFiles(t *testing.T) {
	mapFS := fstest.MapFS{
		"readme.md":   &fstest.MapFile{Data: []byte("# stray")},
		"v1/index.md": &fstest.MapFile{Data: []byte("---\ntitle: V1\ndescription: d\n---\n\nBody.")},
	}
	r := NewRouter()
	err := r.MarkdownVersionedCollectionFS("/docs", mapFS, ".", VersionedCollectionConfig{Current: "v1"})
	if err == nil {
		t.Fatal("a stray file at the collection root becomes the index of every version")
	}
}

func TestVersionedCollectionsPairTheirVersionsInDrawers(t *testing.T) {
	mapFS := fstest.MapFS{
		"v1/index.md": &fstest.MapFile{Data: []byte("---\ntitle: One\ndescription: d\n---\n\nBody.")},
		"v2/index.md": &fstest.MapFile{Data: []byte("---\ntitle: Two\ndescription: d\n---\n\nBody.")},
	}
	r := NewRouter()
	if err := r.MarkdownVersionedCollectionFS("/docs", mapFS, ".", VersionedCollectionConfig{Current: "v2", Versions: []string{"v1", "v2"}}); err != nil {
		t.Fatal(err)
	}
	r.MustPage("/", PageConfig{Title: "Home", Description: "d", Order: 1, Source: "# H"})
	for _, name := range r.NavigationDrawerNames() {
		if strings.Contains(name, "v1") {
			return
		}
	}
	t.Fatalf("version drawers missing: %v", r.NavigationDrawerNames())
}
