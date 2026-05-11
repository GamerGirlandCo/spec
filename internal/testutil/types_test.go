package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssertGolden(t *testing.T) {
	originalUpdate := *Update
	defer func() { *Update = originalUpdate }()

	t.Run("compares against existing golden file", func(t *testing.T) {
		*Update = false
		dir := t.TempDir()
		golden := filepath.Join(dir, "schema.yaml")
		err := os.WriteFile(golden, []byte("openapi: 3.1.2\n"), 0o600)
		require.NoError(t, err)

		AssertGolden(t, []byte("openapi: 3.1.2\n"), golden)
	})

	t.Run("creates golden file when update is enabled", func(t *testing.T) {
		*Update = true
		dir := t.TempDir()
		golden := filepath.Join(dir, "nested", "schema.yaml")
		want := []byte("openapi: 3.1.2\ninfo:\n  title: Test\n")

		AssertGolden(t, want, golden)

		got, err := os.ReadFile(golden)
		require.NoError(t, err)
		assert.Equal(t, string(want), string(got))
	})
}
