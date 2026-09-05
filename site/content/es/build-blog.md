---
tags: [contenido, publicacion, rss]
locale: es
---

# Blog y RSS

Usa `MarkdownBlog` cuando un sitio de documentación también publique novedades.
Registra una ruta de archivo en el prefijo y convierte cada archivo Markdown del
directorio en una entrada. Las entradas usan el mismo contrato de front matter,
así que el mismo Router se ocupa de títulos, descripciones, autores, fechas,
etiquetas, borradores, idiomas, versiones, búsqueda, navegación y SEO.

```go
if err := router.MarkdownBlog("/blog", "content/blog", docs.BlogConfig{
    Title:          "Novedades",
    Description:    "Notas de versión y novedades del producto.",
    Order:          5,
    Offline:        true,
    PostOrderStart: 10,
}); err != nil {
    return err
}
```

`content/blog/index.md` es el contenido del archivo. Un archivo como
`content/blog/first-post.md` se convierte en `/blog/first-post`. Pon `slug` en
su front matter cuando el nombre del archivo no deba decidir la URL:

```yaml
---
title: Una nota de versión más larga
description: Qué cambió en esta versión.
date: 2026-08-30
authors: [Equipo del proyecto]
tags: [version]
slug: versiones/agosto
---
```

`BlogPosts("/blog")` devuelve las entradas publicadas, de la más reciente a la
más antigua. Incluye las marcadas `noindex` para archivos a medida, pero deja
fuera los borradores salvo que el Router esté en modo vista previa.

## Una superficie de publicación aparte

El blog es hermano de la documentación en la envoltura global. No reutiliza el
árbol de contenidos de la sección de documentación ni su navegación interna. El
layout generado ofrece:

- `/blog` — la portada de últimas entradas, con destacada y paginación
- `/blog/archive` y `/blog/archive/:año` — vistas cronológicas
- `/blog/tags` y `/blog/tags/:etiqueta` — índices por tema y vistas filtradas
- `/blog/authors` y `/blog/authors/:autor` — índices por autor y vistas filtradas
- `/blog/search?q=...` — búsqueda de la publicación renderizada en el servidor
- `/blog/feed.xml` — el punto de entrada RSS que monta el host

La búsqueda, la taxonomía, las tarjetas, las entradas relacionadas, el tiempo de
lectura, las migas de pan, los metadatos del artículo y el cajón móvil salen
todos de la misma colección del Router. Las páginas de entrada reciben una
plantilla orientada a la lectura con índice interno opcional; las de archivo no
heredan ese índice.

Las páginas de entrada son artículos de verdad, no contenedores con estilo. El
marcado generado expone la relación del título, las fechas de publicación y
actualización, una región de migas de pan y de índice etiquetada, acciones de
tamaño adecuado para el teclado y avisos en vivo para las tecnologías de apoyo.
La acción de compartir usa la Web Share API cuando el dispositivo la ofrece y,
si no, copia la URL canónica.

{{< warning title="La interfaz del blog aún no se traduce" >}}
Las etiquetas del blog, como archivo, etiquetas o tiempo de lectura, se resuelven
una sola vez para todo el sitio, no por página. Una entrada en español mostraría
la interfaz en inglés. Está anotado como limitación conocida en
[Idiomas y traducción](/es/docs/operate/i18n).
{{< /warning >}}
