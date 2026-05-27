package main

import (
	"os"

	"github.com/owainlewis/pair-cli/internal/cli"
)

func main() {
	if err := cli.NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
