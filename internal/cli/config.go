package cli

import "github.com/spf13/cobra"

func newConfigCommand(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage PAIR CLI configuration",
	}

	set := &cobra.Command{
		Use:   "set",
		Short: "Set a configuration value",
	}
	set.AddCommand(
		placeholderCommand("base-url <url>", "Set the PAIR API base URL", opts),
		placeholderCommand("token <token>", "Set the PAIR bearer token", opts),
	)

	cmd.AddCommand(set)

	return cmd
}
