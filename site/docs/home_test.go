package docs

import (
	"strings"
	"testing"

	fastrdocs "github.com/DonaldMurillo/fastr-docs"
)

func TestHomeBodyReflectsTheRouterTree(t *testing.T) {
	router := fastrdocs.NewRouter(fastrdocs.WithSiteName("Test Docs"))
	router.MustPage("/", fastrdocs.PageConfig{Title: "Test Docs", Source: "# Home", Order: 1})
	guides := router.MustGroup("/guides", fastrdocs.GroupConfig{Title: "Guides", Order: 2})
	guides.MustPage("start", fastrdocs.PageConfig{Title: "Start here", Source: "# Start", Order: 1})

	markup := string(HomeBody(router))
	for _, want := range []string{
		">2 routes<",
		">Test Docs</b><code>/</code>",
		">Guides</b><code>/guides/*</code>",
	} {
		if !strings.Contains(markup, want) {
			t.Fatalf("HomeBody() missing %q in %s", want, markup)
		}
	}
	if strings.Contains(markup, "18 routes") {
		t.Fatal("HomeBody() retained the old hardcoded route count")
	}
}
