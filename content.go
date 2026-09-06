package docs

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	pathpkg "path"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// ContentMetadata is the portable metadata contract shared by Markdown,
// typed screens, search, SEO, drafts, versions, and localized content.
// Markdown pages can provide the same fields in a YAML front matter block.
type ContentMetadata struct {
	Title         string   `json:"title,omitempty" yaml:"title,omitempty"`
	Slug          string   `json:"slug,omitempty" yaml:"slug,omitempty"`
	Description   string   `json:"description,omitempty" yaml:"description,omitempty"`
	Excerpt       string   `json:"excerpt,omitempty" yaml:"excerpt,omitempty"`
	Draft         bool     `json:"draft,omitempty" yaml:"draft,omitempty"`
	NoIndex       bool     `json:"noIndex,omitempty" yaml:"noindex,omitempty"`
	EditURL       string   `json:"editUrl,omitempty" yaml:"edit_url,omitempty"`
	CanonicalURL  string   `json:"canonicalUrl,omitempty" yaml:"canonical,omitempty"`
	Image         string   `json:"image,omitempty" yaml:"image,omitempty"`
	Authors       []string `json:"authors,omitempty" yaml:"authors,omitempty"`
	DatePublished string   `json:"datePublished,omitempty" yaml:"date,omitempty"`
	DateModified  string   `json:"dateModified,omitempty" yaml:"last_updated,omitempty"`
	Locale        string   `json:"locale,omitempty" yaml:"locale,omitempty"`
	Version       string   `json:"version,omitempty" yaml:"version,omitempty"`
	// TranslationOf names the route this page translates, so the two pair as
	// variants of one family whatever their paths are. Without it a
	// translation pairs by path shape alone: /es/docs/guide is the Spanish
	// /docs/guide because stripping the locale segment leaves the same path.
	// A translated slug such as /es/docs/guia breaks that, and this is how it
	// says which page it is.
	TranslationOf string            `json:"translationOf,omitempty" yaml:"translation_of,omitempty"`
	Alternates    map[string]string `json:"alternates,omitempty" yaml:"alternates,omitempty"`
	Tags          []string          `json:"tags,omitempty" yaml:"tags,omitempty"`
	Redirects     []string          `json:"redirects,omitempty" yaml:"redirects,omitempty"`
	Order         int               `json:"order,omitempty" yaml:"order,omitempty"`
	// PageTemplate selects the page shell. Empty is the standard
	// documentation page; PageTemplateSplash is the landing-page shell.
	PageTemplate string    `json:"template,omitempty" yaml:"template,omitempty"`
	Hero         *PageHero `json:"hero,omitempty" yaml:"hero,omitempty"`
}

// PageTemplateSplash renders a landing page: a hero from front matter, no
// table of contents, no breadcrumbs, and a wider content column.
const PageTemplateSplash = "splash"

// PageHero is the front-matter-driven hero of a splash page, so a landing page
// does not require writing a typed screen.
type PageHero struct {
	Eyebrow string           `json:"eyebrow,omitempty" yaml:"eyebrow,omitempty"`
	Title   string           `json:"title,omitempty" yaml:"title,omitempty"`
	Tagline string           `json:"tagline,omitempty" yaml:"tagline,omitempty"`
	Actions []PageHeroAction `json:"actions,omitempty" yaml:"actions,omitempty"`
}

// PageHeroAction is one call-to-action button in a hero.
type PageHeroAction struct {
	Text    string `json:"text,omitempty" yaml:"text,omitempty"`
	Link    string `json:"link,omitempty" yaml:"link,omitempty"`
	Variant string `json:"variant,omitempty" yaml:"variant,omitempty"`
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

// CollectionConfig configures MarkdownCollection and MarkdownCollectionFS.
// Files are discovered in lexical path order, while each document can
// override its order in front matter. Drafts remain registered for validation
// and preview, but IncludeDrafts or WithIncludeDrafts makes them publishable.
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
	return r.markdownCollection(prefix, dir, cfg, "", false)
}

func (r *Router) markdownCollection(prefix, dir string, cfg CollectionConfig, version string, prefixVersion bool) error {
	if r == nil {
		return errors.New("docs: MarkdownCollection requires a Router")
	}
	if strings.TrimSpace(dir) == "" {
		return errors.New("docs: MarkdownCollection requires a directory")
	}
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
		rel = collectionRelativePath(rel)
		if err := r.registerCollectionDocument(prefix, dir, rel, path, document, index, cfg, version, prefixVersion); err != nil {
			return err
		}
	}
	return nil
}

// MarkdownCollectionFS registers Markdown content from any fs.FS, including
// embed.FS, fstest.MapFS, and application-provided virtual filesystems. `root`
// is an fs.ValidPath relative to content; use "." for the filesystem root.
// Unlike MarkdownCollection, the document body is captured at registration so
// a virtual filesystem does not need to masquerade as an OS path.
func (r *Router) MarkdownCollectionFS(prefix string, content fs.FS, root string, cfg CollectionConfig) error {
	return r.markdownCollectionFS(prefix, content, root, cfg, "", false)
}

func (r *Router) markdownCollectionFS(prefix string, content fs.FS, root string, cfg CollectionConfig, version string, prefixVersion bool) error {
	if r == nil {
		return errors.New("docs: MarkdownCollectionFS requires a Router")
	}
	if content == nil {
		return errors.New("docs: MarkdownCollectionFS requires an fs.FS")
	}
	root = strings.TrimSpace(root)
	if root == "" {
		root = "."
	}
	if !fs.ValidPath(root) {
		return fmt.Errorf("docs: MarkdownCollectionFS root %q is not a valid fs path", root)
	}
	var files []string
	err := fs.WalkDir(content, root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if entry.Name() != "." && strings.HasPrefix(entry.Name(), ".") {
				return fs.SkipDir
			}
			return nil
		}
		if strings.HasPrefix(entry.Name(), ".") || strings.HasPrefix(entry.Name(), "_") || !strings.EqualFold(pathpkg.Ext(path), ".md") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return fmt.Errorf("docs: scan Markdown collection FS %q: %w", root, err)
	}
	sort.Strings(files)
	if cfg.OrderStart < 1 {
		cfg.OrderStart = 1
	}
	for index, filePath := range files {
		body, err := fs.ReadFile(content, filePath)
		if err != nil {
			return fmt.Errorf("docs: load collection file %q: %w", filePath, err)
		}
		document, err := ParseMarkdown(string(body))
		if err != nil {
			return fmt.Errorf("docs: load collection file %q: %w", filePath, err)
		}
		rel, err := collectionFSRelativePath(root, filePath)
		if err != nil {
			return fmt.Errorf("docs: resolve collection file %q: %w", filePath, err)
		}
		rel = strings.TrimSuffix(rel, pathpkg.Ext(rel))
		rel = collectionRelativePath(rel)
		if err := r.registerCollectionDocument(prefix, root, rel, "", document, index, cfg, version, prefixVersion); err != nil {
			return err
		}
	}
	return nil
}

func collectionRelativePath(rel string) string {
	if strings.HasSuffix(rel, "/index") {
		return strings.TrimSuffix(rel, "/index")
	}
	if rel == "index" {
		return ""
	}
	return rel
}

func collectionFSRelativePath(root, filePath string) (string, error) {
	root = pathpkg.Clean(root)
	filePath = pathpkg.Clean(filePath)
	if root == "." {
		return filePath, nil
	}
	if filePath == root {
		return "", nil
	}
	prefix := root + "/"
	if !strings.HasPrefix(filePath, prefix) {
		return "", fmt.Errorf("path is outside collection root %q", root)
	}
	return strings.TrimPrefix(filePath, prefix), nil
}

func (r *Router) registerCollectionDocument(prefix, sourceRoot, rel, sourcePath string, document MarkdownDocument, index int, cfg CollectionConfig, version string, prefixVersion bool) error {
	meta := document.Metadata
	if meta.Slug != "" {
		slug, err := collectionSlug(meta.Slug)
		if err != nil {
			return fmt.Errorf("docs: collection document %q: %w", rel, err)
		}
		rel = slug
	}
	if meta.Locale == "" {
		meta.Locale = cfg.DefaultLocale
	}
	if version != "" {
		if meta.Version != "" && meta.Version != version {
			return fmt.Errorf("docs: collection document %q declares version %q, want snapshot version %q", rel, meta.Version, version)
		}
		meta.Version = version
	} else if meta.Version == "" {
		meta.Version = cfg.DefaultVersion
	}
	routePrefix := normalizePath(prefix)
	if cfg.LocalePrefix && meta.Locale != "" && meta.Locale != cfg.DefaultLocale && !strings.HasPrefix(rel, meta.Locale+"/") {
		routePrefix = joinPath(routePrefix, meta.Locale)
	}
	if prefixVersion && version != "" && !strings.HasPrefix(rel, version+"/") {
		routePrefix = joinPath(routePrefix, version)
	} else if cfg.VersionPrefix && meta.Version != "" && meta.Version != cfg.DefaultVersion && !strings.HasPrefix(rel, meta.Version+"/") {
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
			title = collectionRootTitle(sourceRoot)
		}
	}
	description := meta.Description
	if description == "" {
		description = firstParagraph(document.Body)
	}
	config := PageConfig{
		Title:       title,
		Description: description,
		Order:       collectionOrder(meta.Order, cfg.OrderStart, index),
		Offline:     cfg.Offline,
		Hidden:      false,
		Metadata:    meta,
	}
	if sourcePath != "" {
		config.SourcePath = sourcePath
	} else {
		config.Source = document.Body
	}
	if err := r.Page(routePath, config); err != nil {
		return err
	}
	r.routes[normalizePath(routePath)].includeDrafts = cfg.IncludeDrafts
	return nil
}

func collectionSlug(raw string) (string, error) {
	raw = strings.TrimSpace(strings.ReplaceAll(raw, "\\", "/"))
	if raw == "" || raw == "." {
		return "", errors.New("slug must not be empty")
	}
	if raw == "/" {
		return "", nil
	}
	if strings.ContainsAny(raw, "?#") || strings.HasPrefix(raw, "//") {
		return "", fmt.Errorf("slug %q must be a relative path without query or fragment", raw)
	}
	raw = strings.TrimPrefix(raw, "/")
	for _, segment := range strings.Split(raw, "/") {
		if segment == "." || segment == ".." {
			return "", fmt.Errorf("slug %q must stay inside the collection", raw)
		}
	}
	clean := pathpkg.Clean(raw)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") || strings.Contains(clean, "/../") {
		return "", fmt.Errorf("slug %q must stay inside the collection", raw)
	}
	for _, segment := range strings.Split(clean, "/") {
		if segment == "" || segment == "." || segment == ".." || strings.HasPrefix(segment, ".") || strings.HasPrefix(segment, "_") {
			return "", fmt.Errorf("slug %q contains an unusable path segment", raw)
		}
	}
	return clean, nil
}

func collectionOrder(explicit, start, index int) int {
	if explicit > 0 {
		return explicit
	}
	return start + index
}

func collectionRootTitle(sourceRoot string) string {
	cleaned := strings.TrimSpace(strings.TrimRight(sourceRoot, `/\`))
	if cleaned == "" || cleaned == "." {
		return "Documentation"
	}
	return humanizeContentName(filepath.Base(cleaned))
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
	Slug          string            `yaml:"slug"`
	Description   string            `yaml:"description"`
	Excerpt       string            `yaml:"excerpt"`
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
	TranslationOf string            `yaml:"translation_of"`
	Alternates    map[string]string `yaml:"alternates"`
	Tags          stringList        `yaml:"tags"`
	Redirects     stringList        `yaml:"redirects"`
	Order         int               `yaml:"order"`
	Template      string            `yaml:"template"`
	Hero          *frontMatterHero  `yaml:"hero"`
}

type frontMatterHero struct {
	Eyebrow string                  `yaml:"eyebrow"`
	Title   string                  `yaml:"title"`
	Tagline string                  `yaml:"tagline"`
	Actions []frontMatterHeroAction `yaml:"actions"`
}

type frontMatterHeroAction struct {
	Text    string `yaml:"text"`
	Link    string `yaml:"link"`
	Variant string `yaml:"variant"`
}

func (h *frontMatterHero) metadata() *PageHero {
	if h == nil {
		return nil
	}
	hero := &PageHero{Eyebrow: h.Eyebrow, Title: h.Title, Tagline: h.Tagline}
	for _, action := range h.Actions {
		hero.Actions = append(hero.Actions, PageHeroAction{Text: action.Text, Link: action.Link, Variant: action.Variant})
	}
	return hero
}

func (f frontMatter) metadata() ContentMetadata {
	canonical := f.CanonicalURL
	if canonical == "" {
		canonical = f.CanonicalLong
	}
	return ContentMetadata{
		Title: f.Title, Slug: f.Slug, Description: f.Description, Excerpt: f.Excerpt, Draft: f.Draft,
		NoIndex: f.NoIndex || f.NoIndexSnake, EditURL: f.EditURL,
		CanonicalURL: canonical, Image: f.Image, Authors: append([]string(nil), f.Authors...),
		DatePublished: f.DatePublished, DateModified: f.DateModified,
		Locale: f.Locale, Version: f.Version, TranslationOf: f.TranslationOf,
		Alternates: cloneStringMap(f.Alternates), Tags: append([]string(nil), f.Tags...),
		Redirects: append([]string(nil), f.Redirects...), Order: f.Order,
		PageTemplate: strings.ToLower(strings.TrimSpace(f.Template)), Hero: f.Hero.metadata(),
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
	if override.Excerpt != "" {
		merged.Excerpt = override.Excerpt
	}
	if override.Slug != "" {
		merged.Slug = override.Slug
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
	if override.TranslationOf != "" {
		merged.TranslationOf = override.TranslationOf
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
