package cmd

import (
	"fmt"
	"github.com/clambin/go-common/charmer"
	"github.com/clambin/seals/internal/inventory"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"log/slog"
	"os"
	"os/exec"
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
			return seal(inv, viper.GetViper(), charmer.GetLogger(cmd))
		},
	}
)

func seal(inv inventory.Inventory, v *viper.Viper, l *slog.Logger) error {
	for _, secret := range inv.Secrets {
		if err := maybeSeal(inv, secret, v, l.With("secret", secret.Source)); err != nil {
			return fmt.Errorf("failed to seal %q: %w", secret.Source, err)
		}
	}
	return nil
}

func maybeSeal(inv inventory.Inventory, secret inventory.Secret, v *viper.Viper, l *slog.Logger) error {
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

	err = sealer(
		fOut,
		fIn,
		secret.Namespace,
		v.GetString("controller-name"),
		v.GetString("controller-namespace"),
	)
	l.Debug("kubeseal result", "err", err)
	return err
}

// allow to override kubeseal during unit testing
var sealer = doSeal

func doSeal(w io.Writer, r io.Reader, namespace, controllerName, controllerNamespace string) error {
	cmd := exec.Command("/usr/local/bin/kubeseal",
		"--controller-name", controllerName,
		"--controller-namespace", controllerNamespace,
		"--namespace", namespace,
		"-o", "yaml",
	)
	cmd.Stdin = r
	cmd.Stdout = w
	return cmd.Run()
}
