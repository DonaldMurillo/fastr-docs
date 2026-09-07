package docs

import (
	"errors"
	"fmt"
	"io/fs"
	"strings"
)

// MarkdownCollectionPlugin turns a directory of Markdown files into Router
// routes. It is the reusable content plugin behind generated projects and is
// intentionally independent of the CLI.
type MarkdownCollectionPlugin struct {
	// Path is the route prefix the collection mounts under.
	Path string
	// Dir is the disk directory of the collection; FS replaces it.
	Dir string
	// FS serves the collection from an embedded filesystem instead
	// of disk.
	FS fs.FS
	// Root is the subdirectory of FS the collection lives in.
	Root string
	// Config carries the collection's ordering, locale, and view
	// settings.
	Config CollectionConfig
}

// Name identifies the plugin in diagnostics.
func (p MarkdownCollectionPlugin) Name() string { return "markdown-collection" }

// Apply mounts the collection's Markdown tree onto the Router, from disk
// or from the embedded FS.
func (p MarkdownCollectionPlugin) Apply(r *Router) error {
	if r == nil {
		return errors.New("router is nil")
	}
	if p.FS != nil {
		if err := r.MarkdownCollectionFS(p.Path, p.FS, p.Root, p.Config); err != nil {
			return fmt.Errorf("load collection: %w", err)
		}
		return nil
	}
	if strings.TrimSpace(p.Dir) == "" {
		return errors.New("Dir or FS is required")
	}
	if err := r.MarkdownCollection(p.Path, p.Dir, p.Config); err != nil {
		return fmt.Errorf("load collection: %w", err)
	}
	return nil
}

// MarkdownComponentsPlugin registers a shared component vocabulary for
// Markdown shortcodes. It lets a plugin own both its component implementation
// and the authoring contract without requiring pages to repeat a map.
type MarkdownComponentsPlugin struct {
	// Components maps each shortcode name to its implementation;
	// applying the plugin retires any earlier meaning the names had.
	Components map[string]MarkdownComponent
}

// Name identifies the plugin in diagnostics.
func (p MarkdownComponentsPlugin) Name() string { return "markdown-components" }

// Apply registers every shortcode component in the map under its own name,
// retiring any earlier meaning the name had.
func (p MarkdownComponentsPlugin) Apply(r *Router) error {
	if r == nil {
		return errors.New("router is nil")
	}
	if len(p.Components) == 0 {
		return errors.New("Components is required")
	}
	for name, component := range p.Components {
		if err := r.RegisterMarkdownComponent(name, component); err != nil {
			return err
		}
	}
	return nil
}
