---
tags: [ejemplos, arbol-de-rutas, navegacion]
locale: es
translation_of: /examples/route-tree
---

# Ejemplo de árbol de rutas

Este sitio usa un único Router para su manual, sus ejemplos y su referencia
OpenAPI. Las secciones de primer nivel se convierten en grupos de cabecera; cada
sección gobierna una rama de la barra lateral contextual.

```text
/
├── docs
│   ├── getting-started
│   ├── concepts
│   │   ├── router
│   │   ├── content
│   │   └── layouts
│   ├── build
│   │   ├── screens
│   │   ├── framework-ui
│   │   ├── openapi
│   │   └── plugins
│   └── operate
├── es
│   └── docs
│       └── ...
├── examples
│   ├── playground
│   ├── route-tree
│   └── custom-surface
└── api-reference
```

El árbol en español es una rama más del mismo Router, no un segundo sitio. Por
eso la búsqueda, la exportación y la validación lo cubren sin nada aparte.

El Router respeta el orden explícito entre rutas hermanas. La sección activa
estrecha el raíl de escritorio, mientras el cajón móvil expone la misma rama,
despliega el grupo activo y marca la página actual.

Abre [Primeros pasos](/es/docs/getting-started) y luego usa la cabecera para
moverte entre Documentación, Ejemplos y la referencia de la API.
