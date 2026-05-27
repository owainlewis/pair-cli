package cli

import "github.com/spf13/cobra"

func newTasksCommand(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tasks",
		Short: "Manage PAIR tasks",
	}

	cmd.AddCommand(
		placeholderCommand("list", "List tasks", opts),
		placeholderCommand("show <task-id>", "Show a task", opts),
		placeholderCommand("create", "Create a task", opts),
		placeholderCommand("status <task-id> <status>", "Update a task status", opts),
		placeholderCommand("comment <task-id>", "Comment on a task", opts),
		placeholderCommand("link-doc <task-id> <document-id>", "Link a document to a task", opts),
		placeholderCommand("publish <task-id>", "Create and link a document to a task", opts),
		placeholderCommand("unlink-doc <task-id> <document-id>", "Unlink a document from a task", opts),
		placeholderCommand("delete <task-id>", "Delete a task", opts),
	)

	return cmd
}
