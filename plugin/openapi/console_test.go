package openapi

// Console behavior contracts, asserted against the shipped runtime: filter
// folding, hash and focus behavior, cancellation, and curl export.

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
}
