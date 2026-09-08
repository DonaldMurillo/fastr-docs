package docs

import (
	"fmt"
	"net/url"
	pathpkg "path"
	"regexp"
	"sort"
	"strings"
)

// ContentIssue is a source-level problem found in a Markdown page. The
// Router reports these during strict validation so a broken link can be fixed
// in the file that authored it instead of being discovered after deployment.
type ContentIssue struct {
	// RoutePath is the page the issue belongs to; SourcePath names the
	// file when the issue came from content.
	RoutePath  string
	SourcePath string
	// Link is the offending anchor, for issues found by link checking.
	Link string
	// Line and Column locate the issue inside the source file; zero
	// when unknown.
	Line   int
	Column int
	// Message describes the issue in one line.
	Message string
}

// Error renders the issue with its location, for Validate's aggregated error.
func (i ContentIssue) Error() string {
	location := i.RoutePath
	if i.SourcePath != "" {
		location = i.SourcePath
	}
	if i.Line > 0 {
		location = fmt.Sprintf("%s:%d", location, i.Line)
		if i.Column > 0 {
			location += fmt.Sprintf(":%d", i.Column)
		}
	}
	if i.Link != "" {
		return fmt.Sprintf("content %s: %s (%s)", location, i.Message, i.Link)
	}
	return fmt.Sprintf("content %s: %s", location, i.Message)
}

// ContentIssues checks internal Markdown links and heading fragments without
// making network requests. External HTTP links are intentionally left to a
// deployment-time checker because a documentation build must remain
// deterministic and usable offline.
func (r *Router) ContentIssues() []ContentIssue {
	if r == nil {
		return nil
	}
	var issues []ContentIssue
	issues = append(issues, r.translationIssues()...)
	issues = append(issues, r.labelIssues()...)
	for _, route := range r.Routes() {
		if route == nil || route.page == nil || route.page.Body != nil {
			continue
		}
		source := pageSource(route.page)
		issues = append(issues, shortcodeIssues(route, source)...)
		for _, link := range markdownLinks(source) {
			resolved, fragment, kind, err := resolveContentLink(route.Path, link.Target)
			base := ContentIssue{
				RoutePath: route.Path, SourcePath: route.page.SourcePath,
				Link: link.Target, Line: link.Line, Column: link.Column,
			}
			if err != nil {
				base.Message = err.Error()
				issues = append(issues, base)
				continue
			}
			if kind == contentLinkExternal {
				continue
			}
			target := r.routes[resolved]
			if target == nil || !r.routePublished(target) {
				base.Message = fmt.Sprintf("internal link resolves to unpublished route %q", resolved)
				issues = append(issues, base)
				continue
			}
			if fragment == "" || target.page == nil || target.page.Body != nil {
				continue
			}
			if !markdownAnchorIDs(pageSource(target.page))[fragment] {
				base.Message = fmt.Sprintf("heading anchor %q does not exist on %q", fragment, resolved)
				issues = append(issues, base)
			}
		}
	}
	return issues
}

// translationIssues reports every TranslationOf that cannot pair: a path
// nothing serves, the page itself, or a page in the same language, which is a
// duplicate rather than a translation. Each is written by hand and fails
// silently otherwise, with the page simply missing from the selector.
func (r *Router) translationIssues() []ContentIssue {
	var issues []ContentIssue
	paths := make([]string, 0, len(r.routes))
	for path := range r.routes {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		route := r.routes[path]
		ref := strings.TrimSpace(route.Metadata.TranslationOf)
		if ref == "" {
			continue
		}
		issue := ContentIssue{RoutePath: route.Path, Link: ref}
		if route.page != nil {
			issue.SourcePath = route.page.SourcePath
		}
		target := r.routes[normalizePath(ref)]
		switch {
		case target == nil:
			issue.Message = fmt.Sprintf("translation_of points at %q, which no route serves", ref)
		case target == route:
			issue.Message = "translation_of points at the page itself"
		case target.Metadata.Draft:
			issue.Message = fmt.Sprintf("translation_of points at %q, which is a draft; publish the original before translating it", ref)
		case r.effectiveLocale(target) == r.effectiveLocale(route):
			issue.Message = fmt.Sprintf("translation_of points at %q, which is in the same language (%q); a translation needs its own locale", ref, r.effectiveLocale(route))
		default:
			continue
		}
		issues = append(issues, issue)
	}
	return issues
}

// ValidateContent returns all Markdown link and anchor problems as one error.
// Callers that need structured diagnostics should use ContentIssues directly.
func (r *Router) ValidateContent() error {
	issues := r.ContentIssues()
	if len(issues) == 0 {
		return nil
	}
	parts := make([]string, 0, len(issues))
	for _, issue := range issues {
		parts = append(parts, issue.Error())
	}
	return fmt.Errorf("docs: content validation failed: %s", strings.Join(parts, "; "))
}

type contentLinkKind uint8

const (
	contentLinkInternal contentLinkKind = iota
	contentLinkExternal
)

type markdownLink struct {
	Target string
	Line   int
	Column int
}

var markdownLinkPattern = regexp.MustCompile(`(?:^|[^!])\[[^\]\r\n]*\]\(\s*(?:<([^>\r\n]+)>|([^\s)]+))`)

func markdownLinks(source string) []markdownLink {
	var links []markdownLink
	inFence := false
	lineStart := 0
	for _, line := range strings.SplitAfter(strings.ReplaceAll(source, "\r\n", "\n"), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inFence = !inFence
			lineStart += len(line)
			continue
		}
		if !inFence {
			for _, match := range markdownLinkPattern.FindAllStringSubmatchIndex(line, -1) {
				valueStart, valueEnd := match[2], match[3]
				if valueStart < 0 {
					valueStart, valueEnd = match[4], match[5]
				}
				if valueStart < 0 || valueEnd <= valueStart {
					continue
				}
				links = append(links, markdownLink{
					Target: line[valueStart:valueEnd],
					Line:   strings.Count(source[:lineStart], "\n") + 1,
					Column: match[0] + 1,
				})
			}
		}
		lineStart += len(line)
	}
	return links
}

func resolveContentLink(currentPath, raw string) (string, string, contentLinkKind, error) {
	raw = strings.TrimSpace(raw)
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", "", contentLinkInternal, fmt.Errorf("invalid Markdown link: %v", err)
	}
	if parsed.Scheme != "" {
		switch strings.ToLower(parsed.Scheme) {
		case "http", "https", "mailto", "tel":
			return "", "", contentLinkExternal, nil
		default:
			return "", "", contentLinkInternal, fmt.Errorf("unsupported link scheme %q", parsed.Scheme)
		}
	}
	if parsed.Host != "" || strings.HasPrefix(raw, "//") {
		return "", "", contentLinkExternal, nil
	}
	fragment, err := url.PathUnescape(parsed.Fragment)
	if err != nil {
		return "", "", contentLinkInternal, fmt.Errorf("invalid link fragment: %v", err)
	}
	if parsed.Path == "" {
		return normalizePath(currentPath), fragment, contentLinkInternal, nil
	}
	linkPath, err := url.PathUnescape(parsed.Path)
	if err != nil {
		return "", "", contentLinkInternal, fmt.Errorf("invalid link path: %v", err)
	}
	if strings.HasPrefix(linkPath, "/") {
		return normalizePath(pathpkg.Clean(linkPath)), fragment, contentLinkInternal, nil
	}
	base := pathpkg.Dir(normalizePath(currentPath))
	return normalizePath(pathpkg.Join(base, linkPath)), fragment, contentLinkInternal, nil
}

func markdownAnchorIDs(source string) map[string]bool {
	ids := make(map[string]bool)
	counts := make(map[string]int)
	inFence := false
	for _, line := range strings.Split(strings.ReplaceAll(source, "\r\n", "\n"), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inFence = !inFence
			continue
		}
		if inFence || len(trimmed) < 2 || trimmed[0] != '#' {
			continue
		}
		level := 0
		for level < len(trimmed) && trimmed[level] == '#' {
			level++
		}
		if level < 1 || level > 6 || (level < len(trimmed) && trimmed[level] != ' ') {
			continue
		}
		title := strings.TrimSpace(strings.TrimLeft(trimmed, "# "))
		id := headingSlug(title)
		if id == "" {
			continue
		}
		counts[id]++
		if counts[id] > 1 {
			id = fmt.Sprintf("%s-%d", id, counts[id])
		}
		ids[id] = true
	}
	return ids
}

// calloutVariants are the values the sanitized callout and card components
// accept; anything else reaches a fallback silently, so it is named here at
// build time instead.
var calloutVariants = map[string]bool{
	"info": true, "warning": true, "danger": true, "success": true,
	"neutral": true, "accent": true, "": true,
}

var shortcodePropValuePattern = regexp.MustCompile(`\{\{<\s*(callout|card|steps)[^>]*?variant="([^"]*)"[^>]*>`)
var shortcodeTabsBlockPattern = regexp.MustCompile(`(?s)\{\{<\s*tabs[^>]*>\}\}(.*?)\{\{<\s*/tabs\s*>\}\}`)
var shortcodeTabLabelPattern = regexp.MustCompile(`\{\{<\s*tab[^>]*?label="([^"]*)"`)

// shortcodeIssues checks the built-in shortcodes' prop values: variant
// values the sanitizer would quietly replace, and tab labels repeated
// within one tabs block.
func shortcodeIssues(route *Route, source string) []ContentIssue {
	var issues []ContentIssue
	for _, match := range shortcodePropValuePattern.FindAllStringSubmatch(source, -1) {
		name, variant := match[1], match[2]
		if !calloutVariants[variant] {
			issues = append(issues, ContentIssue{RoutePath: route.Path, SourcePath: route.page.SourcePath,
				Message: fmt.Sprintf("shortcode %q: unknown variant %q (info, warning, danger, success, neutral, accent)", name, variant)})
		}
	}
	for _, block := range shortcodeTabsBlockPattern.FindAllStringSubmatch(source, -1) {
		seen := map[string]bool{}
		for _, label := range shortcodeTabLabelPattern.FindAllStringSubmatch(block[1], -1) {
			if seen[label[1]] {
				issues = append(issues, ContentIssue{RoutePath: route.Path, SourcePath: route.page.SourcePath,
					Message: fmt.Sprintf("tabs block: two tabs share the label %q; a reader cannot tell them apart", label[1])})
			}
			seen[label[1]] = true
		}
	}
	return issues
}
