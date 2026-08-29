package docs

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// ContentMetadata is the portable metadata contract shared by Markdown,
// typed screens, search, SEO, drafts, versions, and localized content.
// Markdown pages can provide the same fields in a YAML front matter block.
type ContentMetadata struct {
	Title         string            `json:"title,omitempty" yaml:"title,omitempty"`
	Description   string            `json:"description,omitempty" yaml:"description,omitempty"`
	Draft         bool              `json:"draft,omitempty" yaml:"draft,omitempty"`
	NoIndex       bool              `json:"noIndex,omitempty" yaml:"noindex,omitempty"`
	EditURL       string            `json:"editUrl,omitempty" yaml:"edit_url,omitempty"`
	CanonicalURL  string            `json:"canonicalUrl,omitempty" yaml:"canonical,omitempty"`
	Image         string            `json:"image,omitempty" yaml:"image,omitempty"`
	Authors       []string          `json:"authors,omitempty" yaml:"authors,omitempty"`
	DatePublished string            `json:"datePublished,omitempty" yaml:"date,omitempty"`
	DateModified  string            `json:"dateModified,omitempty" yaml:"last_updated,omitempty"`
	Locale        string            `json:"locale,omitempty" yaml:"locale,omitempty"`
	Version       string            `json:"version,omitempty" yaml:"version,omitempty"`
	Alternates    map[string]string `json:"alternates,omitempty" yaml:"alternates,omitempty"`
	Tags          []string          `json:"tags,omitempty" yaml:"tags,omitempty"`
	Redirects     []string          `json:"redirects,omitempty" yaml:"redirects,omitempty"`
	Order         int               `json:"order,omitempty" yaml:"order,omitempty"`
}

// MarkdownDocument is a parsed Markdown file with its front matter removed
// from Body. Keeping the two values together prevents search, TOC, and render
// paths from disagreeing about what the reader actually sees.
type MarkdownDocument struct {
	Metadata ContentMetadata
	Body     string
}

// ParseMarkdown parses an optional YAML front matter block. A document
// without front matter remains valid Markdown and returns zero metadata.
func ParseMarkdown(source string) (MarkdownDocument, error) {
	source = strings.TrimPrefix(source, "\ufeff")
	normalized := strings.ReplaceAll(source, "\r\n", "\n")
	if !strings.HasPrefix(normalized, "---\n") {
		return MarkdownDocument{Body: source}, nil
	}
	lines := strings.Split(normalized, "\n")
	closing := -1
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" || lines[i] == "..." {
			closing = i
			break
		}
	}
	if closing < 0 {
		return MarkdownDocument{}, errors.New("docs: Markdown front matter starts with --- but has no closing ---")
	}

	var raw frontMatter
	decoder := yaml.NewDecoder(strings.NewReader(strings.Join(lines[1:closing], "\n")))
	decoder.KnownFields(true)
	if err := decoder.Decode(&raw); err != nil {
		return MarkdownDocument{}, fmt.Errorf("docs: invalid Markdown front matter: %w", err)
	}
	body := strings.Join(lines[closing+1:], "\n")
	return MarkdownDocument{Metadata: raw.metadata(), Body: body}, nil
}

// LoadMarkdownFile reads and parses a Markdown document from disk.
func LoadMarkdownFile(path string) (MarkdownDocument, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return MarkdownDocument{}, err
	}
	return ParseMarkdown(string(body))
}

// CollectionConfig configures MarkdownCollection. Files are discovered in
// lexical path order, while each document can override its order in front
// matter. Drafts remain registered for validation and preview, but are not
// published unless the Router is configured with WithIncludeDrafts.
type CollectionConfig struct {
	OrderStart     int
	Offline        bool
	IncludeDrafts  bool
	DefaultLocale  string
	DefaultVersion string
	// LocalePrefix and VersionPrefix make non-default slices addressable at
	// /<prefix>/<locale>/... and /<prefix>/<version>/..., allowing several
	// localized/versioned documents to coexist in one Router.
	LocalePrefix  bool
	VersionPrefix bool
}

// MarkdownCollection registers every Markdown file beneath dir. `index.md`
// maps to the collection prefix; nested files map to their relative path.
// This is the reusable content-collection primitive behind generated sites.
func (r *Router) MarkdownCollection(prefix, dir string, cfg CollectionConfig) error {
	if r == nil {
		return errors.New("docs: MarkdownCollection requires a Router")
	}
	if strings.TrimSpace(dir) == "" {
		return errors.New("docs: MarkdownCollection requires a directory")
	}
	prefix = normalizePath(prefix)
	var files []string
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if entry.Name() != "." && strings.HasPrefix(entry.Name(), ".") {
				return fs.SkipDir
			}
			return nil
		}
		if strings.HasPrefix(entry.Name(), ".") || strings.HasPrefix(entry.Name(), "_") || !strings.EqualFold(filepath.Ext(path), ".md") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return fmt.Errorf("docs: scan Markdown collection %q: %w", dir, err)
	}
	sort.Strings(files)
	if cfg.OrderStart < 1 {
		cfg.OrderStart = 1
	}
	for index, path := range files {
		document, err := LoadMarkdownFile(path)
		if err != nil {
			return fmt.Errorf("docs: load collection file %q: %w", path, err)
		}
		meta := document.Metadata
		if meta.Locale == "" {
			meta.Locale = cfg.DefaultLocale
		}
		if meta.Version == "" {
			meta.Version = cfg.DefaultVersion
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return fmt.Errorf("docs: resolve collection file %q: %w", path, err)
		}
		rel = filepath.ToSlash(rel)
		rel = strings.TrimSuffix(rel, filepath.Ext(rel))
		if strings.HasSuffix(rel, "/index") {
			rel = strings.TrimSuffix(rel, "/index")
		} else if rel == "index" {
			rel = ""
		}
		routePrefix := prefix
		if cfg.LocalePrefix && meta.Locale != "" && meta.Locale != cfg.DefaultLocale && !strings.HasPrefix(rel, meta.Locale+"/") {
			routePrefix = joinPath(routePrefix, meta.Locale)
		}
		if cfg.VersionPrefix && meta.Version != "" && meta.Version != cfg.DefaultVersion && !strings.HasPrefix(rel, meta.Version+"/") {
			routePrefix = joinPath(routePrefix, meta.Version)
		}
		routePath := routePrefix
		if rel != "" {
			routePath = joinPath(routePrefix, rel)
		}
		title := meta.Title
		if title == "" {
			title = humanizeContentName(filepath.Base(rel))
			if rel == "" {
				title = humanizeContentName(filepath.Base(dir))
			}
		}
		description := meta.Description
		if description == "" {
			description = firstParagraph(document.Body)
		}
		order := meta.Order
		if order < 1 {
			order = cfg.OrderStart + index
		}
		if err := r.Page(routePath, PageConfig{
			Title:       title,
			Description: description,
			SourcePath:  path,
			Order:       order,
			Offline:     cfg.Offline,
			Hidden:      false,
			Metadata:    meta,
		}); err != nil {
			return err
		}
	}
	return nil
}

// pageMetadata reads front matter once during registration and combines it
// with explicit PageConfig fields. Explicit configuration wins, while the
// Markdown body is stored without front matter so render, search, and TOC
// all consume the same content.
func pageMetadata(cfg PageConfig) (ContentMetadata, string, error) {
	var document MarkdownDocument
	var err error
	source := cfg.Source
	if source != "" {
		document, err = ParseMarkdown(source)
	} else if cfg.SourcePath != "" {
		document, err = LoadMarkdownFile(cfg.SourcePath)
	}
	if err != nil {
		return ContentMetadata{}, "", err
	}
	metadata := mergeContentMetadata(document.Metadata, cfg.Metadata)
	metadata = mergeContentMetadata(metadata, ContentMetadata{
		Title: cfg.Title, Description: cfg.Description, Order: cfg.Order,
		Tags: cfg.Tags,
	})
	if cfg.Source != "" || cfg.SourcePath != "" {
		source = document.Body
	}
	return metadata, source, nil
}

type frontMatter struct {
	Title         string            `yaml:"title"`
	Description   string            `yaml:"description"`
	Draft         bool              `yaml:"draft"`
	NoIndex       bool              `yaml:"noindex"`
	NoIndexSnake  bool              `yaml:"no_index"`
	EditURL       string            `yaml:"edit_url"`
	CanonicalURL  string            `yaml:"canonical"`
	CanonicalLong string            `yaml:"canonical_url"`
	Image         string            `yaml:"image"`
	Authors       stringList        `yaml:"authors"`
	DatePublished string            `yaml:"date"`
	DateModified  string            `yaml:"last_updated"`
	Locale        string            `yaml:"locale"`
	Version       string            `yaml:"version"`
	Alternates    map[string]string `yaml:"alternates"`
	Tags          stringList        `yaml:"tags"`
	Redirects     stringList        `yaml:"redirects"`
	Order         int               `yaml:"order"`
}

func (f frontMatter) metadata() ContentMetadata {
	canonical := f.CanonicalURL
	if canonical == "" {
		canonical = f.CanonicalLong
	}
	return ContentMetadata{
		Title: f.Title, Description: f.Description, Draft: f.Draft,
		NoIndex: f.NoIndex || f.NoIndexSnake, EditURL: f.EditURL,
		CanonicalURL: canonical, Image: f.Image, Authors: append([]string(nil), f.Authors...),
		DatePublished: f.DatePublished, DateModified: f.DateModified,
		Locale: f.Locale, Version: f.Version, Alternates: cloneStringMap(f.Alternates), Tags: append([]string(nil), f.Tags...),
		Redirects: append([]string(nil), f.Redirects...), Order: f.Order,
	}
}

type stringList []string

func (s *stringList) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		*s = stringList{node.Value}
		return nil
	}
	var values []string
	if err := node.Decode(&values); err != nil {
		return err
	}
	*s = stringList(values)
	return nil
}

func mergeContentMetadata(base, override ContentMetadata) ContentMetadata {
	merged := base
	if override.Title != "" {
		merged.Title = override.Title
	}
	if override.Description != "" {
		merged.Description = override.Description
	}
	if override.Draft {
		merged.Draft = true
	}
	if override.NoIndex {
		merged.NoIndex = true
	}
	if override.EditURL != "" {
		merged.EditURL = override.EditURL
	}
	if override.CanonicalURL != "" {
		merged.CanonicalURL = override.CanonicalURL
	}
	if override.Image != "" {
		merged.Image = override.Image
	}
	if len(override.Authors) > 0 {
		merged.Authors = append([]string(nil), override.Authors...)
	}
	if override.DatePublished != "" {
		merged.DatePublished = override.DatePublished
	}
	if override.DateModified != "" {
		merged.DateModified = override.DateModified
	}
	if override.Locale != "" {
		merged.Locale = override.Locale
	}
	if override.Version != "" {
		merged.Version = override.Version
	}
	if len(override.Alternates) > 0 {
		merged.Alternates = cloneStringMap(override.Alternates)
	}
	if len(override.Tags) > 0 {
		merged.Tags = append([]string(nil), override.Tags...)
	}
	if len(override.Redirects) > 0 {
		merged.Redirects = append([]string(nil), override.Redirects...)
	}
	if override.Order > 0 {
		merged.Order = override.Order
	}
	return merged
}

func cloneStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func firstParagraph(source string) string {
	for _, block := range strings.Split(strings.ReplaceAll(source, "\r\n", "\n"), "\n\n") {
		line := strings.TrimSpace(block)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "```") {
			continue
		}
		line = strings.Join(strings.Fields(line), " ")
		if len(line) > 180 {
			return line[:177] + "..."
		}
		return line
	}
	return "Documentation page."
}

func humanizeContentName(value string) string {
	value = strings.ReplaceAll(strings.ReplaceAll(value, "-", " "), "_", " ")
	value = strings.TrimSpace(value)
	if value == "" {
		return "Documentation"
	}
	return strings.ToUpper(value[:1]) + value[1:]
}
