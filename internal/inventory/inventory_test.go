package inventory_test

import (
	"bytes"
	"github.com/clambin/seals/internal/inventory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInventory(t *testing.T) {
	inv, err := inventory.Read(bytes.NewBufferString(`
secrets_dir: ../../secrets
destination_dir: ../manifests
secrets: []
`))
	require.NoError(t, err)
	assert.Equal(t, "../../secrets", inv.SecretsDir)
	assert.Equal(t, "../manifests", inv.DestinationDir)

	assert.False(t, inv.Delete("foo.yaml"))
	inv.Add(inventory.Secret{Source: "foo.yaml", Destination: "sealed-foo.yaml", Namespace: "default"})

	var out bytes.Buffer
	_ = inv.List(&out)
	assert.Equal(t, "foo.yaml => sealed-foo.yaml (default)\n", out.String())

	out.Reset()
	assert.NoError(t, inv.Write(&out))
	assert.Equal(t, `secrets_dir: ../../secrets
destination_dir: ../manifests
secrets:
  - source: foo.yaml
    destination: sealed-foo.yaml
    namespace: default
`, out.String())

	assert.False(t, inv.Delete("bar.yaml"))
	assert.True(t, inv.Delete("foo.yaml"))
}
