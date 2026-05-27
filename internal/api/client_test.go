package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDoJSONSendsAuthAndDecodesResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer pair_secret" {
			t.Fatalf("missing auth header: %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("unexpected content type: %q", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Fatalf("unexpected accept: %q", r.Header.Get("Accept"))
		}

		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload["name"] != "Draft" {
			t.Fatalf("unexpected payload: %#v", payload)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"doc_123"}`))
	}))
	defer server.Close()

	client := Client{BaseURL: server.URL, Token: "pair_secret"}
	var out struct {
		ID string `json:"id"`
	}
	if err := client.DoJSON(context.Background(), http.MethodPost, "/api/v1/documents", map[string]string{"name": "Draft"}, &out); err != nil {
		t.Fatalf("DoJSON() error = %v", err)
	}
	if out.ID != "doc_123" {
		t.Fatalf("expected decoded response, got %#v", out)
	}
}

func TestDoMarkdownSendsMarkdownContentType(t *testing.T) {
	const body = "# Title\n\nBody\n"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Content-Type"); got != "text/markdown; charset=utf-8" {
			t.Fatalf("unexpected content type: %q", got)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer pair_secret" {
			t.Fatalf("unexpected authorization: %q", got)
		}
		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if string(data) != body {
			t.Fatalf("body changed: %q", string(data))
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := Client{BaseURL: server.URL, Token: "pair_secret"}
	if _, err := client.DoMarkdown(context.Background(), http.MethodPut, "/api/v1/documents/doc_123/content", []byte(body)); err != nil {
		t.Fatalf("DoMarkdown() error = %v", err)
	}
}

func TestNoContentIsSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := Client{BaseURL: server.URL}
	if err := client.DoJSON(context.Background(), http.MethodDelete, "/api/v1/documents/doc_123", nil, nil); err != nil {
		t.Fatalf("DoJSON() error = %v", err)
	}
}

func TestJSONAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not_found","message":"Missing"}`))
	}))
	defer server.Close()

	client := Client{BaseURL: server.URL, Token: "pair_secret"}
	err := client.DoJSON(context.Background(), http.MethodGet, "/api/v1/documents/missing", nil, nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T %v", err, err)
	}
	if apiErr.StatusCode != http.StatusNotFound || apiErr.Code != "not_found" || apiErr.Message != "Missing" {
		t.Fatalf("unexpected APIError: %#v", apiErr)
	}
	if strings.Contains(err.Error(), "pair_secret") {
		t.Fatalf("token leaked in error: %v", err)
	}
}

func TestTextAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad markdown"))
	}))
	defer server.Close()

	client := Client{BaseURL: server.URL}
	err := client.DoJSON(context.Background(), http.MethodGet, "/api/v1/documents/doc_123/content", nil, nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T %v", err, err)
	}
	if apiErr.Message != "bad markdown" {
		t.Fatalf("expected text error, got %#v", apiErr)
	}
}

func TestPathEscape(t *testing.T) {
	if got := PathEscape("doc/with space"); got != "doc%2Fwith%20space" {
		t.Fatalf("PathEscape() = %q", got)
	}
}
