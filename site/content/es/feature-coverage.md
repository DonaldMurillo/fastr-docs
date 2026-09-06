---
tags: [referencia, funciones, arquitectura]
locale: es
---

# Qué cubre fastr-docs

fastr-docs cubre las tareas centrales que se esperan de un generador de
documentación: escribir contenido, organizarlo en secciones navegables,
buscarlo, publicarlo y mantener el mismo modelo de rutas disponible para
personas y para agentes.

Esta página explica qué está soportado, dónde la implementación se aparta a
propósito de las herramientas centradas en JavaScript, y qué decisiones
pertenecen al proyecto que usa fastr-docs.

## Soportado en el Router

| Capacidad | Cómo lo hace fastr-docs |
| --- | --- |
| Jerarquía de contenido | Un `docs.Router` con grupos anidados, páginas, pantallas tipadas y rutas de plugin. |
| Escritura en Markdown | Markdown en disco, `fs.FS` embebido, front matter, slugs propios, enlaces y un vocabulario de shortcodes registrado por defecto. |
| Contenido interactivo | Pantallas y componentes nativos de GoFastr en el mismo árbol que el Markdown. |
| Composición de layouts | Layouts globales y de sección, incluidos layouts anidados dentro de otros. |
| Navegación | `Order` explícito entre hermanas, estado de ancestro activo, migas de pan, enlaces anterior/siguiente, cajones adaptables y navegación por encabezados. |
| Búsqueda | Índice JSON en desarrollo, exportación con Pagefind y la paleta de comandos nativa con `Ctrl+K` o `⌘K`. |
| Localización | Metadatos de idioma, emparejado por ruta o por `translation_of`, alternativas `hreflang` derivadas, filtros, interfaz totalmente traducible por idioma, formatos de fecha configurables, fallback a un idioma por defecto e informe de cobertura. |
| Versiones | Metadatos de versión, filtros, selectores conscientes de la ruta, `MarkdownVersionedCollection` e inventario en el manifiesto. |
| SEO y publicación | Títulos, descripciones, URLs canónicas, `noindex`, autores, fechas, imágenes sociales, redirecciones, sitemap, robots, blogs en Markdown y feeds RSS. |
| Entrega sin conexión | Integración PWA de GoFastr y exportación estática de las rutas y recursos elegibles. |
| Referencias de API | Plugin OpenAPI incluido, con contratos JSON o YAML y URL de servidor sustituible. |
| Extensibilidad | Plugins, tres formas de shortcode, factorías de layout propias, validación, aportes a la búsqueda y una tubería de recursos común. |
| Plantillas de página | Una envoltura `splash` desde el front matter para portadas, además de pantallas tipadas para lo que no exprese. |
| Bloques de código | Opciones de valla para título, numeración, resaltado de líneas y desplazamiento interno. |
| Metadatos del repositorio | `last_updated` y `edit_url` derivados del historial de git, con el front matter por encima. |
| Flujos con agentes | `agents/claude.md` generado, guía de escritura, `/llms.txt`, tarjeta de agente y descubrimiento MCP opcional. |
| Puertas de calidad | Validación estricta de rutas y contenido, comprobaciones de CLI y de exportación, y pruebas de navegador en escritorio y móvil. |

## Las APIs que moldean un proyecto

Los valores por defecto sirven para un sitio generado, pero la API pública está
pensada para que la extienda el proyecto dueño del contenido.

### Traducir la interfaz

`WithLocaleUIStrings` traduce las etiquetas de la envoltura por idioma, sin
tocar los títulos de ruta ni los metadatos de contenido:

```go
router := docs.NewRouter(
    docs.WithLocaleUIStrings("es", docs.UIStrings{
        Contents: "Contenido",
        Search:   "Buscar",
        Language: "Idioma",
    }),
)
```

### Conservar las capacidades de ruta de GoFastr

`PageConfig.Preload` y `ScreenConfig.Preload` aceptan los modos `hover`,
`visible` y `eager` de GoFastr. Las pantallas tipadas conservan su loader, sus
rutas estáticas, sus acciones, su ID de componente y su HTML de cabecera cuando
el Router añade los metadatos de documentación.

### Personalizar la envoltura

`WithLayouts` deja que un proyecto aporte factorías de layout global y de
sección mientras el Router sigue gobernando la selección de rutas:

```go
router := docs.NewRouter(docs.WithLayouts(docs.LayoutConfig{
    Global:  buildGlobalLayout,
    Section: buildSectionLayout,
}))
```

### Hacer útiles las rutas que faltan

Instala `NotFoundScreen` en el host de GoFastr, o sustitúyelo por una pantalla
con la marca del proyecto que siga el mismo contrato. El host en vivo devuelve
un 404 con la ruta pedida y un enlace al inicio. Las exportaciones estáticas
pueden llamar a `docs.WriteStaticNotFound` después de `ExportStatic`.

### Publicar versiones congeladas

Usa `MarkdownVersionedCollection` cuando cada release tenga su directorio de
contenido. La versión actual conserva la familia de rutas normal y las
archivadas reciben un segmento de versión.

## En qué se diferencia de Docusaurus y Starlight

Docusaurus y Starlight son buenas referencias de arquitectura de la información,
escritura en Markdown, búsqueda, localización, versiones y puntos de extensión.
fastr-docs conserva esas expectativas donde encajan con el runtime de GoFastr,
pero no copia cada función del ecosistema.

| Área | fastr-docs hoy | La decisión |
| --- | --- | --- |
| Barras laterales | Grupos del Router y orden explícito | El árbol de rutas es la fuente de la verdad, en vez de mantener un segundo archivo de barra lateral. |
| Componentes en Markdown | Shortcodes y pantallas tipadas | Componentes nativos de Go. No hacen falta React, MDX ni Markdoc. |
| Búsqueda | JSON local, Pagefind y paleta nativa | La búsqueda se despliega sin depender de un servicio alojado. |
| i18n | Metadatos, filtros, alternativas, selectores, interfaz traducida por idioma y fallback para lo no traducido | El proyecto aporta el contenido traducido; el framework no inventa traducciones ni traduce a máquina. |
| Versiones | Metadatos, filtros, selectores, colecciones versionadas e inventario | Congelar una versión sigue siendo un flujo de contenido o de control de versiones. |
| Plugins | Los plugins de Go aportan rutas, componentes, validación, recursos y datos de búsqueda | Las extensiones viven dentro del mismo ciclo del Router. |
| Blog y RSS | `MarkdownBlog`, `BlogPosts`, `RSSXML`, feed en vivo, salida estática y una colección por idioma | Las entradas viven en el Router para que metadatos, navegación, búsqueda, filtros y exportación no se desincronicen. |
| Analítica e integraciones | Responsabilidad del host o de un plugin | El núcleo de marca blanca se mantiene libre de cuentas de proveedor y de supuestos de rastreo. |

## Qué demuestra el proyecto generado

El proyecto inicial contiene páginas de Markdown, grupos anidados, un playground
tipado, ejemplos de UI del framework, una referencia OpenAPI, búsqueda local,
exportación PWA, referencias para agentes y pruebas de navegador. El fixture sin
CLI ejercita a mano la misma API pública del Router.

Sigue por [El Router](/es/docs/concepts/router),
[Escribir contenido](/es/docs/concepts/content),
[UI del framework](/es/docs/build/framework-ui) o
[Pruebas](/es/docs/operate/testing).
