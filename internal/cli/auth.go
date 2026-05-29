package cli

import (
	"fmt"

	pairconfig "github.com/owainlewis/pair-cli/internal/config"
	"github.com/spf13/cobra"
)

func newAuthCommand(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Inspect PAIR CLI authentication",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show configured authentication status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := pairconfig.Resolve(pairconfig.Overrides{
				BaseURL: opts.BaseURL,
				Token:   opts.Token,
			})
			if err != nil {
				return err
			}

			// The base URL is not a secret, so show its value to confirm
			// which backend (prod default vs. dev override) is in effect.
			fmt.Fprintf(cmd.OutOrStdout(), "base URL: %s (%s)\n", resolved.BaseURL, resolved.BaseURLSource)
			fmt.Fprintf(cmd.OutOrStdout(), "token: %s (%s)\n", configuredLabel(resolved.Token), resolved.TokenSource)
			return nil
		},
	})

	return cmd
}

func configuredLabel(value string) string {
	if value == "" {
		return "not configured"
	}
	return "configured"
}
