package cli

import (
	"fmt"

	pairconfig "github.com/owainlewis/pair-cli/internal/config"
	"github.com/spf13/cobra"
)

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
		setConfigValueCommand("base-url <url>", "Set the PAIR API base URL", "base-url"),
		setConfigValueCommand("token <token>", "Set the PAIR bearer token", "token"),
	)

	cmd.AddCommand(set)

	return cmd
}

func setConfigValueCommand(use, short, key string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := pairconfig.DefaultPath()
			if err != nil {
				return err
			}

			cfg, err := pairconfig.Load(path)
			if err != nil {
				return err
			}

			switch key {
			case "base-url":
				cfg.BaseURL = args[0]
				if err := pairconfig.Save(path, cfg); err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), "base URL saved")
			case "token":
				cfg.Token = args[0]
				if err := pairconfig.Save(path, cfg); err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), "token saved")
			default:
				return fmt.Errorf("unknown config key %q", key)
			}

			return nil
		},
	}
}
