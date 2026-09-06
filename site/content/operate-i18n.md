---
tags: [i18n, localization, operate]
locale: en
---

# Languages and translation

Two separate things get translated, and conflating them is the usual source of
confusion.

{{< tabs >}}
{{< tab label="Content" >}}
Your pages. Each one declares `locale:` in front matter, and pages that share a
route family are linked to each other. The Router never invents a translation.
{{< /tab >}}
{{< tab label="Chrome" >}}
The framework's own labels: search, on-this-page, updated, the blog's archive
and tag views. `WithUIStrings` sets them for the whole site;
`WithLocaleUIStrings` sets them per language, which is what a single build
serving several languages needs.
{{< /tab >}}
{{< /tabs >}}

## Route families

A family is a page's path with its locale and version segments removed. This
site's Spanish pages sit under `/es`, so `/es/docs/getting-started` and
`/docs/getting-started` are one family and the header shows a language selector
between them.

{{< filetree >}}
```
content/
  getting-started.md          locale: en
  concepts-router.md          locale: en
  operate-i18n.md             locale: en
  es/
    getting-started.md        locale: es
    concepts-router.md        locale: es
    operate-i18n.md           locale: es
```
{{< /filetree >}}

The `es/` directory is a convention, not a rule. Nothing reads the directory
name; the pairing comes from `locale:` in front matter and from the `/es` path
segment.

{{< warning title="Both sides need a locale" >}}
A page with no `locale:` is not in any language, so it pairs with nothing. Set
it on the original as well as the translation, or the selector silently fails to
appear and there is no error to tell you why.
{{< /warning >}}

Try it: the selector at the top of this page moves you to
[the Spanish version](/es/docs/operate/i18n), chrome and all. On a page with no
translation, like [Math](/docs/build/math), the selector is absent, because
there is nowhere to go.

## Translating the chrome

Empty fields keep their English defaults, so you can translate one label at a
time rather than all of them at once.

```go title="docs/router.go" {2,5}
router := docs.NewRouter(docs.WithUIStrings(docs.UIStrings{
    Search:     "Buscar",
    DateFormat: "02/01/2006",
    Blog: docs.BlogStrings{
        Archive:    "Archivo",
        ResultsFor: "Resultados para «%s»",
    },
    NotFound: docs.NotFoundStrings{Heading: "Página no encontrada"},
}))
```

Labels containing `%s` or `%d` are format strings. A translation may reorder
the words around the placeholder, and dropping it entirely is tolerated rather
than printing Go's `%!(EXTRA ...)` into the page.

{{< note title="Why no CLDR" >}}
`DateFormat` is a Go time layout you choose. The standard library ships no CLDR
data, so deriving a date format from the locale would mean inventing one.
{{< /note >}}

### When one build serves several languages

`WithUIStrings` sets labels for the whole Router, which is right when a build
serves one language. This site serves English and Spanish from a single build,
so it needs a set per locale instead. Without that, a Spanish page arrives
wrapped in English furniture: "Contents", "On this page", "Search".

```go title="docs/router.go"
docs.WithLocaleNames(map[string]string{"en": "English", "es": "Español"}),
docs.WithLocaleUIStrings("es", docs.UIStrings{
    Contents:   "Contenido",
    OnThisPage: "En esta página",
    Search:     "Buscar",
    Language:   "Idioma",
}),
```

Resolution follows the page, not the build, so `/es/docs/getting-started` gets
Spanish chrome and `/docs/getting-started` gets English from the same binary.
Anything a locale leaves out falls back to `WithUIStrings`, then to English.

`WithLocaleNames` is what puts "Español" in the selector instead of "es". Go
ships no locale display names, so you supply them rather than the framework
guessing.

One piece cannot follow the page: the command palette modal is mounted once for
the whole site, so its placeholder stays in the default language. The search
button in the header does follow the page.

A translated section is the same section, so it sits behind the language
selector rather than beside the original as its own nav tab. Route titles stay
in the language you registered them in.

## Building one locale at a time

`WithLocale` slices the site to a single content locale. On its own that
deletes untranslated pages from the build, which leaves holes.

```go title="docs/router.go" {3}
router := docs.NewRouter(
    docs.WithLocale(os.Getenv("DOCS_CONTENT_LOCALE")),
    docs.WithLocaleFallback("en"),
)
```

With a fallback, a family that has no page in the requested locale serves the
default-locale page instead. It never shadows a real translation: when the
translated page exists, that one wins.

## Knowing what is missing

{{< steps >}}
1. `Router.LocaleCoverage()` returns each locale and the families it is missing.
2. `Router.UntranslatedFamilies("es")` answers for one locale.
3. `fastr-docs check .` logs the gaps.
{{< /steps >}}

Coverage is advisory. A partially translated site is a normal state, not a
broken build, so nothing here fails validation.

## What this site actually translates

Five of about thirty pages are in Spanish. That is on purpose: a half-translated
site is what every real one looks like, and it is the state worth demonstrating.

| Surface | Follows the page's language |
| --- | --- |
| Page content | Yes, when a translation exists |
| Sidebar, on-this-page, search button, dates | Yes |
| Language selector | Yes, and it names languages rather than codes |
| Section tabs and groups in the header and sidebar | Yes, where a translation of that section exists |
| Command palette placeholder | No, it is mounted once for the whole site |
| Blog archive, tags, and post chrome | No, those labels are site-wide |
| Previous/next pager | Yes, labels and neighbours both |
| Search results | Yes, narrowed to the page's language |
| The 404 page | No, it is built once and has no route to read a locale from |

The last three are limitations, not decisions, and they are written down in
`AGENTS.md` rather than hidden here.

## Translating a section, not just a page

A page declares `locale:` in front matter. A group has no front matter, so it
declares its language in Go:

```go title="docs/router.go"
conceptos := router.MustGroup("/es/docs/concepts", docs.GroupConfig{
    Title:  "Conceptos",
    Locale: "es",
})
```

Without that, the group is in no language, pairs with nothing, and the header
tab and sidebar group stay in the original language while the article beneath
them is translated.

Original pages need no annotation. When `WithLocaleFallback` names a default
locale, a page with no `locale:` counts as being in it, so adding a language
does not mean editing every existing file.

## Search across languages

One index holds every language, so results are narrowed to the page you are
reading. A Spanish query answers with Spanish pages; following one does not
quietly drop you back into English.

Pagefind is separate, and stricter: it decides which language index a page
belongs to by reading `<html lang>`. GoFastr writes that from a single
site-wide value, so every page claims the same language and Pagefind builds one
index with the wrong stemming. `WriteExportLocales` stamps each exported page
with its route's locale after the export, before Pagefind runs:

```go title="main.go"
router.WriteExportLocales(dist)
```

With it, Pagefind reports `Discovered 2 languages: en, es` and writes a separate
chunked index for each.

This fixes the export, which is what Pagefind reads. The live server still
serves one language attribute for the whole site, and that needs a change in
GoFastr.
