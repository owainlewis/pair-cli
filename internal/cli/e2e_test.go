package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestE2EDocumentCreateReadReplace(t *testing.T) {
	var sawCreate, sawRead, sawReplace bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireAuth(t, r)
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/documents":
			sawCreate = true
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode create: %v", err)
			}
			if payload["body"] != "# Draft" {
				t.Fatalf("create body = %#v", payload)
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"id": "doc_123", "title": "Draft"})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/documents/doc_123/content":
			sawRead = true
			w.Header().Set("Content-Type", "text/markdown")
			_, _ = w.Write([]byte("# Draft\n"))
		case r.Method == http.MethodPut && r.URL.Path == "/api/v1/documents/doc_123/content":
			sawReplace = true
			if got := r.Header.Get("Content-Type"); got != "text/markdown; charset=utf-8" {
				t.Fatalf("replace content type = %q", got)
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	withAPIEnv(t, server.URL)

	stdout, stderr, code := executeCLI("docs", "create", "--body", "# Draft", "--json")
	if code != 0 {
		t.Fatalf("docs create failed: code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	if !strings.Contains(stdout, "doc_123") {
		t.Fatalf("expected created doc id, got %q", stdout)
	}

	stdout, stderr, code = executeCLI("docs", "read", "doc_123")
	if code != 0 || stdout != "# Draft\n" {
		t.Fatalf("docs read failed: code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}

	_, stderr, code = executeCLI("docs", "replace", "doc_123", "--body", "# Updated\n")
	if code != 0 {
		t.Fatalf("docs replace failed: code=%d stderr=%q", code, stderr)
	}
	if !sawCreate || !sawRead || !sawReplace {
		t.Fatalf("expected all document endpoints, create=%v read=%v replace=%v", sawCreate, sawRead, sawReplace)
	}
}

func TestE2ETaskStatusCommentPublish(t *testing.T) {
	var sawStatus, sawComment, sawPublish bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireAuth(t, r)
		switch {
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v1/tasks/task_123":
			sawStatus = true
			_ = json.NewEncoder(w).Encode(map[string]string{"id": "task_123", "title": "Demo", "status": "review"})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/tasks/task_123/comments":
			sawComment = true
			_ = json.NewEncoder(w).Encode(map[string]string{"id": "comment_123", "body": "Looks good"})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/tasks/task_123/documents":
			sawPublish = true
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": "task_123", "title": "Demo", "status": "review",
				"documents": []map[string]string{{"id": "doc_123", "title": "Notes"}},
			})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	withAPIEnv(t, server.URL)

	for _, args := range [][]string{
		{"tasks", "status", "task_123", "review"},
		{"tasks", "comment", "task_123", "--body", "Looks good"},
		{"tasks", "publish", "task_123", "--body", "# Notes", "--tag", "demo"},
	} {
		stdout, stderr, code := executeCLI(args...)
		if code != 0 {
			t.Fatalf("%v failed: code=%d stdout=%q stderr=%q", args, code, stdout, stderr)
		}
	}
	if !sawStatus || !sawComment || !sawPublish {
		t.Fatalf("expected all task endpoints, status=%v comment=%v publish=%v", sawStatus, sawComment, sawPublish)
	}
}

func TestE2ECollectionPublish(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireAuth(t, r)
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/collections/col_123/documents" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": "col_123", "name": "Demo", "documents": []map[string]string{{"id": "doc_123", "title": "Notes"}},
		})
	}))
	defer server.Close()
	withAPIEnv(t, server.URL)

	stdout, stderr, code := executeCLI("collections", "publish", "col_123", "--body", "# Notes", "--tag", "demo", "--json")
	if code != 0 {
		t.Fatalf("collections publish failed: code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	if !strings.Contains(stdout, "doc_123") {
		t.Fatalf("expected document in output, got %q", stdout)
	}
}

func TestE2EFailures(t *testing.T) {
	t.Run("missing token", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", t.TempDir())
		t.Setenv("PAIR_BASE_URL", "http://127.0.0.1")
		t.Setenv("PAIR_TOKEN", "")
		_, stderr, code := executeCLI("docs", "list")
		if code != 1 || !strings.Contains(stderr, "missing token") {
			t.Fatalf("expected missing token config error, code=%d stderr=%q", code, stderr)
		}
	})

	t.Run("json api error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requireAuth(t, r)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"not_found","message":"Nope"}`))
		}))
		defer server.Close()
		withAPIEnv(t, server.URL)
		_, stderr, code := executeCLI("docs", "show", "missing")
		if code != 2 || !strings.Contains(stderr, "not_found") || strings.Contains(stderr, "pair_secret") {
			t.Fatalf("expected token-safe API error, code=%d stderr=%q", code, stderr)
		}
	})

	t.Run("text api error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requireAuth(t, r)
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("bad markdown"))
		}))
		defer server.Close()
		withAPIEnv(t, server.URL)
		_, stderr, code := executeCLI("docs", "read", "doc_123")
		if code != 2 || !strings.Contains(stderr, "bad markdown") {
			t.Fatalf("expected text API error, code=%d stderr=%q", code, stderr)
		}
	})

	t.Run("invalid status", func(t *testing.T) {
		withAPIEnv(t, "http://127.0.0.1")
		_, stderr, code := executeCLI("tasks", "status", "task_123", "blocked")
		if code != 1 || !strings.Contains(stderr, "invalid status") {
			t.Fatalf("expected local validation error, code=%d stderr=%q", code, stderr)
		}
	})

	t.Run("destructive without yes", func(t *testing.T) {
		withAPIEnv(t, "http://127.0.0.1")
		_, stderr, code := executeCLI("tasks", "delete", "task_123")
		if code != 1 || !strings.Contains(stderr, "requires --yes") {
			t.Fatalf("expected --yes error, code=%d stderr=%q", code, stderr)
		}
	})
}

func withAPIEnv(t *testing.T, baseURL string) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("PAIR_BASE_URL", baseURL)
	t.Setenv("PAIR_TOKEN", "pair_secret")
}

func requireAuth(t *testing.T, r *http.Request) {
	t.Helper()
	if got := r.Header.Get("Authorization"); got != "Bearer pair_secret" {
		t.Fatalf("authorization = %q", got)
	}
}
