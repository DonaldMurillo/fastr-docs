package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestInitAndCheckGenerateACompleteProject(t *testing.T) {
	target := t.TempDir()
	var output bytes.Buffer
	if err := Run([]string{"init", target, "--name", "Northstar Docs", "--module", "example.com/northstar-docs"}, &output, &output); err != nil {
		t.Fatalf("init error = %v\n%s", err, output.String())
	}
	if !strings.Contains(output.String(), "created fastr-docs project") {
		t.Fatalf("init output = %q", output.String())
	}
	if err := Run([]string{"check", target}, &output, &output); err != nil {
		t.Fatalf("check error = %v\n%s", err, output.String())
	}
	if err := Run([]string{"doctor", target}, &output, &output); err != nil {
		t.Fatalf("doctor error = %v\n%s", err, output.String())
	}
	for _, name := range []string{"main.go", "docs/router.go", "docs/icon.go", "content/index.md", "openapi.json", "agents/claude.md", ".agents/skills/docs-authoring/SKILL.md"} {
		if _, err := os.Stat(filepath.Join(target, filepath.FromSlash(name))); err != nil {
			t.Fatalf("generated file %s: %v", name, err)
		}
	}
	readme, err := os.ReadFile(filepath.Join(target, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(readme), "# Northstar Docs") || strings.Contains(string(readme), "{{.SiteName}}") {
		t.Fatalf("generated README was not rendered: %s", readme)
	}
	index, err := os.ReadFile(filepath.Join(target, "content", "index.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(index), "# Northstar Docs") || strings.Contains(string(index), "{{.SiteName}}") {
		t.Fatalf("generated index was not rendered: %s", index)
	}
	spec, err := os.ReadFile(filepath.Join(target, "openapi.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(spec), "Northstar Docs API") || strings.Contains(string(spec), "{{.SiteName}}") {
		t.Fatalf("generated OpenAPI spec was not rendered: %s", spec)
	}
	routerSource, err := os.ReadFile(filepath.Join(target, "docs", "router.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(routerSource), `os.Getenv("API_SERVER_URL")`) {
		t.Fatal("generated router did not expose the API_SERVER_URL override")
	}
	if !strings.Contains(string(routerSource), "MarkdownComponentsPlugin") || !strings.Contains(string(routerSource), `os.Getenv("DOCS_SEARCH_BACKEND")`) {
		t.Fatalf("generated router did not include the reusable authoring/search integrations: %s", routerSource)
	}
}

func TestHelpDocumentsPagefindExport(t *testing.T) {
	var output bytes.Buffer
	if err := Run([]string{"help"}, &output, &output); err != nil {
		t.Fatalf("help error = %v", err)
	}
	if !strings.Contains(output.String(), "--pagefind") {
		t.Fatalf("help did not document Pagefind export: %s", output.String())
	}
}

func TestInitEscapesSiteNameForGeneratedGoAndJSON(t *testing.T) {
	target := t.TempDir()
	name := `Acme "Docs"`
	if err := Run([]string{"init", target, "--name", name, "--module", "example.com/quoted-docs"}, nil, nil); err != nil {
		t.Fatalf("init error = %v", err)
	}
	for _, file := range []string{"main.go", filepath.Join("docs", "router.go"), filepath.Join("docs", "icon.go")} {
		if _, err := parser.ParseFile(token.NewFileSet(), filepath.Join(target, file), nil, parser.AllErrors); err != nil {
			t.Fatalf("generated %s is not valid Go: %v", file, err)
		}
	}
	var spec map[string]any
	body, err := os.ReadFile(filepath.Join(target, "openapi.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(body, &spec); err != nil {
		t.Fatalf("generated OpenAPI JSON is invalid: %v", err)
	}
	if !strings.Contains(string(body), `Acme \"Docs\" API`) {
		t.Fatalf("generated OpenAPI JSON did not preserve escaped site name: %s", body)
	}
}

func TestGeneratedProjectCompilesAgainstTheLocalWorkspace(t *testing.T) {
	if runtime.GOOS == "windows" && os.Getenv("FASTR_DOCS_SKIP_GENERATED_BUILD") == "1" {
		t.Skip("generated build disabled by environment")
	}
	target := t.TempDir()
	if err := Run([]string{"init", target, "--name", "Build Test", "--module", "example.com/build-test"}, nil, nil); err != nil {
		t.Fatalf("init error = %v", err)
	}
	goMod := filepath.Join(target, "go.mod")
	modBody, err := os.ReadFile(goMod)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(modBody), "replace github.com/DonaldMurillo/fastr-docs =>") {
		t.Fatalf("generated go.mod did not contain the local source replace: %s", modBody)
	}
	smoke := fmt.Sprintf(`package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	docsite %q
	uiapp "github.com/DonaldMurillo/gofastr/core-ui/app"
)

func TestGeneratedRouterSmoke(t *testing.T) {
	t.Setenv("API_SERVER_URL", "https://staging.example.com/api/")
	router := docsite.NewRouter()
	if err := router.Validate(); err != nil {
		t.Fatalf("router validation: %%v", err)
	}
	if len(router.Routes()) != 18 {
		t.Fatalf("route count: %%d", len(router.Routes()))
	}
	site := uiapp.NewApp("Build Test")
	if err := router.Mount(site, router.Layout()); err != nil {
		t.Fatalf("mount: %%v", err)
	}
	html, err := site.RenderPage(context.Background(), "/docs/getting-started")
	if err != nil {
		t.Fatalf("render: %%v", err)
	}
	if !strings.Contains(string(html), "Getting started") {
		t.Fatalf("rendered page did not contain title")
	}
	for _, marker := range []string{"ui-sidebar", "data-fui-open=\"fastr-docs-sections\"", "fastr-docs-command-trigger", "ui-doc-layout", "ui-anchored-rail", "data-fui-scrollspy", "/docs/getting-started", "/api-reference"} {
		if !strings.Contains(string(html), marker) {
			t.Fatalf("rendered navigation missing %%q", marker)
		}
	}
	apiHTML, err := site.RenderPage(context.Background(), "/api-reference")
	if err != nil {
		t.Fatalf("render api reference: %%v", err)
	}
	if !strings.Contains(string(apiHTML), "data-openapi-reference") || !strings.Contains(string(apiHTML), "data-openapi-server-url=\"https://staging.example.com/api\"") || !strings.Contains(string(apiHTML), "data-openapi-try") || !strings.Contains(string(apiHTML), "listProjects") {
		t.Fatalf("rendered API reference missing rich screen")
	}
}

func TestGeneratedLiveHostRoutes(t *testing.T) {
	built, err := buildSite()
	if err != nil {
		t.Fatalf("build site: %%v", err)
	}
	checks := []struct {
		path   string
		marker string
	}{
		{path: "/", marker: "ui-sidebar"},
		{path: "/docs/getting-started", marker: "data-fui-scrollspy"},
		{path: "/__fastr-docs/docs.js", marker: "initTocSelect"},
		{path: "/__fastr-docs/openapi.js", marker: "data-openapi-filter"},
		{path: "/__fastr-docs/manifest.json", marker: "fastr-docs/v1"},
		{path: "/core-ui/widget/fastr-docs-sections/chrome", marker: "ui-sidebar"},
		{path: "/__gofastr/runtime/scrollspy.js", marker: "data-fui-scrollspy"},
		{path: "/__gofastr/widgets", marker: "fastr-docs-sections"},
		{path: "/manifest.webmanifest", marker: "icon-192.png"},
		{path: "/llms.txt", marker: "## When to use"},
		{path: "/.well-known/agent-card.json", marker: "Build Test"},
		{path: "/service-worker.js", marker: "__fastr-docs/search.json"},
		{path: "/sitemap.xml", marker: "/docs/getting-started"},
		{path: "/robots.txt", marker: "Disallow: /__fastr-docs/"},
		{path: "/assets/favicon.svg", marker: "<svg"},
	}
	var cookies []*http.Cookie
	for _, check := range checks {
		req := httptest.NewRequest(http.MethodGet, check.path, nil)
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}
		res := httptest.NewRecorder()
		built.server.Router().ServeHTTP(res, req)
		if res.Code != http.StatusOK {
			t.Fatalf("GET %%s status = %%d, want 200; body: %%s", check.path, res.Code, res.Body.String())
		}
		if !strings.Contains(res.Body.String(), check.marker) {
			t.Fatalf("GET %%s missing %%q: %%s", check.path, check.marker, res.Body.String())
		}
		cookies = append(cookies, res.Result().Cookies()...)
	}
	iconReq := httptest.NewRequest(http.MethodGet, "/__gofastr/icons/icon-192.png", nil)
	iconRes := httptest.NewRecorder()
	built.server.Router().ServeHTTP(iconRes, iconReq)
	if iconRes.Code != http.StatusOK || !strings.HasPrefix(iconRes.Header().Get("Content-Type"), "image/png") {
		t.Fatalf("generated PWA icon response = %%d/%%q, want image/png", iconRes.Code, iconRes.Header().Get("Content-Type"))
	}
}
`, "example.com/build-test/docs")
	if err := os.WriteFile(filepath.Join(target, "generated_smoke_test.go"), []byte(smoke), 0o644); err != nil {
		t.Fatal(err)
	}

	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = target
	if output, err := tidy.CombinedOutput(); err != nil {
		t.Fatalf("generated go mod tidy: %v\n%s", err, output)
	}
	build := exec.Command("go", "test", "./...")
	build.Dir = target
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("generated go test: %v\n%s", err, output)
	}
	dist := filepath.Join(target, "dist")
	var exportOutput bytes.Buffer
	if err := Run([]string{"export", target, "--out", "dist"}, &exportOutput, &exportOutput); err != nil {
		t.Fatalf("generated static export: %v\n%s", err, exportOutput.String())
	}
	for _, name := range []string{"index.html", filepath.Join("docs", "index.html"), filepath.Join("docs", "getting-started", "index.html"), filepath.Join("api-reference", "index.html"), "manifest.webmanifest", "service-worker.js", "sitemap.xml", "robots.txt", "llms.txt", filepath.Join(".well-known", "agent-card.json"), filepath.Join("assets", "favicon.svg"), filepath.Join("__gofastr", "icons", "icon-192.png"), filepath.Join("__gofastr", "icons", "icon-512.png"), filepath.Join("__gofastr", "widgets.json"), filepath.Join("__gofastr", "runtime", "scrollspy.js"), filepath.Join("core-ui", "widget", "fastr-docs-sections", "chrome"), filepath.Join("__fastr-docs", "docs.js"), filepath.Join("__fastr-docs", "openapi.js"), filepath.Join("__fastr-docs", "search.json"), filepath.Join("__fastr-docs", "manifest.json")} {
		if _, err := os.Stat(filepath.Join(dist, name)); err != nil {
			t.Fatalf("static export missing %s: %v", name, err)
		}
	}
	manifest, err := os.ReadFile(filepath.Join(dist, "manifest.webmanifest"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(manifest), "icon-192.png") || !strings.Contains(string(manifest), "icon-512.png") {
		t.Fatal("static PWA manifest did not include generated install icons")
	}
	serviceWorker, err := os.ReadFile(filepath.Join(dist, "service-worker.js"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(serviceWorker), "/__fastr-docs/docs.js") || !strings.Contains(string(serviceWorker), "/__fastr-docs/openapi.js") || !strings.Contains(string(serviceWorker), "/__fastr-docs/search.json") || !strings.Contains(string(serviceWorker), "/__fastr-docs/manifest.json") || !strings.Contains(string(serviceWorker), "fastr-docs-sections/chrome") || !strings.Contains(string(serviceWorker), "runtime/scrollspy.js") {
		t.Fatal("static PWA service worker did not precache the OpenAPI runtime")
	}
	searchIndex, err := os.ReadFile(filepath.Join(dist, "__fastr-docs", "search.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(searchIndex), "listProjects") {
		t.Fatal("static search index did not include OpenAPI operation text")
	}

	baseDist := filepath.Join(target, "dist-base")
	var baseOutput bytes.Buffer
	if err := Run([]string{"export", target, "--out", baseDist, "--base", "/docs"}, &baseOutput, &baseOutput); err != nil {
		t.Fatalf("base-path static export: %v\n%s", err, baseOutput.String())
	}
	baseHTML, err := os.ReadFile(filepath.Join(baseDist, "docs", "getting-started", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	for _, marker := range []string{"/docs/__fastr-docs/docs.js", "/docs/__fastr-docs/openapi.js"} {
		if !strings.Contains(string(baseHTML), marker) {
			t.Fatalf("base-path HTML missing %q", marker)
		}
	}
	baseRuntime, err := os.ReadFile(filepath.Join(baseDist, "__fastr-docs", "docs.js"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(baseRuntime), "initTocSelect") {
		t.Fatal("base-path export did not include the docs runtime")
	}
	baseServiceWorker, err := os.ReadFile(filepath.Join(baseDist, "service-worker.js"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(baseServiceWorker), "/docs/__fastr-docs/search.json") {
		t.Fatal("base-path service worker did not precache the search index")
	}
	baseLLMs, err := os.ReadFile(filepath.Join(baseDist, "docs", "llms.txt"))
	if err != nil {
		t.Fatalf("base-path export missing llms.txt: %v", err)
	}
	if !strings.Contains(string(baseLLMs), "](/docs/") {
		t.Fatalf("base-path llms.txt did not rewrite root-relative links: %s", baseLLMs)
	}
}
