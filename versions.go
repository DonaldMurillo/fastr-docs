package docs

import (
	"fmt"
	"io/fs"
	"os"
	pathpkg "path"
	"path/filepath"
	"sort"
	"strings"
)

// VersionedCollectionConfig describes a content directory with one child
// directory per documentation version. The current version is mounted at the
// collection prefix; older versions receive a /<version>/ path segment.
//
// For example, a collection rooted at content/versions with Current "v2"
// mounts content/versions/v2/guide.md at /docs/guide and
// content/versions/v1/guide.md at /docs/v1/guide. Versions may be supplied to
// make the source contract explicit. When omitted, immediate child
// directories are discovered in sorted order.
type VersionedCollectionConfig struct {
	Current    string
	Versions   []string
	Collection CollectionConfig
}

// MarkdownVersionedCollection registers disk-backed Markdown snapshots from
// one directory per version. It preserves the ordinary MarkdownCollection
// behavior for front matter, locale prefixes, drafts, search, and validation.
func (r *Router) MarkdownVersionedCollection(prefix, dir string, cfg VersionedCollectionConfig) error {
	if r == nil {
		return fmt.Errorf("docs: MarkdownVersionedCollection requires a Router")
	}
	if strings.TrimSpace(dir) == "" {
		return fmt.Errorf("docs: MarkdownVersionedCollection requires a directory")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("docs: scan version directory %q: %w", dir, err)
	}
	available := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			available = append(available, entry.Name())
		}
	}
	versions, err := normalizeVersionedCollection(cfg, available)
	if err != nil {
		return err
	}
	for index, version := range versions {
		versionDir := filepath.Join(dir, version)
		info, err := os.Stat(versionDir)
		if err != nil {
			return fmt.Errorf("docs: version %q: %w", version, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("docs: version %q is not a directory", version)
		}
		if !markdownDirectoryHasIndex(versionDir) {
			groupPath := joinPath(prefix, version)
			if _, err := r.Group(groupPath, GroupConfig{
				Title:       version,
				Description: formatLabel(r.UIStrings().VersionDescription, version),
				Order:       r.nextChildOrder(groupPath, cfg.Collection.OrderStart+index),
			}); err != nil {
				return fmt.Errorf("docs: register version group %q: %w", version, err)
			}
		}
		collection := versionedCollectionConfig(cfg.Collection, cfg.Current, version)
		if err := r.markdownCollection(prefix, versionDir, collection, version, version != strings.TrimSpace(cfg.Current)); err != nil {
			return fmt.Errorf("docs: register version %q: %w", version, err)
		}
	}
	return nil
}

// MarkdownVersionedCollectionFS registers versioned Markdown snapshots from
// an fs.FS, including embed.FS. The root has the same child-directory layout
// as MarkdownVersionedCollection.
func (r *Router) MarkdownVersionedCollectionFS(prefix string, content fs.FS, root string, cfg VersionedCollectionConfig) error {
	if r == nil {
		return fmt.Errorf("docs: MarkdownVersionedCollectionFS requires a Router")
	}
	if content == nil {
		return fmt.Errorf("docs: MarkdownVersionedCollectionFS requires an fs.FS")
	}
	root = strings.TrimSpace(root)
	if root == "" {
		root = "."
	}
	if !fs.ValidPath(root) {
		return fmt.Errorf("docs: MarkdownVersionedCollectionFS root %q is not a valid fs path", root)
	}
	entries, err := fs.ReadDir(content, root)
	if err != nil {
		return fmt.Errorf("docs: scan version FS directory %q: %w", root, err)
	}
	available := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			available = append(available, entry.Name())
		}
	}
	versions, err := normalizeVersionedCollection(cfg, available)
	if err != nil {
		return err
	}
	for index, version := range versions {
		versionRoot := pathpkg.Join(root, version)
		info, err := fs.Stat(content, versionRoot)
		if err != nil {
			return fmt.Errorf("docs: version %q: %w", version, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("docs: version %q is not a directory", version)
		}
		if !versionFSHasIndex(content, versionRoot) {
			groupPath := joinPath(prefix, version)
			if _, err := r.Group(groupPath, GroupConfig{
				Title:       version,
				Description: formatLabel(r.UIStrings().VersionDescription, version),
				Order:       r.nextChildOrder(groupPath, cfg.Collection.OrderStart+index),
			}); err != nil {
				return fmt.Errorf("docs: register version group %q: %w", version, err)
			}
		}
		collection := versionedCollectionConfig(cfg.Collection, cfg.Current, version)
		if err := r.markdownCollectionFS(prefix, content, versionRoot, collection, version, version != strings.TrimSpace(cfg.Current)); err != nil {
			return fmt.Errorf("docs: register version %q: %w", version, err)
		}
	}
	return nil
}

func normalizeVersionedCollection(cfg VersionedCollectionConfig, available []string) ([]string, error) {
	current := strings.TrimSpace(cfg.Current)
	if current == "" {
		return nil, fmt.Errorf("docs: versioned collection requires Current")
	}
	versions := append([]string(nil), cfg.Versions...)
	if len(versions) == 0 {
		versions = append([]string(nil), available...)
		sort.Strings(versions)
	}
	if len(versions) == 0 {
		return nil, fmt.Errorf("docs: versioned collection has no version directories")
	}
	seen := make(map[string]bool, len(versions))
	for i, version := range versions {
		version = strings.TrimSpace(version)
		if err := validateVersionSegment(version); err != nil {
			return nil, fmt.Errorf("docs: version %q: %w", version, err)
		}
		if seen[version] {
			return nil, fmt.Errorf("docs: versioned collection repeats version %q", version)
		}
		seen[version] = true
		versions[i] = version
	}
	if !seen[current] {
		return nil, fmt.Errorf("docs: Current version %q is not in Versions", current)
	}
	return versions, nil
}

func validateVersionSegment(version string) error {
	if version == "" || version == "." || version == ".." {
		return fmt.Errorf("must be a non-empty path segment")
	}
	if strings.ContainsAny(version, `/\\`) {
		return fmt.Errorf("must not contain a path separator")
	}
	return nil
}

func versionedCollectionConfig(collection CollectionConfig, current, version string) CollectionConfig {
	collection.DefaultVersion = version
	collection.VersionPrefix = version != strings.TrimSpace(current)
	return collection
}

func markdownDirectoryHasIndex(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.EqualFold(entry.Name(), "index.md") {
			return true
		}
	}
	return false
}

func versionFSHasIndex(content fs.FS, root string) bool {
	entries, err := fs.ReadDir(content, root)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.EqualFold(entry.Name(), "index.md") {
			return true
		}
	}
	return false
}

func (r *Router) nextChildOrder(path string, start int) int {
	if start < 1 {
		start = 1
	}
	parent := r.parentFor(normalizePath(path))
	for {
		used := false
		for _, route := range r.Routes() {
			if route.Parent == parent && route.Order == start {
				used = true
				break
			}
		}
		if !used {
			return start
		}
		start++
	}
}
