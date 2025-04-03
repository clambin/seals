package cmd

import (
	"codeberg.org/clambin/go-common/charmer"
	"github.com/clambin/seals/internal/clilogger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log/slog"
	"os"
	"runtime/debug"
)

var (
	RootCmd = &cobra.Command{
		Use:   "seals",
		Short: "seals is a tool to manage the seal-secrets playbook configuration",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			level := slog.LevelInfo
			if viper.GetBool("debug") {
				level = slog.LevelDebug
			}
			charmer.SetLogger(cmd, slog.New(clilogger.NewHandler(os.Stdout, level)))
		},
	}

	commonArgs = charmer.Arguments{
		"debug":     {Default: false, Help: "log debug information"},
		"ansible":   {Default: "", Help: "ansible root directory"},
		"inventory": {Default: "", Help: "ansible secrets inventory path"},
	}
)

func init() {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		panic("could not read build info")
	}
	RootCmd.Version = buildInfo.Main.Version
	if err := charmer.SetPersistentFlags(RootCmd, viper.GetViper(), commonArgs); err != nil {
		panic("failed to set command line flags: " + err.Error())
	}
	if err := charmer.SetPersistentFlags(sealCmd, viper.GetViper(), sealArgs); err != nil {
		panic("failed to set command line flags: " + err.Error())
	}
	viper.SetEnvPrefix("SEALS")
	viper.AutomaticEnv()
	RootCmd.AddCommand(listCmd, addCmd, sealCmd)
}
