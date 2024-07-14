package inventory_test

import (
	"bytes"
	"github.com/clambin/seals/internal/inventory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInventory_Delete(t *testing.T) {
	inv, err := inventory.Read(bytes.NewBufferString(`
secrets_dir: ../../secrets
destination_dir: ../manifests
secrets: []
`))
	require.NoError(t, err)

	assert.False(t, inv.Delete("foo.yaml"))
	inv.Add(inventory.Secret{Source: "foo.yaml", Destination: "sealed-foo.yaml", Namespace: "default"})
	assert.False(t, inv.Delete("bar.yaml"))
	assert.True(t, inv.Delete("foo.yaml"))
}
