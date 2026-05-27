package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

// Options holds global command flags shared by the eventual API-backed commands.
type Options struct {
	BaseURL string
	Token   string
	JSON    bool
}

// NewRootCommand builds the pair CLI command tree.
func NewRootCommand() *cobra.Command {
	opts := &Options{}

	root := &cobra.Command{
		Use:           "pair",
		Short:         "Operate a PAIR agent workspace",
		Long:          "pair is a CLI for AI agents and humans to operate a PAIR workspace from a terminal.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringVar(&opts.BaseURL, "base-url", "", "PAIR API base URL")
	root.PersistentFlags().StringVar(&opts.Token, "token", "", "PAIR bearer token")
	root.PersistentFlags().BoolVar(&opts.JSON, "json", false, "print machine-readable JSON")

	root.AddCommand(
		newAuthCommand(opts),
		newConfigCommand(opts),
		newTasksCommand(opts),
		newDocsCommand(opts),
		newCollectionsCommand(opts),
	)

	return root
}

// Execute runs the CLI and writes actionable command errors to stderr.
func Execute(args []string, stdout, stderr io.Writer) int {
	cmd := NewRootCommand()
	cmd.SetArgs(args)
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	return 0
}

func placeholderCommand(use, short string, opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			printPlaceholder(cmd.OutOrStdout(), cmd.CommandPath())
			return nil
		},
	}
}

func printPlaceholder(w io.Writer, commandPath string) {
	fmt.Fprintf(w, "%s is not implemented yet.\n", commandPath)
}
