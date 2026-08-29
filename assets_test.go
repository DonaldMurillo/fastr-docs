package docs

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	gofastrRouter "github.com/DonaldMurillo/gofastr/core/router"
)

func TestMountAssetsServesFSWithSafePrefix(t *testing.T) {
	router := NewRouter()
	httpRouter := gofastrRouter.New()
	assets := fstest.MapFS{"favicon.svg": &fstest.MapFile{Data: []byte("<svg></svg>")}}
	if err := router.MountAssets(httpRouter, AssetConfig{FS: assets, Prefix: "/assets"}); err != nil {
		t.Fatalf("MountAssets() error = %v", err)
	}
	request := httptest.NewRequest(http.MethodGet, "/assets/favicon.svg", nil)
	response := httptest.NewRecorder()
	httpRouter.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "<svg") {
		t.Fatalf("asset response = %d %q", response.Code, response.Body.String())
	}
	traversal := httptest.NewRecorder()
	httpRouter.ServeHTTP(traversal, httptest.NewRequest(http.MethodGet, "/assets/../secret", nil))
	if traversal.Code == http.StatusOK {
		t.Fatal("asset handler accepted a traversal path")
	}
}
