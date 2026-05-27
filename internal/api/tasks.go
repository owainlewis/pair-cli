package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// Task is a PAIR task record.
type Task struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description,omitempty"`
	Status      string        `json:"status"`
	Comments    []TaskComment `json:"comments,omitempty"`
	Documents   []Document    `json:"documents,omitempty"`
	CreatedAt   string        `json:"created_at,omitempty"`
	UpdatedAt   string        `json:"updated_at,omitempty"`
}

// TaskComment is a comment attached to a task.
type TaskComment struct {
	ID        string `json:"id,omitempty"`
	Body      string `json:"body"`
	Author    string `json:"author,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

// TaskListOptions filters task lists.
type TaskListOptions struct {
	Status string
}

// TaskCreateRequest creates a task.
type TaskCreateRequest struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
}

func ValidTaskStatus(status string) bool {
	switch status {
	case "todo", "doing", "review", "done":
		return true
	default:
		return false
	}
}

func (c Client) ListTasks(ctx context.Context, opts TaskListOptions) ([]Task, error) {
	query := url.Values{}
	if opts.Status != "" {
		query.Set("status", opts.Status)
	}
	path := "/api/v1/tasks"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var tasks []Task
	if err := c.DoJSON(ctx, http.MethodGet, path, nil, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (c Client) ShowTask(ctx context.Context, id string) (Task, error) {
	var task Task
	err := c.DoJSON(ctx, http.MethodGet, taskPath(id), nil, &task)
	return task, err
}

func (c Client) CreateTask(ctx context.Context, req TaskCreateRequest) (Task, error) {
	var task Task
	err := c.DoJSON(ctx, http.MethodPost, "/api/v1/tasks", req, &task)
	return task, err
}

func (c Client) UpdateTaskStatus(ctx context.Context, id, status string) (Task, error) {
	var task Task
	err := c.DoJSON(ctx, http.MethodPatch, taskPath(id), map[string]string{"status": status}, &task)
	return task, err
}

func (c Client) DeleteTask(ctx context.Context, id string) error {
	return c.DoJSON(ctx, http.MethodDelete, taskPath(id), nil, nil)
}

func (c Client) CommentTask(ctx context.Context, id string, body []byte) (TaskComment, error) {
	var comment TaskComment
	err := c.DoJSON(ctx, http.MethodPost, taskPath(id)+"/comments", map[string]string{
		"body": string(body),
	}, &comment)
	return comment, err
}

func (c Client) LinkTaskDocument(ctx context.Context, taskID, documentID string) (Task, error) {
	var task Task
	err := c.DoJSON(ctx, http.MethodPost, taskPath(taskID)+"/documents", map[string]string{
		"document_id": documentID,
	}, &task)
	return task, err
}

func (c Client) PublishTaskDocument(ctx context.Context, taskID string, body []byte, tags []string) (Task, error) {
	var task Task
	err := c.DoJSON(ctx, http.MethodPost, taskPath(taskID)+"/documents", DocumentCreateRequest{
		Body: string(body),
		Tags: tags,
	}, &task)
	return task, err
}

func (c Client) UnlinkTaskDocument(ctx context.Context, taskID, documentID string) error {
	return c.DoJSON(ctx, http.MethodDelete, taskPath(taskID)+"/documents/"+PathEscape(documentID), nil, nil)
}

func taskPath(id string) string {
	return fmt.Sprintf("/api/v1/tasks/%s", PathEscape(id))
}
