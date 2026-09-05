---
tags: [diagramas, mermaid, plugins]
locale: es
---

# Diagramas

Los diagramas de Mermaid se renderizan desde Markdown sin relajar la política de
contenido que protege al resto del sitio.

```go title="docs/router.go"
router.Use(mermaid.Plugin{})
```

Esa es toda la configuración. El plugin aporta sus propios recursos, así que
nada de `main.go` cambia.

## Escribir uno

{{< mermaid title="Cómo se renderiza una página" >}}
```
graph LR
    MD[Archivo Markdown] --> R[docs.Router]
    R --> NAV[Navegación]
    R --> IDX[Índice de búsqueda]
    R --> PAGE[Página renderizada]
    PAGE --> EXP[Exportación estática]
```
{{< /mermaid >}}

La valla dentro del shortcode es opcional. Mantiene la fuente legible en un
editor de Markdown normal, y el plugin la quita.

## Por qué corre dentro de un marco

Mermaid renderiza inyectando un elemento `<style>` y emitiendo SVG con estilos
en línea. Las páginas de documentación se sirven bajo `default-src 'self'`, que
bloquea las dos cosas, así que ejecutar Mermaid en la página significaría
debilitar esa política en todas partes.

En su lugar, el diagrama se renderiza dentro de un documento cargado en un
`<iframe sandbox="allow-scripts">` sin `allow-same-origin`. Ese marco tiene un
origen opaco: no alcanza el DOM, las cookies ni el almacenamiento de la página,
y `postMessage` es el único canal entre ambos. Solo ese documento lleva la
política de estilos relajada; las páginas siguen estrictas.

## Sin JavaScript

Quien no ejecute JavaScript ve el código fuente del diagrama como texto
preformateado, que es más útil que una caja vacía.

## Temas

El marco no puede leer la hoja de estilos de la página, así que el modo oscuro
se refleja pidiendo un nuevo renderizado, no con CSS. Cambia el tema en la
cabecera y el diagrama se redibuja.

## Qué descarga una página

Nada, hasta que un diagrama se acerca a la ventana visible. El adaptador crea
cada marco con un `IntersectionObserver`; `loading="lazy"` por sí solo no basta,
porque el umbral del navegador es lo bastante generoso como para cargar todos
los diagramas de una página de inmediato.

El paquete del marco está dividido por tipo de diagrama, así que un diagrama de
flujo descarga el código de diagramas de flujo y no todo Mermaid. Esta página
transfiere alrededor de 1,7 MB entre sus dos diagramas, frente a 6,9 MB antes de
dividirlo.

Dos diagramas cuestan aproximadamente el doble que uno, y eso no es un fallo que
podamos arreglar. Cada marco es un origen opaco distinto, que es lo que lo
mantiene aislado, y eso mismo le da su propia partición de caché HTTP: los
marcos no pueden compartir una descarga, ni una recarga puede reaprovecharla.
Aislamiento y caché son el mismo compromiso visto por sus dos caras.
