---
title: Un árbol para docs y publicación
description: Por qué el archivo del blog y el feed RSS usan el mismo Router que la documentación.
date: 2026-08-28
authors: [fastr-docs]
tags: [arquitectura, rutas]
locale: es
translation_of: /blog/route-tree
---

# Un árbol para docs y publicación

El blog es una colección de Markdown que pertenece al mismo Router que
gobierna el árbol de documentación. Así la navegación, la búsqueda, los
metadatos, la exportación estática y el RSS viven en el mismo conjunto de
rutas.

## Una fuente, muchas superficies

El archivo, las páginas de taxonomía, el feed RSS y la plantilla de entrada
leen los mismos metadatos de ruta publicados. No hay una segunda lista que
mantener sincronizada.

## Una mejor superficie de lectura

Las entradas añaden marcas de artículo, acciones de compartir y copiar,
lecturas relacionadas y un índice adaptable, sin dejar de sentirse parte del
mismo sitio.
