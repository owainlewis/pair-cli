package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/owainlewis/pair-cli/internal/api"
	"github.com/owainlewis/pair-cli/internal/output"
	"github.com/spf13/cobra"
)

func newDocsCommand(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "docs",
		Short: "Manage PAIR documents",
	}

	cmd.AddCommand(
		newDocsListCommand(opts),
		newDocsShowCommand(opts),
		newDocsReadCommand(opts),
		newDocsCreateCommand(opts),
		newDocsReplaceCommand(opts),
		newDocsDeleteCommand(opts),
	)

	return cmd
}

func newDocsListCommand(opts *Options) *cobra.Command {
	var query string
	var tags []string
	var since string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List documents",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newAPIClient(opts)
			if err != nil {
				return err
			}
			docs, err := client.ListDocuments(context.Background(), api.DocumentListOptions{
				Query: query,
				Tags:  tags,
				Since: since,
			})
			if err != nil {
				return err
			}
			return writeDocuments(cmd, opts.JSON, docs)
		},
	}
	cmd.Flags().StringVar(&query, "query", "", "filter by search query")
	cmd.Flags().StringArrayVar(&tags, "tag", nil, "filter by tag")
	cmd.Flags().StringVar(&since, "since", "", "filter by relative age such as 7d or all")
	return cmd
}

func newDocsShowCommand(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "show <document-id>",
		Short: "Show document metadata",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newAPIClient(opts)
			if err != nil {
				return err
			}
			doc, err := client.ShowDocument(context.Background(), args[0])
			if err != nil {
				return err
			}
			return writeDocument(cmd, opts.JSON, doc)
		},
	}
}

func newDocsReadCommand(opts *Options) *cobra.Command {
	var outputPath string

	cmd := &cobra.Command{
		Use:   "read <document-id>",
		Short: "Read document markdown",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newAPIClient(opts)
			if err != nil {
				return err
			}
			body, err := client.ReadDocumentContent(context.Background(), args[0])
			if err != nil {
				return err
			}
			if outputPath != "" {
				return os.WriteFile(outputPath, body, 0o600)
			}
			_, err = cmd.OutOrStdout().Write(body)
			return err
		},
	}
	cmd.Flags().StringVar(&outputPath, "output", "", "write markdown to a file")
	return cmd
}

func newDocsCreateCommand(opts *Options) *cobra.Command {
	var body string
	var file string
	var tags []string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a document",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newAPIClient(opts)
			if err != nil {
				return err
			}
			markdown, err := ReadMarkdownInput(body, file, os.Stdin)
			if err != nil {
				return err
			}
			doc, err := client.CreateDocument(context.Background(), markdown, tags)
			if err != nil {
				return err
			}
			return writeDocument(cmd, opts.JSON, doc)
		},
	}
	cmd.Flags().StringVar(&body, "body", "", "markdown body")
	cmd.Flags().StringVar(&file, "file", "", "markdown file or - for stdin")
	cmd.Flags().StringArrayVar(&tags, "tag", nil, "document tag")
	return cmd
}

func newDocsReplaceCommand(opts *Options) *cobra.Command {
	var body string
	var file string

	cmd := &cobra.Command{
		Use:   "replace <document-id>",
		Short: "Replace document markdown",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newAPIClient(opts)
			if err != nil {
				return err
			}
			markdown, err := ReadMarkdownInput(body, file, os.Stdin)
			if err != nil {
				return err
			}
			if err := client.ReplaceDocumentContent(context.Background(), args[0], markdown); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "document replaced")
			return nil
		},
	}
	cmd.Flags().StringVar(&body, "body", "", "markdown body")
	cmd.Flags().StringVar(&file, "file", "", "markdown file or - for stdin")
	return cmd
}

func newDocsDeleteCommand(opts *Options) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <document-id>",
		Short: "Delete a document",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ConfirmDestructive(yes, os.Stdin, cmd.ErrOrStderr(), "Delete document "+args[0]+"?"); err != nil {
				return err
			}
			client, err := newAPIClient(opts)
			if err != nil {
				return err
			}
			if err := client.DeleteDocument(context.Background(), args[0]); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "document deleted")
			return nil
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "confirm deletion")
	return cmd
}

func writeDocuments(cmd *cobra.Command, asJSON bool, docs []api.Document) error {
	if asJSON {
		return output.WriteJSON(cmd.OutOrStdout(), docs)
	}
	rows := make([][]string, 0, len(docs))
	for _, doc := range docs {
		rows = append(rows, []string{doc.ID, doc.Title, api.TagsString(doc.Tags), doc.UpdatedAt})
	}
	return output.WriteTable(cmd.OutOrStdout(), []string{"ID", "TITLE", "TAGS", "UPDATED"}, rows)
}

func writeDocument(cmd *cobra.Command, asJSON bool, doc api.Document) error {
	if asJSON {
		return output.WriteJSON(cmd.OutOrStdout(), doc)
	}
	return output.WriteTable(cmd.OutOrStdout(), []string{"ID", "TITLE", "TAGS", "UPDATED"}, [][]string{{
		doc.ID,
		doc.Title,
		api.TagsString(doc.Tags),
		doc.UpdatedAt,
	}})
}
