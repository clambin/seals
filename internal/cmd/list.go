package cmd

import (
	"fmt"
	"github.com/clambin/seals/internal/inventory"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var (
	listCmd = &cobra.Command{
		Use:   "list",
		Short: "Lists all secrets",
		RunE: func(cmd *cobra.Command, args []string) error {
			inv, err := inventory.ReadFromFile(viper.GetString("inventory"))
			if err != nil {
				return fmt.Errorf("unable to load ansible inventory file: %w", err)
			}
			return inv.List(os.Stdout)
		},
	}
)
