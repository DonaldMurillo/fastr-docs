---
tags: [contenido, markdown, front-matter]
locale: es
---

# Escribir contenido

Usa Markdown para la documentación que deba durar, poder buscarse, exportarse y
revisarse con calma. fastr-docs quita el front matter antes de renderizar, así
que el contenido de la página, el texto de búsqueda y los encabezados no se
desincronizan.

## Los metadatos van en el front matter

```md
---
title: Configurar el despliegue
description: Define la URL pública y la base estática.
tags: [despliegue, hosting]
order: 4
locale: es
---

# Configurar el despliegue
```

Se admiten títulos, slugs, descripciones, borradores, reglas de no-index, URLs
de edición y canónicas, autores, fechas, idioma, versión, alternativas,
etiquetas, redirecciones y orden.

El campo que importa aquí es `locale`. Sin él, una página traducida no se
empareja con su original y el selector de idioma no tiene a dónde ir.

## Registrar un archivo

```go title="docs/router.go"
router.MustPage("/es/docs/concepts/content", docs.PageConfig{
    Title:      "Escribir contenido",
    SourcePath: "content/es/concepts-content.md",
    Order:      3,
    Offline:    true,
})
```

Usa `PageFile` cuando la fuente se elige en tiempo de ejecución, o
`MarkdownCollection` cuando un directorio entero debe convertirse en una familia
de rutas. Las rutas de una colección heredan el front matter y pueden aplicar
prefijos de idioma o de versión, que es la forma más corta de publicar una
traducción completa.

## Un slug cuando el nombre del archivo no sirve

```md
---
title: El Router
slug: concepts/router
---
```

Ese archivo se publica en `/docs/concepts/router` cuando el prefijo de la
colección es `/docs`. Para las páginas registradas a mano, la ruta que pasas al
`Router` sigue siendo la fuente de la verdad.

## Contenido empaquetado

```go
import "embed"

//go:embed content
var contentFS embed.FS

router.MarkdownCollectionFS(contentFS, "content", docs.CollectionConfig{Prefix: "/docs"})
```

Mismo comportamiento de rutas y metadatos cuando el contenido viene de un
`embed.FS`, de un sistema de archivos de prueba o de cualquier otra fuente
virtual. Es lo que permite que un binario lleve la documentación dentro.

{{< note title="Traducir no cambia nada de esto" >}}
Una página en español es una página normal con `locale: es`. No hay un formato
aparte para las traducciones, ni un directorio con significado especial: el
prefijo `/es` de este sitio es una convención de rutas, no una regla del
framework.
{{< /note >}}
