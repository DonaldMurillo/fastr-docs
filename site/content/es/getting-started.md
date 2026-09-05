---
tags: [primeros-pasos, instalacion]
locale: es
---

# Primeros pasos

Esta página es una traducción real dentro del sitio de fastr-docs, no un
ejemplo inventado. Sirve para probar que el selector de idioma, la navegación y
la búsqueda funcionan con contenido en varios idiomas.

## Crear un proyecto

{{< steps >}}
1. Ejecuta `go run ./cmd/fastr-docs init mis-docs --name "Mis Docs"`.
2. Entra en el directorio y ejecuta `fastr-docs check .`.
3. Levanta el servidor de desarrollo con `fastr-docs dev .`.
{{< /steps >}}

El proyecto generado ya trae páginas de ejemplo, un blog, una referencia
OpenAPI y las habilidades para agentes.

## Qué se traduce y qué no

Hay dos cosas distintas, y conviene no confundirlas.

{{< tabs >}}
{{< tab label="El contenido" >}}
Lo escribes tú. Cada página lleva `locale:` en su front matter, y las páginas
que comparten familia se enlazan entre sí. Esta página es
`content/es/getting-started.md` y su pareja es `content/getting-started.md`.
{{< /tab >}}
{{< tab label="La interfaz" >}}
Son las etiquetas del marco: «Buscar», «En esta página», «Actualizado», los
títulos del blog. Se traducen con `WithUIStrings`, y los campos vacíos
conservan el inglés, así que puedes traducir de una en una.
{{< /tab >}}
{{< /tabs >}}

## Páginas sin traducir

Cuando compilas el sitio para un idioma concreto, una página sin traducir
desaparecería del resultado. `WithLocaleFallback` evita eso:

```go title="docs/router.go" {3}
router := docs.NewRouter(
    docs.WithLocale(os.Getenv("DOCS_CONTENT_LOCALE")),
    docs.WithLocaleFallback("en"),
)
```

Con esa opción, una familia sin versión en el idioma pedido usa la del idioma
por defecto. Nunca tapa una traducción existente: si la página traducida
existe, gana ella.

{{< warning title="Revisa lo que falta" >}}
`Router.LocaleCoverage()` dice qué páginas siguen sin traducir, y
`fastr-docs check` las registra. Es informativo: un sitio traducido a medias es
un estado normal, no un error de compilación.
{{< /warning >}}

## Seguir leyendo

La versión completa de la documentación está en inglés. Empieza por
[el Router](/docs/concepts/router) o por [la guía en inglés](/docs/getting-started).
