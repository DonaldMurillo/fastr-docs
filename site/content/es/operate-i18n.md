---
tags: [i18n, localizacion, operate]
locale: es
---

# Idiomas y traducción

Esta página es la traducción de [Languages and translation](/docs/operate/i18n),
y sirve de ejemplo de sí misma: la estás leyendo con el contenido y la interfaz
en español.

Se traducen dos cosas distintas, y confundirlas es el origen habitual de los
problemas.

{{< tabs >}}
{{< tab label="El contenido" >}}
Tus páginas. Cada una declara `locale:` en el front matter, y las páginas que
comparten familia de ruta quedan enlazadas entre sí. El Router nunca inventa una
traducción.
{{< /tab >}}
{{< tab label="La interfaz" >}}
Las etiquetas del framework: la búsqueda, el índice de la página, la fecha de
actualización, el 404. `WithLocaleUIStrings` las traduce por idioma.
{{< /tab >}}
{{< /tabs >}}

## Familias de rutas

Una familia es la ruta de una página sin sus segmentos de idioma y de versión.
Las páginas en español de este sitio viven bajo `/es`, así que
`/es/docs/operate/i18n` y `/docs/operate/i18n` son la misma familia, y por eso
la cabecera muestra un selector de idioma entre ellas.

````md
content/
  operate-i18n.md             locale: en
  es/
    operate-i18n.md           locale: es
````

{{< warning title="El locale va en las dos" >}}
Una página sin `locale:` no está en ningún idioma, así que no se empareja con
nada. Ponlo en el original además de en la traducción, o el selector no aparece
y no hay ningún error que te diga por qué.
{{< /warning >}}

## Traducir también el slug

El emparejamiento por ruta necesita que la ruta traducida sea la original más
un segmento de idioma. Sirve para un árbol traducido carpeta por carpeta, que
es como se empareja casi todo este sitio. No puede emparejar
`/es/ejemplos/arbol-de-rutas` con `/examples/route-tree`, porque al quitar
`es` queda otra ruta.

Una página con el slug traducido nombra su original:

```md title="content/es/examples-route-tree.md"
---
locale: es
translation_of: /examples/route-tree
---
```

Un grupo hace lo mismo en Go con `TranslationOf: "/examples"`. La página entra
en la familia del original, así que la sigue todo lo que lee familias: el
selector de idioma, la pestaña de la cabecera, la sección de la barra lateral,
el paginador, los resultados de búsqueda y el informe de cobertura. Las dos
páginas reciben enlaces `hreflang` la una a la otra sin escribir `alternates:`.

`fastr-docs check` falla si `translation_of` apunta a una ruta que nadie
sirve, a la propia página o a una página del mismo idioma.

Pruébalo: el [ejemplo de árbol de rutas](/es/ejemplos/arbol-de-rutas) está
emparejado así, y su selector lleva a `/examples/route-tree`.

Pruébalo con el selector de arriba. En una página sin traducción, como
[Math](/docs/build/math), el selector no aparece: no hay nada a lo que cambiar.

## Traducir la interfaz

Un solo binario sirve los dos idiomas, así que las etiquetas se resuelven por
página, no por sitio.

```go title="docs/router.go"
docs.WithLocaleNames(map[string]string{"en": "English", "es": "Español"}),
docs.WithLocaleUIStrings("es", docs.UIStrings{
    Contents:   "Contenido",
    OnThisPage: "En esta página",
    Search:     "Buscar",
    Language:   "Idioma",
}),
```

Lo que un idioma no traduce cae de vuelta a `WithUIStrings` y luego al inglés,
así que puedes traducir una etiqueta cada vez en lugar de todas de golpe.

{{< note title="Por qué no hay CLDR" >}}
`DateFormat` es un layout de fecha de Go que eliges tú. La biblioteca estándar
no trae datos CLDR, así que deducir el formato a partir del idioma sería
inventárselo. Go además escribe los meses solo en inglés, así que un layout que
los nombra toma los nombres de `Months` (doce, empezando por enero) o de
`ShortMonths`. Esta página usa `"2 de January de 2006"` con los meses en
español y sale `30 de agosto de 2026`.
{{< /note >}}

`WithLocaleNames` es lo que pone "Español" en el selector en lugar de "es". Go
no incluye nombres de idioma, así que los aporta el proyecto en vez de que el
framework los adivine.

## Cuando falta una traducción

`WithLocale` recorta el sitio a un solo idioma de contenido. Por sí solo,
eso borra del build las páginas sin traducir y deja huecos.

```go title="docs/router.go" {3}
router := docs.NewRouter(
    docs.WithLocale(os.Getenv("DOCS_CONTENT_LOCALE")),
    docs.WithLocaleFallback("en"),
)
```

Con el fallback, una familia que no tiene página en el idioma pedido sirve la
del idioma por defecto. Nunca tapa una traducción real: si la página traducida
existe, gana esa.

## Saber qué falta

{{< steps >}}
1. `Router.LocaleCoverage()` devuelve cada idioma y las familias que le faltan.
2. `Router.UntranslatedFamilies("es")` responde para un solo idioma.
3. `fastr-docs check .` registra los huecos.
{{< /steps >}}

La cobertura es informativa. Un sitio traducido a medias es un estado normal, no
un build roto, así que nada de esto invalida la validación.

{{< tip title="Este sitio" >}}
Toda la documentación está en los dos idiomas. El blog no, y la razón está
anotada en la versión en inglés de esta página.
{{< /tip >}}
