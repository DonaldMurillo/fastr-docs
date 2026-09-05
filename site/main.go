package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	fastrdocs "github.com/DonaldMurillo/fastr-docs"
	docsite "github.com/DonaldMurillo/fastr-docs/site/docs"
	uiapp "github.com/DonaldMurillo/gofastr/core-ui/app"
	"github.com/DonaldMurillo/gofastr/core/middleware"
	"github.com/DonaldMurillo/gofastr/framework"
	"github.com/DonaldMurillo/gofastr/framework/uihost"
)

func main() {
	built, err := buildSite()
	if err != nil {
		panic(err)
	}
	server := built.server
	if dir := exportDir(os.Args[1:]); dir != "" {
		base := normalizeBase(exportBase(os.Args[1:]))
		if err := server.ExportStatic(context.Background(), dir, base); err != nil {
			panic(err)
		}
		if err := fastrdocs.WriteStaticRSS(dir, base, "/blog/feed.xml", built.rss); err != nil {
			panic(err)
		}
		if err := fastrdocs.WriteStaticNotFound(dir, base, built.notFound, built.notFoundCSS); err != nil {
			panic(err)
		}
		if err := built.router.WriteRuntimeAssets(dir, built.router.AssetPrefix(), normalizeBase(exportBase(os.Args[1:]))); err != nil {
			panic(err)
		}
		if err := copyPublicAssets(dir); err != nil {
			panic(err)
		}
		if err := fastrdocs.WriteAgentAssets(dir, base, server.Router()); err != nil {
			panic(err)
		}
		if err := rewriteRuntimeURLs(dir, base); err != nil {
			panic(err)
		}
		if err := fastrdocs.RewriteStaticCSP(dir, fastrdocs.ContentSecurityPolicy(built.serverConnectOrigins...)); err != nil {
			panic(err)
		}
		fmt.Println("static site exported to " + dir)
		return
	}
	addr := ":3079"
	if port := os.Getenv("PORT"); port != "" {
		addr = listenAddress(port, addr)
	}
	fmt.Println("fastr-docs" + " listening on " + listenURL(addr))
	if err := server.Start(addr); err != nil {
		panic(err)
	}
}

func listenAddress(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	if strings.Contains(value, ":") {
		return value
	}
	return ":" + value
}

func listenURL(addr string) string {
	if strings.HasPrefix(addr, ":") {
		return "http://localhost" + addr
	}
	return "http://" + addr
}

func sitePath(parts ...string) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return filepath.Join(parts...)
	}
	return filepath.Join(append([]string{filepath.Dir(file)}, parts...)...)
}

type generatedSite struct {
	server               *framework.App
	router               *fastrdocs.Router
	searchIndex          []byte
	manifest             []byte
	rss                  []byte
	serverConnectOrigins []string
	notFound             fastrdocs.NotFoundScreen
	notFoundCSS          string
}

func buildSite() (*generatedSite, error) {
	router := docsite.NewRouter()
	notFound := fastrdocs.NotFoundScreen{SiteName: router.SiteName()}
	site := uiapp.NewApp("fastr-docs").WithTheme(router.Theme()).WithLang(router.Language())
	if err := router.Mount(site, router.Layout()); err != nil {
		return nil, err
	}
	assetNames, err := router.RuntimeAssetNames(normalizeBase(exportBase(os.Args[1:])))
	if err != nil {
		return nil, err
	}
	// Page scripts are asked for by name rather than filtered by suffix: a
	// plugin may ship JavaScript that belongs somewhere other than the page,
	// such as the Mermaid bundle the sandboxed frame loads for itself.
	scriptNames, err := router.RuntimeScriptNames(normalizeBase(exportBase(os.Args[1:])))
	if err != nil {
		return nil, err
	}
	precache := make([]string, 0, len(assetNames)+1)
	for _, name := range assetNames {
		precache = append(precache, router.AssetPrefix()+"/"+name)
	}
	if router.SearchBackend() == fastrdocs.SearchBackendPagefind {
		precache = append(precache, "/pagefind/pagefind.js")
	}
	host := uihost.New(site,
		uihost.WithDescription("fastr-docs"+" — reusable documentation built with GoFastr."),
		uihost.WithCustomCSS(router.CSS()+"\n"+router.BrandCSS()+"\n"+docsite.ExamplesCSS()+"\n"+docsite.OpenAPICSS()),
		uihost.WithNotFoundScreen(notFound),
		uihost.WithPublicLLMMD(),
		uihost.WithAgentReady(uihost.AgentReadyConfig{
			Title:     "fastr-docs",
			Summary:   "Searchable documentation with a router-owned content tree.",
			WhenToUse: "Use this site to find the project's guides, API reference, and implementation examples.",
			AgentCard: &uihost.AgentCardConfig{
				Name:        "fastr-docs",
				Description: "Searchable documentation with a router-owned content tree.",
				MCPEndpoint: "/mcp",
			},
			CLI: &uihost.CLIToolConfig{
				Name:    "fastr-docs",
				Install: "go install github.com/DonaldMurillo/fastr-docs/cmd/fastr-docs@latest",
				Docs:    "/docs/getting-started",
			},
		}),
		uihost.WithHeadHTML(`<link rel="icon" href="/assets/favicon.svg">`),
		uihost.WithThemeColor(router.ThemeColor()),
		uihost.WithSitemap(uihost.SitemapConfig{BaseURL: publicSiteURL(), ExcludePaths: append([]string{"/__fastr-docs/"}, router.SitemapExcludePaths()...)}),
		uihost.WithRobots(uihost.RobotsConfig{Disallow: []string{"/__fastr-docs/"}}),
		uihost.WithExtraScripts(assetURLs(router, scriptNames)...),
		uihost.WithAppIcon(docsite.DefaultIconPNG()),
		uihost.WithPWA(uihost.PWAConfig{Name: "fastr-docs", ShortName: "fastr-docs", Precache: precache}),
	)
	server := framework.NewUIHostApp(host,
		framework.WithConfig(framework.AppConfig{
			Name: "fastr-docs",
			SecurityHeaders: middleware.SecurityHeadersConfig{
				ContentSecurityPolicy: fastrdocs.ContentSecurityPolicy(router.ConnectOrigins()...),
			},
		}),
		framework.WithMCP(),
		framework.WithMCPIntrospection(),
	)
	if err := router.MountNavigation(server.Router()); err != nil {
		return nil, err
	}
	if err := router.MountCommandPalette(server.Router()); err != nil {
		return nil, err
	}
	if err := router.MountPluginAssets(server.Router()); err != nil {
		return nil, err
	}
	feedConfig := fastrdocs.RSSConfig{
		Prefix: "/blog", Title: "fastr-docs updates",
		Description: "Release notes and implementation updates for fastr-docs.",
		SiteURL:     publicSiteURL(), Limit: 20,
	}
	if err := router.MountRSS(server.Router(), "/blog/feed.xml", feedConfig); err != nil {
		return nil, err
	}
	rss, err := router.RSSXML(feedConfig)
	if err != nil {
		return nil, err
	}
	// One call serves the docs runtime, the search index, the export manifest,
	// and every plugin's assets.
	if err := router.MountRuntimeAssets(server.Router(), router.AssetPrefix(), normalizeBase(exportBase(os.Args[1:]))); err != nil {
		return nil, err
	}
	publicDir := sitePath("public")
	if _, err := os.Stat(publicDir); err == nil {
		if err := router.MountAssets(server.Router(), fastrdocs.AssetConfig{FS: os.DirFS(publicDir), Prefix: "/assets", MaxAge: 365 * 24 * 60 * 60 * 1e9}); err != nil {
			return nil, err
		}
	}
	return &generatedSite{server: server, router: router, rss: rss, serverConnectOrigins: router.ConnectOrigins(), notFound: notFound, notFoundCSS: router.CSS() + "\n" + router.BrandCSS()}, nil
}

func copyPublicAssets(dir string) error {
	publicDir := sitePath("public")
	if _, err := os.Stat(publicDir); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return filepath.WalkDir(publicDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(publicDir, path)
		if err != nil {
			return err
		}
		destination := filepath.Join(dir, "assets", rel)
		if entry.IsDir() {
			return os.MkdirAll(destination, 0o755)
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
			return err
		}
		return os.WriteFile(destination, body, 0o644)
	})
}

func publicSiteURL() string {
	if value := strings.TrimSpace(os.Getenv("PUBLIC_SITE_URL")); value != "" {
		return strings.TrimRight(value, "/")
	}
	return "http://localhost:3079"
}

func rewriteRuntimeURLs(dir, base string) error {
	base = normalizeBase(base)
	if base == "" {
		return nil
	}
	return filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".html" {
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		updated := strings.ReplaceAll(string(body), `"/__fastr-docs/`, `"`+base+`/__fastr-docs/`)
		updated = strings.ReplaceAll(updated, `data-fastr-docs-pagefind-path="/pagefind/"`, `data-fastr-docs-pagefind-path="`+base+`/pagefind/"`)
		if updated == string(body) {
			return nil
		}
		return os.WriteFile(path, []byte(updated), 0o644)
	})
}

func rewriteStaticCSP(dir, policy string) error {
	return fastrdocs.RewriteStaticCSP(dir, policy)
}

func normalizeBase(base string) string {
	base = strings.TrimSpace(base)
	if base == "" || base == "/" {
		return ""
	}
	if !strings.HasPrefix(base, "/") {
		base = "/" + base
	}
	return strings.TrimRight(base, "/")
}

func exportDir(args []string) string {
	for i := 0; i < len(args); i++ {
		if args[i] == "--export" && i+1 < len(args) {
			return args[i+1]
		}
		if strings.HasPrefix(args[i], "--export=") {
			return strings.TrimPrefix(args[i], "--export=")
		}
	}
	return ""
}

func exportBase(args []string) string {
	for i := 0; i < len(args); i++ {
		if args[i] == "--export-base" && i+1 < len(args) {
			return args[i+1]
		}
		if strings.HasPrefix(args[i], "--export-base=") {
			return strings.TrimPrefix(args[i], "--export-base=")
		}
	}
	return ""
}

// assetURLs turns runtime asset names into the URLs the host serves them
// from.
func assetURLs(router *fastrdocs.Router, names []string) []string {
	urls := make([]string, 0, len(names))
	for _, name := range names {
		urls = append(urls, router.AssetPrefix()+"/"+name)
	}
	return urls
}
