package cli

import "github.com/spf13/cobra"

func newCollectionsCommand(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collections",
		Short: "Manage PAIR collections",
	}

	cmd.AddCommand(
		placeholderCommand("list", "List collections", opts),
		placeholderCommand("show <collection-id>", "Show a collection", opts),
		placeholderCommand("create", "Create a collection", opts),
		placeholderCommand("rename <collection-id> <name>", "Rename a collection", opts),
		placeholderCommand("link-doc <collection-id> <document-id>", "Link a document to a collection", opts),
		placeholderCommand("publish <collection-id>", "Create and link a document to a collection", opts),
		placeholderCommand("unlink-doc <collection-id> <document-id>", "Unlink a document from a collection", opts),
		placeholderCommand("delete <collection-id>", "Delete a collection", opts),
	)

	return cmd
}
