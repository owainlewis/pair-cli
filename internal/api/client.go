package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// HTTPClient is the subset of *http.Client used by Client.
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// Client talks to the PAIR HTTP API.
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient HTTPClient
}

// APIError is returned for non-2xx API responses.
type APIError struct {
	StatusCode int
	Code       string
	Message    string
	Method     string
	URL        string
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("%s %s: %d %s: %s", e.Method, e.URL, e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("%s %s: %d: %s", e.Method, e.URL, e.StatusCode, e.Message)
}

// PathEscape escapes one URL path segment.
func PathEscape(segment string) string {
	return url.PathEscape(segment)
}

// DoJSON sends a JSON request and decodes a JSON response into out.
func (c Client) DoJSON(ctx context.Context, method, path string, in, out any) error {
	var body io.Reader
	if in != nil {
		data, err := json.Marshal(in)
		if err != nil {
			return fmt.Errorf("encode request: %w", err)
		}
		body = bytes.NewReader(data)
	}

	req, err := c.newRequest(ctx, method, path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.do(req, func(resp *http.Response) error {
		if out == nil || resp.StatusCode == http.StatusNoContent {
			return nil
		}
		return json.NewDecoder(resp.Body).Decode(out)
	})
}

// DoMarkdown sends raw markdown and optionally returns the raw response body.
func (c Client) DoMarkdown(ctx context.Context, method, path string, markdown []byte) ([]byte, error) {
	req, err := c.newRequest(ctx, method, path, bytes.NewReader(markdown))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "text/markdown, text/plain")
	req.Header.Set("Content-Type", "text/markdown; charset=utf-8")

	var response []byte
	err = c.do(req, func(resp *http.Response) error {
		if resp.StatusCode == http.StatusNoContent {
			return nil
		}
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read response: %w", err)
		}
		response = data
		return nil
	})
	return response, err
}

// DoRaw gets a raw response body without JSON decoding.
func (c Client) DoRaw(ctx context.Context, method, path string) ([]byte, error) {
	req, err := c.newRequest(ctx, method, path, nil)
	if err != nil {
		return nil, err
	}

	var response []byte
	err = c.do(req, func(resp *http.Response) error {
		if resp.StatusCode == http.StatusNoContent {
			return nil
		}
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read response: %w", err)
		}
		response = data
		return nil
	})
	return response, err
}

func (c Client) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	baseURL, err := url.Parse(strings.TrimRight(c.BaseURL, "/"))
	if err != nil {
		return nil, fmt.Errorf("parse base URL: %w", err)
	}
	pathURL, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("parse request path: %w", err)
	}
	requestURL := baseURL.ResolveReference(pathURL)

	req, err := http.NewRequestWithContext(ctx, method, requestURL.String(), body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	return req, nil
}

func (c Client) do(req *http.Request, decode func(*http.Response) error) error {
	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%s %s: %w", req.Method, sanitizeURL(req.URL), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return decodeAPIError(req, resp)
	}

	return decode(resp)
}

func decodeAPIError(req *http.Request, resp *http.Response) error {
	apiErr := &APIError{
		StatusCode: resp.StatusCode,
		Method:     req.Method,
		URL:        sanitizeURL(req.URL),
		Message:    http.StatusText(resp.StatusCode),
	}

	contentType := resp.Header.Get("Content-Type")
	data, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		apiErr.Message = readErr.Error()
		return apiErr
	}

	if strings.Contains(contentType, "application/json") {
		var payload struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal(data, &payload); err == nil {
			apiErr.Code = payload.Error
			if payload.Message != "" {
				apiErr.Message = payload.Message
			}
			return apiErr
		}
	}

	if text := strings.TrimSpace(string(data)); text != "" {
		apiErr.Message = text
	}
	return apiErr
}

func sanitizeURL(u *url.URL) string {
	clean := *u
	clean.User = nil
	query := clean.Query()
	for key := range query {
		if strings.Contains(strings.ToLower(key), "token") {
			query.Set(key, "REDACTED")
		}
	}
	clean.RawQuery = query.Encode()
	return clean.String()
}
