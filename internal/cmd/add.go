package cmd

import (
	"fmt"
	"github.com/clambin/go-common/charmer"
	"github.com/clambin/seals/internal/inventory"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"log/slog"
	"os"
	"path/filepath"
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
			inv, err := inventory.ReadFromFile(viper.GetString("inventory"))
			if err != nil {
				return fmt.Errorf("unable to load ansible inventory file: %w", err)
			}
			if err = addToInventory(&inv, args[0], args[1], args[2], l); err != nil {
				return fmt.Errorf("failed to add secret: %w", err)
			}
			return inv.WriteToFile(viper.GetString("inventory"))
		},
	}
)

func addToInventory(inv *inventory.Inventory, source, destination, namespace string, l *slog.Logger) (err error) {
	secret := inventory.Secret{
		Source:      source,
		Destination: destination,
		Namespace:   namespace,
	}

	// check the target directory is writeable
	fInfo, err := os.Stat(filepath.Dir(secret.Destination))
	if err != nil || !fInfo.IsDir() {
		return fmt.Errorf("invalid directory for %s", secret.Destination)
	}

	// check that source is below SecretsDir
	if secret.Source, err = makeRelativePath(filepath.Join(viper.GetString("ansible"), inv.SecretsDir), secret.Source); err != nil {
		return fmt.Errorf("source does not appear to be below secrets subdirectory: %w", err)
	}
	if escapes(secret.Source) {
		l.Warn("secret isn't below secrets directory " + inv.SecretsDir)
	}

	// check that destination is below DestinationDir
	if secret.Destination, err = makeRelativePath(filepath.Join(viper.GetString("ansible"), inv.DestinationDir), secret.Destination); err != nil {
		return fmt.Errorf("destination does not appear to be below destinations subdirectory: %w", err)
	}
	if escapes(secret.Destination) {
		l.Warn("sealed secret isn't below manifests directory " + inv.DestinationDir)
	}

	// read the secret
	namespaceFromSecret, err := getNamespaceFromSecret(source)
	if err != nil {
		return fmt.Errorf("unable to read secret %q: %w", source, err)
	}

	// check namespace
	if namespaceFromSecret != secret.Namespace {
		l.Warn("secret namespace doesn't match command line argument", "secret", namespaceFromSecret, "namespace", secret.Namespace)
	}

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
