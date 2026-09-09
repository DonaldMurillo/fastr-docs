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
	"reflect"
	"slices"
	"strings"
	"testing"
)

// initProject scaffolds a generated project and returns its directory plus
// the init output, the fixture every command test starts from.
func initProject(t *testing.T, name, module string) (string, string) {
	t.Helper()
	target := t.TempDir()
	var output bytes.Buffer
	if err := Run([]string{"init", target, "--name", name, "--module", module}, &output, &output); err != nil {
		t.Fatalf("init %s: %v\n%s", name, err, output.String())
	}
	return target, output.String()
}

// lookupStub stands in for exec.LookPath, resolving only the executables a
// test wants installed.
func lookupStub(paths map[string]string) func(string) (string, error) {
	return func(name string) (string, error) {
		if path, ok := paths[name]; ok {
			return path, nil
		}
		return "", fmt.Errorf("%s is not installed", name)
	}
}

func TestInitAndCheckGenerateACompleteProject(t *testing.T) {
	target, initOutput := initProject(t, "Northstar Docs", "example.com/northstar-docs")
	if !strings.Contains(initOutput, "created fastr-docs project") {
		t.Fatalf("init output = %q", initOutput)
	}
	var output bytes.Buffer
	if err := Run([]string{"check", target}, &output, &output); err != nil {
		t.Fatalf("check error = %v\n%s", err, output.String())
	}
	if err := Run([]string{"doctor", target}, &output, &output); err != nil {
		t.Fatalf("doctor error = %v\n%s", err, output.String())
	}
	for _, name := range []string{"main.go", "docs/router.go", "docs/icon.go", "content/index.md", "content/build-themes.md", "openapi.json", "agents/claude.md", ".agents/skills/docs-authoring/SKILL.md"} {
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
	if !strings.Contains(string(routerSource), "MarkdownComponentsPlugin") || !strings.Contains(string(routerSource), `os.Getenv("DOCS_SEARCH_BACKEND")`) || !strings.Contains(string(routerSource), `os.Getenv("DOCS_TEMPLATE")`) {
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
	if !strings.Contains(output.String(), "browser refresh") {
		t.Fatalf("help did not document the dev reload loop: %s", output.String())
	}
}

func TestVersionUsesBuildMetadata(t *testing.T) {
	previous := Version
	Version = "v0.2.0"
	t.Cleanup(func() { Version = previous })
	var output bytes.Buffer
	if err := Run([]string{"version"}, &output, nil); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(output.String()); got != "fastr-docs v0.2.0" {
		t.Fatalf("version output = %q", got)
	}
}

func TestSplitInitArgsAcceptsEqualsForm(t *testing.T) {
	flags, positionals, err := splitInitArgs([]string{"./docs", "--name=Manual Docs", "--module=example.com/manual", "--force"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(positionals, []string{"./docs"}) || !reflect.DeepEqual(flags, []string{"--name", "Manual Docs", "--module", "example.com/manual", "--force"}) {
		t.Fatalf("splitInitArgs() = %#v, %#v", flags, positionals)
	}
}

func TestCreateDevReloadMarkerRemovesStaleMarkers(t *testing.T) {
	target := t.TempDir()
	stale := filepath.Join(target, ".fastr-docs-dev-reload-stale.go")
	if err := os.WriteFile(stale, []byte("// stale"), 0o644); err != nil {
		t.Fatal(err)
	}
	marker, err := createDevReloadMarker(target)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatalf("stale marker still exists: %v", err)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("new marker missing: %v", err)
	}
}

func TestGofastrDevCommandPrefersInstalledCLIAndForwardsFlags(t *testing.T) {
	target := filepath.Join(t.TempDir(), "docs")
	want := []string{"dev", "--dir", target, "--addr", "localhost:4173", "--no-a11y"}
	name, got, err := gofastrCommandWithLookup("dev", target, want[3:], lookupStub(map[string]string{"gofastr": `C:\tools\gofastr.exe`}))
	if err != nil {
		t.Fatalf("gofastrCommand error = %v", err)
	}
	if name != `C:\tools\gofastr.exe` {
		t.Fatalf("command name = %q", name)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("command args = %#v, want %#v", got, want)
	}
}

func TestGofastrDevCommandFallsBackToProjectGoModule(t *testing.T) {
	target := filepath.Join(t.TempDir(), "docs")
	name, got, err := gofastrCommandWithLookup("dev", target, []string{"--pkg", "./cmd/docs"}, lookupStub(map[string]string{"go": `C:\Go\bin\go.exe`}))
	if err != nil {
		t.Fatalf("gofastrCommand fallback error = %v", err)
	}
	if name != `C:\Go\bin\go.exe` {
		t.Fatalf("fallback command name = %q", name)
	}
	want := []string{"run", "-mod=mod", "github.com/DonaldMurillo/gofastr/cmd/gofastr", "dev", "--dir", target, "--pkg", "./cmd/docs"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("fallback command args = %#v, want %#v", got, want)
	}
}

func TestGofastrBuildCommandUsesInstalledCLI(t *testing.T) {
	target := filepath.Join(t.TempDir(), "docs")
	name, got, err := gofastrCommandWithLookup("build", target, []string{"--no-a11y"}, lookupStub(map[string]string{"gofastr": `C:\tools\gofastr.exe`}))
	if err != nil {
		t.Fatalf("gofastr build command error = %v", err)
	}
	if name != `C:\tools\gofastr.exe` {
		t.Fatalf("command name = %q", name)
	}
	want := []string{"build", "--no-a11y"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("command args = %#v, want %#v", got, want)
	}
}

func TestGofastrUpgradeCommandPassesProjectRoot(t *testing.T) {
	target := filepath.Join(t.TempDir(), "docs")
	name, got, err := gofastrCommandWithLookup("upgrade", target, []string{"--apply"}, lookupStub(map[string]string{"gofastr": `C:\tools\gofastr.exe`}))
	if err != nil {
		t.Fatalf("gofastr upgrade command error = %v", err)
	}
	if name != `C:\tools\gofastr.exe` {
		t.Fatalf("command name = %q", name)
	}
	want := []string{"upgrade", target, "--apply"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("command args = %#v, want %#v", got, want)
	}
}

func TestDevExtraFileScanWatchesJSONAndYAMLSpecs(t *testing.T) {
	target := t.TempDir()
	if err := os.Mkdir(filepath.Join(target, "dist"), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{
		"openapi.json":                       "{}",
		"openapi.yaml":                       "openapi: 3.1.0",
		"config.yml":                         "name: docs",
		"content.md":                         "# docs",
		filepath.Join("dist", "export.json"): "{}",
	} {
		if err := os.WriteFile(filepath.Join(target, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	files := scanDevExtraModTimes(target)
	for _, name := range []string{"openapi.json", "openapi.yaml", "config.yml"} {
		if _, ok := files[filepath.Join(target, name)]; !ok {
			t.Fatalf("extra dev watcher did not include %s", name)
		}
	}
	for _, name := range []string{"content.md", filepath.Join("dist", "export.json")} {
		if _, ok := files[filepath.Join(target, name)]; ok {
			t.Fatalf("extra dev watcher unexpectedly included %s", name)
		}
	}
}

func TestInitEscapesSiteNameForGeneratedGoAndJSON(t *testing.T) {
	name := `Acme "Docs"`
	target, _ := initProject(t, name, "example.com/quoted-docs")
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

func TestInitRejectsNamesWithPathSeparators(t *testing.T) {
	var out, errOut strings.Builder
	if err := Run([]string{"init", "--name", "a/b", t.TempDir()}, &out, &errOut); err == nil {
		t.Fatal("a name carrying a path separator writes into the filesystem")
	}
}

func TestCheckRunsStrictRouterValidation(t *testing.T) {
	target, _ := initProject(t, "Validation Docs", "example.com/validation-docs")
	routerPath := filepath.Join(target, "docs", "router.go")
	routerSource, err := os.ReadFile(routerPath)
	if err != nil {
		t.Fatal(err)
	}
	routerText := string(routerSource)
	start := strings.Index(routerText, `router.MustPage("/docs/getting-started"`)
	orderOffset := -1
	if start >= 0 {
		orderOffset = strings.Index(routerText[start:], "Order: 1,")
	}
	if orderOffset < 0 {
		t.Fatal("test fixture did not find the Getting started order")
	}
	orderOffset += start
	updated := routerText[:orderOffset] + "Order: 0," + routerText[orderOffset+len("Order: 1,"):]
	if err := os.WriteFile(routerPath, []byte(updated), 0o644); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	err = Run([]string{"check", target}, &output, &output)
	if err == nil || !strings.Contains(output.String(), `route "/docs/getting-started": a positive explicit Order is required`) {
		t.Fatalf("check error = %v, output = %s", err, output.String())
	}
}

func TestGeneratedProjectCompilesAgainstTheLocalWorkspace(t *testing.T) {
	target, _ := initProject(t, "Build Test", "example.com/build-test")
	goMod := filepath.Join(target, "go.mod")
	modBody, err := os.ReadFile(goMod)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(modBody), "replace github.com/DonaldMurillo/fastr-docs =>") {
		t.Fatalf("generated go.mod did not contain the local source replace: %s", modBody)
	}
	if strings.Contains(string(modBody), "replace github.com/DonaldMurillo/gofastr =>") {
		t.Fatalf("generated go.mod should use the released GoFastr module: %s", modBody)
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
	if len(router.Routes()) < 30 {
		t.Fatalf("route count: %%d, want the generated docs and publication surfaces", len(router.Routes()))
	}
	for _, path := range []string{"/blog", "/blog/search", "/blog/archive", "/blog/tags", "/blog/authors"} {
		found := false
		for _, route := range router.Routes() {
			if route.Path == path {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("generated publication route missing: %%s", path)
		}
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
		{path: "/__gofastr/pwa/register.js", marker: "gofastr:pwa-update"},
		{path: "/__gofastr/pwa/offline", marker: "offline"},
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

	goExecutable, err := findGoExecutable()
	if err != nil {
		t.Fatal(err)
	}
	tidy := exec.Command(goExecutable, "mod", "tidy")
	tidy.Dir = target
	if output, err := tidy.CombinedOutput(); err != nil {
		t.Fatalf("generated go mod tidy: %v\n%s", err, output)
	}
	build := exec.Command(goExecutable, "test", "./...")
	build.Dir = target
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("generated go test: %v\n%s", err, output)
	}
	dist := filepath.Join(target, "dist")
	var exportOutput bytes.Buffer
	if err := Run([]string{"export", target, "--out", "dist"}, &exportOutput, &exportOutput); err != nil {
		t.Fatalf("generated static export: %v\n%s", err, exportOutput.String())
	}
	for _, name := range []string{"index.html", "404.html", "404.css", filepath.Join("docs", "index.html"), filepath.Join("docs", "getting-started", "index.html"), filepath.Join("api-reference", "index.html"), filepath.Join("blog", "index.html"), filepath.Join("blog", "search", "index.html"), filepath.Join("blog", "archive", "index.html"), filepath.Join("blog", "tags", "index.html"), filepath.Join("blog", "authors", "index.html"), "manifest.webmanifest", "service-worker.js", "sitemap.xml", "robots.txt", "llms.txt", filepath.Join(".well-known", "agent-card.json"), filepath.Join("assets", "favicon.svg"), filepath.Join("__gofastr", "icons", "icon-192.png"), filepath.Join("__gofastr", "icons", "icon-512.png"), filepath.Join("__gofastr", "widgets.json"), filepath.Join("__gofastr", "runtime", "scrollspy.js"), filepath.Join("__gofastr", "pwa", "register.js"), filepath.Join("__gofastr", "pwa", "offline", "index.html"), filepath.Join("core-ui", "widget", "fastr-docs-sections", "chrome"), filepath.Join("core-ui", "widget", "fastr-docs-blog-sections", "chrome"), filepath.Join("__fastr-docs", "docs.js"), filepath.Join("__fastr-docs", "openapi.js"), filepath.Join("__fastr-docs", "search.json"), filepath.Join("__fastr-docs", "manifest.json")} {
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
	if !strings.Contains(string(serviceWorker), "gofastr-pwa-static-") || !strings.Contains(string(serviceWorker), "/__fastr-docs/docs.js") || !strings.Contains(string(serviceWorker), "/__fastr-docs/openapi.js") || !strings.Contains(string(serviceWorker), "/__fastr-docs/search.json") || !strings.Contains(string(serviceWorker), "/__fastr-docs/manifest.json") || !strings.Contains(string(serviceWorker), "fastr-docs-sections/chrome") || !strings.Contains(string(serviceWorker), "runtime/scrollspy.js") {
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

func TestInitShipsSkillsToBothAgentDirectoriesAndTheGitignore(t *testing.T) {
	target, _ := initProject(t, "Skill Docs", "example.com/skill-docs")

	skills := starterSkills()
	if len(skills) < 2 {
		t.Fatalf("starterSkills() = %v, want the shipped skill set", skills)
	}
	for _, skill := range skills {
		agent := filepath.Join(target, filepath.FromSlash(agentSkillsDir), skill, "SKILL.md")
		claude := filepath.Join(target, filepath.FromSlash(claudeSkillsDir), skill, "SKILL.md")
		agentBody, err := os.ReadFile(agent)
		if err != nil {
			t.Fatalf("read %s: %v", agent, err)
		}
		claudeBody, err := os.ReadFile(claude)
		if err != nil {
			t.Fatalf("read %s: %v", claude, err)
		}
		if !bytes.Equal(agentBody, claudeBody) {
			t.Fatalf("%s differs between %s and %s", skill, agentSkillsDir, claudeSkillsDir)
		}
		if !bytes.Contains(agentBody, []byte("name: "+skill)) {
			t.Fatalf("%s front matter does not declare its own name:\n%s", skill, agentBody)
		}
	}

	// A plain embed glob silently drops dotfiles, which used to leave generated
	// projects without the ignore rules for dev-loop and export artifacts.
	ignore, err := os.ReadFile(filepath.Join(target, ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	for _, rule := range []string{"/dist/", ".fastr-docs-dev-reload-*.go"} {
		if !strings.Contains(string(ignore), rule) {
			t.Fatalf(".gitignore missing %q:\n%s", rule, ignore)
		}
	}
}

func TestCheckReportsEveryKindOfSkillDrift(t *testing.T) {
	target, _ := initProject(t, "Drift Docs", "example.com/drift-docs")
	var output bytes.Buffer
	if drift, err := skillDrift(target); err != nil || len(drift) != 0 {
		t.Fatalf("skillDrift() on a fresh project = %v, %v; want none", drift, err)
	}

	// One of each: an edit that landed only in the authored copy, a deleted
	// mirror, and a file that exists only in the mirror.
	edited := filepath.Join(target, filepath.FromSlash(agentSkillsDir), "docs-blog", "SKILL.md")
	body, err := os.ReadFile(edited)
	if err != nil {
		t.Fatalf("read %s: %v", edited, err)
	}
	if err := os.WriteFile(edited, append(body, "\nA later edit.\n"...), 0o644); err != nil {
		t.Fatalf("write %s: %v", edited, err)
	}
	if err := os.Remove(filepath.Join(target, filepath.FromSlash(claudeSkillsDir), "docs-theming", "SKILL.md")); err != nil {
		t.Fatalf("remove mirrored skill: %v", err)
	}
	orphan := filepath.Join(target, filepath.FromSlash(claudeSkillsDir), "orphan")
	if err := os.MkdirAll(orphan, 0o755); err != nil {
		t.Fatalf("create orphan: %v", err)
	}
	if err := os.WriteFile(filepath.Join(orphan, "SKILL.md"), []byte("stale"), 0o644); err != nil {
		t.Fatalf("write orphan: %v", err)
	}

	drift, err := skillDrift(target)
	if err != nil {
		t.Fatalf("skillDrift: %v", err)
	}
	for _, want := range []string{
		"docs-blog/SKILL.md: differs between " + agentSkillsDir + " and " + claudeSkillsDir,
		"docs-theming/SKILL.md: missing from " + claudeSkillsDir,
		"orphan/SKILL.md: present in " + claudeSkillsDir + " but not " + agentSkillsDir,
	} {
		if !slices.Contains(drift, want) {
			t.Fatalf("skillDrift() = %v, missing %q", drift, want)
		}
	}

	output.Reset()
	if err := Run([]string{"check", target}, &output, &output); err == nil {
		t.Fatal("check passed on a drifted project")
	} else if !strings.Contains(err.Error(), "agent skills have drifted") {
		t.Fatalf("check error did not name the drift: %v", err)
	}

	// sync-skills alone repairs the stale and missing copies but leaves the
	// orphan, because deleting is opt-in.
	output.Reset()
	if err := Run([]string{"sync-skills", target}, &output, &output); err != nil {
		t.Fatalf("sync-skills: %v\n%s", err, output.String())
	}
	drift, err = skillDrift(target)
	if err != nil {
		t.Fatalf("skillDrift after sync: %v", err)
	}
	if len(drift) != 1 || !strings.HasPrefix(drift[0], "orphan/SKILL.md") {
		t.Fatalf("skillDrift() after sync = %v, want only the orphan", drift)
	}

	output.Reset()
	if err := Run([]string{"sync-skills", target, "--prune"}, &output, &output); err != nil {
		t.Fatalf("sync-skills --prune: %v\n%s", err, output.String())
	}
	if drift, err := skillDrift(target); err != nil || len(drift) != 0 {
		t.Fatalf("skillDrift() after prune = %v, %v; want none", drift, err)
	}
	if _, err := os.Stat(orphan); !os.IsNotExist(err) {
		t.Fatalf("prune left the empty orphan directory behind: %v", err)
	}
}

// TestDogfoodSiteSkillsMatchTheTemplate guards the copies inside this
// repository. site/ is a checked-in generated project, so its two skill
// directories have to stay identical to the template they came from; nothing
// else would notice if an edit landed in only one of the three.
func TestDogfoodSiteSkillsMatchTheTemplate(t *testing.T) {
	siteRoot := filepath.Join("..", "..", "site")
	if _, err := os.Stat(siteRoot); os.IsNotExist(err) {
		t.Skip("no dogfood site in this checkout")
	}
	siteAgents, err := skillFiles(filepath.Join(siteRoot, filepath.FromSlash(agentSkillsDir)))
	if err != nil {
		t.Fatalf("read site %s: %v", agentSkillsDir, err)
	}
	siteClaude, err := skillFiles(filepath.Join(siteRoot, filepath.FromSlash(claudeSkillsDir)))
	if err != nil {
		t.Fatalf("read site %s: %v", claudeSkillsDir, err)
	}

	template := make(map[string][]byte)
	for _, skill := range starterSkills() {
		rel := skill + "/SKILL.md"
		body, err := starterTemplates.ReadFile("templates/starter/" + agentSkillsDir + "/" + rel)
		if err != nil {
			t.Fatalf("read template skill %s: %v", rel, err)
		}
		template[rel] = body
	}

	for _, pair := range []struct {
		name  string
		files map[string][]byte
	}{
		{"site/" + agentSkillsDir, siteAgents},
		{"site/" + claudeSkillsDir, siteClaude},
	} {
		if len(pair.files) != len(template) {
			t.Fatalf("%s has %d skill files, template has %d", pair.name, len(pair.files), len(template))
		}
		for rel, want := range template {
			got, ok := pair.files[rel]
			if !ok {
				t.Fatalf("%s is missing %s", pair.name, rel)
			}
			// The site copies are checked in, so on Windows they arrive with
			// CRLF while the embedded template keeps LF. Compare content, not
			// the checkout's line-ending policy.
			if !bytes.Equal(normalizeNewlines(got), normalizeNewlines(want)) {
				t.Fatalf("%s/%s has drifted from the starter template; copy the template version across", pair.name, rel)
			}
		}
	}
}

func normalizeNewlines(body []byte) []byte {
	return bytes.ReplaceAll(body, []byte("\r\n"), []byte("\n"))
}
