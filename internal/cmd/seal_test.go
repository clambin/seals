package cmd

import (
	"github.com/clambin/seals/internal/inventory"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func Test_seal(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpdir) }()

	v := viper.New()
	v.Set("ansible", tmpdir)
	v.Set("controller-name", "seal-secrets")
	v.Set("controller-namespace", "seal-secrets")

	require.NoError(t, os.WriteFile(filepath.Join(tmpdir, "test"), []byte("test"), 0644))
	var inv inventory.Inventory
	inv.SecretsDir = "."
	inv.DestinationDir = "."
	inv.Add(inventory.Secret{Source: "test", Destination: "sealed-test", Namespace: "default"})

	sealer = func(w io.Writer, r io.Reader, namespace, controllerName, controllerNamespace string) error {
		assert.Equal(t, "default", namespace)
		assert.Equal(t, "seal-secrets", controllerName)
		assert.Equal(t, "seal-secrets", controllerNamespace)
		_, err := io.Copy(w, r)
		return err
	}

	assert.NoError(t, seal(inv, v, slog.Default()))

	body, err := os.ReadFile(filepath.Join(tmpdir, "sealed-test"))
	require.NoError(t, err)
	assert.Equal(t, "test", string(body))
}
