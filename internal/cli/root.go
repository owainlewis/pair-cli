package cli

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/owainlewis/pair-cli/internal/api"
	"github.com/owainlewis/pair-cli/internal/version"
	"github.com/spf13/cobra"
)

// Options holds global command flags shared by the eventual API-backed commands.
type Options struct {
	BaseURL string
	Token   string
	JSON    bool
	Stdin   io.Reader
}

// NewRootCommand builds the pair CLI command tree.
func NewRootCommand() *cobra.Command {
	return newRootCommand(os.Stdin)
}

func newRootCommand(stdin io.Reader) *cobra.Command {
	opts := &Options{Stdin: stdin}

	root := &cobra.Command{
		Use:           "pair",
		Short:         "Operate a PAIR agent workspace",
		Long:          "pair is a CLI for AI agents and humans to operate a PAIR workspace from a terminal.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version.String(),
	}
	root.SetVersionTemplate("pair {{.Version}}\n")

	root.PersistentFlags().StringVar(&opts.BaseURL, "base-url", "", "PAIR API base URL")
	root.PersistentFlags().StringVar(&opts.Token, "token", "", "PAIR bearer token")
	root.PersistentFlags().BoolVar(&opts.JSON, "json", false, "print machine-readable JSON")

	root.AddCommand(
		newVersionCommand(),
		newAuthCommand(opts),
		newConfigCommand(opts),
		newTasksCommand(opts),
		newDocsCommand(opts),
		newCollectionsCommand(opts),
	)

	return root
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the pair version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "pair %s\n", version.String())
			return nil
		},
	}
}

// Execute runs the CLI and writes actionable command errors to stderr.
func Execute(args []string, stdout, stderr io.Writer) int {
	return ExecuteWithInput(args, os.Stdin, stdout, stderr)
}

// ExecuteWithInput runs the CLI with explicit standard streams.
func ExecuteWithInput(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	cmd := newRootCommand(stdin)
	cmd.SetArgs(args)
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(stderr, err)
		var apiErr *api.APIError
		if errors.As(err, &apiErr) {
			return 2
		}
		return 1
	}

	return 0
}
