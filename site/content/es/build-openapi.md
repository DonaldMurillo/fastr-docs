---
tags: [openapi, plugin, referencia-api]
locale: es
---

# Referencia OpenAPI

El plugin de OpenAPI convierte un contrato en una ruta del mismo Router. Dibuja
un índice de operaciones, tarjetas por endpoint, parámetros, esquemas y una
consola de peticiones opcional. La referencia es una ruta normal: participa en
la navegación superior, en la barra lateral contextual, en la búsqueda local, en
la exportación estática y en la envoltura PWA.

## Montar el plugin

```go
router.Use(openapi.Plugin{
    SpecPath:  "openapi.json",
    Path:      "/api-reference",
    Title:     "Referencia de la API",
    ServerURL: os.Getenv("API_SERVER_URL"),
    Order:     3,
})
```

El plugin acepta contratos en JSON y en YAML. Por defecto usa la primera URL de
`servers`. `ServerURL` la sustituye para staging o producción sin tocar el
contrato versionado ni el árbol de rutas.

## Usar la consola de peticiones

La consola sigue a la operación seleccionada y dibuja los valores que describe
el contrato:

- parámetros de ruta, de consulta, de cabecera y de cookie
- marcas de obligatorio y comprobaciones en el cliente
- cuerpos de petición con tipo de contenido y validación JSON
- salida de respuesta legible y errores de red o de CORS

Los parámetros de ruta se sustituyen en la URL, los valores de consulta se
codifican y el cuerpo se envía solo cuando la operación declara uno. Los
parámetros de cookie se muestran como referencia, pero no se pueden inyectar en
una petición de navegador entre orígenes.

## Peticiones explícitas

La consola se ejecuta en el navegador y necesita que el servidor de la API
permita CORS. El Router registra el origen resuelto para que el host de GoFastr
pueda incluirlo en su política estricta de `connect-src`. Si no hay URL de
servidor configurada, la referencia sigue siendo legible y la consola explica
por qué no se pueden hacer peticiones.

Este sitio incluye un contrato de ejemplo pequeño y un servidor simulado local
para poder ejercitar el plugin de principio a fin. Son fixtures, no una API HTTP
que ofrezca fastr-docs.

## Extender el plugin

El plugin aporta una ruta de pantalla normal. Un proyecto puede fijar su título,
descripción, orden, URL de servidor y distintivo igual que en cualquier otra
ruta. Para otro adaptador de API, implementa `docs.Plugin` y regístralo en el
mismo Router.
