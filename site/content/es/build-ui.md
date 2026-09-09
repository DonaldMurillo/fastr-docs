---
tags: [ui, componentes, gofastr, referencia]
locale: es
---

# UI del framework

fastr-docs usa los componentes nativos de GoFastr para la envoltura, los bloques
de contenido, los formularios, las vistas de datos y los estados de interacción.
Una página y una pantalla tipada comparten la misma biblioteca, así que una
superficie propia no necesita un segundo sistema visual.

## La envoltura de documentación

El layout por defecto del Router aporta estas piezas nativas:

| Área | Componentes | Para qué |
| --- | --- | --- |
| Envoltura del sitio | `SiteHeader`, `SiteFooter`, `DocLayout` | Marca, navegación de sección, marco del artículo |
| Navegación | `Sidebar`, `AnchoredRail`, `TableOfContents`, `StepRail` | Árbol de rutas, encabezados internos, secuencias guiadas |
| Búsqueda | `CommandPalette`, `GlobalSearch`, `SearchInput` | Búsqueda de rutas con `Ctrl+K` o `⌘K` |
| Utilidades | `ThemeToggle`, `BackToTop`, `SkipLink`, `Tooltip`, `Menu` | Tema, acceso por teclado, controles de página |

El Router aporta los datos de las rutas. GoFastr gobierna el comportamiento del
cajón, el foco, los atajos de teclado, el bloqueo del desplazamiento y el ciclo
de vida de los componentes.

## Catálogo de componentes nativos

Los nombres son los constructores públicos de GoFastr y no se traducen.

### Envoltura y navegación

| Componentes | Para qué |
| --- | --- |
| `AnchoredRail`, `StepRail`, `TableOfContents` | Navegación fija interna, por pasos y por encabezados |
| `BackToTop`, `SkipLink` | Páginas largas y navegación por teclado |
| `CommandPalette`, `GlobalSearch`, `SearchInput`, `ShortcutHint` | Búsqueda, acceso a comandos y pistas de teclado |
| `DocLayout`, `Sidebar`, `SidebarBody` | Marco de la página y navegación adaptable |
| `Menu`, `Tooltip` | Menús compactos y ayuda contextual |
| `SiteHeader`, `SiteFooter`, `ThemeToggle` | Envoltura, identidad y modo claro/oscuro/automático |

### Lectura y contenido técnico

| Componentes | Para qué |
| --- | --- |
| `Markdown`, `Section`, `PageHeader`, `Hero`, `HeroSplit` | Prosa, secciones, entradillas y portadas |
| `CodeBlock`, `CodeTabs`, `CopyButton`, `TerminalBlock` | Código, variantes de lenguaje, portapapeles y transcripciones |
| `Callout`, `Banner`, `FactBox`, `StatusBadge`, `StatusPill`, `Muted` | Contexto, estado e información secundaria |
| `Card`, `DetailList`, `MetricBand`, `RecordSummary`, `StatCard` | Explicaciones estructuradas y datos compactos |
| `Collapsible`, `Tabs`, `DiffViewer`, `JSONViewer` | Detalle opcional, alternativas, cambios y salida estructurada |
| `Divider`, `EmptyState` | Separación y estados sin contenido |

### Layout y composición

| Componentes | Para qué |
| --- | --- |
| `Container`, `Stack`, `Cluster`, `Grid`, `Center`, `Box`, `Sticky`, `Responsive` | Composición adaptable sin trucos de flex por página |
| `PaneHost`, `Workbench`, `Toolbar` | Vistas de referencia densas, inspectores y acciones agrupadas |
| `AspectRatioComponent` (`AspectRatio`), `Spacer` | Geometría de medios estable y espacio intencionado |
| `Themed` | Tema acotado a un subárbol de componentes |

### Formularios y controles

| Componentes | Para qué |
| --- | --- |
| `AuthCard`, `SignOut` | Envolturas de autenticación y cierre de sesión |
| `Form`, `FormSection`, `FormField`, `ValidationSummary` | Estructura, agrupación y validación |
| `TextField`, `TextArea`, `PasswordInput`, `NumberField`, `DateField`, `TimePicker`, `ColorField` | Entradas nativas tipadas |
| `Select`, `Checkbox`, `CheckboxGroup`, `Radio`, `RadioGroup`, `Switch` | Selección y controles booleanos |
| `Slider`, `RangeSlider`, `RatingInput`, `SegmentedControl` | Valores acotados, valoraciones y opciones excluyentes |
| `InputGroup`, `TagInput`, `Tag`, `FileUpload`, `FileDropzone` | Campos compuestos, etiquetas y subida de archivos |
| `ConditionalField`, `FormRepeater`, `Repeater`, `StepWizard` | Flujos condicionales, repetidos y por pasos |

### Datos, acciones y respuesta

| Componentes | Para qué |
| --- | --- |
| `DataTable`, `FilterToolbar`, `FilterChipBar` | Datos ordenables y filtrado desde la URL |
| `Button`, `Link`, `LinkButton`, `ConfirmAction`, `ToggleAction`, `OptimisticAction` | Acciones seguras y navegación |
| `Notification`, `NotificationBell`, `NetworkRetryBanner`, `PollingIndicator` | Respuesta, estado sin leer, reintento y sondeo activo |
| `Spinner`, `SkeletonRow`, `SkeletonCard`, `ProgressSteps` | Estados de carga y de progreso |
| `Counter`, `AnimatedCounter` | Valores por señal y por desplazamiento |

### Visualización y medios

| Componentes | Para qué |
| --- | --- |
| `LineChart`, `BarChart`, `PieChart`, `Sparkline` | Gráficas SVG y tendencias compactas |
| `Gallery`, `Lightbox`, `Carousel`, `Avatar`, `AvatarGroup` | Imágenes, identidad y navegación visual |
| `OptimizedImage`, `PipelineImage` | Imágenes adaptables y variantes del framework |
| `PricingCard`, `Timeline` | Planes, hitos y registros cronológicos |

El inventario es una referencia a propósito, no un segundo sistema de
componentes. GoFastr gobierna el renderizado, la accesibilidad, el foco, el
ciclo de vida y los tokens de tema. fastr-docs aporta los datos de ruta y la
composición del contenido alrededor.

## Bloques de contenido en Markdown

Una página de Markdown puede usar el renderizador nativo y sus bloques de código
enmarcados. Un proyecto puede exponer cualquier componente del catálogo mediante
un adaptador de shortcode tipado:

```go
router.Use(docs.MarkdownComponentsPlugin{Components: map[string]docs.MarkdownComponent{
    "note": func(props map[string]string, body render.HTML) render.HTML {
        return ui.Callout(ui.CalloutConfig{
            Title: props["title"], Variant: ui.StatusInfo,
        }, body)
    },
}})
```

```md
{{< note title="Importante" >}}
Este cuerpo sigue siendo Markdown y admite enlaces y énfasis.
{{< /note >}}
```

## Pantallas interactivas

Las pantallas tipadas usan los mismos controles nativos:

- `Form`, `FormField`, `TextField`, `TextArea`, `Select`, `Checkbox`, `Radio` y `Switch`
- `FilterToolbar`, `FilterChipBar`, `SegmentedControl` y `SearchInput`
- `DataTable`, `JSONViewer`, `LineChart`, `BarChart`, `PieChart` y `Sparkline`
- `Notification`, `Banner`, `Spinner`, `Skeleton`, `ProgressSteps` y `PollingIndicator`

Registra la pantalla en el mismo Router:

```go
router.MustScreen("/examples/playground", docs.ScreenConfig{
    Title:       "Playground",
    Description: "Ejercita la biblioteca de componentes.",
    Component:   &PlaygroundScreen{},
    Order:       1,
})
```

La ruta se sigue pudiendo buscar, aparece en la navegación de su sección, recibe
migas de pan y enlaces anterior/siguiente, y entra en la exportación estática y
en el precaché de la PWA.

## Interfaces del framework

La biblioteca de UI es solo una capa. Un sitio completo depende además de las
interfaces del Router:

- Un árbol de rutas para páginas, pantallas, grupos, plugins, búsqueda y exportación
- Orden explícito entre hermanas y validación estricta
- Front matter para títulos, descripciones, borradores, etiquetas, versiones, idiomas, redirecciones y metadatos SEO
- Búsqueda JSON local en desarrollo y Pagefind para índices estáticos troceados
- Páginas de referencia OpenAPI con consola de peticiones opcional
- Tokens claros y oscuros, marca blanca, entrega PWA y contenido sin conexión

Consulta [El Router](/es/docs/concepts/router),
[Escribir contenido](/es/docs/concepts/content),
[Referencia OpenAPI](/es/docs/build/openapi) y [Pruebas](/es/docs/operate/testing)
para las interfaces que hay detrás de estos componentes.

## Cuando el conjunto nativo no basta

Deja el comportamiento propio del producto en una pantalla tipada o en un
plugin. Un plugin puede registrar rutas, añadir componentes de Markdown, validar
su configuración, aportar datos de búsqueda, declarar orígenes de navegador y
montar recursos a través del router de GoFastr. Así las capacidades nuevas
entran en la misma tubería de navegación, búsqueda, exportación y validación en
lugar de crear una aplicación paralela.
