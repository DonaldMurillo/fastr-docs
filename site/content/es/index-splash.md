---
template: splash
locale: es
hero:
  eyebrow: Documentación reutilizable para GoFastr
  title: Documentación construida sobre un árbol de rutas.
  tagline: Las páginas, las pantallas, la búsqueda, OpenAPI y el modo sin conexión salen todos del mismo Router explícito.
  actions:
    - text: Leer la documentación
      link: /docs
      variant: primary
    - text: Empezar
      link: /es/docs/getting-started
      variant: secondary
---

## Un router, todas las superficies

Elige Markdown para contenido duradero o una pantalla tipada para lo
interactivo. La navegación, la búsqueda, las migas de pan y la exportación no
se desincronizan porque leen el mismo árbol.

{{< cards >}}
{{< card title="Páginas" description="Markdown con front matter, encabezados, búsqueda local e índice de contenidos." href="/docs/concepts/content" >}}
{{< /card >}}
{{< card title="Pantallas" description="Componentes tipados de GoFastr para playgrounds y flujos de producto." href="/docs/build/screens" >}}
{{< /card >}}
{{< card title="Complementos" description="Extiende el mismo router sin bifurcar la interfaz." href="/docs/build/plugins" >}}
{{< /card >}}
{{< /cards >}}

## Esta página es la demostración

Solo tres páginas de este sitio están traducidas al español. El resto sigue en
inglés a propósito: así se ve cómo se comporta un sitio traducido a medias, que
es el estado normal de cualquier proyecto real.

{{< note title="Cambia de idioma" >}}
El selector de idioma de la cabecera aparece solo cuando una página tiene
variantes. Te lleva a la misma página en el otro idioma, no al inicio.
{{< /note >}}

## Empieza aquí

{{< steps >}}
1. `go run ./cmd/fastr-docs init mis-docs` crea un proyecto estricto.
2. Edita `content/` y registra las rutas en `docs/router.go`.
3. `fastr-docs check .` valida enlaces, anclas, orden y traducciones pendientes.
{{< /steps >}}
