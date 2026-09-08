package cli

// CLI contract from the third audit cycle: init validates its inputs.

import (
	"strings"
	"testing"
)

func TestRed289CLI(t *testing.T) {
	t.Run("init rejects names with path separators", func(t *testing.T) {
		var out, errOut strings.Builder
		err := Run([]string{"init", "--name", "a/b", t.TempDir()}, &out, &errOut)
		if err == nil {
			t.Fatal("a name carrying a path separator writes into the filesystem")
		}
	})
}
