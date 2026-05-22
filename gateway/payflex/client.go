package payflex

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

func (c *httpClient) postForm(ctx context.Context, endpoint string, fields map[string]string) ([]byte, error) {
	form := url.Values{}
	for k, v := range fields {
		form.Set(k, v)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("payflex: create form request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("payflex: form post: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("payflex: read form response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("payflex: form status %d: %s", resp.StatusCode, string(data))
	}
	return data, nil
}

func (c *httpClient) postXML(ctx context.Context, endpoint string, xmlBody []byte) ([]byte, error) {
	form := url.Values{}
	form.Set("prmstr", string(xmlBody))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("payflex: create xml request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("payflex: xml post: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("payflex: read xml response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("payflex: xml status %d: %s", resp.StatusCode, string(data))
	}
	return data, nil
}

func (c *httpClient) postSearch(ctx context.Context, endpoint string, xmlBody []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(xmlBody))
	if err != nil {
		return nil, fmt.Errorf("payflex: create search request: %w", err)
	}
	req.Header.Set("Content-Type", "application/xml; charset=utf-8")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("payflex: search post: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("payflex: read search response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("payflex: search status %d: %s", resp.StatusCode, string(data))
	}
	return data, nil
}
