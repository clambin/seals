package main

import (
	"fmt"
	"github.com/clambin/seals/internal/cmd"
	"os"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
}
