package docs

import (
	"errors"
	"io/fs"
	"strings"
	"time"

	"github.com/DonaldMurillo/gofastr/core/router"
	"github.com/DonaldMurillo/gofastr/core/static"
)

// AssetConfig mounts a project's public fs.FS through GoFastr's hardened
// static handler. It works with embed.FS, os.DirFS, or any other fs.FS.
type AssetConfig struct {
	FS     fs.FS
	Prefix string
	MaxAge time.Duration
}

// MountAssets serves public assets from the same host as the docs app. The
// framework static handler supplies MIME types, ETags, cache headers, and
// traversal/dotfile protection.
func (r *Router) MountAssets(httpRouter *router.Router, cfg AssetConfig) error {
	if httpRouter == nil {
		return errors.New("docs: MountAssets requires a GoFastr router")
	}
	if cfg.FS == nil {
		return errors.New("docs: MountAssets requires an fs.FS")
	}
	prefix := strings.TrimSpace(cfg.Prefix)
	if prefix == "" {
		prefix = "/assets"
	}
	if !strings.HasPrefix(prefix, "/") || strings.ContainsAny(prefix, "?#") {
		return errors.New("docs: asset prefix must be an absolute path without query or fragment data")
	}
	httpRouter.Get(strings.TrimRight(prefix, "/")+"/{path...}", static.Handler(static.Config{
		FS: cfg.FS, Prefix: prefix, MaxAge: cfg.MaxAge,
	}))
	return nil
}
