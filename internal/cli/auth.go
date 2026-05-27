package cli

import "github.com/spf13/cobra"

func newAuthCommand(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Inspect PAIR CLI authentication",
	}

	cmd.AddCommand(placeholderCommand("status", "Show configured authentication status", opts))

	return cmd
}
