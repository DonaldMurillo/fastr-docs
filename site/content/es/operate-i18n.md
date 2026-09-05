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
inventárselo.
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
De unas treinta páginas en inglés hay cinco en español. Es a propósito: enseña
el selector, y enseña cómo se ve un sitio traducido a medias, que es como está
todo sitio real.
{{< /tip >}}
