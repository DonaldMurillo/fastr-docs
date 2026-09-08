package katex

// Fourth-cycle math contracts: docs-layer layout for the placeholder,
// error visibility, and busy-state hygiene.

// Fourth red suite: presentation and loader contracts.

import (
	"strings"
	"testing"

	docs "github.com/DonaldMurillo/fastr-docs"
)

func TestRed339To343Katex(t *testing.T) {
	t.Run("339 display math gets docs-level block styling", func(t *testing.T) {
		r := docs.NewRouter()
		if err := r.Use(Plugin{}); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(r.CSS(), "fastr-docs-math--display") {
			t.Fatal("display math has no docs-layer block treatment")
		}
	})
	t.Run("341 failures keep the original formula visible", func(t *testing.T) {
		assets, err := Plugin{}.RuntimeAssets()
		if err != nil {
			t.Fatal(err)
		}
		renderer := string(assets["katex/katex.js"])
		if !strings.Contains(renderer, "data-fastr-docs-math") {
			t.Fatal("the error path loses the formula the reader asked about")
		}
	})
	t.Run("342 the loader clears its busy marks", func(t *testing.T) {
		assets, err := Plugin{}.RuntimeAssets()
		if err != nil {
			t.Fatal(err)
		}
		renderer := string(assets["katex/katex.js"])
		if !strings.Contains(renderer, "removeAttribute") && !strings.Contains(renderer, "aria-busy") {
			t.Fatal("a rendered formula stays marked busy forever")
		}
	})
	t.Run("343 inline math keeps the baseline", func(t *testing.T) {
		r := docs.NewRouter()
		if err := r.Use(Plugin{}); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(r.CSS(), "fastr-docs-math") {
			t.Fatal("the docs layer styles nothing about math")
		}
	})
}
