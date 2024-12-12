package cmd

import (
	"context"
	"crypto/rsa"
	"fmt"
	"github.com/bitnami-labs/sealed-secrets/pkg/apis/sealedsecrets/v1alpha1"
	"github.com/bitnami-labs/sealed-secrets/pkg/kubeseal"
	"github.com/clambin/go-common/charmer"
	"github.com/clambin/seals/internal/inventory"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"log/slog"
	"os"
	"path/filepath"
)

var (
	sealArgs = charmer.Arguments{
		"controller-name":      {Default: "sealed-secrets", Help: "Name of sealed-secrets controller"},
		"controller-namespace": {Default: "sealed-secrets", Help: "Namespace of sealed-secrets controller"},
		"force":                {Default: false, Help: "Seal secrets even if the secret has not been updated"},
	}

	sealCmd = &cobra.Command{
		Use:   "seal",
		Short: "Seal all secrets",
		RunE: func(cmd *cobra.Command, args []string) error {
			inv, err := inventory.ReadFromFile(viper.GetString("inventory"))
			if err != nil {
				return fmt.Errorf("unable to load ansible inventory file: %w", err)
			}
			s := newKubeSealer(viper.GetString("controller-namespace"), viper.GetString("controller-namespace"))
			return seal(s, inv, viper.GetViper(), charmer.GetLogger(cmd))
		},
	}
)

// sealer interface so we can stub during unit testing
type sealer interface {
	seal(w io.Writer, r io.Reader, namespace string) error
}

func seal(s sealer, inv inventory.Inventory, v *viper.Viper, l *slog.Logger) error {
	for _, secret := range inv.Secrets {
		if err := maybeSeal(s, inv, secret, v, l.With("secret", secret.Source)); err != nil {
			return fmt.Errorf("failed to seal %q: %w", secret.Source, err)
		}
	}
	return nil
}

func maybeSeal(s sealer, inv inventory.Inventory, secret inventory.Secret, v *viper.Viper, l *slog.Logger) error {
	ansibleDir := v.GetString("ansible")

	secretFile := filepath.Join(ansibleDir, inv.SecretsDir, secret.Source)
	sealedSecretFile := filepath.Join(ansibleDir, inv.DestinationDir, secret.Destination)

	if !v.GetBool("force") {
		update, err := shouldUpdate(secretFile, sealedSecretFile)
		if err != nil {
			return err
		}
		if !update {
			l.Debug("secret is already sealed")
			return nil
		}
	}

	l.Info("sealing secret")

	var fIn *os.File
	var err error
	if fIn, err = os.Open(secretFile); err != nil {
		return fmt.Errorf("unable to open secret: %w", err)
	}
	defer func(f *os.File) { _ = f.Close() }(fIn)

	var fOut *os.File
	if fOut, err = os.OpenFile(sealedSecretFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644); err != nil {
		return fmt.Errorf("unable to open sealed secret: %w", err)
	}
	defer func(f *os.File) { _ = f.Close() }(fOut)

	err = s.seal(fOut, fIn, secret.Namespace)
	l.Debug("kubeseal result", "err", err)
	return err
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type kubeSealer struct {
	clientConfig        kubeseal.ClientConfig
	controllerNamespace string
	controllerName      string
	publicKey           *rsa.PublicKey
}

func newKubeSealer(controllerNamespace, controllerName string) *kubeSealer {
	return &kubeSealer{
		clientConfig:        initClient(),
		controllerNamespace: controllerNamespace,
		controllerName:      controllerName,
	}
}

func initClient() clientcmd.ClientConfig {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	return clientcmd.NewInteractiveDeferredLoadingClientConfig(loadingRules, nil, nil)
}

func (s *kubeSealer) getPublicKey() error {
	r, err := kubeseal.OpenCert(context.Background(), s.clientConfig, s.controllerName, s.controllerNamespace, "")
	if err == nil {
		s.publicKey, err = kubeseal.ParseKey(r)
		_ = r.Close()
	}
	return err
}

func (s *kubeSealer) seal(w io.Writer, r io.Reader, namespace string) error {
	if s.publicKey == nil {
		if err := s.getPublicKey(); err != nil {
			return err
		}
	}
	return kubeseal.Seal(s.clientConfig, "yaml", r, w, scheme.Codecs, s.publicKey, v1alpha1.DefaultScope, true, "", namespace)
}
