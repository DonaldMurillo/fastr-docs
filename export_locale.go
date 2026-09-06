package docs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// htmlLangAttribute matches the opening <html> tag's lang attribute, which is
// the only one that should ever be rewritten: an <html> string inside page prose
// is content.
var htmlLangAttribute = regexp.MustCompile(`(?i)^(\s*(?:<!doctype[^>]*>\s*)?<html\b[^>]*?\blang=")([^"]*)(")`)

// WriteExportLocales stamps each exported page with the language it is written
// in.
//
// GoFastr takes the document language from one host-wide value, so every page
// of a multilingual site is exported as the same language. Two things break as
// a result, and neither is visible until someone looks:
//
// Pagefind decides which language index a page belongs to by reading
// <html lang>. Told every page is English, it builds a single English index and
// stems Spanish text with English rules, so searching in Spanish barely works
// and the per-language chunking never happens.
//
// A screen reader reads the whole site with English pronunciation rules, which
// is a WCAG 3.1.1 failure on every translated page.
//
// This rewrites the attribute per route after the export, before Pagefind runs.
// It cannot fix the live server; that needs a per-page language in GoFastr.
func (r *Router) WriteExportLocales(dir string) error {
	if r == nil {
		return errors.New("docs: WriteExportLocales requires a Router")
	}
	if strings.TrimSpace(dir) == "" {
		return errors.New("docs: WriteExportLocales requires an output directory")
	}
	for _, route := range r.PublishedRoutes() {
		locale := strings.TrimSpace(r.effectiveLocale(route))
		if locale == "" {
			continue
		}
		file := exportedPagePath(dir, route.Path)
		body, err := os.ReadFile(file)
		if err != nil {
			// A route with no exported page is not an error here. Screens,
			// redirects and plugin routes may not produce one.
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("docs: read %s: %w", file, err)
		}
		rewritten := htmlLangAttribute.ReplaceAll(body, []byte("${1}"+locale+"${3}"))
		if string(rewritten) == string(body) {
			continue
		}
		if err := os.WriteFile(file, rewritten, 0o644); err != nil {
			return fmt.Errorf("docs: write %s: %w", file, err)
		}
	}
	return nil
}

// exportedPagePath maps a route to the file the static export wrote for it.
//
// The export base is not part of this. A base path rewrites the URLs inside the
// pages; the files still land at the route's own path under the output
// directory.
func exportedPagePath(dir, routePath string) string {
	rel := strings.Trim(normalizePath(routePath), "/")
	if rel == "" {
		return filepath.Join(dir, "index.html")
	}
	return filepath.Join(dir, filepath.FromSlash(rel), "index.html")
}
