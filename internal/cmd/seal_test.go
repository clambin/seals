package cmd

import (
	"bytes"
	ssv1alpha1 "github.com/bitnami-labs/sealed-secrets/pkg/apis/sealedsecrets/v1alpha1"
	"github.com/bitnami-labs/sealed-secrets/pkg/kubeseal"
	"github.com/clambin/seals/internal/inventory"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_seal(t *testing.T) {
	tmpdir := t.TempDir()

	v := viper.New()
	v.Set("ansible", tmpdir)
	v.Set("controller-name", "seal-secrets")
	v.Set("controller-namespace", "seal-secrets")

	const body = "test"
	require.NoError(t, os.WriteFile(filepath.Join(tmpdir, "test"), []byte(body), 0644))
	var inv inventory.Inventory
	inv.SecretsDir = "."
	inv.DestinationDir = "."
	inv.Add(inventory.Secret{Source: "test", Destination: "sealed-test", Namespace: "default"})

	var s fakeSealer
	assert.NoError(t, seal(s, inv, v, slog.Default()))

	result, err := os.ReadFile(filepath.Join(tmpdir, "sealed-test"))
	require.NoError(t, err)
	assert.Equal(t, body, string(result))
}

var _ sealer = fakeSealer{}

type fakeSealer struct{}

func (f fakeSealer) seal(w io.Writer, r io.Reader, _ string) error {
	_, err := io.Copy(w, r)
	return err
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

const testCert = `
-----BEGIN CERTIFICATE-----
MIIErTCCApWgAwIBAgIQBekz48i8NbrzIpIrLMIULTANBgkqhkiG9w0BAQsFADAA
MB4XDTE3MDYyMDA0MzI0NVoXDTI3MDYxODA0MzI0NVowADCCAiIwDQYJKoZIhvcN
AQEBBQADggIPADCCAgoCggIBAL6ISW4MnHAmC6MdmJOwo9C6YYhKYDwPD2tF+j4p
I2duB3y7DLF+zWNHgbUlBZck8CudacJTuxOJFEqr4umqm0f4EGgRPwZgFvFLHKSZ
/hxUFnMcGVhY1qsk55peSghPHarOYyBhhHDtCu7qdMu9MqPZB68y16HdPvwWPadI
dBKSxDLvwYfjDnG/ZHX9rmlDKej7jPGdvqAY5VJteP30w6YHb1Uc4whppNcDSc2l
gOuKAWtQ5WfZbB0NpMhj4framNeXMYwjZytEdC1c/4O45zm5eK4FNPueCfxOlzFQ
D3y34OuQlJwlrPE4KmdMHtE1a8x0ihbglInJrtqcXK3vEdUJ2c/BKWgFtPOTz6Du
jV4j0OMVVGnk5jUmh+yfbgielIkPcpSTWP1cIPwK3eWbrvMziq6sv0x7QoOD3Pzm
GBE8Y9sa5uy+bJZt5MywbamZ3xWaxoQbSN8RPoxRhTe0DEpx6utCXSWpapT7kWZ3
R1PTuVx+Ktyz7MRoDUWvxfpMJ2hsJ71Az0AuUZ4N4fmmGdUcM81GPUOiMZ4uqySQ
A2phgikbJaTzcT85RcNFYSi4eKc5mYFNqr5xVa6uHhZ+OGeGy1yyOEWLgIZV3A/8
4eZshOyYtRlZjCkaGZTfXNft+8QJi8rEZRcJtVhqLzezBVRsL7pt6P/mQj4+XHsE
VSBrAgMBAAGjIzAhMA4GA1UdDwEB/wQEAwIAATAPBgNVHRMBAf8EBTADAQH/MA0G
CSqGSIb3DQEBCwUAA4ICAQCSizqBB3bjHCSGk/8lpqIyHJQR5u4Cf7LRrC9U8mxe
pvC3Fx3/RlVe87Y4cUb37xZc/TmB6Bq10Y6R7ydS3oe8PCh4UQRnEfBgtJ6m59ha
t3iPX0NdQVYz/D+yEiHjpI7gpyFNuGkd4/78JE51SO4yGYvWk/ChHoMvbLcxzfdK
PI2Ymf3MWtGfoF/TQ1jy/Biy+qumDPSz23MynQG39cdUInSK26oemUbTH0koLulN
fNl4TwSEdSm2DRl0la+vkrzu7SvF9SJ2ES6wMWVjYiJLNpApjGuF9/ZOFw9DvSSH
m+UYXn+IC7rTgvXKvXTlG//z/14Lx0GFIY+ZjdENwLH//orBQLg37TZatKEpaWO6
uRzFUxZVw3ic3RxoHfEbRA9vQlQdKnV+BpZe/Pb08RAh82OZyujqqyK7cPPOW5Vi
T9y+NeMwfKH8H4un7mQWkgWFw3LMIspYY5uHWp6jBwU9u/mjoK4+Y219dkaAhAcx
D+YIZRXwxc6ehLCavGF2DIepybzDlJbiCe8JxUDsrE/Xkm6x28uq35oZ3UQznubU
7LfAeRSI99sNvFnq0TqhSlp+CUDs8Z1LvDXzAHX4UeZQl4g+H+w1KudCvjO0mPPp
R9bIjJLIvp7CQPDkdRzJSjvetrKtI0l97VjsjbRB9v6ZekGY9SFI49KzKUTk8fsF
/A==
-----END CERTIFICATE-----
`

func TestKubeSeal(t *testing.T) {
	ks := newKubeSealer("sealed-secrets", "sealed-secret")
	var err error
	ks.publicKey, err = kubeseal.ParseKey(strings.NewReader(testCert))
	assert.NoError(t, err)

	var output bytes.Buffer
	const mySecret = `
apiVersion: v1
kind: Secret
metadata:
  name: my-secret
type: Opaque
stringData:
  PASS: "1234"
`

	err = ks.seal(&output, strings.NewReader(mySecret), "my-namespace")
	assert.NoError(t, err)

	var sealedSecret ssv1alpha1.SealedSecret
	err = runtime.DecodeInto(scheme.Codecs.UniversalDecoder(), output.Bytes(), &sealedSecret)
	require.NoError(t, err)
	assert.Equal(t, "my-secret", sealedSecret.GetName())
	assert.Equal(t, "my-namespace", sealedSecret.GetNamespace())
	assert.Contains(t, sealedSecret.Spec.EncryptedData, "PASS")
	assert.NotEqual(t, "1234", sealedSecret.Spec.EncryptedData)
}
