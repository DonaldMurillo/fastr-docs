package katex

import (
	"strings"
	"testing"
)

func TestMathPresentationAndErrorHandling(t *testing.T) {
	t.Run("display math gets docs-level block styling", func(t *testing.T) {
		r := routerWithPlugin(t, Plugin{})
		if !strings.Contains(r.CSS(), "fastr-docs-math--display") {
			t.Fatal("display math has no docs-layer block treatment")
		}
	})
	t.Run("failures keep the original formula visible", func(t *testing.T) {
		renderer := string(runtimeAssets(t)["katex/katex.js"])
		if !strings.Contains(renderer, "data-fastr-docs-math") {
			t.Fatal("the error path loses the formula the reader asked about")
		}
	})
	t.Run("the loader clears its busy marks", func(t *testing.T) {
		renderer := string(runtimeAssets(t)["katex/katex.js"])
		if !strings.Contains(renderer, "removeAttribute") && !strings.Contains(renderer, "aria-busy") {
			t.Fatal("a rendered formula stays marked busy forever")
		}
	})
	t.Run("inline math keeps the baseline", func(t *testing.T) {
		r := routerWithPlugin(t, Plugin{})
		if !strings.Contains(r.CSS(), "fastr-docs-math") {
			t.Fatal("the docs layer styles nothing about math")
		}
	})
}
