package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/owainlewis/pair-cli/internal/api"
	"github.com/owainlewis/pair-cli/internal/output"
	"github.com/spf13/cobra"
)

func newTasksCommand(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tasks",
		Short: "Manage PAIR tasks",
	}

	cmd.AddCommand(
		newTasksListCommand(opts),
		newTasksShowCommand(opts),
		newTasksCreateCommand(opts),
		newTasksStatusCommand(opts),
		newTasksCommentCommand(opts),
		newTasksLinkDocCommand(opts),
		newTasksPublishCommand(opts),
		newTasksUnlinkDocCommand(opts),
		newTasksDeleteCommand(opts),
	)

	return cmd
}

func newTasksCommentCommand(opts *Options) *cobra.Command {
	var body string
	var file string

	cmd := &cobra.Command{
		Use:   "comment <task-id>",
		Short: "Comment on a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			markdown, err := ReadMarkdownInput(body, file, opts.Stdin)
			if err != nil {
				return err
			}
			if strings.TrimSpace(string(markdown)) == "" {
				return fmt.Errorf("comment body cannot be blank")
			}
			client, err := newAPIClient(opts)
			if err != nil {
				return err
			}
			comment, err := client.CommentTask(context.Background(), args[0], markdown)
			if err != nil {
				return err
			}
			if opts.JSON {
				return output.WriteJSON(cmd.OutOrStdout(), comment)
			}
			return output.WriteTable(cmd.OutOrStdout(), []string{"COMMENT", "AUTHOR", "BODY"}, [][]string{{comment.ID, comment.Author, comment.Body}})
		},
	}
	cmd.Flags().StringVar(&body, "body", "", "comment body")
	cmd.Flags().StringVar(&file, "file", "", "comment file or - for stdin")
	return cmd
}

func newTasksLinkDocCommand(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "link-doc <task-id> <document-id>",
		Short: "Link a document to a task",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newAPIClient(opts)
			if err != nil {
				return err
			}
			task, err := client.LinkTaskDocument(context.Background(), args[0], args[1])
			if err != nil {
				return err
			}
			return writeTaskDetail(cmd, opts.JSON, task)
		},
	}
}

func newTasksPublishCommand(opts *Options) *cobra.Command {
	var body string
	var file string
	var tags []string

	cmd := &cobra.Command{
		Use:   "publish <task-id>",
		Short: "Create and link a document to a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			markdown, err := ReadMarkdownInput(body, file, opts.Stdin)
			if err != nil {
				return err
			}
			client, err := newAPIClient(opts)
			if err != nil {
				return err
			}
			task, err := client.PublishTaskDocument(context.Background(), args[0], markdown, tags)
			if err != nil {
				return err
			}
			return writeTaskDetail(cmd, opts.JSON, task)
		},
	}
	cmd.Flags().StringVar(&body, "body", "", "document markdown body")
	cmd.Flags().StringVar(&file, "file", "", "document markdown file or - for stdin")
	cmd.Flags().StringArrayVar(&tags, "tag", nil, "document tag")
	return cmd
}

func newTasksUnlinkDocCommand(opts *Options) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "unlink-doc <task-id> <document-id>",
		Short: "Unlink a document from a task",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ConfirmDestructive(yes, opts.Stdin, cmd.ErrOrStderr(), "Unlink document "+args[1]+" from task "+args[0]+"?"); err != nil {
				return err
			}
			client, err := newAPIClient(opts)
			if err != nil {
				return err
			}
			if err := client.UnlinkTaskDocument(context.Background(), args[0], args[1]); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "document unlinked")
			return nil
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "confirm unlink")
	return cmd
}

func newTasksListCommand(opts *Options) *cobra.Command {
	var status string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if status != "" && !api.ValidTaskStatus(status) {
				return fmt.Errorf("invalid status %q", status)
			}
			client, err := newAPIClient(opts)
			if err != nil {
				return err
			}
			tasks, err := client.ListTasks(context.Background(), api.TaskListOptions{Status: status})
			if err != nil {
				return err
			}
			return writeTasks(cmd, opts.JSON, tasks)
		},
	}
	cmd.Flags().StringVar(&status, "status", "", "filter by status")
	return cmd
}

func newTasksShowCommand(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "show <task-id>",
		Short: "Show a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newAPIClient(opts)
			if err != nil {
				return err
			}
			task, err := client.ShowTask(context.Background(), args[0])
			if err != nil {
				return err
			}
			return writeTaskDetail(cmd, opts.JSON, task)
		},
	}
}

func newTasksCreateCommand(opts *Options) *cobra.Command {
	var title string
	var description string
	var status string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a task",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if title == "" {
				return fmt.Errorf("--title is required")
			}
			if status != "" && !api.ValidTaskStatus(status) {
				return fmt.Errorf("invalid status %q", status)
			}
			client, err := newAPIClient(opts)
			if err != nil {
				return err
			}
			task, err := client.CreateTask(context.Background(), api.TaskCreateRequest{
				Title:       title,
				Description: description,
				Status:      status,
			})
			if err != nil {
				return err
			}
			return writeTaskDetail(cmd, opts.JSON, task)
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "task title")
	cmd.Flags().StringVar(&description, "description", "", "task description")
	cmd.Flags().StringVar(&status, "status", "todo", "task status")
	return cmd
}

func newTasksStatusCommand(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "status <task-id> <status>",
		Short: "Update a task status",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !api.ValidTaskStatus(args[1]) {
				return fmt.Errorf("invalid status %q", args[1])
			}
			client, err := newAPIClient(opts)
			if err != nil {
				return err
			}
			task, err := client.UpdateTaskStatus(context.Background(), args[0], args[1])
			if err != nil {
				return err
			}
			return writeTaskDetail(cmd, opts.JSON, task)
		},
	}
}

func newTasksDeleteCommand(opts *Options) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <task-id>",
		Short: "Delete a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ConfirmDestructive(yes, opts.Stdin, cmd.ErrOrStderr(), "Delete task "+args[0]+"?"); err != nil {
				return err
			}
			client, err := newAPIClient(opts)
			if err != nil {
				return err
			}
			if err := client.DeleteTask(context.Background(), args[0]); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "task deleted")
			return nil
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "confirm deletion")
	return cmd
}

func writeTasks(cmd *cobra.Command, asJSON bool, tasks []api.Task) error {
	if asJSON {
		return output.WriteJSON(cmd.OutOrStdout(), tasks)
	}
	rows := make([][]string, 0, len(tasks))
	for _, task := range tasks {
		rows = append(rows, []string{task.ID, task.Title, task.Status, task.UpdatedAt})
	}
	return output.WriteTable(cmd.OutOrStdout(), []string{"ID", "TITLE", "STATUS", "UPDATED"}, rows)
}

func writeTaskDetail(cmd *cobra.Command, asJSON bool, task api.Task) error {
	if asJSON {
		return output.WriteJSON(cmd.OutOrStdout(), task)
	}
	if err := output.WriteTable(cmd.OutOrStdout(), []string{"ID", "TITLE", "STATUS", "UPDATED"}, [][]string{{
		task.ID,
		task.Title,
		task.Status,
		task.UpdatedAt,
	}}); err != nil {
		return err
	}
	if len(task.Comments) > 0 {
		fmt.Fprintln(cmd.OutOrStdout())
		rows := make([][]string, 0, len(task.Comments))
		for _, comment := range task.Comments {
			rows = append(rows, []string{comment.ID, comment.Author, comment.Body})
		}
		if err := output.WriteTable(cmd.OutOrStdout(), []string{"COMMENT", "AUTHOR", "BODY"}, rows); err != nil {
			return err
		}
	}
	if len(task.Documents) > 0 {
		fmt.Fprintln(cmd.OutOrStdout())
		rows := make([][]string, 0, len(task.Documents))
		for _, doc := range task.Documents {
			rows = append(rows, []string{doc.ID, doc.Title})
		}
		return output.WriteTable(cmd.OutOrStdout(), []string{"DOCUMENT", "TITLE"}, rows)
	}
	return nil
}
