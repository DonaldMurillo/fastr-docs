package docs

// Label and locale inventory contracts: nested Keys, month
// calendars, plural forms, hreflang defaults, and RTL coverage.
import (
	"strings"
	"testing"
)

func TestLabelInventory(t *testing.T) {
	t.Run("Keys lists the nested label groups", func(t *testing.T) {
		keys := UIStrings{}.Keys()
		joined := strings.Join(keys, ",")
		if !strings.Contains(joined, "Blog.") || !strings.Contains(joined, "NotFound.") {
			t.Fatalf("Keys() omits nested groups: %v", keys)
		}
	})

	t.Run("an empty month name is flagged", func(t *testing.T) {
		months := []string{"", "febrero", "marzo", "abril", "mayo", "junio", "julio", "agosto", "septiembre", "octubre", "noviembre", "diciembre"}
		r := NewRouter(WithLocaleUIStrings("es", UIStrings{Months: months}))
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G"})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "month") {
			t.Fatalf("Validate() = %v, want an empty-month complaint", err)
		}
	})

	t.Run("month names compare trimmed", func(t *testing.T) {
		months := []string{"enero ", "enero", "marzo", "abril", "mayo", "junio", "julio", "agosto", "septiembre", "octubre", "noviembre", "diciembre"}
		r := NewRouter(WithLocaleUIStrings("es", UIStrings{Months: months}))
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G"})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "month") {
			t.Fatal("Validate() wants a duplicate-month complaint once trimmed")
		}
	})

	t.Run("pages carry an x-default hreflang", func(t *testing.T) {
		r := r2Bilingual()
		head := r.metadataHeadHTML(r.routeAtPath("/docs/guide"))
		if !strings.Contains(head, `hreflang="x-default"`) {
			t.Fatalf("head alternates lack x-default: %s", head)
		}
	})

	t.Run("more rtl locales are recognized", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/ckb/x", PageConfig{Title: "X", Description: "d", Order: 1, Source: "# X", Metadata: ContentMetadata{Locale: "ckb"}})
		if got := r.DirectionFor("/ckb/x"); got != "rtl" {
			t.Fatalf("DirectionFor(ckb) = %q, want rtl", got)
		}
	})

	t.Run("a bad-shape locale label set is flagged", func(t *testing.T) {
		r := NewRouter(WithLocaleUIStrings("e", UIStrings{Home: "x"}))
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G"})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "locale") {
			t.Fatalf("Validate() = %v, want a locale-shape complaint", err)
		}
	})

	t.Run("placeholder checks walk the nested blog labels", func(t *testing.T) {
		r := NewRouter(WithLocaleUIStrings("es", UIStrings{Blog: BlogStrings{PostCount: "entradas"}}))
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G"})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "PostCount") {
			t.Fatalf("Validate() = %v, want the dropped placeholder in Blog.PostCount named", err)
		}
	})

	t.Run("the 404 alternates carry x-default", func(t *testing.T) {
		r := NewRouter(WithLocaleUIStrings("es", UIStrings{Home: "Inicio"}))
		r.MustPage("/", PageConfig{Title: "Home", Description: "d", Order: 1, Source: "# H"})
		r.MustPage("/es", PageConfig{Title: "Inicio", Description: "d", Order: 2, Source: "# H", Metadata: ContentMetadata{Locale: "es"}})
		html := string(r.NotFoundScreen().RenderNotFound("/missing"))
		if !strings.Contains(html, `hreflang="x-default"`) {
			t.Fatalf("404 alternates lack x-default: %s", html)
		}
	})

	t.Run("counted labels accept a zero form", func(t *testing.T) {
		if got := formatCount("0 posts|1 post|%d posts", 0); got != "0 posts" {
			t.Fatalf("formatCount three-form = %q, want %q", got, "0 posts")
		}
	})
}
