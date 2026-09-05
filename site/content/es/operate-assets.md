---
tags: [recursos, plugins, exportacion, operate]
locale: es
---

# Recursos de ejecución y de plugin

Una sola llamada recoge todo lo que se sirve bajo el prefijo de recursos: el
runtime de navegador, el índice de búsqueda, el manifiesto de exportación y lo
que aporte cada plugin.

```go title="main.go"
router.MountRuntimeAssets(server.Router(), router.AssetPrefix(), exportBase)
router.WriteRuntimeAssets(dist, router.AssetPrefix(), exportBase)
```

`RuntimeAssetNames` devuelve esa misma lista, así que la lista de precaché de la
PWA y las etiquetas `<script>` del host se derivan de ella en lugar de ser una
segunda copia que se desincroniza.

## Aportar desde un plugin

Un plugin se suma implementando `RuntimeAssetPlugin`:

```go title="plugin/openapi/openapi.go" {3}
func (Plugin) RuntimeAssets() (map[string][]byte, error) {
    body, err := fs.ReadFile(runtimeFS, "openapi.js")
    return map[string][]byte{"openapi.js": body}, err
}
```

Ese es todo el contrato. Nada del `main.go` generado cambia cuando añades un
plugin que trae código de navegador.

{{< warning title="Los nombres se validan" >}}
Los nombres de archivo vienen del plugin, no del proyecto, así que se comprueban.
Un plugin no puede sobrescribir `docs.js` ni escapar del directorio de recursos
con `..`.
{{< /warning >}}

## Dónde acaban los archivos

{{< filetree >}}
```
dist/
  __fastr-docs/
    docs.js
    openapi.js
    search.json
    manifest.json
```
{{< /filetree >}}

`AssetPrefix` sigue a `WithSearchIndexPath`, así que el prefijo y la URL del
índice no pueden discrepar. Cambia uno y el otro le sigue.

## Fechas y enlaces de edición

Ninguno hace falta escribirlo en la página. `WithGitMetadata` lee un solo
`git log` para todo el repositorio y rellena `last_updated` y `edit_url` a
partir del historial:

```go title="docs/router.go"
docs.WithGitMetadata(docs.GitMetadataConfig{
    RepoURL: os.Getenv("DOCS_REPO_URL"),
    Branch:  "main",
})
```

El front matter sigue ganando donde una página declare cualquiera de los dos.
Degrada en silencio a propósito: sin binario de git, con un clon superficial o
con contenido servido desde un `embed.FS`, cada página queda tal como se
escribió, porque un build de documentación no debe exigir historial de control
de versiones.
