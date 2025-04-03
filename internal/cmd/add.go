package cmd

import (
	"codeberg.org/clambin/go-common/charmer"
	"fmt"
	"github.com/clambin/seals/internal/inventory"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

var (
	addCmd = &cobra.Command{
		Use:   "add [flags] <secret> <sealed-secret> <namespace>",
		Short: "Add a secret",
		RunE: func(cmd *cobra.Command, args []string) error {
			l := charmer.GetLogger(cmd)
			if len(args) != 3 {
				return fmt.Errorf("expected 3 arguments, got %d", len(args))
			}
			inventoryFile := viper.GetString("inventory")
			inv, err := inventory.ReadFromFile(inventoryFile)
			if err != nil {
				return fmt.Errorf("unable to load ansible inventory file: %w", err)
			}
			if err = addToInventory(&inv, args[0], args[1], args[2], viper.GetViper(), l); err != nil {
				return fmt.Errorf("failed to add secret: %w", err)
			}
			return inv.WriteToFile(inventoryFile)
		},
	}
)

func addToInventory(inv *inventory.Inventory, source, destination, namespace string, v *viper.Viper, l *slog.Logger) error {
	// check source is readable
	namespaceFromSecret, err := getNamespaceFromSecret(source)
	if err != nil {
		return fmt.Errorf("unable to read secret %q: %w", source, err)
	}
	// if source secret namespace is set and differs from namespace, give a warning
	if namespaceFromSecret != "" && namespaceFromSecret != namespace {
		l.Warn("secret namespace doesn't match command line argument. Ignoring command line argument", "secret", namespaceFromSecret, "namespace", namespace)
		namespace = namespaceFromSecret
	}

	// check destination dir is writable
	if err = isWritableDirectory(filepath.Dir(destination)); err != nil {
		return fmt.Errorf("unable to check if destination directory exists: %w", err)
	}

	// make secret with relative paths
	secret := inventory.Secret{Namespace: namespace}
	if secret.Source, err = makeRelativePath(filepath.Join(v.GetString("ansible"), inv.SecretsDir), source); err != nil {
		return fmt.Errorf("failed to make relative path: %w", err)
	}
	if secret.Destination, err = makeRelativePath(filepath.Join(v.GetString("ansible"), inv.DestinationDir), destination); err != nil {
		return fmt.Errorf("failed to make relative path: %w", err)
	}

	// if paths escape, warn
	if strings.HasPrefix(secret.Source, ".."+string(os.PathSeparator)) {
		l.Warn("secret isn't below secrets directory " + inv.SecretsDir)
	}
	if strings.HasPrefix(secret.Destination, ".."+string(os.PathSeparator)) {
		l.Warn("sealed secret isn't below manifests directory " + inv.DestinationDir)
	}

	// add the secret
	inv.Add(secret)
	return nil
}

func getNamespaceFromSecret(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	var secret struct {
		Kind     string `yaml:"kind"`
		Metadata struct {
			Namespace string `yaml:"namespace"`
		} `yaml:"metadata"`
	}
	if err = yaml.NewDecoder(f).Decode(&secret); err == nil {
		if secret.Kind != "Secret" {
			err = fmt.Errorf("secret kind in %q must be 'Secret', got %q", filename, secret.Kind)
		}
	}
	return secret.Metadata.Namespace, err
}
