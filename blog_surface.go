package docs

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"unicode"

	uiapp "github.com/DonaldMurillo/gofastr/core-ui/app"
	corehtml "github.com/DonaldMurillo/gofastr/core-ui/html"
	"github.com/DonaldMurillo/gofastr/core/render"
	"github.com/DonaldMurillo/gofastr/framework/ui"
)

// isBlogPrefix reports whether path is the root of a registered blog
// collection. Keeping this lookup in Router means a project can mount more
// than one blog without any global singleton or special URL convention.
func (r *Router) isBlogPrefix(path string) bool {
	if r == nil {
		return false
	}
	_, ok := r.blogs[normalizePath(path)]
	return ok
}

func (r *Router) isBlogView(path string) bool {
	if r == nil {
		return false
	}
	_, ok := r.blogViews[normalizePath(path)]
	return ok
}

func (r *Router) blogCollection(prefix string) blogCollection {
	return r.blogs[normalizePath(prefix)]
}

func (r *Router) blogPrefixes() string {
	if r == nil || len(r.blogs) == 0 {
		return ""
	}
	return strings.Join(r.blogPrefixesList(), ",")
}

func (r *Router) blogPrefixesList() []string {
	if r == nil || len(r.blogs) == 0 {
		return nil
	}
	prefixes := make([]string, 0, len(r.blogs))
	for prefix := range r.blogs {
		prefixes = append(prefixes, prefix)
	}
	sort.Strings(prefixes)
	return prefixes
}

func (r *Router) blogView(path string) blogView {
	return r.blogViews[normalizePath(path)]
}

func (r *Router) blogPublicPosts(prefix string) []*Route {
	all := r.BlogPosts(prefix)
	posts := make([]*Route, 0, len(all))
	for _, post := range all {
		if post.Metadata.NoIndex || !r.routePublished(post) {
			continue
		}
		posts = append(posts, post)
	}
	return posts
}

// registerBlogViews adds the aggregation surfaces that make a Markdown
// collection feel like a publication: search, archive, tags, authors, and
// archive pagination. They are ordinary Router routes, so they participate in
// navigation, search, static export, active state, and the same section
// layout as posts.
func (r *Router) registerBlogViews(prefix string) error {
	collection := r.blogCollection(prefix)
	cfg := collection.cfg
	posts := r.blogPublicPosts(prefix)
	nextOrder := cfg.PostOrderStart + len(posts) + 100
	for _, route := range r.Routes() {
		if route.Parent != nil && route.Parent.Path == prefix && route.Order >= nextOrder {
			nextOrder = route.Order + 1
		}
	}

	add := func(path string, view blogView, title, description string, order int, contextBody func(context.Context) render.HTML) error {
		view.prefix = prefix
		r.blogViews[normalizePath(path)] = view
		cfg := PageConfig{
			Title:       title,
			Description: description,
			Order:       order,
			Offline:     collection.cfg.Offline,
			DisableTOC:  true,
			Body:        func() render.HTML { return r.renderBlogView(path) },
			ContextBody: contextBody,
		}
		if err := r.Page(path, cfg); err != nil {
			delete(r.blogViews, normalizePath(path))
			return err
		}
		r.routes[normalizePath(path)].Blog = true
		return nil
	}

	if !cfg.DisableSearch {
		if err := add(joinPath(prefix, "search"), blogView{kind: blogViewSearch}, "Search", "Search the published posts.", nextOrder, func(ctx context.Context) render.HTML {
			return r.renderBlogSearch(prefix, ctx)
		}); err != nil {
			return err
		}
		nextOrder++
	}
	if !cfg.DisableArchive {
		if err := add(joinPath(prefix, "archive"), blogView{kind: blogViewArchive}, "Archive", "Browse every published post by year.", nextOrder, nil); err != nil {
			return err
		}
		nextOrder++
		for yearIndex, year := range blogYears(posts) {
			path := joinPath(prefix, "archive/"+year)
			if err := add(path, blogView{kind: blogViewArchive, term: year}, year+" archive", "Posts published in "+year+".", yearIndex+1, nil); err != nil {
				return err
			}
		}
	}
	if !cfg.DisableTags {
		if err := add(joinPath(prefix, "tags"), blogView{kind: blogViewTags}, "Tags", "Browse posts by topic.", nextOrder, nil); err != nil {
			return err
		}
		nextOrder++
		for index, tag := range blogTerms(posts, false) {
			path := joinPath(prefix, "tags/"+blogSlug(tag))
			if err := add(path, blogView{kind: blogViewTag, term: tag}, tag, "Posts tagged "+tag+".", index+1, nil); err != nil {
				return err
			}
		}
	}
	if !cfg.DisableAuthors {
		if err := add(joinPath(prefix, "authors"), blogView{kind: blogViewAuthors}, "Authors", "Browse posts by author.", nextOrder, nil); err != nil {
			return err
		}
		nextOrder++
		for index, author := range blogTerms(posts, true) {
			path := joinPath(prefix, "authors/"+blogSlug(author))
			if err := add(path, blogView{kind: blogViewAuthor, term: author}, author, "Posts by "+author+".", index+1, nil); err != nil {
				return err
			}
		}
	}

	for page := 2; page <= blogPageCount(len(posts), cfg.PostsPerPage); page++ {
		path := joinPath(prefix, "page/"+strconv.Itoa(page))
		if err := add(path, blogView{kind: blogViewPage, page: page}, "Blog · Page "+strconv.Itoa(page), "More posts from "+collection.cfg.Title+".", nextOrder, nil); err != nil {
			return err
		}
		nextOrder++
	}
	return nil
}

func blogPageCount(total, perPage int) int {
	if perPage < 1 {
		perPage = defaultBlogPostsPerPage
	}
	if total == 0 {
		return 1
	}
	return (total + perPage - 1) / perPage
}

func blogYears(posts []*Route) []string {
	seen := map[string]bool{}
	for _, post := range posts {
		if date, ok := parseBlogDate(post.Metadata.DatePublished); ok {
			seen[strconv.Itoa(date.Year())] = true
		}
	}
	years := make([]string, 0, len(seen))
	for year := range seen {
		years = append(years, year)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(years)))
	return years
}

func blogTerms(posts []*Route, authors bool) []string {
	seen := map[string]string{}
	for _, post := range posts {
		values := post.Metadata.Tags
		if authors {
			values = post.Metadata.Authors
		}
		for _, value := range values {
			value = strings.TrimSpace(value)
			if value == "" {
				continue
			}
			key := strings.ToLower(value)
			if _, exists := seen[key]; !exists {
				seen[key] = value
			}
		}
	}
	values := make([]string, 0, len(seen))
	for _, value := range seen {
		values = append(values, value)
	}
	sort.SliceStable(values, func(i, j int) bool { return strings.ToLower(values[i]) < strings.ToLower(values[j]) })
	return values
}

func blogSlug(value string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(strings.TrimSpace(value)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash && b.Len() > 0 {
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func (r *Router) renderBlogView(path string) render.HTML {
	view := r.blogView(path)
	switch view.kind {
	case blogViewArchive:
		return r.renderBlogArchiveView(view.prefix, view.term)
	case blogViewPage:
		return r.renderBlogArchivePage(view.prefix, view.page)
	case blogViewTags:
		return r.renderBlogTerms(view.prefix, false)
	case blogViewTag:
		return r.renderBlogTerm(view.prefix, false, view.term)
	case blogViewAuthors:
		return r.renderBlogTerms(view.prefix, true)
	case blogViewAuthor:
		return r.renderBlogTerm(view.prefix, true, view.term)
	case blogViewSearch:
		return r.renderBlogSearch(view.prefix, context.Background())
	default:
		return render.Text("")
	}
}

func (r *Router) renderBlogArchive(prefix string, page int, indexBody string) render.HTML {
	collection := r.blogCollection(prefix)
	posts := r.blogPublicPosts(prefix)
	intro := strings.TrimSpace(stripLeadingMarkdownTitle(indexBody))
	return r.blogArchiveShell(prefix, "Blog", collection.cfg.Description, intro, posts, page, true)
}

func (r *Router) renderBlogArchiveView(prefix, year string) render.HTML {
	posts := r.blogPublicPosts(prefix)
	if year != "" {
		posts = blogPostsForYear(posts, year)
	}
	description := "Browse every published post by year."
	if year != "" {
		description = "Posts published in " + year + "."
	}
	return r.blogArchiveShell(prefix, "Archive", description, "", posts, 1, false)
}

func (r *Router) renderBlogArchivePage(prefix string, page int) render.HTML {
	collection := r.blogCollection(prefix)
	return r.blogArchiveShell(prefix, "Page "+strconv.Itoa(page), "More posts from "+collection.cfg.Title+".", "", r.blogPublicPosts(prefix), page, false)
}

func blogPostsForYear(posts []*Route, year string) []*Route {
	filtered := make([]*Route, 0, len(posts))
	for _, post := range posts {
		if date, ok := parseBlogDate(post.Metadata.DatePublished); ok && strconv.Itoa(date.Year()) == year {
			filtered = append(filtered, post)
		}
	}
	return filtered
}

func (r *Router) blogArchiveShell(prefix, title, description, intro string, posts []*Route, page int, showFeatured bool) render.HTML {
	collection := r.blogCollection(prefix)
	if page < 1 {
		page = 1
	}
	perPage := collection.cfg.PostsPerPage
	if perPage < 1 {
		perPage = defaultBlogPostsPerPage
	}
	pageCount := blogPageCount(len(posts), perPage)
	if page > pageCount {
		page = pageCount
	}
	start := (page - 1) * perPage
	end := start + perPage
	if start > len(posts) {
		start = len(posts)
	}
	if end > len(posts) {
		end = len(posts)
	}
	selected := posts[start:end]
	listTitle := title
	if title == "Blog" {
		listTitle = "Latest posts"
	}

	children := []render.HTML{
		render.Tag("header", map[string]string{"class": "fastr-docs-blog__header"},
			render.Tag("p", map[string]string{"class": "fastr-docs-blog__eyebrow"}, render.Text("Publication")),
			render.Tag("h1", nil, render.Text(title)),
			render.Tag("p", map[string]string{"class": "fastr-docs-blog__lede"}, render.Text(description)),
			render.Tag("div", map[string]string{"class": "fastr-docs-blog__toolbar"}, r.blogToolbar(prefix)),
		),
	}
	if intro != "" {
		children = append(children, render.Tag("div", map[string]string{"class": "fastr-docs-blog__intro ui-markdown"}, renderDocsMarkdown(intro, nil)))
	}
	if showFeatured && len(posts) > 0 {
		children = append(children,
			render.Tag("section", map[string]string{"class": "fastr-docs-blog__featured", "aria-labelledby": "fastr-docs-blog-featured"},
				render.Tag("p", map[string]string{"class": "fastr-docs-blog__section-label", "id": "fastr-docs-blog-featured"}, render.Text("Featured")),
				r.blogPostCard(posts[0], true),
			),
		)
	}
	children = append(children,
		render.Tag("section", map[string]string{"class": "fastr-docs-blog__listing", "aria-labelledby": "fastr-docs-blog-list"},
			render.Tag("div", map[string]string{"class": "fastr-docs-blog__section-head"},
				render.Tag("h2", map[string]string{"id": "fastr-docs-blog-list"}, render.Text(listTitle)),
				render.Tag("span", map[string]string{"class": "fastr-docs-blog__count"}, render.Text(strconv.Itoa(len(posts))+" posts")),
			),
			render.Tag("div", map[string]string{"class": "fastr-docs-blog__cards"}, r.blogPostCards(selected, false)...),
		),
	)
	if len(selected) == 0 {
		children = append(children, render.Tag("p", map[string]string{"class": "fastr-docs-blog__empty"}, render.Text("No posts are published here yet.")))
	}
	if pageCount > 1 {
		children = append(children, r.blogPagination(prefix, page, pageCount))
	}
	return render.Tag("div", map[string]string{"class": "fastr-docs-blog-page fastr-docs-blog-page--archive", "data-blog-view": "archive"}, children...)
}

func (r *Router) blogToolbar(prefix string) render.HTML {
	collection := r.blogCollection(prefix)
	links := []render.HTML{
		render.Tag("a", map[string]string{"href": prefix, "class": "fastr-docs-blog__toolbar-link"}, render.Text("Latest")),
	}
	if !collection.cfg.DisableArchive {
		links = append(links, render.Tag("a", map[string]string{"href": joinPath(prefix, "archive"), "class": "fastr-docs-blog__toolbar-link"}, render.Text("Archive")))
	}
	if !collection.cfg.DisableTags {
		links = append(links, render.Tag("a", map[string]string{"href": joinPath(prefix, "tags"), "class": "fastr-docs-blog__toolbar-link"}, render.Text("Tags")))
	}
	if !collection.cfg.DisableAuthors {
		links = append(links, render.Tag("a", map[string]string{"href": joinPath(prefix, "authors"), "class": "fastr-docs-blog__toolbar-link"}, render.Text("Authors")))
	}
	if !collection.cfg.DisableSearch {
		links = append(links, render.Tag("a", map[string]string{"href": joinPath(prefix, "search"), "class": "fastr-docs-blog__toolbar-link"}, render.Text("Search")))
	}
	links = append(links, render.Tag("a", map[string]string{"href": joinPath(prefix, "feed.xml"), "class": "fastr-docs-blog__feed-link", "type": "application/rss+xml"}, render.Text("RSS")))
	return render.Tag("nav", map[string]string{"class": "fastr-docs-blog__toolbar-nav", "aria-label": "Blog views"}, links...)
}

func (r *Router) blogPostCards(posts []*Route, featured bool) []render.HTML {
	cards := make([]render.HTML, 0, len(posts))
	for _, post := range posts {
		cards = append(cards, r.blogPostCard(post, featured))
	}
	return cards
}

func (r *Router) blogPostCard(post *Route, featured bool) render.HTML {
	if post == nil {
		return render.Text("")
	}
	meta := render.Tag("div", map[string]string{"class": "fastr-docs-blog-card__meta"}, render.Text(formatBlogDate(post.Metadata.DatePublished)+" · "+blogReadingTime(post)+" min read"))
	header := render.Join(meta, render.Tag("h3", nil, render.Text(post.Title)))
	excerpt := post.Metadata.Excerpt
	if excerpt == "" {
		excerpt = post.Description
	}
	tags := r.blogTagLinks(post)
	className := "fastr-docs-blog-card"
	if featured {
		className += " fastr-docs-blog-card--featured"
	}
	card := ui.Card(ui.CardConfig{
		Href:    post.Path,
		Header:  header,
		Variant: ui.CardOutlined,
		Class:   className,
	}, render.Tag("p", map[string]string{"class": "fastr-docs-blog-card__excerpt"}, render.Text(excerpt)))
	// Tags are links in their own right. Keep them beside the card instead
	// of inside its anchor so the generated HTML remains valid and both the
	// post link and each topic filter stay independently keyboard accessible.
	tagRow := render.Tag("div", map[string]string{"class": "fastr-docs-blog-card__tags", "aria-label": "Topics"},
		render.Tag("span", map[string]string{"class": "ui-visually-hidden"}, render.Text("Topics: ")),
		render.Join(tags...))
	return render.Tag("div", map[string]string{"class": "fastr-docs-blog-card-wrap"}, card, tagRow)
}

func (r *Router) blogTagLinks(post *Route) []render.HTML {
	tags := make([]render.HTML, 0, len(post.Metadata.Tags))
	prefix := r.blogPrefixForPost(post)
	linkTags := !r.blogCollection(prefix).cfg.DisableTags
	for _, tag := range post.Metadata.Tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		href := ""
		if linkTags {
			href = joinPath(prefix, "tags/"+blogSlug(tag))
		}
		tags = append(tags, ui.Tag(ui.TagConfig{Label: tag, Href: href, Class: "fastr-docs-blog-card__tag"}))
	}
	return tags
}

func (r *Router) blogAuthorLinks(post *Route) []render.HTML {
	authors := make([]render.HTML, 0, len(post.Metadata.Authors))
	prefix := r.blogPrefixForPost(post)
	linkAuthors := !r.blogCollection(prefix).cfg.DisableAuthors
	for _, author := range post.Metadata.Authors {
		author = strings.TrimSpace(author)
		if author == "" {
			continue
		}
		if !linkAuthors || blogSlug(author) == "" {
			authors = append(authors, render.Text(author))
			continue
		}
		authors = append(authors, render.Tag("a", map[string]string{
			"href":  joinPath(prefix, "authors/"+blogSlug(author)),
			"class": "fastr-docs-blog-post__author",
		}, render.Text(author)))
	}
	return authors
}

func blogShareTargetID(route *Route) string {
	return "fastr-docs-blog-share-" + blogSlug(route.Path)
}

func (r *Router) blogPostActions(route *Route) render.HTML {
	targetID := blogShareTargetID(route)
	return corehtml.Group(corehtml.GroupConfig{
		Role:       "group",
		AriaLabel:  "Post actions",
		Class:      "fastr-docs-blog-post__actions",
		ExtraAttrs: corehtml.Attrs{"data-fastr-docs-share-surface": ""},
	},
		ui.Button(ui.ButtonConfig{
			Label:     "Share",
			Variant:   ui.ButtonGhost,
			Class:     "fastr-docs-blog-post__share",
			AriaLabel: "Share this post",
			ExtraAttrs: corehtml.Attrs{
				"data-fastr-docs-share":       "",
				"data-fastr-docs-share-title": route.Title,
				"data-fastr-docs-share-text":  route.Description,
				"data-fastr-docs-share-url":   route.Path,
			},
		}),
		ui.CopyButton(ui.CopyButtonConfig{
			Target:       "#" + targetID,
			IconOnly:     true,
			AriaLabel:    "Copy link",
			AnnounceText: "Link copied",
			ToastOnCopy:  true,
			ToastTitle:   "Link copied",
			Class:        "fastr-docs-blog-post__copy",
		}),
		render.Tag("span", map[string]string{
			"id":                           targetID,
			"class":                        "ui-visually-hidden",
			"data-fastr-docs-share-target": "",
			"data-fastr-docs-share-path":   route.Path,
		}, render.Text(route.Path)),
		render.Tag("span", map[string]string{
			"class":                        "ui-visually-hidden",
			"role":                         "status",
			"aria-live":                    "polite",
			"data-fastr-docs-share-status": "",
		}),
	)
}

func (r *Router) blogPrefixForPost(post *Route) string {
	for current := post; current != nil; current = current.Parent {
		if r.isBlogPrefix(current.Path) {
			return current.Path
		}
	}
	return "/blog"
}

func (r *Router) blogPagination(prefix string, page, pageCount int) render.HTML {
	links := make([]render.HTML, 0, 2)
	if page > 1 {
		links = append(links, render.Tag("a", map[string]string{"href": blogPageHref(prefix, page-1), "class": "fastr-docs-blog__pager-link"}, render.Text("← Newer posts")))
	}
	if page < pageCount {
		links = append(links, render.Tag("a", map[string]string{"href": blogPageHref(prefix, page+1), "class": "fastr-docs-blog__pager-link"}, render.Text("Older posts →")))
	}
	return render.Tag("nav", map[string]string{"class": "fastr-docs-blog__pagination", "aria-label": "Blog pagination"}, links...)
}

func blogPageHref(prefix string, page int) string {
	if page <= 1 {
		return prefix
	}
	return joinPath(prefix, "page/"+strconv.Itoa(page))
}

func (r *Router) renderBlogSearch(prefix string, ctx context.Context) render.HTML {
	query := ""
	if ctx != nil {
		query = strings.TrimSpace(uiapp.QueryFromContext(ctx).Get("q"))
	}
	posts := r.blogPublicPosts(prefix)
	results := posts
	if query != "" {
		results = make([]*Route, 0, len(posts))
		for _, post := range posts {
			if blogQueryMatch(post, query) {
				results = append(results, post)
			}
		}
	}
	label := "Search the publication"
	if query != "" {
		label = "Results for “" + query + "”"
	}
	body := []render.HTML{
		render.Tag("header", map[string]string{"class": "fastr-docs-blog__header"},
			render.Tag("p", map[string]string{"class": "fastr-docs-blog__eyebrow"}, render.Text("Publication")),
			render.Tag("h1", nil, render.Text("Search")),
			render.Tag("p", map[string]string{"class": "fastr-docs-blog__lede"}, render.Text("Find release notes, essays, and implementation updates.")),
			render.Tag("form", map[string]string{"class": "fastr-docs-blog-search", "action": joinPath(prefix, "search"), "method": "get", "role": "search", "data-fastr-docs-blog-search-form": ""},
				render.Tag("label", map[string]string{"for": "fastr-docs-blog-search-input"}, render.Text("Search posts")),
				render.Tag("input", map[string]string{"id": "fastr-docs-blog-search-input", "name": "q", "type": "search", "value": query, "placeholder": "Search posts", "autocomplete": "off", "data-fastr-docs-blog-search-input": ""}),
				render.Tag("button", map[string]string{"type": "submit"}, render.Text("Search")),
			),
			render.Tag("div", map[string]string{"class": "fastr-docs-blog__toolbar"}, r.blogToolbar(prefix)),
			render.Tag("div", map[string]string{"class": "fastr-docs-blog__search-summary", "data-fastr-docs-blog-search-summary": ""}, render.Text(label+" · "+strconv.Itoa(len(results))+" matches")),
		),
	}
	searchItems := r.blogSearchPostCards(posts, query)
	if len(searchItems) > 0 {
		body = append(body, render.Tag("div", map[string]string{"class": "fastr-docs-blog__cards", "data-fastr-docs-blog-search-results": ""}, searchItems...))
	}
	emptyAttrs := map[string]string{"class": "fastr-docs-blog__empty", "data-fastr-docs-blog-search-empty": ""}
	if len(results) > 0 {
		emptyAttrs["hidden"] = ""
	}
	body = append(body, render.Tag("p", emptyAttrs, render.Text("No posts matched that search.")))
	return render.Tag("div", map[string]string{"class": "fastr-docs-blog-page fastr-docs-blog-page--search", "data-blog-view": "search", "data-fastr-docs-blog-search": ""}, body...)
}

func (r *Router) blogSearchPostCards(posts []*Route, query string) []render.HTML {
	cards := make([]render.HTML, 0, len(posts))
	for _, post := range posts {
		if post == nil {
			continue
		}
		attrs := map[string]string{
			"class":                            "fastr-docs-blog-search-item",
			"data-fastr-docs-blog-search-item": blogSearchText(post),
			"data-pagefind-ignore":             "",
		}
		if strings.TrimSpace(query) != "" && !blogQueryMatch(post, query) {
			attrs["hidden"] = ""
		}
		cards = append(cards, render.Tag("div", attrs, r.blogPostCard(post, false)))
	}
	return cards
}

func blogSearchText(post *Route) string {
	if post == nil {
		return ""
	}
	return strings.Join([]string{
		post.Title, post.Description, post.Metadata.Excerpt,
		strings.Join(post.Metadata.Tags, " "), strings.Join(post.Metadata.Authors, " "),
		truncateBlogSearchText(pageSource(post.page), 800),
	}, " ")
}

func truncateBlogSearchText(value string, limit int) string {
	value = strings.TrimSpace(value)
	if limit < 1 || len(value) <= limit {
		return value
	}
	return value[:limit] + " …"
}

func blogQueryMatch(post *Route, query string) bool {
	terms := strings.Fields(strings.ToLower(query))
	if len(terms) == 0 {
		return true
	}
	text := strings.ToLower(blogSearchText(post))
	for _, term := range terms {
		if !strings.Contains(text, term) {
			return false
		}
	}
	return true
}

func (r *Router) renderBlogTerms(prefix string, authors bool) render.HTML {
	label := "Tags"
	termLabel := "topics"
	if authors {
		label = "Authors"
		termLabel = "authors"
	}
	terms := blogTerms(r.blogPublicPosts(prefix), authors)
	items := make([]render.HTML, 0, len(terms))
	for _, term := range terms {
		pathPart := "tags"
		if authors {
			pathPart = "authors"
		}
		count := len(r.blogPostsForTerm(prefix, authors, term))
		items = append(items, render.Tag("a", map[string]string{"href": joinPath(prefix, pathPart+"/"+blogSlug(term)), "class": "fastr-docs-blog-term"},
			render.Tag("strong", nil, render.Text(term)), render.Tag("span", nil, render.Text(strconv.Itoa(count)+" posts"))))
	}
	body := []render.HTML{
		render.Tag("header", map[string]string{"class": "fastr-docs-blog__header"},
			render.Tag("p", map[string]string{"class": "fastr-docs-blog__eyebrow"}, render.Text("Publication")),
			render.Tag("h1", nil, render.Text(label)),
			render.Tag("p", map[string]string{"class": "fastr-docs-blog__lede"}, render.Text("Explore "+termLabel+" across the publication.")),
			render.Tag("div", map[string]string{"class": "fastr-docs-blog__toolbar"}, r.blogToolbar(prefix)),
		),
		render.Tag("div", map[string]string{"class": "fastr-docs-blog-terms"}, items...),
	}
	if len(items) == 0 {
		body = append(body, render.Tag("p", map[string]string{"class": "fastr-docs-blog__empty"}, render.Text("Nothing has been classified yet.")))
	}
	return render.Tag("div", map[string]string{"class": "fastr-docs-blog-page fastr-docs-blog-page--terms", "data-blog-view": strings.ToLower(label)}, body...)
}

func (r *Router) renderBlogTerm(prefix string, authors bool, term string) render.HTML {
	posts := r.blogPostsForTerm(prefix, authors, term)
	label := "Tag"
	if authors {
		label = "Author"
	}
	return r.blogArchiveShell(prefix, label+": "+term, "Published posts in this collection.", "", posts, 1, false)
}

func (r *Router) blogPostsForTerm(prefix string, authors bool, term string) []*Route {
	term = strings.ToLower(strings.TrimSpace(term))
	posts := make([]*Route, 0)
	for _, post := range r.blogPublicPosts(prefix) {
		values := post.Metadata.Tags
		if authors {
			values = post.Metadata.Authors
		}
		for _, value := range values {
			if strings.ToLower(strings.TrimSpace(value)) == term {
				posts = append(posts, post)
				break
			}
		}
	}
	return posts
}

func (r *Router) wrapBlogPost(route *Route, body render.HTML, source string) render.HTML {
	headings := markdownHeadings(source)
	meta := route.Metadata
	metaParts := []render.HTML{}
	if date := formatBlogDate(meta.DatePublished); date != "" {
		metaParts = append(metaParts, render.Tag("span", map[string]string{"class": "fastr-docs-blog-post__meta-item"},
			render.Text("Published "), render.Tag("time", map[string]string{"datetime": meta.DatePublished}, render.Text(date))))
	}
	if authors := r.blogAuthorLinks(route); len(authors) > 0 {
		authorParts := []render.HTML{render.Text("By ")}
		for index, author := range authors {
			if index > 0 {
				authorParts = append(authorParts, render.Text(", "))
			}
			authorParts = append(authorParts, author)
		}
		metaParts = append(metaParts, render.Tag("span", map[string]string{"class": "fastr-docs-blog-post__meta-item"}, render.Join(authorParts...)))
	}
	if modified := formatBlogDate(meta.DateModified); modified != "" && modified != formatBlogDate(meta.DatePublished) {
		metaParts = append(metaParts, render.Tag("span", map[string]string{"class": "fastr-docs-blog-post__meta-item"},
			render.Text("Updated "), render.Tag("time", map[string]string{"datetime": meta.DateModified}, render.Text(modified))))
	}
	metaParts = append(metaParts, render.Text(blogReadingTime(route)+" min read"))
	titleID := blogShareTargetID(route) + "-title"
	header := render.Tag("header", map[string]string{"class": "fastr-docs-blog-post__header"},
		render.Tag("p", map[string]string{"class": "fastr-docs-blog__eyebrow"}, render.Text("Publication")),
		render.Tag("h1", map[string]string{"id": titleID}, render.Text(route.Title)),
		render.Tag("p", map[string]string{"class": "fastr-docs-blog-post__lede"}, render.Text(route.Description)),
		render.Tag("div", map[string]string{"class": "fastr-docs-blog-post__meta"}, joinWithDot(metaParts...)),
		render.Tag("div", map[string]string{"class": "fastr-docs-blog-post__tags"}, r.blogTagLinks(route)...),
		r.blogPostActions(route),
	)
	content := render.Tag("div", map[string]string{"class": "fastr-docs-blog-post__content ui-markdown", "data-blog-content": route.Path}, header, body)
	article := corehtml.Article(corehtml.ArticleConfig{
		Class: "fastr-docs-blog-post__article",
		ID:    blogShareTargetID(route) + "-article",
		ExtraAttrs: corehtml.Attrs{
			"aria-labelledby":                titleID,
			"data-fastr-docs-reader-article": "",
		},
	}, content)
	children := []render.HTML{render.Tag("div", map[string]string{"class": "fastr-docs-blog-post__crumbs"}, r.blogCrumbs(route)...)}
	if len(headings) > 0 {
		items := make([]ui.RailItem, 0, len(headings))
		for _, heading := range headings {
			items = append(items, ui.RailItem{Anchor: heading.ID, Text: heading.Title})
		}
		rail := ui.AnchoredRail(ui.AnchoredRailConfig{
			Label: "On this page", Items: items,
			ObserveSelector: ".fastr-docs-blog-post__content",
			TargetSelector:  "h2[id], h3[id]",
			Class:           "fastr-docs-toc fastr-docs-toc--rail",
		})
		children = append(children, render.Tag("div", map[string]string{"class": "fastr-docs-blog-post__grid"}, article, corehtml.Aside(corehtml.AsideConfig{Label: "On this page", Class: "fastr-docs-blog-post__toc"}, rail, r.docsTocSelect(headings))))
	} else {
		children = append(children, article)
	}
	if editURL := safeMetadataURL(route.Metadata.EditURL); editURL != "" {
		children = append(children, render.Tag("p", map[string]string{"class": "fastr-docs-blog-post__edit"}, render.Tag("a", map[string]string{"href": editURL, "rel": "nofollow noopener", "target": "_blank"}, render.Text(r.UIStrings().EditPage))))
	}
	if pager := r.blogPager(route); pager != nil {
		children = append(children, ui.DocPrevNext(*pager))
	}
	if related := r.blogRelated(route); len(related) > 0 {
		children = append(children, corehtml.Section(corehtml.SectionConfig{Class: "fastr-docs-blog-post__related", LabelledBy: "fastr-docs-blog-related"},
			render.Tag("h2", map[string]string{"id": "fastr-docs-blog-related"}, render.Text("Keep reading")),
			render.Tag("div", map[string]string{"class": "fastr-docs-blog__cards"}, r.blogPostCards(related, false)...),
		))
	}
	return render.Tag("div", map[string]string{"class": "fastr-docs-blog-page fastr-docs-blog-page--post", "data-blog-view": "post"}, children...)
}

func joinWithDot(parts ...render.HTML) render.HTML {
	if len(parts) == 0 {
		return ""
	}
	joined := make([]render.HTML, 0, len(parts)*2-1)
	for index, part := range parts {
		if index > 0 {
			joined = append(joined, render.Text(" · "))
		}
		joined = append(joined, part)
	}
	return render.Join(joined...)
}

func (r *Router) blogCrumbs(route *Route) []render.HTML {
	crumbs := []render.HTML{render.Tag("a", map[string]string{"href": "/"}, render.Text(r.SiteName()))}
	if prefix := r.blogPrefixForPost(route); prefix != "" {
		crumbs = append(crumbs, render.Text(" / "), render.Tag("a", map[string]string{"href": prefix}, render.Text(r.blogCollection(prefix).cfg.Title)))
	}
	crumbs = append(crumbs, render.Text(" / "), render.Tag("span", map[string]string{"aria-current": "page"}, render.Text(route.Title)))
	return []render.HTML{corehtml.Nav(corehtml.NavConfig{Label: "Breadcrumb"}, crumbs...)}
}

func (r *Router) blogPager(route *Route) *ui.DocPager {
	prefix := r.blogPrefixForPost(route)
	posts := r.blogPublicPosts(prefix)
	index := -1
	for i, post := range posts {
		if post.Path == route.Path {
			index = i
			break
		}
	}
	if index < 0 {
		return nil
	}
	pager := &ui.DocPager{PrevHref: prefix, PrevLabel: "All posts"}
	if index+1 < len(posts) {
		pager.PrevHref = posts[index+1].Path
		pager.PrevLabel = posts[index+1].Title
	}
	if index > 0 {
		pager.NextHref = posts[index-1].Path
		pager.NextLabel = posts[index-1].Title
	}
	return pager
}

func (r *Router) blogRelated(route *Route) []*Route {
	prefix := r.blogPrefixForPost(route)
	limit := r.blogCollection(prefix).cfg.RelatedPosts
	if limit < 1 {
		limit = 3
	}
	posts := r.blogPublicPosts(prefix)
	related := make([]*Route, 0, limit)
	for _, candidate := range posts {
		if candidate.Path == route.Path {
			continue
		}
		if !blogSharesTag(route, candidate) {
			continue
		}
		related = append(related, candidate)
		if len(related) == limit {
			return related
		}
	}
	return related
}

func blogSharesTag(left, right *Route) bool {
	for _, a := range left.Metadata.Tags {
		for _, b := range right.Metadata.Tags {
			if strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b)) && strings.TrimSpace(a) != "" {
				return true
			}
		}
	}
	return false
}

func blogReadingTime(route *Route) string {
	words := len(strings.Fields(pageSource(route.page)))
	minutes := (words + 199) / 200
	if minutes < 1 {
		minutes = 1
	}
	return strconv.Itoa(minutes)
}

func formatBlogDate(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if date, ok := parseBlogDate(raw); ok {
		return date.Format("Jan 2, 2006")
	}
	return raw
}

func stripLeadingMarkdownTitle(source string) string {
	lines := strings.Split(strings.ReplaceAll(source, "\r\n", "\n"), "\n")
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[0]), "# ") {
		lines = lines[1:]
		for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
			lines = lines[1:]
		}
	}
	return strings.Join(lines, "\n")
}

type blogSidebar struct {
	router *Router
	prefix string
}

func (s *blogSidebar) Render() render.HTML { return s.render("") }

func (s *blogSidebar) RenderCtx(ctx context.Context) render.HTML {
	currentPath := ""
	if request := uiapp.RequestFromContext(ctx); request != nil && request.URL != nil {
		currentPath = request.URL.Path
	}
	return s.render(currentPath)
}

func (s *blogSidebar) render(currentPath string) render.HTML {
	return ui.Sidebar(s.config(currentPath, blogDrawerName(s.prefix))).Render()
}

func (s *blogSidebar) config(currentPath, drawer string) ui.SidebarConfig {
	if drawer == "" {
		drawer = blogDrawerName(s.prefix)
	}
	posts := s.router.blogPublicPosts(s.prefix)
	items := []ui.SidebarItem{
		{Label: "All posts", Href: s.prefix, Icon: sidebarIcon(KindPage, NavBadge{}, s.prefix, false), Active: strings.TrimSpace(currentPath) == "" || normalizePath(currentPath) == normalizePath(s.prefix)},
	}
	collection := s.router.blogCollection(s.prefix)
	if !collection.cfg.DisableArchive {
		items = append(items, ui.SidebarItem{Label: "Archive", Href: joinPath(s.prefix, "archive"), Icon: sidebarIcon(KindPage, NavBadge{}, joinPath(s.prefix, "archive"), false), Active: normalizePath(currentPath) == joinPath(s.prefix, "archive"), MatchPath: joinPath(s.prefix, "archive")})
	}
	if !collection.cfg.DisableTags {
		items = append(items, ui.SidebarItem{Label: "Tags", Href: joinPath(s.prefix, "tags"), Icon: sidebarIcon(KindPage, NavBadge{}, joinPath(s.prefix, "tags"), false), MatchPath: joinPath(s.prefix, "tags")})
	}
	if !collection.cfg.DisableAuthors {
		items = append(items, ui.SidebarItem{Label: "Authors", Href: joinPath(s.prefix, "authors"), Icon: sidebarIcon(KindPage, NavBadge{}, joinPath(s.prefix, "authors"), false), MatchPath: joinPath(s.prefix, "authors")})
	}
	if !collection.cfg.DisableSearch {
		items = append(items, ui.SidebarItem{Label: "Search", Href: joinPath(s.prefix, "search"), Icon: sidebarIcon(KindPage, NavBadge{}, joinPath(s.prefix, "search"), false), MatchPath: joinPath(s.prefix, "search")})
	}
	if len(posts) > 0 {
		recent := posts
		if len(recent) > 5 {
			recent = recent[:5]
		}
		recentItems := make([]ui.SidebarItem, 0, len(recent))
		for _, post := range recent {
			recentItems = append(recentItems, ui.SidebarItem{Label: post.Title, Href: post.Path, Icon: sidebarIcon(KindPage, NavBadge{}, post.Path, false), Active: normalizePath(currentPath) == normalizePath(post.Path)})
		}
		items = append(items, ui.SidebarItem{Label: "Recent posts", Icon: sidebarIcon(KindGroup, NavBadge{}, s.prefix, false), Children: recentItems})
	}
	return ui.SidebarConfig{
		Title: "Blog", NavLabel: "Blog navigation", Items: items,
		CurrentPath: currentPath, DrawerName: drawer, SuppressDrawerTrigger: true,
	}
}

func blogDrawerName(prefix string) string {
	if normalizePath(prefix) == "/blog" {
		return "fastr-docs-blog-sections"
	}
	value := strings.Trim(normalizePath(prefix), "/")
	value = strings.NewReplacer("/", "-", "_", "-", ".", "-").Replace(value)
	if value == "" {
		value = "publication"
	}
	return "fastr-docs-blog-" + value + "-sections"
}

func (r *Router) blogDrawerNames() string {
	prefixes := make([]string, 0, len(r.blogs))
	for prefix := range r.blogs {
		prefixes = append(prefixes, prefix)
	}
	sort.Strings(prefixes)
	pairs := make([]string, 0, len(prefixes))
	for _, prefix := range prefixes {
		pairs = append(pairs, prefix+"="+blogDrawerName(prefix))
	}
	return strings.Join(pairs, ";")
}
