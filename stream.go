package sapixdb

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// StreamEvent is a realtime write event pushed via SSE from SapixDB.
type StreamEvent struct {
	AgentID     string          `json:"agent_id"`
	EventType   string          `json:"event_type"`
	RecordID    string          `json:"record_id"`
	ContentHash string          `json:"content_hash"`
	TimestampMs int64           `json:"timestamp_ms"`
	Payload     json.RawMessage `json:"payload,omitempty"`
}

// SubscribeOptions controls an SSE subscription.
type SubscribeOptions struct {
	// Since is an HLC timestamp; the server replays records written after
	// this value before switching to the live stream.
	Since *uint64
	// Filter limits events to those whose JSON payload contains this field name.
	Filter string
	// Agents restricts the global stream to specific agent IDs.
	// Only used by SubscribeGlobal.
	Agents []string
}

// Subscription is an active SSE connection. Receive events from Events and
// call Close to stop the stream.
type Subscription struct {
	// Events is a buffered, read-only channel of incoming stream events.
	Events <-chan StreamEvent
	cancel context.CancelFunc
	done   <-chan struct{}
}

// Close cancels the subscription and waits for the background goroutine to exit.
func (s *Subscription) Close() {
	s.cancel()
	<-s.done
}

// Subscribe opens an SSE stream for the named agent.
//
//	sub, err := db.Subscribe(ctx, "orders", sapixdb.SubscribeOptions{})
//	defer sub.Close()
//	for event := range sub.Events {
//	    fmt.Println(event.AgentID, event.RecordID)
//	}
func (c *Client) Subscribe(ctx context.Context, agentID string, opts SubscribeOptions) (*Subscription, error) {
	return c.openStream(ctx, "/v1/agents/"+agentID+"/stream", opts)
}

// SubscribeGlobal opens an SSE stream that receives writes from all agents.
// Use opts.Agents to filter to a subset.
func (c *Client) SubscribeGlobal(ctx context.Context, opts SubscribeOptions) (*Subscription, error) {
	return c.openStream(ctx, "/v1/stream", opts)
}

func (c *Client) openStream(ctx context.Context, path string, opts SubscribeOptions) (*Subscription, error) {
	subCtx, cancel := context.WithCancel(ctx)

	rawURL := c.http.base + path
	params := url.Values{}
	if opts.Since != nil {
		params.Set("since", fmt.Sprintf("%d", *opts.Since))
	}
	if opts.Filter != "" {
		params.Set("filter", opts.Filter)
	}
	if len(opts.Agents) > 0 {
		params.Set("agents", strings.Join(opts.Agents, ","))
	}
	if len(params) > 0 {
		rawURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(subCtx, "GET", rawURL, nil)
	if err != nil {
		cancel()
		return nil, &SapixNetworkError{Cause: err}
	}
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	for k, v := range c.http.headers {
		req.Header.Set(k, v)
	}

	// Use a dedicated client with no timeout for long-lived streaming.
	streamHTTP := &http.Client{}
	resp, err := streamHTTP.Do(req)
	if err != nil {
		cancel()
		return nil, &SapixNetworkError{Cause: err}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		cancel()
		return nil, &SapixError{
			Message: fmt.Sprintf("HTTP %d", resp.StatusCode),
			Status:  resp.StatusCode,
		}
	}

	ch := make(chan StreamEvent, 64)
	done := make(chan struct{})

	go func() {
		defer close(done)
		defer close(ch)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		var eventName string
		var dataBuf []string

		for scanner.Scan() {
			line := scanner.Text()
			switch {
			case strings.HasPrefix(line, "event:"):
				eventName = strings.TrimSpace(line[6:])
			case strings.HasPrefix(line, "data:"):
				dataBuf = append(dataBuf, strings.TrimSpace(line[5:]))
			case line == "":
				if len(dataBuf) > 0 && eventName != "heartbeat" {
					var ev StreamEvent
					if json.Unmarshal([]byte(strings.Join(dataBuf, "\n")), &ev) == nil {
						select {
						case ch <- ev:
						case <-subCtx.Done():
							return
						}
					}
				}
				eventName = ""
				dataBuf = dataBuf[:0]
			}
		}
	}()

	return &Subscription{Events: ch, cancel: cancel, done: done}, nil
}
