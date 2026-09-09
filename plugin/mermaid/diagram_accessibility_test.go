package mermaid

import (
	"strings"
	"testing"
)

// diagramHTML renders the placeholder markup the host adapter replaces with a
// sandboxed iframe, using the default mounted frame path.
func diagramHTML(source string) string {
	return string(diagram(source, "/__fastr-docs/mermaid/frame.html", "Flow"))
}

func TestDiagramAnnouncesItselfAsAnImage(t *testing.T) {
	html := diagramHTML("graph TD\nA-->B")
	if !strings.Contains(html, `role="img"`) {
		t.Fatal("screen readers hear the raw graph source instead of an image")
	}
}

func TestDiagramFramesCarryAccessibleTitles(t *testing.T) {
	adapter := string(adapterJS)
	if !strings.Contains(adapter, "'title'") && !strings.Contains(adapter, "title=") {
		t.Fatal("the created iframe has no title for assistive tech")
	}
}

func TestDiagramsMarkThemselvesBusyWhileLoading(t *testing.T) {
	adapter := string(adapterJS)
	if !strings.Contains(adapter, "aria-busy") {
		t.Fatal("a mounting diagram is silent about its loading state")
	}
}
