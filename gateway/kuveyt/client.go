package kuveyt

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

func (c *httpClient) postXML(ctx context.Context, endpoint string, body []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("kuveyt: create request: %w", err)
	}
	req.Header.Set("Content-Type", "text/xml; charset=UTF-8")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("kuveyt: post: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("kuveyt: read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("kuveyt: status %d: %s", resp.StatusCode, string(data))
	}
	return data, nil
}

func (c *httpClient) postSOAP(ctx context.Context, endpoint, soapAction string, body []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("kuveyt: create soap request: %w", err)
	}
	req.Header.Set("Content-Type", "text/xml; charset=UTF-8")
	req.Header.Set("SOAPAction", soapAction)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("kuveyt: soap post: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("kuveyt: read soap response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("kuveyt: soap status %d: %s", resp.StatusCode, string(data))
	}
	return data, nil
}
