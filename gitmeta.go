package docs

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// GitMetadataConfig fills in edit links and last-updated dates from the
// repository, so neither has to be typed into every page's front matter and
// then left to rot.
//
// Front matter always wins. A page that sets edit_url or last_updated keeps
// what it declared.
type GitMetadataConfig struct {
	// Dir is where git runs. Empty uses the working directory.
	Dir string
	// RepoURL is the repository's web URL, such as
	// https://github.com/acme/docs. Empty leaves edit links alone.
	RepoURL string
	// Branch is the branch edit links point at. Empty uses "main".
	Branch string
	// EditURLFor replaces the built-in GitHub-style link for hosts that shape
	// their edit URLs differently. It receives the file's repository-relative,
	// slash-separated path.
	EditURLFor func(path string) string
	// Timeout caps the single git invocation. Empty uses 10 seconds.
	Timeout time.Duration
}

// WithGitMetadata derives last-updated dates, and optionally edit links, from
// git history.
//
// It reads one `git log` for the whole repository rather than one per page, and
// it degrades silently: no git binary, no repository, or a shallow clone leaves
// every page exactly as authored. A documentation build must not depend on
// version-control history being present.
func WithGitMetadata(config GitMetadataConfig) Option {
	return func(r *Router) {
		r.gitMeta = &config
	}
}

// resolveGitMetadata fills in metadata for every disk-backed page. It runs once,
// from Validate, which Mount also calls.
func (r *Router) resolveGitMetadata() {
	if r == nil || r.gitMeta == nil || r.gitMetaResolved {
		return
	}
	r.gitMetaResolved = true

	dir := r.gitMeta.Dir
	if strings.TrimSpace(dir) == "" {
		dir = "."
	}
	root, ok := gitRepositoryRoot(dir, r.gitMeta.timeout())
	if !ok {
		return
	}
	modified := gitLastModified(root, r.gitMeta.timeout())
	if len(modified) == 0 {
		return
	}

	for _, route := range r.routes {
		if route.page == nil || strings.TrimSpace(route.page.SourcePath) == "" {
			continue
		}
		relative, ok := repositoryRelativePath(root, route.page.SourcePath)
		if !ok {
			continue
		}
		if route.Metadata.DateModified == "" {
			if when, ok := modified[relative]; ok {
				route.Metadata.DateModified = when
			}
		}
		if route.Metadata.EditURL == "" {
			if link := r.gitMeta.editURL(relative); link != "" {
				route.Metadata.EditURL = link
			}
		}
	}
}

func (c *GitMetadataConfig) timeout() time.Duration {
	if c == nil || c.Timeout <= 0 {
		return 10 * time.Second
	}
	return c.Timeout
}

func (c *GitMetadataConfig) editURL(relative string) string {
	if c == nil {
		return ""
	}
	if c.EditURLFor != nil {
		return c.EditURLFor(relative)
	}
	base := strings.TrimRight(strings.TrimSpace(c.RepoURL), "/")
	if base == "" {
		return ""
	}
	branch := strings.TrimSpace(c.Branch)
	if branch == "" {
		branch = "main"
	}
	return base + "/edit/" + branch + "/" + relative
}

func gitRepositoryRoot(dir string, timeout time.Duration) (string, bool) {
	out, ok := runGit(dir, timeout, "rev-parse", "--show-toplevel")
	if !ok {
		return "", false
	}
	root := strings.TrimSpace(out)
	if root == "" {
		return "", false
	}
	return root, true
}

// gitLastModified maps every tracked file to the committer date of the most
// recent commit that touched it, from a single log walk.
func gitLastModified(root string, timeout time.Duration) map[string]string {
	// The unit separator prefixes each commit's date so a file named like a
	// timestamp cannot be mistaken for one.
	out, ok := runGit(root, timeout, "log", "--pretty=format:\x1f%cI", "--name-only")
	if !ok {
		return nil
	}
	modified := make(map[string]string)
	current := ""
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.HasPrefix(line, "\x1f") {
			current = strings.TrimPrefix(line, "\x1f")
			continue
		}
		if line == "" || current == "" {
			continue
		}
		// Commits are newest first, so the first sighting of a path is its most
		// recent change.
		if _, seen := modified[line]; !seen {
			modified[line] = current
		}
	}
	return modified
}

func repositoryRelativePath(root, sourcePath string) (string, bool) {
	absolute, err := filepath.Abs(sourcePath)
	if err != nil {
		return "", false
	}
	relative, err := filepath.Rel(root, absolute)
	if err != nil {
		return "", false
	}
	relative = filepath.ToSlash(relative)
	if strings.HasPrefix(relative, "../") {
		// Outside the repository, so git knows nothing about it.
		return "", false
	}
	return relative, true
}

func runGit(dir string, timeout time.Duration, args ...string) (string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	command := exec.CommandContext(ctx, "git", args...)
	command.Dir = dir
	out, err := command.Output()
	if err != nil {
		return "", false
	}
	return string(out), true
}
