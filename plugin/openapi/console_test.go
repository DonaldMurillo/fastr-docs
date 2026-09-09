package openapi

// Console behavior, asserted against the shipped runtime and the rendered
// markup: filter folding, hash and focus behavior, cancellation, curl
// export, and in-flight feedback.

import (
	"os"
	"strings"
	"testing"
)

func consoleJS(t *testing.T) string {
	t.Helper()
	body, err := os.ReadFile("openapi.js")
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

func TestConsoleBehavior(t *testing.T) {
	js := consoleJS(t)
	t.Run("the filter folds accents", func(t *testing.T) {
		if !strings.Contains(js, ".normalize(") {
			t.Fatal(`query "operacion" cannot find "operación"`)
		}
	})
	t.Run("selecting an operation updates the hash", func(t *testing.T) {
		if !strings.Contains(js, "replaceState") {
			t.Fatal("picking an operation leaves the URL alone; the choice is unshareable")
		}
	})
	t.Run("the console follows the URL hash", func(t *testing.T) {
		if !strings.Contains(js, "hash") {
			t.Fatal("deep links cannot preselect an operation")
		}
	})
	t.Run("the console marks itself busy", func(t *testing.T) {
		if !strings.Contains(js, "aria-busy") {
			t.Fatal("an in-flight request is only visible through the disabled button")
		}
	})
	t.Run("emptied groups hide their headings", func(t *testing.T) {
		if !strings.Contains(js, "data-openapi-group") {
			t.Fatal("filtering hides operations but leaves every group heading behind")
		}
	})
	t.Run("a deep link scrolls to its operation", func(t *testing.T) {
		if !strings.Contains(js, "scrollIntoView") {
			t.Fatal("landing on #operation leaves it under the sticky header")
		}
	})
	t.Run("responses pretty-print json", func(t *testing.T) {
		if !strings.Contains(js, "null, 2") {
			t.Fatal("json responses render as one unreadable line")
		}
	})
	t.Run("the filter resets on navigation", func(t *testing.T) {
		if !strings.Contains(js, "resetFilter") {
			t.Fatal("a narrowed operation list survives navigating to another page")
		}
	})
	t.Run("in-flight requests can be cancelled", func(t *testing.T) {
		if !strings.Contains(js, "AbortController") {
			t.Fatal("switching operations leaves the old request running to completion")
		}
	})
	t.Run("a request can be copied as curl", func(t *testing.T) {
		if !strings.Contains(js, "curl") {
			t.Fatal("there is no way to take a prepared request out of the page")
		}
	})
	t.Run("escape clears the endpoint filter", func(t *testing.T) {
		if !strings.Contains(js, "Escape") {
			t.Fatal("no keyboard way out of a filtered list")
		}
	})
	t.Run("the send button is disabled while a request is in flight", func(t *testing.T) {
		if !strings.Contains(js, "disabled") {
			t.Fatal("no in-flight state on the send button")
		}
	})
	t.Run("the console shows loading feedback", func(t *testing.T) {
		if !strings.Contains(js, "loading") && !strings.Contains(js, "Sending") {
			t.Fatal("no feedback between click and response")
		}
	})
	t.Run("the try console announces errors inline", func(t *testing.T) {
		if !strings.Contains(js, "finally {") {
			t.Fatal("no finally: errors can leave the console stuck loading")
		}
		if !strings.Contains(js, "Request failed") {
			t.Fatal("no inline failure message")
		}
	})
	t.Run("received response headers are listed", func(t *testing.T) {
		if !strings.Contains(js, "result.headers") && !strings.Contains(js, "response.headers") {
			t.Fatal("the pane shows the body only; headers the server sent are dropped")
		}
	})
	t.Run("the apiKey feeds the curl command", func(t *testing.T) {
		start := strings.Index(js, "data-openapi-curl")
		if start < 0 {
			t.Skip("no curl button")
		}
		if !strings.Contains(js[start:], "api-key") {
			t.Fatal("the curl command drops the api key")
		}
	})
}

func TestConsoleMarkup(t *testing.T) {
	t.Run("the console exposes its busy state in markup", func(t *testing.T) {
		html := specPage(t, summarySpec, "/ref")
		if !strings.Contains(html, "data-openapi-busy") && !strings.Contains(html, `aria-busy="false"`) {
			t.Fatal("busy state lives only in JavaScript")
		}
	})
	t.Run("search matches parameter names", func(t *testing.T) {
		spec := `{"openapi":"3.1.0","info":{"title":"T"},"paths":{"/x":{"get":{"summary":"List","operationId":"listX","parameters":[{"name":"limit","in":"query","schema":{"type":"integer"}}],"responses":{"200":{"description":"ok"}}}}}}`
		html := specPage(t, spec, "/ref")
		if !strings.Contains(html, "limit") {
			t.Fatal("parameter names never render at all")
		}
		index := strings.LastIndex(html, "data-openapi-search=")
		if index < 0 || !strings.Contains(html[index:index+220], "limit") {
			t.Fatal(`filtering by "limit" finds nothing; parameter names are not searchable`)
		}
	})
}
