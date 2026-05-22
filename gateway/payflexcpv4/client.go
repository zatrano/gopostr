package payflexcpv4

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultHTTPTimeout = 30 * time.Second

type httpClient struct {
	client *http.Client
}

func newHTTPClient() *httpClient {
	return &httpClient{client: &http.Client{Timeout: defaultHTTPTimeout}}
}

func (c *httpClient) postForm(ctx context.Context, endpoint, body string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("payflexcpv4: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "text/xml")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("payflexcpv4: post: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("payflexcpv4: read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("payflexcpv4: status %d: %s", resp.StatusCode, string(data))
	}
	return data, nil
}
