package docs

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRuntimeJSSyntax(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node is not installed")
	}
	path := filepath.Join(t.TempDir(), "docs.js")
	if err := os.WriteFile(path, []byte(RuntimeJS()), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(node, "--check", path)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("RuntimeJS is not valid JavaScript: %v\n%s", err, output)
	}
}
