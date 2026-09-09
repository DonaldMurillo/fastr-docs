package mermaid

import (
	"os"
	"strings"
	"testing"

	docs "github.com/DonaldMurillo/fastr-docs"
	"github.com/DonaldMurillo/gofastr/core/render"
)

// diagramHTML renders the placeholder markup the host adapter replaces with a
// sandboxed iframe, using the default mounted frame path.
func diagramHTML(source string) string {
	return string(diagram(source, "/__fastr-docs/mermaid/frame.html", "Flow"))
}

// adapterSource returns the host-side script that mounts diagram frames.
func adapterSource(t *testing.T) string {
	t.Helper()
	body, err := os.ReadFile("host/adapter.js")
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

func TestDiagramAnnouncesItselfAsAnImage(t *testing.T) {
	html := diagramHTML("graph TD\nA-->B")
	if !strings.Contains(html, `role="img"`) {
		t.Fatal("screen readers hear the raw graph source instead of an image")
	}
}

func TestDiagramFramesCarryAccessibleTitles(t *testing.T) {
	adapter := adapterSource(t)
	if !strings.Contains(adapter, "'title'") && !strings.Contains(adapter, "title=") {
		t.Fatal("the created iframe has no title for assistive tech")
	}
}

func TestDiagramsMarkThemselvesBusyWhileLoading(t *testing.T) {
	adapter := adapterSource(t)
	if !strings.Contains(adapter, "aria-busy") {
		t.Fatal("a mounting diagram is silent about its loading state")
	}
}

var _ render.HTML
var _ docs.Plugin = Plugin{}
