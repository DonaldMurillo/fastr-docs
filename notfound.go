package docs

import (
	"errors"
	"fmt"
	stdhtml "html"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/DonaldMurillo/gofastr/core/render"
)

// NotFoundScreen is a small, framework-neutral 404 component for generated
// docs hosts. Mount it with uihost.WithNotFoundScreen so missing URLs keep the
// site's header, theme, and recovery link instead of falling back to a bare
// server response.
type NotFoundScreen struct {
	SiteName string
	HomeHref string
	// Strings translates the page. Empty fields keep the English defaults, so
	// a host that does not localize passes nothing.
	Strings NotFoundStrings
	// Locales translates the page per language, matched against the URL that
	// was not found.
	//
	// The 404 is built once for the whole site, so unlike every other surface
	// it has no route to read a language from. The requested path is all there
	// is, and it is enough: a miss under /es is a Spanish reader who mistyped a
	// Spanish URL, and answering in English strands them.
	Locales []LocaleNotFound
}

// LocaleNotFound is the 404 for URLs beneath one language's path prefix.
type LocaleNotFound struct {
	// Prefix is the language's path prefix, such as "/es". The longest
	// matching prefix wins.
	Prefix string
	// HomeHref is that language's home, so the recovery link does not drop the
	// reader into another language.
	HomeHref string
	Strings  NotFoundStrings
}

func (s NotFoundScreen) labels() NotFoundStrings {
	return s.labelsFor("")
}

// labelsFor resolves the labels for the URL that was not found, layering the
// language's overrides on the site-wide ones and then the English defaults.
func (s NotFoundScreen) labelsFor(path string) NotFoundStrings {
	merged := defaultUIStrings.NotFound
	overlayStrings(reflect.ValueOf(&merged).Elem(), reflect.ValueOf(s.Strings))
	if locale := s.localeFor(path); locale != nil {
		overlayStrings(reflect.ValueOf(&merged).Elem(), reflect.ValueOf(locale.Strings))
	}
	return merged
}

func (s NotFoundScreen) localeFor(path string) *LocaleNotFound {
	path = normalizePath(path)
	if path == "" {
		return nil
	}
	var best *LocaleNotFound
	for i := range s.Locales {
		prefix := normalizePath(s.Locales[i].Prefix)
		if prefix == "" || !pathActive(prefix, path) {
			continue
		}
		if best == nil || len(prefix) > len(normalizePath(best.Prefix)) {
			best = &s.Locales[i]
		}
	}
	return best
}

// Render draws the 404 without a request path; the host uses it when no
// path-specific variant is available.
func (s NotFoundScreen) Render() render.HTML {
	return s.render("")
}

// RenderNotFound adds the unmatched path as escaped diagnostic text. The
// host passes the path per request, so the component remains safe to share
// between concurrent requests.
func (s NotFoundScreen) RenderNotFound(path string) render.HTML {
	return s.render(path)
}

func (s NotFoundScreen) render(path string) render.HTML {
	home := s.HomeHref
	labels := s.labelsFor(path)
	if locale := s.localeFor(path); locale != nil && strings.TrimSpace(locale.HomeHref) != "" {
		home = locale.HomeHref
	}
	if home == "" {
		home = "/"
	}
	name := s.SiteName
	if name == "" {
		name = labels.SiteFallback
	}
	message := render.Escape(labels.Message)
	if path != "" {
		message = render.Escape(formatLabel(labels.MessageForURL, path))
	}
	return render.Raw(`<div class="fastr-docs-not-found">` +
		`<p class="fastr-docs-not-found__code">404</p>` +
		`<h1>` + render.Escape(labels.Heading) + `</h1>` +
		`<p class="fastr-docs-not-found__message">` + message + `</p>` +
		`<a class="fastr-docs-not-found__link" href="` + render.Escape(home) + `">` +
		render.Escape(formatLabel(labels.BackTo, name)) + `</a>` +
		`</div>`)
}

// WriteStaticNotFound writes the branded fallback page used by static
// deployments. A project's public/404.html can replace this file later in
// the export pipeline when it needs a fully custom fallback.
func WriteStaticNotFound(dir, basePath string, screen NotFoundScreen, css string) error {
	if strings.TrimSpace(dir) == "" {
		return fmt.Errorf("docs: static not-found directory is empty")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("docs: create static not-found directory: %w", err)
	}
	basePath = normalizeBasePath(basePath)
	if screen.HomeHref == "" {
		screen.HomeHref = basePath + "/"
	}
	if screen.SiteName == "" {
		screen.SiteName = screen.labels().SiteFallback
	}
	if strings.TrimSpace(css) == "" {
		css = `body{box-sizing:border-box;margin:0;min-height:100vh;background:#0d1411;color:#eef1ed}.fastr-docs-not-found{box-sizing:border-box;max-width:620px;margin:0 auto;padding:12vh 19px;color:#eef1ed;font:16px/1.6 system-ui,sans-serif}.fastr-docs-not-found__code{margin:0 0 12px;color:#ff8b43;font:650 11px/1.2 ui-monospace,monospace;letter-spacing:.12em}.fastr-docs-not-found h1{margin:0 0 14px;font-size:clamp(40px,7vw,68px);line-height:1;letter-spacing:-.06em}.fastr-docs-not-found__message{margin:0 0 26px;color:#9ba79f}.fastr-docs-not-found__link{display:inline-flex;padding:10px 13px;border:1px solid #35423a;border-radius:7px;color:#ff8b43;background:#19211d;font-weight:650;text-decoration:none}`
	}
	body := screen.RenderNotFound("")
	title := stdhtml.EscapeString("404: " + screen.SiteName)
	document := `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>` + title + `</title><link rel="stylesheet" href="` + stdhtml.EscapeString(basePath+"/404.css") + `"></head><body>` + string(body) + `</body></html>`
	if err := os.WriteFile(filepath.Join(dir, "404.html"), []byte(document), 0o644); err != nil {
		return fmt.Errorf("docs: write static 404.html: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "404.css"), []byte(css), 0o644); err != nil {
		return fmt.Errorf("docs: write static 404.css: %w", err)
	}
	return nil
}

var staticMetaTagPattern = regexp.MustCompile(`(?is)<meta\b[^>]*>`)
var staticHTTPEquivPattern = regexp.MustCompile(`(?i)\bhttp-equiv\s*=\s*["']([^"']+)["']`)

// RewriteStaticCSP replaces the generated Content-Security-Policy metadata
// and any hosting-header policy in an exported site. It matches the element
// by its semantic http-equiv attribute, so changes to attribute order or
// unrelated GoFastr metadata do not silently leave the default policy behind.
func RewriteStaticCSP(dir, policy string) error {
	if strings.TrimSpace(dir) == "" {
		return errors.New("docs: RewriteStaticCSP requires an output directory")
	}
	if strings.TrimSpace(policy) == "" {
		return errors.New("docs: RewriteStaticCSP requires a policy")
	}
	found := false
	if err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(path), ".html") {
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		updated := staticMetaTagPattern.ReplaceAllStringFunc(string(body), func(tag string) string {
			match := staticHTTPEquivPattern.FindStringSubmatch(tag)
			if len(match) != 2 || !strings.EqualFold(strings.TrimSpace(match[1]), "Content-Security-Policy") {
				return tag
			}
			start := strings.Index(strings.ToLower(tag), "content")
			if start < 0 {
				return tag
			}
			remainder := tag[start+len("content"):]
			eq := strings.IndexByte(remainder, '=')
			if eq < 0 {
				return tag
			}
			valueStart := start + len("content") + eq + 1
			for valueStart < len(tag) && (tag[valueStart] == ' ' || tag[valueStart] == '\t' || tag[valueStart] == '\r' || tag[valueStart] == '\n') {
				valueStart++
			}
			if valueStart >= len(tag) || (tag[valueStart] != '"' && tag[valueStart] != '\'') {
				return tag
			}
			quote := tag[valueStart]
			valueEnd := strings.IndexByte(tag[valueStart+1:], quote)
			if valueEnd < 0 {
				return tag
			}
			valueEnd += valueStart + 1
			found = true
			return tag[:valueStart+1] + stdhtml.EscapeString(policy) + tag[valueEnd:]
		})
		if updated != string(body) {
			return os.WriteFile(path, []byte(updated), 0o644)
		}
		return nil
	}); err != nil {
		return err
	}

	headerPath := filepath.Join(dir, "_headers")
	headerBody, err := os.ReadFile(headerPath)
	if err == nil {
		lines := strings.Split(string(headerBody), "\n")
		for i, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "Content-Security-Policy:") {
				lines[i] = "  Content-Security-Policy: " + policy
				found = true
			}
		}
		if err := os.WriteFile(headerPath, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	if !found {
		return errors.New("docs: exported site has no Content-Security-Policy metadata or header")
	}
	return nil
}

// NotFoundScreen builds the 404 for this Router, already translated.
//
// A project can construct NotFoundScreen itself, but it would have to restate
// what the Router already knows: which languages exist, where each one's home
// is, and what its labels are. Getting any of those wrong is invisible until
// someone mistypes a URL.
func (r *Router) NotFoundScreen() NotFoundScreen {
	screen := NotFoundScreen{SiteName: r.SiteName(), Strings: r.UIStrings().NotFound}
	if r == nil {
		return screen
	}
	for locale := range r.localeUI {
		home := r.findVariant(r.roots, nil, "", locale)
		if home == nil {
			continue
		}
		screen.Locales = append(screen.Locales, LocaleNotFound{
			Prefix:   home.Path,
			HomeHref: home.Path,
			Strings:  r.UIStringsForLocale(locale).NotFound,
		})
	}
	sort.Slice(screen.Locales, func(i, j int) bool {
		return screen.Locales[i].Prefix < screen.Locales[j].Prefix
	})
	return screen
}
