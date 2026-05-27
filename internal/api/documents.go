package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Document is document metadata returned by the PAIR API.
type Document struct {
	ID        string   `json:"id"`
	Title     string   `json:"title,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	CreatedAt string   `json:"created_at,omitempty"`
	UpdatedAt string   `json:"updated_at,omitempty"`
}

// DocumentListOptions are filters for listing documents.
type DocumentListOptions struct {
	Query string
	Tags  []string
	Since string
}

// DocumentCreateRequest creates a markdown document.
type DocumentCreateRequest struct {
	Body string   `json:"body"`
	Tags []string `json:"tags,omitempty"`
}

func (c Client) ListDocuments(ctx context.Context, opts DocumentListOptions) ([]Document, error) {
	query := url.Values{}
	if opts.Query != "" {
		query.Set("q", opts.Query)
	}
	for _, tag := range opts.Tags {
		query.Add("tag", tag)
	}
	if opts.Since != "" {
		query.Set("since", opts.Since)
	}

	path := "/api/v1/documents"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var docs []Document
	if err := c.DoJSON(ctx, http.MethodGet, path, nil, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

func (c Client) ShowDocument(ctx context.Context, id string) (Document, error) {
	var doc Document
	err := c.DoJSON(ctx, http.MethodGet, documentPath(id), nil, &doc)
	return doc, err
}

func (c Client) CreateDocument(ctx context.Context, body []byte, tags []string) (Document, error) {
	var doc Document
	err := c.DoJSON(ctx, http.MethodPost, "/api/v1/documents", DocumentCreateRequest{
		Body: string(body),
		Tags: tags,
	}, &doc)
	return doc, err
}

func (c Client) ReadDocumentContent(ctx context.Context, id string) ([]byte, error) {
	return c.DoRaw(ctx, http.MethodGet, documentPath(id)+"/content")
}

func (c Client) ReplaceDocumentContent(ctx context.Context, id string, body []byte) error {
	_, err := c.DoMarkdown(ctx, http.MethodPut, documentPath(id)+"/content", body)
	return err
}

func (c Client) DeleteDocument(ctx context.Context, id string) error {
	return c.DoJSON(ctx, http.MethodDelete, documentPath(id), nil, nil)
}

func documentPath(id string) string {
	return fmt.Sprintf("/api/v1/documents/%s", PathEscape(id))
}

// TagsString returns a stable human-readable tag list.
func TagsString(tags []string) string {
	return strings.Join(tags, ",")
}
