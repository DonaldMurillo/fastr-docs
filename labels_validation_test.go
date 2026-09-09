package docs

// Labels and locale inventories: nested Keys, month calendars, plural
// forms, hreflang defaults, RTL coverage, and per-route label resolution.
import (
	"reflect"
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
		r := bilingualPair()
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

	t.Run("count labels pluralize", func(t *testing.T) {
		if got := formatCount(defaultUIStrings.Blog.PostCount, 1); got != "1 post" {
			t.Fatalf("singular renders as %q", got)
		}
		if got := formatCount(defaultUIStrings.Blog.PostCount, 5); got != "5 posts" {
			t.Fatalf("plural renders as %q", got)
		}
		if got := formatCount("%d posts", 2); got != "2 posts" {
			t.Fatalf("plain labels keep working: %q", got)
		}
	})

	t.Run("Keys is sorted", func(t *testing.T) {
		keys := UIStrings{}.Keys()
		for i := 1; i < len(keys); i++ {
			if keys[i-1] > keys[i] {
				t.Fatalf("Keys() is not sorted around %q", keys[i])
			}
		}
	})
}

func TestMonthCalendarsAreValidated(t *testing.T) {
	t.Run("a short month list is flagged", func(t *testing.T) {
		r := NewRouter(WithLocaleUIStrings("es", UIStrings{Months: []string{"enero", "febrero"}}))
		r.MustPage("/es/post", PageConfig{Title: "P", Description: "d", Source: "# P", Order: 1,
			Metadata: ContentMetadata{Locale: "es", DatePublished: "2026-12-01"}})
		if err := r.Validate(); err == nil {
			t.Fatal("Validate accepted a two-month calendar")
		}
	})
	t.Run("a short short-months list is flagged", func(t *testing.T) {
		r := NewRouter(WithLocaleUIStrings("es", UIStrings{ShortMonths: []string{"ene"}}))
		r.MustPage("/es/p", PageConfig{Title: "P", Description: "d", Source: "# P", Order: 1, Metadata: ContentMetadata{Locale: "es"}})
		if err := r.Validate(); err == nil {
			t.Fatal("Validate accepted a one-entry ShortMonths")
		}
	})
	t.Run("month names are validated for uniqueness", func(t *testing.T) {
		r := NewRouter(WithLocaleUIStrings("es", UIStrings{Months: []string{"x", "x", "x", "x", "x", "x", "x", "x", "x", "x", "x", "x"}}))
		r.MustPage("/es/p", PageConfig{Title: "P", Description: "d", Source: "# P", Order: 1, Metadata: ContentMetadata{Locale: "es"}})
		if err := r.Validate(); err == nil {
			t.Fatal("Validate accepted twelve identical month names")
		}
	})
}

func TestLocaleLabelsResolvePerRoute(t *testing.T) {
	t.Run("locale labels merge router-wide, locale, then default", func(t *testing.T) {
		r := NewRouter(
			WithUIStrings(UIStrings{Sections: "Router sections"}),
			WithLocaleUIStrings("es", UIStrings{Sections: "Secciones"}),
		)
		r.MustPage("/es/p", PageConfig{Title: "P", Description: "d", Source: "# P", Order: 1, Metadata: ContentMetadata{Locale: "es"}})
		if got := r.uiAt("/es/p").Sections; got != "Secciones" {
			t.Fatalf("locale label = %q", got)
		}
		if got := r.UIStringsForLocale("fr").Sections; got != "Router sections" {
			t.Fatalf("unknown locale label = %q, want the router-wide value", got)
		}
	})
	t.Run("region label sets fall back to the primary locale", func(t *testing.T) {
		r := NewRouter(WithLocaleUIStrings("es", UIStrings{OnThisPage: "En esta página"}))
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G"})
		r.MustPage("/es-MX/g", PageConfig{Title: "G", Description: "d", Order: 2, Source: "# G", Metadata: ContentMetadata{Locale: "es-MX"}})
		if got := r.UIStringsForLocale("es-mx").OnThisPage; got != "En esta página" {
			t.Fatalf("region labels = %q, want the primary locale's set", got)
		}
	})
	t.Run("uiAt accepts query strings", func(t *testing.T) {
		r := bilingualDocsSite(t)
		if got := r.uiAt("/es/docs/guide?x=1").Contents; got != "Contenido" {
			t.Fatalf("uiAt with query = %q", got)
		}
	})
	t.Run("partial blog translations are reported", func(t *testing.T) {
		r := NewRouter(WithLocaleUIStrings("es", UIStrings{Blog: BlogStrings{Title: "Blog"}}))
		r.MustPage("/es/p", PageConfig{Title: "P", Description: "d", Source: "# P", Order: 1, Metadata: ContentMetadata{Locale: "es"}})
		for _, warning := range r.Warnings() {
			if strings.Contains(warning, "blog labels") {
				return
			}
		}
		t.Fatal("a one-field Blog translation raises no warning")
	})
	t.Run("partial chrome translations warn", func(t *testing.T) {
		r := NewRouter(WithLocaleUIStrings("es", UIStrings{Home: "Inicio"}))
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G"})
		for _, warning := range r.Warnings() {
			if strings.Contains(warning, "chrome") || strings.Contains(warning, "labels") {
				return
			}
		}
		t.Fatal("a locale translating one chrome label of seventy passes silently")
	})
	t.Run("the heading anchor label is translatable", func(t *testing.T) {
		labels := UIStrings{}
		field := reflect.ValueOf(&labels).Elem().FieldByName("AnchorLabel")
		if !field.IsValid() || !field.CanSet() {
			t.Fatal("UIStrings has no settable AnchorLabel; the anchor button is English on every page")
		}
		field.SetString("Copiar enlace a esta sección")
		r := NewRouter(WithLocaleUIStrings("es", labels))
		r.MustPage("/es/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "## Sección\n\nTexto.", Metadata: ContentMetadata{Locale: "es"}})
		html := renderRouterPage(t, r, "/es/g")
		if !strings.Contains(html, "Copiar enlace") {
			t.Fatal("the anchor button ignores the locale's label")
		}
	})
}
