package cmd

import (
	"fmt"
	"github.com/clambin/seals/internal/inventory"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	listCmd = &cobra.Command{
		Use:   "list",
		Short: "Lists all secrets",
		RunE: func(cmd *cobra.Command, args []string) error {
			inv, err := inventory.ReadFromFile(viper.GetString("inventory"))
			if err == nil {
				for _, secret := range inv.Secrets {
					fmt.Printf("%s => %s (%s)\n", secret.Source, secret.Destination, secret.Namespace)
				}
			}
			return err
		},
	}
)
