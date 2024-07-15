package cmd

import (
	"github.com/clambin/seals/internal/inventory"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func Test_addToInventory(t *testing.T) {
	tests := []struct {
		name         string
		inv          inventory.Inventory
		source       string
		destination  string
		namespace    string
		secretExists bool
		wantErr      assert.ErrorAssertionFunc
		wantSecret   inventory.Secret
	}{
		{
			name:         "success",
			inv:          inventory.Inventory{SecretsDir: "../secrets", DestinationDir: "../manifests"},
			source:       "secrets/secret.yaml",
			destination:  "manifests/sealed-secret.yaml",
			namespace:    "default",
			secretExists: true,
			wantErr:      assert.NoError,
			wantSecret:   inventory.Secret{Source: "secret.yaml", Destination: "sealed-secret.yaml", Namespace: "default"},
		},
		{
			name:         "secret missing",
			inv:          inventory.Inventory{SecretsDir: "../secrets", DestinationDir: "../manifests"},
			source:       "secrets/secret.yaml",
			destination:  "manifests/sealed-secret.yaml",
			namespace:    "default",
			secretExists: false,
			wantErr:      assert.Error,
		},
		{
			name:         "secret & destination outside of designated directory",
			inv:          inventory.Inventory{SecretsDir: "../secrets", DestinationDir: "../manifests"},
			source:       "secret.yaml",
			destination:  "sealed-secret.yaml",
			namespace:    "default",
			secretExists: true,
			wantErr:      assert.NoError,
			wantSecret:   inventory.Secret{Source: "../secret.yaml", Destination: "../sealed-secret.yaml", Namespace: "default"},
		},
		{
			name:         "invalid destination dir",
			inv:          inventory.Inventory{SecretsDir: "../secrets", DestinationDir: "../manifests"},
			source:       "secrets/secret.yaml",
			destination:  "manifests-bad/sealed-secret.yaml",
			namespace:    "default",
			secretExists: true,
			wantErr:      assert.Error,
		},
	}

	logger := slog.Default()
	tmpdir, err := initFS()
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpdir) }()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := viper.New()
			v.Set("ansible", filepath.Join(tmpdir, "ansible"))

			source := filepath.Join(tmpdir, tt.source)
			destination := filepath.Join(tmpdir, tt.destination)

			if tt.secretExists {
				require.NoError(t, os.WriteFile(source, []byte(`kind: Secret
metadata:
  namespace: `+tt.namespace+`
`), 0644))
			}

			err := addToInventory(&tt.inv, source, destination, tt.namespace, v, logger)
			tt.wantErr(t, err)

			if err == nil {
				require.Len(t, tt.inv.Secrets, 1)
				assert.Equal(t, tt.wantSecret, tt.inv.Secrets[0])
			}

			if tt.secretExists {
				require.NoError(t, os.Remove(source))
			}

		})
	}
}

func initFS() (string, error) {
	tmpdir, err := os.MkdirTemp("", "")
	if err != nil {
		return "", err
	}
	for _, directory := range []string{"secrets", "manifests", "ansible"} {
		if err = os.Mkdir(filepath.Join(tmpdir, directory), 0755); err != nil {
			return "", err
		}
	}
	return tmpdir, nil
}
