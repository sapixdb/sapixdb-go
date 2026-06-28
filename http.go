package sapixdb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

var rpsRe = regexp.MustCompile(`rate limit of (\d+) req/s`)

type httpClient struct {
	base    string
	headers map[string]string
	client  *http.Client
}

func newHTTPClient(base string, headers map[string]string, timeout time.Duration) *httpClient {
	return &httpClient{
		base:    base,
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
			Error   string `json:"error"`
			Message string `json:"message"`
			Code    string `json:"code"`
		}
		_ = json.Unmarshal(respBody, &errBody)
		msg := errBody.Message
		if msg == "" {
			msg = errBody.Error
		}
		if msg == "" {
			msg = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
		if resp.StatusCode == 429 {
			rle := &SapixRateLimitError{Message: msg}
			if m := rpsRe.FindStringSubmatch(msg); m != nil {
				rle.RPSLimit, _ = strconv.Atoi(m[1])
			}
			return rle
		}
		return &SapixError{Message: msg, Status: resp.StatusCode, Code: errBody.Code}
	}

	if out != nil && len(respBody) > 0 {
		return json.Unmarshal(respBody, out)
	}
	return nil
}

func (c *httpClient) doBytes(ctx context.Context, path string, data []byte, out any) error {
	req, err := http.NewRequestWithContext(ctx, "PUT", c.base+path, bytes.NewReader(data))
	if err != nil {
		return &SapixNetworkError{Cause: err}
	}
	req.Header.Set("Content-Type", "application/octet-stream")
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
			Error   string `json:"error"`
			Message string `json:"message"`
			Code    string `json:"code"`
		}
		_ = json.Unmarshal(respBody, &errBody)
		msg := errBody.Message
		if msg == "" {
			msg = errBody.Error
		}
		if msg == "" {
			msg = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
		if resp.StatusCode == 429 {
			rle := &SapixRateLimitError{Message: msg}
			if m := rpsRe.FindStringSubmatch(msg); m != nil {
				rle.RPSLimit, _ = strconv.Atoi(m[1])
			}
			return rle
		}
		return &SapixError{Message: msg, Status: resp.StatusCode, Code: errBody.Code}
	}

	if out != nil && len(respBody) > 0 {
		return json.Unmarshal(respBody, out)
	}
	return nil
}

func (c *httpClient) getBytes(ctx context.Context, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.base+path, nil)
	if err != nil {
		return nil, &SapixNetworkError{Cause: err}
	}
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, &SapixNetworkError{Cause: err}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &SapixNetworkError{Cause: err}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errBody struct {
			Error   string `json:"error"`
			Message string `json:"message"`
			Code    string `json:"code"`
		}
		_ = json.Unmarshal(respBody, &errBody)
		msg := errBody.Message
		if msg == "" {
			msg = errBody.Error
		}
		if msg == "" {
			msg = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
		if resp.StatusCode == 429 {
			rle := &SapixRateLimitError{Message: msg}
			if m := rpsRe.FindStringSubmatch(msg); m != nil {
				rle.RPSLimit, _ = strconv.Atoi(m[1])
			}
			return nil, rle
		}
		return nil, &SapixError{Message: msg, Status: resp.StatusCode, Code: errBody.Code}
	}

	return respBody, nil
}
