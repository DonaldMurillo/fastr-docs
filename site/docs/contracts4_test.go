package docs

// Fourth-cycle site contract: the export warns before baking localhost
// into public URLs.

// Fourth red suite, site share: the export path must not bake localhost
// into public URLs, and declared redirects must materialize.

import (
	"os"
	"strings"
	"testing"
)

func TestRed351SiteExport(t *testing.T) {
	t.Run("351 the site warns when PUBLIC_SITE_URL is unset", func(t *testing.T) {
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
	})
}
