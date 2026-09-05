---
tags: [plugins, extensiones, router]
locale: es
---

# Plugins y extensiones

Un plugin extiende un proyecto sin crear un segundo sistema de contenido. Puede
registrar páginas o pantallas, añadir texto de búsqueda, declarar orígenes
externos, validar reglas del proyecto y montar recursos de navegador.

## El contrato básico

```go
type Plugin interface {
    Name() string
    Apply(*Router) error
}
```

`Apply` debería registrar todo lo que posee la extensión. A partir de ahí, los
títulos, descripciones, orden, visibilidad, metadatos de búsqueda y distintivos
fluyen a los adaptadores normales.

El plugin de colección incorporado acepta un directorio en disco o un sistema de
archivos virtual. Elige una fuente por plugin:

```go
router.Use(docs.MarkdownCollectionPlugin{
    Path: "/docs",
    FS:   contentFS,
    Root: "content",
    Config: docs.CollectionConfig{Offline: true},
})
```

Usa `Dir` para archivos de desarrollo y `FS` más `Root` para un `embed.FS` u
otra fuente empaquetada. El plugin delega en las mismas APIs de colección, así
que los metadatos, el orden, los borradores, la búsqueda y la validación siguen
siendo coherentes.

## Añadir validación o recursos

Usa `ValidatingPlugin` cuando la extensión tenga comprobaciones que deban correr
con la validación estricta del Router. Usa `AssetPlugin` cuando necesite montar
archivos estáticos a través del router HTTP de GoFastr.

```go
type ValidatingPlugin interface {
    Plugin
    Validate(*Router) error
}
```

Llama a `router.MountPluginAssets(server.Router())` cuando ya exista el router
del host. Mantén explícitos los orígenes externos con
`router.AllowConnectOrigin` para que la CSP generada siga siendo estrecha.

## Los ejemplos incluidos

La [referencia OpenAPI](/es/docs/build/openapi) está implementada como plugin
dentro del proyecto. Los [diagramas](/es/docs/build/diagrams) y las
[matemáticas](/es/docs/build/math) también, cada uno con un modelo de
aislamiento distinto. Sus rutas se registran junto a las de Markdown y las
cubren las mismas superficies de búsqueda, navegación, exportación y pruebas de
navegador.
