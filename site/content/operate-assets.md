---
tags: [assets, plugins, export, operate]
locale: en
---

# Runtime and plugin assets

One call collects everything served under the docs asset prefix: the browser
runtime, the search index, the export manifest, and whatever each plugin
contributes.

```go title="main.go"
router.MountRuntimeAssets(server.Router(), router.AssetPrefix(), exportBase)
router.WriteRuntimeAssets(dist, router.AssetPrefix(), exportBase)
```

`RuntimeAssetNames` returns the same list, so the PWA precache list and the
host's `<script>` tags are derived from it rather than being a second copy that
drifts.

## Contributing from a plugin

A plugin joins in by implementing `RuntimeAssetPlugin`:

```go title="plugin/openapi/openapi.go" {3}
func (Plugin) RuntimeAssets() (map[string][]byte, error) {
    body, err := fs.ReadFile(runtimeFS, "openapi.js")
    return map[string][]byte{"openapi.js": body}, err
}
```

That is the whole arrangement. Nothing in the generated `main.go` changes when you
add a plugin that ships browser code.

{{< warning title="Names are validated" >}}
File names come from the plugin, not the project, so they are checked. A plugin
cannot overwrite `docs.js` or escape the asset directory with `..`.
{{< /warning >}}

## Where the files land

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

`AssetPrefix` follows `WithSearchIndexPath`, so the prefix and the search index
URL cannot disagree. Change one and the other follows.

## Dates and edit links

Neither needs to be typed into a page. `WithGitMetadata` reads one `git log`
for the whole repository and fills `last_updated` and `edit_url` from history:

```go title="docs/router.go"
docs.WithGitMetadata(docs.GitMetadataConfig{
    RepoURL: os.Getenv("DOCS_REPO_URL"),
    Branch:  "main",
})
```

Front matter still wins where a page declares either. It degrades silently by
design: no git binary, a shallow clone, or content served from an `embed.FS`
leaves every page exactly as authored, because a docs build must not require
version-control history.
