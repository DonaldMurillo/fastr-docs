package mermaid

// Fourth-cycle adapter contracts: frames announce themselves and their
// loading state.

// Fourth red suite: adapter and frame contracts mined from the host side.

import (
	"os"
	"strings"
	"testing"
)

func r4Adapter(t *testing.T) string {
	t.Helper()
	body, err := os.ReadFile("host/adapter.js")
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

func r4Diagram(t *testing.T, source string) string {
	t.Helper()
	return string(diagram(source, "/frame.html", "Flow"))
}

func TestRed331To338Mermaid(t *testing.T) {
	t.Run("331 frames carry accessible titles", func(t *testing.T) {
		adapter := r4Adapter(t)
		if !strings.Contains(adapter, "'title'") && !strings.Contains(adapter, "title=") {
			t.Fatal("the created iframe has no title for assistive tech")
		}
	})
	t.Run("332 diagrams mark themselves busy while loading", func(t *testing.T) {
		adapter := r4Adapter(t)
		if !strings.Contains(adapter, "aria-busy") {
			t.Fatal("a mounting diagram is silent about its loading state")
		}
	})
}
