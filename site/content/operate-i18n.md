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
and tag views, the 404 page. `WithUIStrings` translates all of them.
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
  es/
    getting-started.md        locale: es
    concepts-router.md        locale: es
```
{{< /filetree >}}

Try it: this page has no Spanish version, so the selector is absent here but
present on [Getting started](/docs/getting-started).

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
