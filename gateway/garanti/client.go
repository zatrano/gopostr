package garanti

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultHTTPTimeout = 30 * time.Second

type httpClient struct {
	client *http.Client
}

func newHTTPClient() *httpClient {
	return &httpClient{client: &http.Client{Timeout: defaultHTTPTimeout}}
}

func (c *httpClient) post(ctx context.Context, url string, body []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("garanti: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/xml; charset=utf-8")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("garanti: http post: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("garanti: read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("garanti: unexpected status %d: %s", resp.StatusCode, string(data))
	}
	return data, nil
}
