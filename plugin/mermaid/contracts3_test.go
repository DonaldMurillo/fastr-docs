package mermaid

// Diagram contract from the third audit cycle: the container announces
// itself as a titled image.

import (
	"strings"
	"testing"

	docs "github.com/DonaldMurillo/fastr-docs"
	"github.com/DonaldMurillo/gofastr/core/render"
)

func r3Diagram(t *testing.T, source string) string {
	t.Helper()
	return string(diagram(source, "/__fastr-docs/mermaid/frame.html", "Flow"))
}

func TestRed288To288(t *testing.T) {
	t.Run("diagrams announce themselves as images", func(t *testing.T) {
		html := r3Diagram(t, "graph TD\nA-->B")
		if !strings.Contains(html, `role="img"`) {
			t.Fatal("screen readers hear the raw graph source instead of an image")
		}
	})
}

var _ render.HTML
var _ docs.Plugin = Plugin{}
