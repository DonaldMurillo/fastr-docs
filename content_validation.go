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
		issues = append(issues, imageIssues(route, source, r)...)
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
			if target == nil && r.redirectClaimed(resolved) {
				// An old path some route claims as its redirect is not a
				// broken link; the host answers it with a 301.
				continue
			}
			if target == nil || !r.routePublished(target) {
				base.Message = fmt.Sprintf("internal link resolves to unpublished route %q", resolved)
				issues = append(issues, base)
				continue
			}
			if fragment == "" || target.page == nil || target.page.Body != nil {
				continue
			}
			if !markdownAnchorIDs(pageSource(target.page))[foldRunes(fragment)] {
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
		if target == route {
			issue.Message = "translation_of points at the page itself"
			issues = append(issues, issue)
			continue
		}
		if target != nil && translationCycle(r, route, target) {
			issue.Message = fmt.Sprintf("translation_of cycle: %q and its target point at each other", route.Path)
			issues = append(issues, issue)
			continue
		}
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

// stripCodeSpans blanks inline code with same-length spaces, so links that
// are documented as syntax inside backticks are not checked while column
// positions stay honest.
func stripCodeSpans(line string) string {
	bytes := []byte(line)
	for i := 0; i < len(bytes); i++ {
		if bytes[i] != '`' {
			continue
		}
		run := 1
		for i+run < len(bytes) && bytes[i+run] == '`' {
			run++
		}
		close := strings.Index(string(bytes[i+run:]), strings.Repeat("`", run))
		if close < 0 {
			return string(bytes)
		}
		for k := i; k < i+run+close+run && k < len(bytes); k++ {
			bytes[k] = ' '
		}
		i = i + run + close + run - 1
	}
	return string(bytes)
}

// markdownReferenceTargets collects reference-style link definitions
// ([ref]: /target), which inline scanning never sees.
var markdownReferenceTargets = regexp.MustCompile(`(?m)^\s{0,3}\[([^\]]+)\]:\s*(\S+)`)

var markdownLinkPattern = regexp.MustCompile(`(?:^|[^!])\[[^\]\r\n]*\]\(\s*(?:<([^>\r\n]+)>|([^\s)]+))`)

func markdownLinks(source string) []markdownLink {
	source = strings.ReplaceAll(source, "\r\n", "\n")
	// Reference definitions behave like links for checking purposes.
	var refs []markdownLink
	for _, match := range markdownReferenceTargets.FindAllStringSubmatch(source, -1) {
		refs = append(refs, markdownLink{Target: match[2], Line: 0, Column: 0})
	}
	source = markdownReferenceTargets.ReplaceAllString(source, "")
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
			scannable := stripCodeSpans(line)
			for _, match := range markdownLinkPattern.FindAllStringSubmatchIndex(scannable, -1) {
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
	return append(links, refs...)
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

// builtinShortcodeProps names the props each built-in shortcode accepts, so
// a typo'd prop is caught at build time instead of doing nothing.
var builtinShortcodeProps = map[string]map[string]bool{
	"callout":  {"variant": true, "title": true, "icon": true},
	"card":     {"variant": true, "title": true, "href": true, "description": true},
	"steps":    {"variant": true},
	"tabs":     {},
	"tab":      {"label": true},
	"filetree": {},
	"diff":     {"file": true, "lang": true, "left": true, "right": true},
	"math":     {"display": true},
	"mermaid":  {"title": true},
}

// shortcodeIssues checks the built-in shortcodes' prop values: variant
// values the sanitizer would quietly replace, tab labels repeated within
// one tabs block, props no built-in knows, and block shortcodes inside
// headings.
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
	issues = append(issues, shortcodePropNameIssues(route, source)...)
	issues = append(issues, shortcodeInHeadingIssues(route, source)...)
	return issues
}

var shortcodeOpenPattern = regexp.MustCompile(`\{\{<\s*([A-Za-z][A-Za-z0-9_-]*)((?:[^>]*)?)>\}\}`)
var shortcodePropScan = regexp.MustCompile(`([A-Za-z][A-Za-z0-9_-]*)\s*=`)

func shortcodePropNameIssues(route *Route, source string) []ContentIssue {
	var issues []ContentIssue
	for _, match := range shortcodeOpenPattern.FindAllStringSubmatch(source, -1) {
		name, attrs := match[1], match[2]
		allowed, known := builtinShortcodeProps[name]
		if !known {
			continue
		}
		for _, prop := range shortcodePropScan.FindAllStringSubmatch(attrs, -1) {
			// Prop names fold case at parse time, so the check compares
			// the folded name.
			if !allowed[strings.ToLower(prop[1])] {
				issues = append(issues, ContentIssue{RoutePath: route.Path, SourcePath: route.page.SourcePath,
					Message: fmt.Sprintf("shortcode %q has no prop %q", name, prop[1])})
			}
		}
	}
	return issues
}

var headingWithShortcodePattern = regexp.MustCompile(`(?m)^#{1,6}.*\{\{<`)

func shortcodeInHeadingIssues(route *Route, source string) []ContentIssue {
	var issues []ContentIssue
	for _, line := range strings.Split(strings.ReplaceAll(source, "\r\n", "\n"), "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "#") || !strings.Contains(trimmed, "{{<") {
			continue
		}
		issues = append(issues, ContentIssue{RoutePath: route.Path, SourcePath: route.page.SourcePath,
			Message: fmt.Sprintf("a heading carries a block shortcode; move it below the heading so the section ids stay stable")})
	}
	return issues
}

// markdownImageRef is one image in a page: its alt text and target.
type markdownImageRef struct {
	Alt, Target string
	Line        int
}

var markdownImagePattern = regexp.MustCompile(`!\[([^\]]*)\]\(\s*([^)\s]+)\s*\)`)

func markdownImages(source string) []markdownImageRef {
	var images []markdownImageRef
	line := 1
	for _, char := range source {
		if char == '\n' {
			line++
		}
	}
	_ = line
	offset := 0
	for _, match := range markdownImagePattern.FindAllStringSubmatchIndex(source, -1) {
		_ = offset
		images = append(images, markdownImageRef{Alt: source[match[2]:match[3]], Target: source[match[4]:match[5]]})
	}
	return images
}

// imageIssues flags images with no alt text and internal image sources no
// route serves.
func imageIssues(route *Route, source string, r *Router) []ContentIssue {
	var issues []ContentIssue
	for _, image := range markdownImages(source) {
		if strings.TrimSpace(image.Alt) == "" {
			issues = append(issues, ContentIssue{RoutePath: route.Path, SourcePath: route.page.SourcePath,
				Link:    image.Target,
				Message: fmt.Sprintf("image %q has no alt text; describe it or mark it decorative with empty brackets and a reason", image.Target)})
			continue
		}
		if strings.HasPrefix(image.Target, "http://") || strings.HasPrefix(image.Target, "https://") || strings.HasPrefix(image.Target, "/__") {
			continue
		}
		resolved, _, kind, err := resolveContentLink(route.Path, image.Target)
		if err != nil || kind == contentLinkExternal {
			continue
		}
		if target := r.routes[resolved]; target == nil || !r.routePublished(target) {
			issues = append(issues, ContentIssue{RoutePath: route.Path, SourcePath: route.page.SourcePath,
				Link:    image.Target,
				Message: fmt.Sprintf("image resolves to unpublished route %q", resolved)})
		}
	}
	return issues
}

// redirectClaimed reports whether any registered route lists path as one of
// its redirect sources.
func (r *Router) redirectClaimed(path string) bool {
	for _, route := range r.routes {
		for _, redirect := range route.Metadata.Redirects {
			if safeRedirectPath(redirect) == normalizePath(path) {
				return true
			}
		}
	}
	return false
}

// translationCycle reports whether following TranslationOf from target ever
// returns to start.
func translationCycle(r *Router, start, target *Route) bool {
	seen := map[*Route]bool{start: true}
	current := target
	for current != nil && current.Metadata.TranslationOf != "" {
		if current == start {
			return true
		}
		if seen[current] {
			return false
		}
		seen[current] = true
		next := r.routes[normalizePath(current.Metadata.TranslationOf)]
		if next == nil {
			return false
		}
		if next == start {
			return true
		}
		current = next
	}
	return false
}
