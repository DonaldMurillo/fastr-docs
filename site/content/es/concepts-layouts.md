---
tags: [layouts, navegacion, responsive]
locale: es
---

# Layouts y navegación

fastr-docs usa layouts anidados de GoFastr. El layout exterior gobierna la
envoltura de marca blanca; cada sección de primer nivel puede gobernar su propia
barra lateral contextual y su grupo de pantallas.

## Layout global y de sección

```go
router.Layout()
```

El layout global dibuja la marca, el control de tema, la paleta de comandos
nativa, la envoltura PWA y los enlaces de cabecera derivados de las rutas de
primer nivel. `Router.Mount` crea un layout de sección para cada rama de rutas
activa.

La barra lateral persistente se estrecha a la sección activa en las páginas de
documentación. En pantallas pequeñas, el cajón de GoFastr expone el mismo árbol
de rutas con su rama activa desplegada y la página actual marcada.

## Los grupos de cabecera no duplican la barra lateral

Los enlaces de cabecera representan las secciones de primer nivel. La barra
lateral representa el árbol local de la sección activa. Un enlace de cabecera
puede seguir activo en páginas anidadas porque la coincidencia usa el prefijo de
la ruta de sección.

## Navegación dentro de la página

El layout de documento deriva los encabezados del Markdown y dibuja un raíl
anclado en pantallas anchas. En anchos menores usa un desplegable. El scrollspy
va marcando el encabezado actual sin alterar el desplazamiento normal del
documento.

Mira el [ejemplo de árbol de rutas](/es/ejemplos/arbol-de-rutas) para ver cómo un
solo registro produce las tres superficies de navegación.
