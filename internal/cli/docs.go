package cli

import "github.com/spf13/cobra"

func newDocsCommand(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "docs",
		Short: "Manage PAIR documents",
	}

	cmd.AddCommand(
		placeholderCommand("list", "List documents", opts),
		placeholderCommand("show <document-id>", "Show document metadata", opts),
		placeholderCommand("read <document-id>", "Read document markdown", opts),
		placeholderCommand("create", "Create a document", opts),
		placeholderCommand("replace <document-id>", "Replace document markdown", opts),
		placeholderCommand("delete <document-id>", "Delete a document", opts),
	)

	return cmd
}
