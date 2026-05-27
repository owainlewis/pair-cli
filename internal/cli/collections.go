package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/owainlewis/pair-cli/internal/api"
	"github.com/owainlewis/pair-cli/internal/output"
	"github.com/spf13/cobra"
)

func newCollectionsCommand(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collections",
		Short: "Manage PAIR collections",
	}

	cmd.AddCommand(
		newCollectionsListCommand(opts),
		newCollectionsShowCommand(opts),
		newCollectionsCreateCommand(opts),
		newCollectionsRenameCommand(opts),
		newCollectionsLinkDocCommand(opts),
		newCollectionsPublishCommand(opts),
		newCollectionsUnlinkDocCommand(opts),
		newCollectionsDeleteCommand(opts),
	)

	return cmd
}

func newCollectionsListCommand(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List collections",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newAPIClient(opts)
			if err != nil {
				return err
			}
			collections, err := client.ListCollections(context.Background())
			if err != nil {
				return err
			}
			return writeCollections(cmd, opts.JSON, collections)
		},
	}
}

func newCollectionsShowCommand(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "show <collection-id>",
		Short: "Show a collection",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newAPIClient(opts)
			if err != nil {
				return err
			}
			collection, err := client.ShowCollection(context.Background(), args[0])
			if err != nil {
				return err
			}
			return writeCollectionDetail(cmd, opts.JSON, collection)
		},
	}
}

func newCollectionsCreateCommand(opts *Options) *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a collection",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			client, err := newAPIClient(opts)
			if err != nil {
				return err
			}
			collection, err := client.CreateCollection(context.Background(), name)
			if err != nil {
				return err
			}
			return writeCollectionDetail(cmd, opts.JSON, collection)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "collection name")
	return cmd
}

func newCollectionsRenameCommand(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "rename <collection-id> <name>",
		Short: "Rename a collection",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newAPIClient(opts)
			if err != nil {
				return err
			}
			collection, err := client.RenameCollection(context.Background(), args[0], args[1])
			if err != nil {
				return err
			}
			return writeCollectionDetail(cmd, opts.JSON, collection)
		},
	}
}

func newCollectionsLinkDocCommand(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "link-doc <collection-id> <document-id>",
		Short: "Link a document to a collection",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newAPIClient(opts)
			if err != nil {
				return err
			}
			collection, err := client.LinkCollectionDocument(context.Background(), args[0], args[1])
			if err != nil {
				return err
			}
			return writeCollectionDetail(cmd, opts.JSON, collection)
		},
	}
}

func newCollectionsPublishCommand(opts *Options) *cobra.Command {
	var body string
	var file string
	var tags []string

	cmd := &cobra.Command{
		Use:   "publish <collection-id>",
		Short: "Create and link a document to a collection",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			markdown, err := ReadMarkdownInput(body, file, os.Stdin)
			if err != nil {
				return err
			}
			client, err := newAPIClient(opts)
			if err != nil {
				return err
			}
			collection, err := client.PublishCollectionDocument(context.Background(), args[0], markdown, tags)
			if err != nil {
				return err
			}
			return writeCollectionDetail(cmd, opts.JSON, collection)
		},
	}
	cmd.Flags().StringVar(&body, "body", "", "document markdown body")
	cmd.Flags().StringVar(&file, "file", "", "document markdown file or - for stdin")
	cmd.Flags().StringArrayVar(&tags, "tag", nil, "document tag")
	return cmd
}

func newCollectionsUnlinkDocCommand(opts *Options) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "unlink-doc <collection-id> <document-id>",
		Short: "Unlink a document from a collection",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ConfirmDestructive(yes, os.Stdin, cmd.ErrOrStderr(), "Unlink document "+args[1]+" from collection "+args[0]+"?"); err != nil {
				return err
			}
			client, err := newAPIClient(opts)
			if err != nil {
				return err
			}
			if err := client.UnlinkCollectionDocument(context.Background(), args[0], args[1]); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "document unlinked")
			return nil
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "confirm unlink")
	return cmd
}

func newCollectionsDeleteCommand(opts *Options) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <collection-id>",
		Short: "Delete a collection",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ConfirmDestructive(yes, os.Stdin, cmd.ErrOrStderr(), "Delete collection "+args[0]+"?"); err != nil {
				return err
			}
			client, err := newAPIClient(opts)
			if err != nil {
				return err
			}
			if err := client.DeleteCollection(context.Background(), args[0]); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "collection deleted")
			return nil
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "confirm deletion")
	return cmd
}

func writeCollections(cmd *cobra.Command, asJSON bool, collections []api.Collection) error {
	if asJSON {
		return output.WriteJSON(cmd.OutOrStdout(), collections)
	}
	rows := make([][]string, 0, len(collections))
	for _, collection := range collections {
		rows = append(rows, []string{collection.ID, collection.Name, fmt.Sprint(collection.DocumentCount), collection.UpdatedAt})
	}
	return output.WriteTable(cmd.OutOrStdout(), []string{"ID", "NAME", "DOCS", "UPDATED"}, rows)
}

func writeCollectionDetail(cmd *cobra.Command, asJSON bool, collection api.Collection) error {
	if asJSON {
		return output.WriteJSON(cmd.OutOrStdout(), collection)
	}
	if err := output.WriteTable(cmd.OutOrStdout(), []string{"ID", "NAME", "DOCS", "UPDATED"}, [][]string{{
		collection.ID,
		collection.Name,
		fmt.Sprint(collection.DocumentCount),
		collection.UpdatedAt,
	}}); err != nil {
		return err
	}
	if len(collection.Documents) == 0 {
		return nil
	}
	fmt.Fprintln(cmd.OutOrStdout())
	rows := make([][]string, 0, len(collection.Documents))
	for _, doc := range collection.Documents {
		rows = append(rows, []string{doc.ID, doc.Title})
	}
	return output.WriteTable(cmd.OutOrStdout(), []string{"DOCUMENT", "TITLE"}, rows)
}
