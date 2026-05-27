package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListDocumentsBuildsQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/documents" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if got := r.URL.Query()["tag"]; len(got) != 2 || got[0] != "one" || got[1] != "two" {
			t.Fatalf("tag query = %#v", got)
		}
		if r.URL.Query().Get("q") != "draft" || r.URL.Query().Get("since") != "7d" {
			t.Fatalf("query = %s", r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode([]Document{{ID: "doc_123", Title: "Draft"}})
	}))
	defer server.Close()

	client := Client{BaseURL: server.URL}
	docs, err := client.ListDocuments(context.Background(), DocumentListOptions{
		Query: "draft",
		Tags:  []string{"one", "two"},
		Since: "7d",
	})
	if err != nil {
		t.Fatalf("ListDocuments() error = %v", err)
	}
	if len(docs) != 1 || docs[0].ID != "doc_123" {
		t.Fatalf("unexpected docs: %#v", docs)
	}
}

func TestReplaceDocumentContentEscapesIDAndPreservesBody(t *testing.T) {
	body := []byte("# Title\n\n")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.EscapedPath() != "/api/v1/documents/doc%2F123/content" {
			t.Fatalf("path = %q", r.URL.EscapedPath())
		}
		if r.Header.Get("Content-Type") != "text/markdown; charset=utf-8" {
			t.Fatalf("content type = %q", r.Header.Get("Content-Type"))
		}
		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if string(data) != string(body) {
			t.Fatalf("body = %q", string(data))
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := Client{BaseURL: server.URL}
	if err := client.ReplaceDocumentContent(context.Background(), "doc/123", body); err != nil {
		t.Fatalf("ReplaceDocumentContent() error = %v", err)
	}
}
