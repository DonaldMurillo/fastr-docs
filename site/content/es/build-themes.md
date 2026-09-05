---
tags: [temas, plantillas, marca]
locale: es
---

# Temas y plantillas

fastr-docs separa el sistema de documentación de su identidad visual. Todas las
plantillas usan el mismo `docs.Router`, el mismo árbol de rutas, el mismo índice
de búsqueda, la misma exportación PWA, el mismo plugin de OpenAPI, el mismo blog
y el mismo comportamiento de accesibilidad. Elige un punto de partida al
construir y luego ajusta los tokens semánticos de GoFastr a tu producto.

## Cinco puntos de partida

- **Editorial** — cálida, compacta y centrada en un ritmo de lectura claro. Es la de por defecto.
- **Terminal** — monoespaciada, de alto contraste, inspirada en la línea de comandos.
- **Blueprint** — de tonos fríos, técnica y ordenada sobre una rejilla.
- **Studio** — expresiva, de esquinas suaves, apropiada para documentación de producto.
- **Notebook** — de tono papel, pensada para leer, con tipografía con serifa.

Cada plantilla incluye valores claros y oscuros. La preferencia de esquema de
color de la persona y el conmutador de tema nativo de GoFastr siguen
funcionando después de elegir una.

## Elegir una plantilla

Pásasela al mismo router que gobierna tus páginas:

```go
router := docs.NewRouter(
    docs.WithTemplate(docs.TemplateBlueprint),
)
```

El proyecto generado también lee `DOCS_TEMPLATE`, así que puedes previsualizar
otro punto de partida sin tocar el árbol de rutas:

```text
DOCS_TEMPLATE=notebook go run .
```

Los valores desconocidos caen a `editorial`.

## Personalizar los tokens semánticos

Usa `WithTheme` cuando la plantilla se acerque pero tu marca necesite otros
colores, tipografías o formas:

```go
router := docs.NewRouter(
    docs.WithTheme(docs.ThemeConfig{
        Template: docs.TemplateStudio,
        Overrides: docs.ThemeOverrides{
            Primary:     "#8B5CF6",
            Accent:      "#F97316",
            FontHeading: "Manrope, Inter, sans-serif",
            RadiusMd:    12,
            DarkColors: map[string]string{
                "primary": "#C4B5FD",
                "accent":  "#FDBA74",
            },
        },
    }),
)
```

`ThemeOverrides` es el conjunto de tokens semánticos de GoFastr. Cubre las
superficies de la página, el texto, los bordes, los colores de estado, los
bloques de código, las familias tipográficas y los radios compartidos. Los
campos vacíos conservan el valor de la plantilla elegida. Usa `DarkColors` para
valores de modo oscuro con contraste seguro en lugar de dar por hecho que un
color claro sirve igual.

## Conectar el tema con el host

Dale a la app de GoFastr el mismo tema del router. Así los componentes del
framework y la envoltura de fastr-docs comparten un solo sistema de tokens:

```go
site := uiapp.NewApp("Acme Docs").
    WithTheme(router.Theme()).
    WithLang(router.Language())
```

El proyecto que genera `fastr-docs init` ya lo hace. Un host propio puede añadir
CSS de confianza mediante `ThemeConfig.CustomCSS`; úsalo para componentes del
proyecto o un ajuste pequeño de layout, y deja la envoltura principal sobre
tokens semánticos.

## El modelo de rutas no cambia

Las plantillas no son árboles de contenido alternativos. Un enlace a
`/docs/getting-started` es el mismo enlace en todas, y los plugins como el de
OpenAPI se siguen montando en el mismo router. Cambiar el punto de partida
visual no invalida marcadores, resultados de búsqueda, RSS ni contenido sin
conexión.
