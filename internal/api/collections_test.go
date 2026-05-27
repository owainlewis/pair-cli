package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRenameCollectionSendsPatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("method = %s", r.Method)
		}
		if r.URL.EscapedPath() != "/api/v1/collections/col%2F123" {
			t.Fatalf("path = %s", r.URL.EscapedPath())
		}
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if payload["name"] != "Research" {
			t.Fatalf("payload = %#v", payload)
		}
		_ = json.NewEncoder(w).Encode(Collection{ID: "col/123", Name: "Research"})
	}))
	defer server.Close()

	client := Client{BaseURL: server.URL}
	collection, err := client.RenameCollection(context.Background(), "col/123", "Research")
	if err != nil {
		t.Fatalf("RenameCollection() error = %v", err)
	}
	if collection.Name != "Research" {
		t.Fatalf("collection = %#v", collection)
	}
}

func TestPublishCollectionDocument(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.EscapedPath() != "/api/v1/collections/col_123/documents" {
			t.Fatalf("path = %s", r.URL.EscapedPath())
		}
		var payload DocumentCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if payload.Body != "# Bundle\n" || len(payload.Tags) != 1 || payload.Tags[0] != "demo" {
			t.Fatalf("payload = %#v", payload)
		}
		_ = json.NewEncoder(w).Encode(Collection{ID: "col_123", Name: "Research"})
	}))
	defer server.Close()

	client := Client{BaseURL: server.URL}
	if _, err := client.PublishCollectionDocument(context.Background(), "col_123", []byte("# Bundle\n"), []string{"demo"}); err != nil {
		t.Fatalf("PublishCollectionDocument() error = %v", err)
	}
}
