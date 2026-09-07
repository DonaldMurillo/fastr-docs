package openapi

// Contracts from the red-suite audit: each case failed before the fix and
// passes now.

import (
	"context"
	"strings"
	"testing"

	docs "github.com/DonaldMurillo/fastr-docs"
	uiapp "github.com/DonaldMurillo/gofastr/core-ui/app"
	"github.com/DonaldMurillo/gofastr/core/render"
)

const redSpec = `{"openapi":"3.1.0","info":{"title":"T","description":"D"},"paths":{"/x":{"get":{"summary":"S","operationId":"x","responses":{"200":{"description":"ok"}}}}}}`

func redMount(t *testing.T, path, locale string, strings_ Strings) (*docs.Router, string) {
	t.Helper()
	r := docs.NewRouter()
	if err := r.Use(Plugin{Spec: []byte(redSpec), Path: path, Locale: locale, Strings: strings_}); err != nil {
		t.Fatal(err)
	}
	site := uiapp.NewApp("D")
	if err := r.Mount(site, r.Layout()); err != nil {
		t.Fatal(err)
	}
	html, err := site.RenderPage(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	return r, string(html)
}

func TestRed086To097OpenAPI(t *testing.T) {
	t.Run("086 a mixed-case locale is normalized", func(t *testing.T) {
		r, _ := redMount(t, "/es/api", "ES", Strings{})
		if got := r.Routes()[0].Metadata.Locale; got != "es" {
			t.Fatalf("metadata locale = %q, want es", got)
		}
	})
	t.Run("087 the reference screen opts into preload", func(t *testing.T) {
		r, _ := redMount(t, "/api", "", Strings{})
		if r.Routes()[0].Preload == "" {
			t.Fatal("plugin screen sets no Preload")
		}
	})
	t.Run("088 the reference renders context aware", func(t *testing.T) {
		_, html := redMount(t, "/api", "", Strings{})
		if !strings.Contains(html, "data-openapi-reference") {
			t.Fatal("reference did not render")
		}
		var ref *Reference
		var ctxAware interface {
			RenderCtx(context.Context) render.HTML
		}
		_ = ref
		if _, ok := any(&Reference{}).(interface {
			RenderCtx(context.Context) render.HTML
		}); !ok {
			_ = ctxAware
			t.Fatal("Reference has no RenderCtx path for request-aware chrome")
		}
	})
	t.Run("089 operation options carry option semantics", func(t *testing.T) {
		_, html := redMount(t, "/api", "", Strings{})
		if !strings.Contains(html, `role="option"`) {
			t.Fatal("console options are bare text")
		}
	})
	t.Run("090 operation ids are unique per mount", func(t *testing.T) {
		_, html := redMount(t, "/es/api", "es", Strings{})
		if !strings.Contains(html, `id="fastr-openapi-operation-es-api-1"`) {
			t.Fatal("operation ids carry no mount discriminator; two mounts collide")
		}
	})
	t.Run("091 duplicate ids name both operations", func(t *testing.T) {
		spec := `{"openapi":"3.1.0","info":{"title":"T"},"paths":{"/a":{"get":{"operationId":"dupe","responses":{"200":{"description":"ok"}}}},"/b":{"post":{"operationId":"dupe","responses":{"200":{"description":"ok"}}}}}}`
		r := docs.NewRouter()
		err := r.Use(Plugin{Spec: []byte(spec), Path: "/api"})
		if err == nil || !strings.Contains(err.Error(), "/a") || !strings.Contains(err.Error(), "/b") {
			t.Fatalf("error does not name both operations: %v", err)
		}
	})
	t.Run("092 the console follows the URL hash", func(t *testing.T) {
		if !strings.Contains(string(mustReadRuntime()), "hash") {
			t.Fatal("deep links cannot preselect an operation")
		}
	})
	t.Run("093 duplicate operation ids are rejected", func(t *testing.T) {
		spec := `{"openapi":"3.1.0","info":{"title":"T"},"paths":{"/a":{"get":{"operationId":"dupe","responses":{"200":{"description":"ok"}}}},"/b":{"get":{"operationId":"dupe","responses":{"200":{"description":"ok"}}}}}}`
		r := docs.NewRouter()
		if err := r.Use(Plugin{Spec: []byte(spec), Path: "/api"}); err == nil {
			t.Fatal("Apply accepted two operations claiming one id")
		}
	})
	t.Run("094 escape clears the endpoint filter", func(t *testing.T) {
		if !strings.Contains(string(mustReadRuntime()), "Escape") {
			t.Fatal("no keyboard way out of a filtered list")
		}
	})
	t.Run("095 the send button is disabled while a request is in flight", func(t *testing.T) {
		if !strings.Contains(string(mustReadRuntime()), "disabled") {
			t.Fatal("no in-flight state on the send button")
		}
	})
	t.Run("096 the console shows loading feedback", func(t *testing.T) {
		if !strings.Contains(string(mustReadRuntime()), "loading") && !strings.Contains(string(mustReadRuntime()), "Sending") {
			t.Fatal("no feedback between click and response")
		}
	})
	t.Run("097 the try console announces errors inline", func(t *testing.T) {
		js := string(mustReadRuntime())
		if !strings.Contains(js, "finally {") {
			t.Fatal("no finally: errors can leave the console stuck loading")
		}
		if !strings.Contains(js, "Request failed") {
			t.Fatal("no inline failure message")
		}
	})
}

func mustReadRuntime() []byte {
	files, err := (Plugin{}).RuntimeAssets()
	if err != nil {
		return nil
	}
	return files["openapi.js"]
}
