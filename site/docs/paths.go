package docs

import (
	"path/filepath"
	"runtime"
)

// siteFile resolves files next to the self-hosted site instead of depending
// on the caller's working directory. This keeps both `go run ./site` from the
// repository root and `go run .` from site/ working the same way.
func siteFile(parts ...string) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return filepath.Join(parts...)
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), ".."))
	return filepath.Join(append([]string{root}, parts...)...)
}

func contentFile(name string) string { return siteFile("content", name) }

func contractFile() string   { return siteFile("openapi.json") }
func contractFileES() string { return siteFile("openapi.es.json") }
