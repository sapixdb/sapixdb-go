// Package sapixdb is the official Go client for SapixDB — the agent-native living database.
//
// Usage:
//
//	db := sapixdb.New(sapixdb.Config{
//	    URL:   "http://localhost:7475",
//	    Agent: "my-app",
//	})
//
//	record, err := db.Collection("products").Write(ctx, map[string]any{
//	    "name":  "Classic T-Shirt",
//	    "price": 29.99,
//	})
package sapixdb

import (
	"context"
	"time"
)

const defaultTimeout = 10 * time.Second

// Config holds client configuration.
type Config struct {
	// URL is the SapixDB agent base URL, e.g. "http://localhost:7475".
	URL string
	// Agent is the agent identifier (matches SAPIX_AGENT_ID on the server).
	Agent string
	// Headers are extra HTTP headers sent with every request.
	Headers map[string]string
	// Timeout overrides the default 10s request timeout.
	Timeout time.Duration
}

// Client is the SapixDB client. Create one with New.
type Client struct {
	http  *httpClient
	agent string
	// Graph provides graph relationship operations.
	Graph *GraphClient
}

// New creates a Client from a Config.
func New(cfg Config) *Client {
	t := cfg.Timeout
	if t == 0 {
		t = defaultTimeout
	}
	h := newHTTPClient(cfg.URL, cfg.Agent, cfg.Headers, t)
	return &Client{
		http:  h,
		agent: cfg.Agent,
		Graph: &GraphClient{http: h, agent: cfg.Agent},
	}
}

// Collection returns a CollectionClient for the named collection.
// Collections are created automatically on first write.
func (c *Client) Collection(name string) *CollectionClient {
	return &CollectionClient{http: c.http, agent: c.agent, name: name}
}

// Ingest writes a record via the ingest endpoint — for AI agents, webhooks, and pipelines.
func (c *Client) Ingest(ctx context.Context, collection string, data map[string]any) (*WriteResult, error) {
	body := map[string]any{"collection": collection, "data": data}
	var out WriteResult
	if err := c.http.do(ctx, "POST", "/v1/"+c.agent+"/ingest", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Health calls the health endpoint and returns the response.
func (c *Client) Health(ctx context.Context) (*HealthResponse, error) {
	var out HealthResponse
	if err := c.http.do(ctx, "GET", "/v1/health", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Ping returns true if the agent is reachable and healthy. Never returns an error.
func (c *Client) Ping(ctx context.Context) bool {
	h, err := c.Health(ctx)
	return err == nil && h.Status == "ok"
}
