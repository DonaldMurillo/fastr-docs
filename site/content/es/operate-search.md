---
tags: [busqueda, pagefind, paleta-de-comandos]
locale: es
---

# Búsqueda

La búsqueda se deriva del Router, así que una ruta se puede buscar en cuanto
queda registrada y publicada. Cada registro incluye su ruta, título,
descripción, tipo, etiquetas, encabezados y texto buscable.

## Usar la paleta de comandos nativa

`Router.MountCommandPalette` monta la paleta de comandos de GoFastr. El botón de
la cabecera y el atajo `Ctrl+K` / `⌘K` abren la misma superficie. Los resultados
navegan con URLs de ruta normales y conservan la envoltura de la aplicación.

```go
if err := router.MountCommandPalette(server.Router()); err != nil {
    return err
}
```

## Elegir un backend

El backend JSON por defecto no tiene dependencias. La paleta descarga el índice
`search.json` generado en su primera consulta, así que el cuerpo de las páginas
se puede buscar tanto en desarrollo como en las exportaciones estáticas. Los
builds estáticos pueden optar por Pagefind y su índice troceado:

```go
router := docs.NewRouter(docs.WithPagefind())
```

```sh
fastr-docs export . --out dist --pagefind
```

Usa `WithPagefindPath` cuando un proxy inverso sirva el paquete de Pagefind bajo
un prefijo de recursos distinto. La paleta se comporta igual de cara a quien la
usa.

## Reglas de la búsqueda

El contenido oculto, en borrador, de un idioma inactivo, de una versión inactiva
o marcado como no indexable queda fuera de la búsqueda pública. Un
`SearchProvider` propio puede sustituir el artefacto JSON sin abandonar el
modelo portable de `SearchIndex`.

{{< note title="La búsqueda ya cubre los dos idiomas" >}}
Las páginas en español son rutas normales de este mismo Router, así que entran
en el índice sin configuración aparte. Busca "traducción" y verás salir esta
sección.
{{< /note >}}
