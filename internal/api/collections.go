package api

import (
	"context"
	"fmt"
	"net/http"
)

// Collection is a PAIR collection record.
type Collection struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	DocumentCount int        `json:"document_count,omitempty"`
	Documents     []Document `json:"documents,omitempty"`
	CreatedAt     string     `json:"created_at,omitempty"`
	UpdatedAt     string     `json:"updated_at,omitempty"`
}

func (c Client) ListCollections(ctx context.Context) ([]Collection, error) {
	var collections []Collection
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/collections", nil, &collections); err != nil {
		return nil, err
	}
	return collections, nil
}

func (c Client) ShowCollection(ctx context.Context, id string) (Collection, error) {
	var collection Collection
	err := c.DoJSON(ctx, http.MethodGet, collectionPath(id), nil, &collection)
	return collection, err
}

func (c Client) CreateCollection(ctx context.Context, name string) (Collection, error) {
	var collection Collection
	err := c.DoJSON(ctx, http.MethodPost, "/api/v1/collections", map[string]string{"name": name}, &collection)
	return collection, err
}

func (c Client) RenameCollection(ctx context.Context, id, name string) (Collection, error) {
	var collection Collection
	err := c.DoJSON(ctx, http.MethodPatch, collectionPath(id), map[string]string{"name": name}, &collection)
	return collection, err
}

func (c Client) LinkCollectionDocument(ctx context.Context, collectionID, documentID string) (Collection, error) {
	var collection Collection
	err := c.DoJSON(ctx, http.MethodPost, collectionPath(collectionID)+"/documents", map[string]string{
		"document_id": documentID,
	}, &collection)
	return collection, err
}

func (c Client) PublishCollectionDocument(ctx context.Context, collectionID string, body []byte, tags []string) (Collection, error) {
	var collection Collection
	err := c.DoJSON(ctx, http.MethodPost, collectionPath(collectionID)+"/documents", DocumentCreateRequest{
		Body: string(body),
		Tags: tags,
	}, &collection)
	return collection, err
}

func (c Client) UnlinkCollectionDocument(ctx context.Context, collectionID, documentID string) error {
	return c.DoJSON(ctx, http.MethodDelete, collectionPath(collectionID)+"/documents/"+PathEscape(documentID), nil, nil)
}

func (c Client) DeleteCollection(ctx context.Context, id string) error {
	return c.DoJSON(ctx, http.MethodDelete, collectionPath(id), nil, nil)
}

func collectionPath(id string) string {
	return fmt.Sprintf("/api/v1/collections/%s", PathEscape(id))
}
