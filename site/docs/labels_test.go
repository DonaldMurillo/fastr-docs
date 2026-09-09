package docs

// The Spanish label set keeps pace with the framework's inventory; the
// drawer select's help line is the current gap this test pins.

import (
	"os"
	"strings"
	"testing"
)

func TestSpanishLabelCoverage(t *testing.T) {
	source, err := os.ReadFile("router.go")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(source), "SectionHelp") {
		t.Fatal("the es label set never learned the drawer select's help line")
	}
}

func TestSkipLinksAreLocalized(t *testing.T) {
	html := renderSitePage(t, "/es/docs/getting-started")
	if !strings.Contains(html, "Saltar") {
		t.Fatal("the skip link stays English on Spanish pages")
	}
}

func TestSpanishReferenceLabelsCoverTheSurface(t *testing.T) {
	t.Run("the Spanish reference labels cover the new surface", func(t *testing.T) {
		source, err := os.ReadFile("router.go")
		if err != nil {
			t.Fatal(err)
		}
		text := string(source)
		for _, label := range []string{"ServerLabel", "TokenPlaceholder", "DeprecatedLabel", "ResponseHeadersLabel", "CopyCurlLabel"} {
			if !strings.Contains(text, label) {
				t.Fatalf("the es reference mount never learned %s", label)
			}
		}
	})
	t.Run("the Spanish anchor label is declared", func(t *testing.T) {
		source, err := os.ReadFile("router.go")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(source), "AnchorLabel") {
			t.Fatal("the es label set has no translation for the heading anchor button")
		}
	})
}
