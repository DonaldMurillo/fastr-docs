package docs

import (
	"strings"
	"testing"
)

func TestNotFoundScreenEscapesRequestedPathAndProvidesRecoveryLink(t *testing.T) {
	html := string(NotFoundScreen{SiteName: "Manual Docs"}.RenderNotFound(`/missing"><script>alert(1)</script>`))
	for _, marker := range []string{"404", "Page not found", "Back to Manual Docs", "&lt;script&gt;"} {
		if !strings.Contains(html, marker) {
			t.Fatalf("NotFoundScreen missing %q: %s", marker, html)
		}
	}
	if strings.Contains(html, "<script>alert") {
		t.Fatalf("NotFoundScreen reflected unsafe markup: %s", html)
	}
}
