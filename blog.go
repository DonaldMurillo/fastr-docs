package docs

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	pathpkg "path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/DonaldMurillo/gofastr/core/render"
	gofastrRouter "github.com/DonaldMurillo/gofastr/core/router"
)

// BlogConfig configures a Router-backed Markdown blog. The directory's
// index.md becomes the archive route at prefix. Every other Markdown file
// becomes a post route under that prefix.
type BlogConfig struct {
	// Title names the collection in navigation and feeds.
	Title string
	// Description is the collection's meta description.
	Description string
	// Order positions the collection among the site sections.
	Order int
	// PostOrderStart is the Order given to the first post, so posts sort
	// after the collection's own views; zero uses Order + 1.
	PostOrderStart int
	// Offline marks every generated view precacheable by the service
	// worker.
	Offline bool
	// IncludeDrafts publishes posts whose front matter marks them as
	// drafts.
	IncludeDrafts bool
	// DefaultLocale is the language the collection's views and unmarked
	// posts count as, so a Spanish collection reads the Spanish labels.
	DefaultLocale string
	// DefaultVersion is the version the collection serves when a post
	// declares none.
	DefaultVersion string
	// LocalePrefix pairs the collection's generated views with their
	// translations by path shape, the way /es/blog pairs with /blog.
	LocalePrefix bool
	// VersionPrefix pairs the collection's views with their versioned
	// variants by path shape.
	VersionPrefix bool
	// PostsPerPage controls archive pagination. Zero uses the built-in
	// default of 10.
	PostsPerPage int
	// The aggregate views are enabled by default. Disable* is explicit so a
	// zero-value BlogConfig still produces a complete publishing surface.
	DisableSearch  bool
	DisableTags    bool
	DisableAuthors bool
	DisableArchive bool
	// RelatedPosts controls the number of related links on a post. Zero uses
	// the built-in default of 3.
	RelatedPosts int
}

const defaultBlogPostsPerPage = 10

type blogCollection struct {
	cfg       BlogConfig
	prefix    string
	indexBody string
}

type blogViewKind string

const (
	blogViewArchive blogViewKind = "archive"
	blogViewPage    blogViewKind = "page"
	blogViewSearch  blogViewKind = "search"
	blogViewTags    blogViewKind = "tags"
	blogViewTag     blogViewKind = "tag"
	blogViewAuthors blogViewKind = "authors"
	blogViewAuthor  blogViewKind = "author"
)

type blogView struct {
	kind   blogViewKind
	prefix string
	term   string
	page   int
}

// BlogPosts returns visible blog posts beneath prefix, newest first. The
// archive route itself is excluded. No-index posts remain available here so a
// project can choose whether an internal archive should link to them; RSSXML
// excludes them from the public feed.
func (r *Router) BlogPosts(prefix string) []*Route {
	if r == nil {
		return nil
	}
	prefix = normalizePath(prefix)
	posts := make([]*Route, 0)
	for _, route := range r.PublishedRoutes() {
		if !route.Blog || route.BlogIndex || route.Path == prefix || r.isBlogView(route.Path) {
			continue
		}
		if prefix != "/" && !strings.HasPrefix(route.Path, prefix+"/") {
			continue
		}
		// A future publish date is a schedule, not a leak: the post
		// stays registered but leaves every public listing.
		if blogPostDatedFuture(route) {
			continue
		}
		posts = append(posts, route)
	}
	sort.SliceStable(posts, func(i, j int) bool {
		left, leftOK := parseBlogDate(posts[i].Metadata.DatePublished)
		right, rightOK := parseBlogDate(posts[j].Metadata.DatePublished)
		if leftOK != rightOK {
			return leftOK
		}
		if leftOK && !left.Equal(right) {
			return left.After(right)
		}
		if posts[i].Order != posts[j].Order {
			return posts[i].Order < posts[j].Order
		}
		return posts[i].seq < posts[j].seq
	})
	return posts
}

// MarkdownBlog registers an archive and all Markdown posts found beneath dir.
// Disk-backed posts retain SourcePath, so the normal development reload loop
// can reread changed content without introducing a second content system.
func (r *Router) MarkdownBlog(prefix, dir string, cfg BlogConfig) error {
	if r == nil {
		return errors.New("docs: MarkdownBlog requires a Router")
	}
	if strings.TrimSpace(dir) == "" {
		return errors.New("docs: MarkdownBlog requires a directory")
	}
	files, err := scanBlogFiles(dir)
	if err != nil {
		return fmt.Errorf("docs: scan Markdown blog %q: %w", dir, err)
	}
	documents := make([]blogDocument, 0, len(files))
	for _, sourcePath := range files {
		document, err := LoadMarkdownFile(sourcePath)
		if err != nil {
			return fmt.Errorf("docs: load blog file %q: %w", sourcePath, err)
		}
		rel, err := filepath.Rel(dir, sourcePath)
		if err != nil {
			return fmt.Errorf("docs: resolve blog file %q: %w", sourcePath, err)
		}
		rel = collectionRelativePath(strings.TrimSuffix(filepath.ToSlash(rel), filepath.Ext(rel)))
		documents = append(documents, blogDocument{rel: rel, sourcePath: sourcePath, document: document})
	}
	return r.registerBlogDocuments(prefix, documents, cfg, dir, false)
}

// MarkdownBlogFS is the embedded or virtual-filesystem form of MarkdownBlog.
// It captures document bodies at registration, matching MarkdownCollectionFS.
func (r *Router) MarkdownBlogFS(prefix string, content fs.FS, root string, cfg BlogConfig) error {
	if r == nil {
		return errors.New("docs: MarkdownBlogFS requires a Router")
	}
	if content == nil {
		return errors.New("docs: MarkdownBlogFS requires an fs.FS")
	}
	root = strings.TrimSpace(root)
	if root == "" {
		root = "."
	}
	if !fs.ValidPath(root) {
		return fmt.Errorf("docs: MarkdownBlogFS root %q is not a valid fs path", root)
	}
	paths, err := collectMarkdownFiles(func(root string, fn fs.WalkDirFunc) error { return fs.WalkDir(content, root, fn) }, root, pathpkg.Ext)
	if err != nil {
		return fmt.Errorf("docs: scan Markdown blog FS %q: %w", root, err)
	}
	documents := make([]blogDocument, 0, len(paths))
	for _, filePath := range paths {
		body, err := fs.ReadFile(content, filePath)
		if err != nil {
			return fmt.Errorf("docs: load blog file %q: %w", filePath, err)
		}
		document, err := ParseMarkdown(string(body))
		if err != nil {
			return fmt.Errorf("docs: load blog file %q: %w", filePath, err)
		}
		rel, err := collectionFSRelativePath(root, filePath)
		if err != nil {
			return fmt.Errorf("docs: resolve blog file %q: %w", filePath, err)
		}
		rel = collectionRelativePath(strings.TrimSuffix(rel, pathpkg.Ext(rel)))
		documents = append(documents, blogDocument{rel: rel, document: document})
	}
	return r.registerBlogDocuments(prefix, documents, cfg, root, true)
}

type blogDocument struct {
	rel        string
	sourcePath string
	document   MarkdownDocument
}

func scanBlogFiles(dir string) ([]string, error) {
	return collectMarkdownFiles(filepath.WalkDir, dir, filepath.Ext)
}

func (r *Router) registerBlogDocuments(prefix string, documents []blogDocument, cfg BlogConfig, sourceRoot string, virtual bool) error {
	prefix = normalizePath(prefix)
	if prefix == "" || prefix == "/" {
		return errors.New("docs: MarkdownBlog prefix must be a non-root path")
	}
	if cfg.Order < 1 {
		cfg.Order = 1
	}
	if cfg.PostOrderStart < 1 {
		cfg.PostOrderStart = cfg.Order + 1
		if cfg.PostOrderStart < 2 {
			cfg.PostOrderStart = 2
		}
	}
	if cfg.PostsPerPage < 1 {
		cfg.PostsPerPage = defaultBlogPostsPerPage
	}
	if cfg.RelatedPosts < 1 {
		cfg.RelatedPosts = 3
	}
	var index *blogDocument
	posts := make([]blogDocument, 0, len(documents))
	for i := range documents {
		if documents[i].rel == "" {
			if index != nil {
				return errors.New("docs: MarkdownBlog contains more than one index.md")
			}
			index = &documents[i]
			continue
		}
		posts = append(posts, documents[i])
	}
	if index == nil {
		return fmt.Errorf("docs: MarkdownBlog %q has no index.md; the archive route has nothing to say", prefix)
	}
	// The collection registers before its documents so path-aware helpers
	// (blog prefixes, slug application) can see it while posts are added.
	r.blogs[prefix] = blogCollection{cfg: cfg, prefix: prefix}
	for i := range posts {
		document := posts[i]
		meta := blogMetadata(document.document.Metadata, cfg, document.rel, document.document.Body)
		routePath, err := blogDocumentPath(prefix, document.rel, meta, cfg)
		if err != nil {
			return fmt.Errorf("docs: blog document %q: %w", document.rel, err)
		}
		if routePath == prefix {
			return fmt.Errorf("docs: blog document %q resolves to the archive path", document.rel)
		}
		page := PageConfig{
			Title:       meta.Title,
			Description: meta.Description,
			Order:       collectionOrder(meta.Order, cfg.PostOrderStart, i),
			Offline:     cfg.Offline,
			Metadata:    meta,
		}
		if document.sourcePath != "" && !virtual {
			page.SourcePath = document.sourcePath
		} else {
			page.Source = document.document.Body
		}
		if page.Title == "" {
			page.Title = humanizeContentName(filepath.Base(document.rel))
		}
		if page.Description == "" {
			page.Description = firstParagraph(document.document.Body)
		}
		if err := r.Page(routePath, page); err != nil {
			return err
		}
		route := r.routes[routePath]
		route.Blog = true
		route.includeDrafts = cfg.IncludeDrafts
	}

	indexMeta := ContentMetadata{}
	indexBody := ""
	if index != nil {
		indexMeta = blogMetadata(index.document.Metadata, cfg, "", index.document.Body)
		indexBody = index.document.Body
	}
	indexTitle := strings.TrimSpace(cfg.Title)
	if indexTitle == "" {
		indexTitle = strings.TrimSpace(indexMeta.Title)
	}
	if indexTitle == "" {
		indexTitle = collectionRootTitle(sourceRoot)
	}
	indexDescription := strings.TrimSpace(cfg.Description)
	if indexDescription == "" {
		indexDescription = strings.TrimSpace(indexMeta.Description)
	}
	if indexDescription == "" {
		indexDescription = firstParagraph(indexBody)
	}
	if indexDescription == "" {
		indexDescription = "Published posts from " + indexTitle + "."
	}
	indexSource := blogIndexSource(indexBody, indexTitle, indexDescription, r.BlogPosts(prefix))
	indexPage := PageConfig{
		Title:       indexTitle,
		Description: indexDescription,
		Source:      indexSource,
		Order:       cfg.Order,
		Offline:     cfg.Offline,
		Metadata:    indexMeta,
	}
	indexPage.Metadata.Title = indexTitle
	indexPage.Metadata.Description = indexDescription
	indexPage.Body = func() render.HTML { return r.renderBlogArchive(prefix, 1, indexBody) }
	indexPage.DisableTOC = true
	// The generated listing is part of the archive, so it must remain in
	// Source instead of using SourcePath. Post files still retain live reload.
	if err := r.Page(prefix, indexPage); err != nil {
		return err
	}
	r.routes[prefix].Blog = true
	r.routes[prefix].BlogIndex = true
	r.routes[prefix].includeDrafts = cfg.IncludeDrafts
	if err := r.registerBlogViews(prefix); err != nil {
		return err
	}
	return nil
}

func blogMetadata(meta ContentMetadata, cfg BlogConfig, rel, body string) ContentMetadata {
	if meta.Locale == "" {
		meta.Locale = cfg.DefaultLocale
	}
	if meta.Version == "" {
		meta.Version = cfg.DefaultVersion
	}
	if meta.DatePublished == "" {
		if date := blogDateFromPath(rel); date != "" {
			meta.DatePublished = date
		}
	}
	if meta.Excerpt == "" {
		meta.Excerpt = blogExcerpt(body)
	}
	return meta
}

func blogDateFromPath(rel string) string {
	name := pathpkg.Base(strings.TrimSuffix(rel, pathpkg.Ext(rel)))
	if len(name) < len("2006-01-02") {
		return ""
	}
	raw := name[:len("2006-01-02")]
	if _, err := time.Parse("2006-01-02", raw); err != nil {
		return ""
	}
	return raw
}

func blogExcerpt(body string) string {
	body = strings.ReplaceAll(body, "\r\n", "\n")
	for _, marker := range []string{"<!-- truncate -->", "<!--more-->", "<!-- more -->"} {
		if index := strings.Index(strings.ToLower(body), marker); index >= 0 {
			body = body[:index]
			break
		}
	}
	return firstParagraph(body)
}

func blogDocumentPath(prefix, rel string, meta ContentMetadata, cfg BlogConfig) (string, error) {
	if meta.Slug != "" {
		slug, err := collectionSlug(meta.Slug)
		if err != nil {
			return "", err
		}
		rel = slug
	}
	routePrefix := prefix
	if cfg.LocalePrefix && meta.Locale != "" && meta.Locale != cfg.DefaultLocale && !strings.HasPrefix(rel, meta.Locale+"/") {
		routePrefix = joinPath(routePrefix, meta.Locale)
	}
	if cfg.VersionPrefix && meta.Version != "" && meta.Version != cfg.DefaultVersion && !strings.HasPrefix(rel, meta.Version+"/") {
		routePrefix = joinPath(routePrefix, meta.Version)
	}
	return joinPath(routePrefix, rel), nil
}

func blogIndexSource(body, title, description string, posts []*Route) string {
	body = strings.TrimSpace(body)
	if body == "" {
		body = "# " + markdownText(title) + "\n\n" + markdownText(description)
	}
	listing := []string{"## Latest posts", ""}
	if len(posts) == 0 {
		listing = append(listing, "No posts have been published yet.")
	} else {
		for _, post := range posts {
			line := "- [" + markdownText(post.Title) + "](" + post.Path + ")"
			if date := strings.TrimSpace(post.Metadata.DatePublished); date != "" {
				line += " · " + markdownText(date)
			}
			if post.Description != "" {
				line += "\n  " + markdownText(post.Description)
			}
			listing = append(listing, line)
		}
	}
	return strings.TrimSpace(body) + "\n\n" + strings.Join(listing, "\n") + "\n"
}

func markdownText(value string) string {
	value = plainImageAlt.ReplaceAllString(value, "$1")
	value = plainLink.ReplaceAllString(value, "$1")
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "[", "\\[")
	value = strings.ReplaceAll(value, "]", "\\]")
	return strings.TrimSpace(value)
}

func parseBlogDate(raw string) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02", time.RFC1123Z, time.RFC1123} {
		if value, err := time.Parse(layout, raw); err == nil {
			return value, true
		}
	}
	return time.Time{}, false
}

// RSSConfig controls the RSS projection of a blog route collection.
type RSSConfig struct {
	// Prefix names the collection whose published posts feed this
	// channel.
	Prefix string
	// Title is the channel title; empty uses the collection's.
	Title string
	// Description is the channel description; empty uses the
	// collection's.
	Description string
	// SiteURL is the absolute origin of the site, needed for valid
	// RSS link elements on a static export.
	SiteURL string
	// FeedPath is the URL the feed itself is served at, used for the
	// channel's atom:link rel="self". Empty derives prefix + "/feed.xml".
	FeedPath string
	// Limit caps the item count; zero includes every published post.
	Limit int
}

// RSSXML returns a valid RSS 2.0 document for the published posts beneath
// cfg.Prefix. Links are absolute when SiteURL is configured and otherwise
// remain route-relative, which keeps local development deterministic.
func (r *Router) RSSXML(cfg RSSConfig) ([]byte, error) {
	if r == nil {
		return nil, errors.New("docs: RSSXML requires a Router")
	}
	if err := r.Validate(); err != nil {
		return nil, err
	}
	if cfg.Limit < 0 {
		return nil, errors.New("docs: RSS limit cannot be negative")
	}
	if !strings.HasPrefix(strings.TrimSpace(cfg.Prefix), "/") {
		return nil, fmt.Errorf("docs: RSS prefix %q must start with a slash", cfg.Prefix)
	}
	prefix := normalizePath(cfg.Prefix)
	if prefix == "" || prefix == "/" {
		return nil, errors.New("docs: RSS prefix must be a non-root path")
	}
	if _, ok := r.blogs[prefix]; !ok {
		return nil, fmt.Errorf("docs: RSS prefix %q has no registered blog collection", prefix)
	}
	siteURL, err := rssSiteURL(cfg.SiteURL)
	if err != nil {
		return nil, err
	}
	title := strings.TrimSpace(cfg.Title)
	if title == "" {
		if collection, ok := r.blogs[prefix]; ok {
			title = strings.TrimSpace(collection.cfg.Title)
		}
	}
	if title == "" {
		title = r.SiteName() + " blog"
	}
	description := strings.TrimSpace(cfg.Description)
	if description == "" {
		description = "Latest posts from " + title + "."
	}
	posts := r.BlogPosts(prefix)
	items := make([]rssItem, 0, len(posts))
	for _, route := range posts {
		if route.Metadata.NoIndex || route.Metadata.Draft || blogPostDatedFuture(route) {
			continue
		}
		items = append(items, rssItem{
			Title:       route.Title,
			Link:        rssLink(siteURL, route.Path),
			GUID:        rssGUID{Value: rssLink(siteURL, route.Path), IsPermaLink: "true"},
			Description: routeDescription(route),
			PubDate:     rssDate(route.Metadata.DatePublished),
			Author:      strings.Join(route.Metadata.Authors, ", "),
			Categories:  cloneStrings(route.Metadata.Tags),
		})
		if cfg.Limit > 0 && len(items) >= cfg.Limit {
			break
		}
	}
	language := ""
	if collection, ok := r.blogs[prefix]; ok {
		language = collection.cfg.DefaultLocale
	}
	lastBuild := ""
	for _, item := range items {
		if item.PubDate != "" && (lastBuild == "" || item.PubDate > lastBuild) {
			lastBuild = item.PubDate
		}
	}
	feedPath := strings.TrimSpace(cfg.FeedPath)
	if feedPath == "" {
		feedPath = joinPath(prefix, "feed.xml")
	}
	document := rssDocument{
		Version:   "2.0",
		XMLNSAtom: "http://www.w3.org/2005/Atom",
		Channel: rssChannel{
			Title:         title,
			Link:          rssLink(siteURL, prefix),
			Description:   description,
			Language:      language,
			Generator:     "fastr-docs",
			LastBuildDate: lastBuild,
			Items:         items,
			// Validators require the channel to name itself; readers
			// use it to detect moved feeds.
			AtomLink: atomLink{Href: rssLink(siteURL, feedPath), Rel: "self", Type: "application/rss+xml"},
		},
	}
	body, err := xml.Marshal(document)
	if err != nil {
		return nil, fmt.Errorf("docs: encode RSS feed: %w", err)
	}
	return append([]byte(xml.Header), body...), nil
}

// MountRSS mounts a live RSS endpoint on the GoFastr HTTP router. The feed is
// rendered per request so a development server sees metadata changes without
// duplicating route registration or content loading.
func (r *Router) MountRSS(httpRouter *gofastrRouter.Router, path string, cfg RSSConfig) error {
	if r == nil {
		return errors.New("docs: MountRSS requires a Router")
	}
	if httpRouter == nil {
		return errors.New("docs: MountRSS requires a GoFastr router")
	}
	path = normalizePath(path)
	if path == "" || path == "/" {
		return errors.New("docs: MountRSS requires a non-root path")
	}
	if _, err := rssSiteURL(cfg.SiteURL); err != nil {
		return err
	}
	httpRouter.Get(path, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		body, err := r.RSSXML(cfg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}))
	return nil
}

// WriteStaticRSS writes a feed to the same base-path layout used by static
// route exports. Call RSSXML first so the feed is generated from the Router.
func WriteStaticRSS(dir, basePath, feedPath string, body []byte) error {
	if strings.TrimSpace(dir) == "" {
		return errors.New("docs: WriteStaticRSS requires an output directory")
	}
	if len(body) == 0 {
		return errors.New("docs: WriteStaticRSS requires feed content")
	}
	if !strings.HasPrefix(strings.TrimSpace(feedPath), "/") {
		return fmt.Errorf("docs: WriteStaticRSS feed path %q must start with a slash", feedPath)
	}
	feedPath, err := safeStaticRoutePath(feedPath)
	if err != nil {
		return err
	}
	basePath = normalizeBasePath(basePath)
	target := filepath.Join(append([]string{dir}, splitStaticPath(basePath+feedPath)...)...)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(target, body, 0o644); err != nil {
		return fmt.Errorf("docs: write static RSS feed %q: %w", target, err)
	}
	return nil
}

func safeStaticRoutePath(raw string) (string, error) {
	clean := normalizePath(raw)
	if clean == "" || clean == "/" || strings.ContainsAny(clean, "?#\\") || strings.HasPrefix(clean, "//") {
		return "", fmt.Errorf("docs: invalid static route path %q", raw)
	}
	for _, segment := range strings.Split(strings.TrimPrefix(clean, "/"), "/") {
		if segment == "" || segment == "." || segment == ".." {
			return "", fmt.Errorf("docs: invalid static route path %q", raw)
		}
	}
	return clean, nil
}

func splitStaticPath(raw string) []string {
	parts := strings.Split(strings.TrimPrefix(raw, "/"), "/")
	filtered := parts[:0]
	for _, part := range parts {
		if part != "" {
			filtered = append(filtered, part)
		}
	}
	return filtered
}

func rssSiteURL(raw string) (string, error) {
	raw = strings.TrimRight(strings.TrimSpace(raw), "/")
	if raw == "" {
		return "", nil
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.User != nil {
		return "", fmt.Errorf("docs: RSS SiteURL must be an http or https base URL, got %q", raw)
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("docs: RSS SiteURL cannot contain a query or fragment, got %q", raw)
	}
	return raw, nil
}

func rssLink(siteURL, routePath string) string {
	if siteURL == "" {
		return routePath
	}
	return siteURL + routePath
}

func routeDescription(route *Route) string {
	if route == nil {
		return ""
	}
	// A curated excerpt wins: the author wrote it for exactly this slot.
	if strings.TrimSpace(route.Metadata.Excerpt) != "" {
		return strings.TrimSpace(route.Metadata.Excerpt)
	}
	if route.Description != "" {
		return route.Description
	}
	if route.page != nil {
		return firstParagraph(pageSource(route.page))
	}
	return ""
}

func rssDate(raw string) string {
	if value, ok := parseBlogDate(raw); ok {
		return value.UTC().Format(time.RFC1123Z)
	}
	return strings.TrimSpace(raw)
}

type rssDocument struct {
	XMLName   xml.Name   `xml:"rss"`
	Version   string     `xml:"version,attr"`
	XMLNSAtom string     `xml:"xmlns:atom,attr"`
	Channel   rssChannel `xml:"channel"`
}

// atomLink is the channel's self reference.
type atomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr,omitempty"`
}

type rssChannel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	// Language tells feed readers which language the channel is in,
	// derived from the collection's declared locale.
	Language string `xml:"language,omitempty"`
	// LastBuildDate is the newest publish date among the items, so a
	// reader can sort feeds by activity without fetching every post.
	Generator     string    `xml:"generator,omitempty"`
	LastBuildDate string    `xml:"lastBuildDate,omitempty"`
	AtomLink      atomLink  `xml:"atom:link"`
	Items         []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string   `xml:"title"`
	Link        string   `xml:"link"`
	GUID        rssGUID  `xml:"guid"`
	Description string   `xml:"description"`
	PubDate     string   `xml:"pubDate,omitempty"`
	Author      string   `xml:"author,omitempty"`
	Categories  []string `xml:"category,omitempty"`
}

type rssGUID struct {
	Value       string `xml:",chardata"`
	IsPermaLink string `xml:"isPermaLink,attr,omitempty"`
}
