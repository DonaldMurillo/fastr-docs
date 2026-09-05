---
tags: [content, publishing, rss]
---

# Blog and RSS

Use `MarkdownBlog` when a docs site also publishes updates. It registers an
archive route at the prefix and turns every other Markdown file in the
directory into a post. Posts use the existing front matter contract, so the
same Router handles titles, descriptions, authors, dates, tags, drafts,
locales, versions, search, navigation, and SEO.

```go
if err := router.MarkdownBlog("/blog", "content/blog", docs.BlogConfig{
    Title:          "Updates",
    Description:    "Release notes and product updates.",
    Order:          5,
    Offline:        true,
    PostOrderStart: 10,
}); err != nil {
    return err
}
```

`content/blog/index.md` is the archive content. A file such as
`content/blog/first-post.md` becomes `/blog/first-post`. Set `slug` in its
front matter when the filename should not determine the URL.

`BlogPosts("/blog")` returns published posts newest first. It includes
`noindex` posts for custom archives, but excludes drafts unless the Router is
in preview mode.

## A separate publication layout

The blog is a sibling of Documentation in the global shell. It has its own
sidebar and responsive post template instead of inheriting the documentation
contents tree. `MarkdownBlog` creates these aggregate routes automatically:

- `/blog` for latest posts and pagination
- `/blog/archive` and `/blog/archive/:year` for chronological browsing
- `/blog/tags` and `/blog/tags/:tag` for topic filters
- `/blog/authors` and `/blog/authors/:author` for author filters
- `/blog/search?q=...` for server-rendered search

The same collection drives the cards, RSS feed, reading time, related posts,
article metadata, mobile navigation, and static export. Aggregate views are
enabled by default and can be disabled individually:

```go
docs.BlogConfig{
    PostsPerPage:   10,
    RelatedPosts:   3,
    DisableAuthors: true,
}
```

Post pages render as real articles with labelled breadcrumbs and table of
contents, published and updated times, keyboard-sized actions, and live
announcements for assistive technology. Share uses the Web Share API when it
is available and copies the page URL otherwise. Copy link uses GoFastr's
clipboard component, including its screen-reader announcement. Tags and
authors link to filtered collections, and shared tags populate a Keep
reading section on related posts.

Use `excerpt` in front matter when the first paragraph is not the right card
summary. `<!-- truncate -->` is also supported. If `date` is omitted,
fastr-docs recognizes a `YYYY-MM-DD` filename prefix.

## Generate the feed

RSS is a projection of the same post routes. It is not a second collection.

```go
feed := docs.RSSConfig{
    Prefix:      "/blog",
    Title:       "Acme updates",
    Description: "Release notes and product updates.",
    SiteURL:     "https://docs.example.com",
    Limit:       20,
}

if err := router.MountRSS(server.Router(), "/blog/feed.xml", feed); err != nil {
    return err
}
```

`MountRSS` serves the feed from the live SSR host. For a static deployment,
call `router.RSSXML(feed)` and `docs.WriteStaticRSS` after GoFastr's static
export. The feed uses the active locale and version, removes drafts and
`noindex` posts, orders by `date` descending, and escapes XML values. Set
`SiteURL` in production so feed readers receive absolute links.
