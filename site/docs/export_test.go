package docs

import (
	"os"
	"strings"
	"testing"
)

func TestSiteExportWarnsWhenPublicSiteURLIsUnset(t *testing.T) {
	body, err := os.ReadFile("../main.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if !strings.Contains(text, "PUBLIC_SITE_URL") {
		t.Skip("no public URL handling")
	}
	// A warning (or hard requirement) must exist for exports; counting
	// occurrences separates a bare default from a check.
	if strings.Count(text, "PUBLIC_SITE_URL") < 3 {
		t.Fatal("the site exports with localhost baked into its public URLs")
	}
}
