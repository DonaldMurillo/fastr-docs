---
tags: [pantallas, componentes, gofastr]
locale: es
---

# Pantallas y componentes

Usa una pantalla tipada cuando una ruta necesita estado, acciones, carga de
datos o una composición de componentes de GoFastr más rica. La ruta sigue
participando en el mismo modelo de navegación, búsqueda, migas de pan, orden y
exportación que una página de Markdown.

## Registrar una pantalla

```go
router.MustScreen("/examples/playground", docs.ScreenConfig{
    Title:       "Playground",
    Description: "Prueba los componentes del framework.",
    Component:   &PlaygroundScreen{},
    Order:       1,
    Offline:     true,
})
```

El `Component` implementa el contrato de renderizado normal de GoFastr. Construye
la superficie con las primitivas de UI del framework y deja el CSS propio de la
ruta junto a la pantalla.

## Dónde está la frontera

Empieza con una página cuando la persona está leyendo o buscando prosa que debe
durar. Asciende la ruta a pantalla cuando necesita un control en vivo, una
consola de peticiones, una gráfica, un formulario o estado que viene del
servidor. Ni la URL ni su sitio en el árbol de rutas tienen que cambiar.

## Probar la superficie

El [playground tipado](/examples/playground) es una pantalla real dentro de este
sitio. Sus pestañas, su desplegable, su contador, su interruptor y su control
segmentado están cubiertos por pruebas de navegador. Usa el mismo enfoque para
las interacciones de tu producto: comprueba el resultado que ve la persona, no
solo el marcado renderizado.
