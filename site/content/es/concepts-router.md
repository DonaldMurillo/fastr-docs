---
tags: [router, conceptos]
locale: es
---

# El Router

`docs.Router` es el objeto central. Conoce qué páginas existen, en qué orden,
cómo se buscan y con qué estrategia se renderiza cada superficie.

## Un solo árbol

Una página puede ser Markdown, una pantalla tipada o la contribución de un
complemento sin crear un segundo sistema de navegación. De ese árbol salen:

{{< cards >}}
{{< card title="Navegación" description="Cabecera, barra lateral, migas de pan y enlaces anterior/siguiente." >}}
{{< /card >}}
{{< card title="Búsqueda" description="Un índice local en JSON, o Pagefind en la exportación estática." >}}
{{< /card >}}
{{< card title="Exportación" description="Un sitio estático instalable con las páginas marcadas como sin conexión." >}}
{{< /card >}}
{{< /cards >}}

## Orden explícito

Cada ruta necesita un `Order` positivo. No se deduce del nombre del archivo ni
del orden de registro, y dos hermanas con el mismo valor fallan la validación.
Es deliberado: el orden de una documentación es una decisión editorial.

## Validación estricta

La validación estricta viene activada. Comprueba enlaces internos rotos,
anclas de encabezado inexistentes, órdenes duplicados y páginas sin título o
sin descripción, y falla durante el arranque en vez de en producción.

{{< tip >}}
Ejecuta `fastr-docs check .` antes de cada commit. Valida las rutas, el
contenido y también que las habilidades de los agentes no se hayan
desincronizado.
{{< /tip >}}
