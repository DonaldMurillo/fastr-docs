package docs

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
)

// WriteAgentAssets materializes GoFastr's opt-in agent discovery responses in
// a static export. GoFastr serves these responses dynamically, while its
// static builder intentionally writes declared page and asset routes. The
// helper keeps generated and hand-authored docs projects from losing /llms.txt
// and the agent card when they switch to offline hosting.
//
// A 404 is ignored so callers can use this with hosts that do not enable the
// agent-ready option. Root-relative Markdown links are prefixed with basePath
// because the response is being moved below the export mount point.
func WriteAgentAssets(dir, basePath string, handler http.Handler) error {
	if handler == nil {
		return fmt.Errorf("docs: WriteAgentAssets requires an HTTP handler")
	}
	basePath = normalizeExportBase(basePath)
	for _, route := range []string{"/llms.txt", "/.well-known/agent-card.json", "/.well-known/agent.json"} {
		req := httptest.NewRequest(http.MethodGet, route, nil)
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		if res.Code == http.StatusNotFound {
			continue
		}
		if res.Code != http.StatusOK {
			return fmt.Errorf("docs: agent asset %s returned HTTP %d", route, res.Code)
		}
		body := res.Body.Bytes()
		if basePath != "" && strings.HasSuffix(route, ".txt") {
			body = []byte(prefixRootLinks(string(body), basePath))
		}
		target := filepath.Join(dir, strings.TrimPrefix(basePath+route, "/"))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return fmt.Errorf("docs: create agent asset directory: %w", err)
		}
		if err := os.WriteFile(target, body, 0o644); err != nil {
			return fmt.Errorf("docs: write agent asset %s: %w", route, err)
		}
	}
	return nil
}

func normalizeExportBase(path string) string {
	path = strings.TrimSpace(path)
	if path == "" || path == "/" {
		return ""
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return strings.TrimRight(path, "/")
}

func prefixRootLinks(body, basePath string) string {
	return strings.ReplaceAll(body, "](/", "]("+basePath+"/")
}
