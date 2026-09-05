package docs

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func hasGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not on PATH")
	}
}

// The repository this test runs in has real history, so the dates come from
// actual commits rather than a fixture.
func TestGitMetadataFillsDatesAndEditLinksFromHistory(t *testing.T) {
	hasGit(t)
	source, err := filepath.Abs("README.md")
	if err != nil {
		t.Fatalf("resolve README: %v", err)
	}
	if _, err := os.Stat(source); err != nil {
		t.Skipf("no README to read history for: %v", err)
	}

	r := NewRouter(WithGitMetadata(GitMetadataConfig{
		RepoURL: "https://github.com/acme/docs",
		Branch:  "trunk",
	}))
	r.MustPage("/readme", PageConfig{
		Title: "Readme", Description: "From history", Order: 1, SourcePath: source,
	})
	if err := r.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	route := r.Routes()[0]
	if route.Metadata.DateModified == "" {
		t.Skip("this checkout has no commit touching README.md (shallow clone?)")
	}
	if _, err := time.Parse(time.RFC3339, route.Metadata.DateModified); err != nil {
		t.Fatalf("DateModified = %q, want RFC 3339: %v", route.Metadata.DateModified, err)
	}
	want := "https://github.com/acme/docs/edit/trunk/README.md"
	if route.Metadata.EditURL != want {
		t.Fatalf("EditURL = %q, want %q", route.Metadata.EditURL, want)
	}
}

func TestFrontMatterBeatsGitMetadata(t *testing.T) {
	hasGit(t)
	source, err := filepath.Abs("README.md")
	if err != nil {
		t.Fatalf("resolve README: %v", err)
	}
	r := NewRouter(WithGitMetadata(GitMetadataConfig{RepoURL: "https://github.com/acme/docs"}))
	r.MustPage("/readme", PageConfig{
		Title: "Readme", Description: "Declared", Order: 1, SourcePath: source,
		Metadata: ContentMetadata{DateModified: "2020-01-01T00:00:00Z", EditURL: "https://example.com/edit"},
	})
	if err := r.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	route := r.Routes()[0]
	if route.Metadata.DateModified != "2020-01-01T00:00:00Z" {
		t.Fatalf("git overwrote an authored date: %q", route.Metadata.DateModified)
	}
	if route.Metadata.EditURL != "https://example.com/edit" {
		t.Fatalf("git overwrote an authored edit URL: %q", route.Metadata.EditURL)
	}
}

// A docs build must not require version-control history to be present.
func TestGitMetadataDegradesSilentlyOutsideARepository(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "page.md")
	if err := os.WriteFile(source, []byte("# Page\n\nBody.\n"), 0o644); err != nil {
		t.Fatalf("write page: %v", err)
	}

	for name, config := range map[string]GitMetadataConfig{
		"no repository": {Dir: dir, RepoURL: "https://github.com/acme/docs"},
		"empty repository": func() GitMetadataConfig {
			if _, err := exec.LookPath("git"); err == nil {
				empty := t.TempDir()
				_ = exec.Command("git", "-C", empty, "init").Run()
				return GitMetadataConfig{Dir: empty, RepoURL: "https://github.com/acme/docs"}
			}
			return GitMetadataConfig{Dir: dir}
		}(),
	} {
		t.Run(name, func(t *testing.T) {
			r := NewRouter(WithGitMetadata(config))
			r.MustPage("/page", PageConfig{Title: "Page", Description: "Plain", Order: 1, SourcePath: source})
			if err := r.Validate(); err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
			route := r.Routes()[0]
			if route.Metadata.DateModified != "" {
				t.Fatalf("invented a date with no history: %q", route.Metadata.DateModified)
			}
		})
	}
}

func TestGitMetadataIgnoresEmbeddedAndOutsideSources(t *testing.T) {
	hasGit(t)
	outside := filepath.Join(t.TempDir(), "elsewhere.md")
	if err := os.WriteFile(outside, []byte("# Outside\n"), 0o644); err != nil {
		t.Fatalf("write page: %v", err)
	}
	r := NewRouter(WithGitMetadata(GitMetadataConfig{RepoURL: "https://github.com/acme/docs"}))
	// No SourcePath at all, as an embed.FS collection produces.
	r.MustPage("/embedded", PageConfig{Title: "Embedded", Description: "From FS", Order: 1, Source: "# Embedded\n"})
	r.MustPage("/outside", PageConfig{Title: "Outside", Description: "Elsewhere", Order: 2, SourcePath: outside})
	if err := r.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	for _, route := range r.Routes() {
		if route.Metadata.EditURL != "" {
			t.Fatalf("route %q got an edit link it cannot support: %q", route.Path, route.Metadata.EditURL)
		}
	}
}

func TestEditURLForOverridesTheGitHubShape(t *testing.T) {
	config := GitMetadataConfig{
		RepoURL:    "https://gitlab.com/acme/docs",
		EditURLFor: func(path string) string { return "https://gitlab.com/acme/docs/-/edit/main/" + path },
	}
	if got := config.editURL("content/index.md"); got != "https://gitlab.com/acme/docs/-/edit/main/content/index.md" {
		t.Fatalf("editURL() = %q", got)
	}
	if got := (&GitMetadataConfig{}).editURL("content/index.md"); got != "" {
		t.Fatalf("editURL() without a RepoURL = %q, want empty", got)
	}
}

func TestGitLogParsingTakesTheMostRecentCommitPerFile(t *testing.T) {
	hasGit(t)
	root, ok := gitRepositoryRoot(".", 10*time.Second)
	if !ok {
		t.Skip("not inside a git repository")
	}
	modified := gitLastModified(root, 30*time.Second)
	if len(modified) == 0 {
		t.Skip("no history in this checkout")
	}
	for path, when := range modified {
		if strings.HasPrefix(path, "\x1f") {
			t.Fatalf("a commit line was parsed as a file path: %q", path)
		}
		if _, err := time.Parse(time.RFC3339, when); err != nil {
			t.Fatalf("path %q has a non-RFC3339 date %q", path, when)
		}
	}
}
