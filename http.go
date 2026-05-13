package sapixdb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type httpClient struct {
	base    string
	agent   string
	headers map[string]string
	client  *http.Client
}

func newHTTPClient(base, agent string, headers map[string]string, timeout time.Duration) *httpClient {
	return &httpClient{
		base:    base,
		agent:   agent,
		headers: headers,
		client:  &http.Client{Timeout: timeout},
	}
}

func (c *httpClient) do(ctx context.Context, method, path string, body, out any) error {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return &SapixNetworkError{Cause: err}
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.base+path, bodyReader)
	if err != nil {
		return &SapixNetworkError{Cause: err}
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return &SapixNetworkError{Cause: err}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return &SapixNetworkError{Cause: err}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errBody struct {
			Error string `json:"error"`
			Code  string `json:"code"`
		}
		_ = json.Unmarshal(respBody, &errBody)
		msg := errBody.Error
		if msg == "" {
			msg = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
		return &SapixError{Message: msg, Status: resp.StatusCode, Code: errBody.Code}
	}

	if out != nil {
		return json.Unmarshal(respBody, out)
	}
	return nil
}
