package docs

// Site contracts from the third audit cycle: the Spanish reference
// labels keep pace with the new surface.

import (
	"os"
	"strings"
	"testing"
)

func TestRed291To292Site(t *testing.T) {
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
