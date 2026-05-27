package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateTaskStatusValidatesRequestPathAndPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("method = %s", r.Method)
		}
		if r.URL.EscapedPath() != "/api/v1/tasks/task%2F123" {
			t.Fatalf("path = %s", r.URL.EscapedPath())
		}
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if payload["status"] != "doing" {
			t.Fatalf("payload = %#v", payload)
		}
		_ = json.NewEncoder(w).Encode(Task{ID: "task/123", Status: "doing"})
	}))
	defer server.Close()

	client := Client{BaseURL: server.URL}
	task, err := client.UpdateTaskStatus(context.Background(), "task/123", "doing")
	if err != nil {
		t.Fatalf("UpdateTaskStatus() error = %v", err)
	}
	if task.Status != "doing" {
		t.Fatalf("task = %#v", task)
	}
}

func TestValidTaskStatus(t *testing.T) {
	for _, status := range []string{"todo", "doing", "review", "done"} {
		if !ValidTaskStatus(status) {
			t.Fatalf("%q should be valid", status)
		}
	}
	if ValidTaskStatus("blocked") {
		t.Fatal("blocked should be invalid")
	}
}
