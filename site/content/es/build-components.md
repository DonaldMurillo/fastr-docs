---
tags: [guia, escritura, componentes]
locale: es
---

# Componentes de Markdown

Todo proyecto de fastr-docs arranca con este vocabulario de shortcodes ya
registrado. Sin código Go, sin imports, sin configuración. Registra el mismo
nombre tú y sustituyes cualquiera de ellos.

## Avisos

Cinco nombres, un componente. Cada uno es un aviso con su tono fijado.

```md
{{< note title="Atención" >}}
`note` e `info` son el mismo aviso.
{{< /note >}}
```

Que se ve así:

{{< note title="Atención" >}}
`note` e `info` son el mismo aviso. Usa el que se lea mejor.
{{< /note >}}

{{< tip >}}
`tip` y `success` comparten tono. El cuerpo sigue siendo Markdown, así que los
[enlaces](/es/docs) y el **énfasis** funcionan.
{{< /tip >}}

{{< warning title="Revisa el orden" >}}
Las rutas hermanas necesitan valores de `Order` explícitos.
{{< /warning >}}

{{< danger >}}
`danger` se dibuja con `role="alert"`.
{{< /danger >}}

`callout` toma el tono como propiedad cuando prefieres no elegir un nombre:

```md
{{< callout variant="warning" title="Importante" >}}
Texto del cuerpo.
{{< /callout >}}
```

## Pestañas

```md
{{< tabs >}}
{{< tab label="Go" >}}
Las pestañas son elementos `<details>` nativos.
{{< /tab >}}
{{< tab label="Shell" >}}
Cada conjunto recibe su propio nombre de grupo.
{{< /tab >}}
{{< /tabs >}}
```

{{< tabs >}}
{{< tab label="Go" >}}
Las pestañas son elementos `<details>` nativos. No corre JavaScript para
cambiarlas.
{{< /tab >}}
{{< tab label="Shell" >}}
Cada conjunto recibe su propio nombre de grupo, así que dos conjuntos en una
página no se estorban.
{{< /tab >}}
{{< tab label="Notas" >}}
La primera pestaña se abre por defecto.
{{< /tab >}}
{{< /tabs >}}

## Tarjetas

```md
{{< cards >}}
{{< card title="Rutas" description="Un árbol gobierna navegación y búsqueda." href="/es/docs/concepts/router" >}}
{{< /card >}}
{{< /cards >}}
```

{{< cards >}}
{{< card title="Rutas" description="Un árbol gobierna navegación y búsqueda." href="/es/docs/concepts/router" >}}
{{< /card >}}
{{< card title="Contenido" description="Markdown con front matter." href="/es/docs/concepts/content" >}}
{{< /card >}}
{{< card title="Publicar" description="Comprobar, exportar, desplegar." href="/es/docs/operate/deploy" >}}
{{< /card >}}
{{< /cards >}}

## Pasos

Una lista numerada normal dentro del shortcode. Los números salen de la lista,
así que reordenar los pasos los renumera.

```md
{{< steps >}}
1. Ejecuta `fastr-docs init mis-docs` para crear el proyecto.
2. Edita `content/` y registra rutas en `docs/router.go`.
{{< /steps >}}
```

{{< steps >}}
1. Ejecuta `fastr-docs init mis-docs` para crear el proyecto.
2. Edita `content/` y registra rutas en `docs/router.go`.
3. Ejecuta `fastr-docs check .` antes de confirmar.
{{< /steps >}}

## Árbol de archivos

Escribe el árbol como un bloque con valla e indéntalo. `filetree` lee su cuerpo
sin renderizar, así que la indentación sobrevive; el anidamiento sale de ella y
una `/` final marca un directorio.

````md
{{< filetree >}}
```
mis-docs/
  content/
    index.md
  main.go
```
{{< /filetree >}}
````

{{< filetree >}}
```
mis-docs/
  content/
    index.md
    getting-started.md
  docs/
    router.go
  main.go
  openapi.json
```
{{< /filetree >}}

## Piezas pequeñas

Secciones plegables, distintivos, etiquetas e iconos son de una línea:

```md
{{< details summary="¿Qué comprueba la validación estricta?" >}}
Enlaces internos rotos y anclas de encabezado que faltan.
{{< /details >}}

{{< badge label="Nuevo" variant="success" />}}
{{< tag label="rutas" href="/es/docs/concepts/router" />}}
{{< icon name="check" label="Soportado" />}}
```

{{< details summary="¿Qué comprueba la validación estricta?" >}}
Enlaces internos rotos, anclas de encabezado que faltan, orden duplicado entre
hermanas y páginas sin título o sin descripción.
{{< /details >}}

Distintivos y etiquetas: {{< badge label="Nuevo" variant="success" />}} {{< badge label="Beta" variant="warning" />}} {{< tag label="rutas" href="/es/docs/concepts/router" />}}

Los iconos vienen del registro de GoFastr: {{< icon name="check" label="Soportado" />}} {{< icon name="info" />}}

## Diferencias

El cuerpo es un parche, leído sin renderizar para que su estructura de líneas
sobreviva.

````md
{{< diff left="Antes" right="Después" >}}
```
 router := docs.NewRouter(
-    docs.WithSiteName("Docs"),
+    docs.WithSiteName("Acme Docs"),
 )
```
{{< /diff >}}
````

{{< diff left="Antes" right="Después" >}}
```
 router := docs.NewRouter(
-    docs.WithSiteName("Docs"),
+    docs.WithSiteName("Acme Docs"),
+    docs.WithTemplate(docs.TemplateBlueprint),
 )
```
{{< /diff >}}

## Opciones de los bloques de código

Las vallas admiten opciones después del lenguaje. Una valla sin opciones se
dibuja exactamente igual que siempre.

{{< filetree >}}
```
title="ruta/al/archivo"   una cabecera con el nombre
{1,4-6}                   resalta esas líneas
showLineNumbers           numera el margen
scroll                    limita la altura y desplaza dentro
```
{{< /filetree >}}

Abrir una valla con `go title="docs/router.go" {3,4} showLineNumbers` da:

```go title="docs/router.go" {3,4} showLineNumbers
router := docs.NewRouter(
    docs.WithSiteName("Acme Docs"),
    docs.WithTemplate(docs.TemplateBlueprint),
    docs.WithSearchIndexPath("/assets/search.json"),
)
```

## Tres formas

Un shortcode se registra como una de tres formas, y un nombre resuelve a
exactamente una.

| Forma | Recibe | La usan |
| --- | --- | --- |
| `MarkdownComponent` | su cuerpo como Markdown renderizado | casi todos |
| `MarkdownContainer` | sus shortcodes anidados como hijos separados | `tabs`, `cards`, `hero` |
| `MarkdownRawComponent` | su cuerpo sin renderizar | `diff`, `filetree`, `mermaid`, `math` |

El contenedor existe porque un componente recibe un solo bloque fundido y no
puede saber dónde acaba un hijo y empieza el siguiente, que es justo lo que
necesita saber un conjunto de pestañas.

La forma cruda existe porque renderizar un cuerpo como Markdown destruye la
estructura de líneas. Un parche se colapsa en una sola línea y la envoltura del
bloque de código acaba dentro del texto.

## Sustituir uno por defecto

Registrar un nombre reemplaza a quien lo tuviera:

```go
router.RegisterMarkdownComponent("note", func(props map[string]string, body render.HTML) render.HTML {
    return ui.Callout(ui.CalloutConfig{Title: props["title"], Variant: ui.StatusNeutral}, body)
})
```

Pasa `docs.WithoutDefaultComponents()` para empezar con un vocabulario vacío,
donde cualquier shortcode no registrado rompe el build en lugar de renderizarse.
