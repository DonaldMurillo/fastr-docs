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
front matter when the filename should not determine the URL:

```yaml
---
title: A longer release note
description: What changed in this release.
date: 2026-08-30
authors: [Project team]
tags: [release]
slug: releases/august
---
```

`BlogPosts("/blog")` returns the published posts newest first. It includes
`noindex` posts for custom archives, but excludes drafts unless the Router is
in preview mode. Use the returned routes to build a custom archive screen when
the generated Markdown archive is not enough.

## A separate publication layout

The blog is a sibling of Documentation in the global shell. It does not reuse
the documentation section's contents tree or its in-page navigation. The
generated blog layout provides:

- `/blog` — the latest-posts landing page with a featured post and pagination
- `/blog/archive` and `/blog/archive/:year` — chronological archive views
- `/blog/tags` and `/blog/tags/:tag` — topic indexes and filtered views
- `/blog/authors` and `/blog/authors/:author` — author indexes and filtered views
- `/blog/search?q=...` — server-rendered publication search
- `/blog/feed.xml` — the RSS entry point mounted by the host

Search, taxonomy, cards, related posts, reading time, breadcrumbs, article
metadata, and the responsive mobile drawer all come from the same Router
collection. Post pages get a reading-oriented template with an optional
in-page table of contents; archive pages do not inherit the docs TOC.

Post pages are real articles, not only styled containers. The generated
markup exposes the title relationship, published and updated times, a
labelled breadcrumb and table-of-contents region, keyboard-sized actions,
and live announcements for assistive technology. The Share action uses the
Web Share API when the device provides it and copies the canonical page URL
as a fallback. Copy link uses GoFastr's clipboard component and its built-in
screen-reader announcement.

Topic tags and authors link to their filtered collections. Posts with shared
tags receive a Keep reading section, so readers can continue into related
work without returning to the archive.

The aggregate views are on by default. A project can opt out of individual
surfaces without replacing the layout:

```go
docs.BlogConfig{
    PostsPerPage:   10,
    RelatedPosts:   3,
    DisableAuthors: true,
}
```

The front matter also accepts `excerpt`. When it is omitted, fastr-docs uses
the first paragraph, or the content before `<!-- truncate -->`, for cards,
search results, and social previews. A `date` remains the canonical date;
`YYYY-MM-DD` at the start of a filename is used as a fallback.

## Generate the feed

RSS is a projection of the same post routes. It is not a second collection or
another publication database.

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
generate the same bytes and write them beside the exported pages:

```go
body, err := router.RSSXML(feed)
if err != nil {
    return err
}
if err := docs.WriteStaticRSS(dist, "/docs", "/blog/feed.xml", body); err != nil {
    return err
}
```

The feed uses the active locale and version, removes drafts and `noindex`
posts, orders by `date` descending, and escapes XML values. Set `SiteURL` in
production so feed readers receive absolute links. GoFastr still owns SSR,
article metadata, hydration, and static page export around these routes.
