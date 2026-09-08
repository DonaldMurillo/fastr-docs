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
