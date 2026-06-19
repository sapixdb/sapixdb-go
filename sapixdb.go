// Package sapixdb is the official Go client for SapixDB — the agent-native living database.
//
// Usage (primary agent):
//
//	db := sapixdb.New(sapixdb.Config{
//	    URL: "http://localhost:7475",
//	})
//	record, err := db.Write(ctx, map[string]any{"name": "Classic T-Shirt", "price": 29.99})
//
// Usage (named agent / collection):
//
//	orders := db.Collection("orders-agt")
//	record, err := orders.Write(ctx, map[string]any{"total": 99.99})
//	results, err := orders.Latest(ctx, 10)
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
	// Headers are extra HTTP headers sent with every request.
	Headers map[string]string
	// Timeout overrides the default 10s request timeout.
	Timeout time.Duration
}

// Client is the SapixDB client. Create one with New.
type Client struct {
	http *httpClient
	// Graph provides graph relationship operations.
	Graph *GraphClient
}

// New creates a Client from a Config.
func New(cfg Config) *Client {
	t := cfg.Timeout
	if t == 0 {
		t = defaultTimeout
	}
	h := newHTTPClient(cfg.URL, cfg.Headers, t)
	return &Client{
		http:  h,
		Graph: &GraphClient{http: h},
	}
}

// Collection returns a CollectionClient for the named agent (collection).
// In SapixDB, every collection is an agent — the agent-as-table paradigm.
// The named agent must have been created first via CreateAgent or SAPIX_AGENTS env var.
func (c *Client) Collection(name string) *CollectionClient {
	return &CollectionClient{http: c.http, name: name}
}

// Agent is an alias for Collection — prefer this name to match SapixDB terminology.
func (c *Client) Agent(id string) *CollectionClient {
	return &CollectionClient{http: c.http, name: id}
}

// Write appends a JSON record to the primary agent's strand.
func (c *Client) Write(ctx context.Context, data map[string]any) (*WriteResponse, error) {
	var out WriteResponse
	if err := c.http.do(ctx, "POST", "/v1/records/json", map[string]any{"data": data}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get fetches a record from the primary agent by content hash.
func (c *Client) Get(ctx context.Context, contentHash string) (*RecordView, error) {
	var out RecordView
	if err := c.http.do(ctx, "GET", "/v1/records/"+contentHash+"/json", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Query executes a structured SaQL query against the primary agent (POST /v1/query).
// body must have at least a "type" key, e.g.:
//
//	{"type": "latest", "limit": 10}
//	{"type": "hash",   "content_hash": "e7f2..."}
//	{"type": "time_range", "from_ts": ..., "to_ts": 18446744073709551615}
//	{"type": "chain_head"}
func (c *Client) Query(ctx context.Context, body map[string]any) (*QueryResult, error) {
	var out QueryResult
	if err := c.http.do(ctx, "POST", "/v1/query", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// QuerySemantic sends a natural-language query to POST /v1/query/semantic.
func (c *Client) QuerySemantic(ctx context.Context, nl string) (*SemanticQueryResult, error) {
	var out SemanticQueryResult
	if err := c.http.do(ctx, "POST", "/v1/query/semantic", map[string]any{"query": nl}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Head returns the primary agent's current chain head and record count.
func (c *Client) Head(ctx context.Context) (*ChainHeadResponse, error) {
	var out ChainHeadResponse
	if err := c.http.do(ctx, "GET", "/v1/strand/head", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Health calls GET /v1/health and returns the response.
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

// CreateAgent creates a new named agent with a genesis block (POST /v1/agents).
// seedHex must be 64 hex characters (32 bytes). zone defaults to "governed".
func (c *Client) CreateAgent(ctx context.Context, agentID, seedHex, zone string) (*CreateAgentResponse, error) {
	if zone == "" {
		zone = "governed"
	}
	body := map[string]any{"agent_id": agentID, "seed_hex": seedHex, "zone": zone}
	var out CreateAgentResponse
	if err := c.http.do(ctx, "POST", "/v1/agents", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListAgents returns all agents running on this node (GET /v1/agents).
func (c *Client) ListAgents(ctx context.Context) (*AgentListResponse, error) {
	var out AgentListResponse
	if err := c.http.do(ctx, "GET", "/v1/agents", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Ingest writes a JSON record to a named agent's strand.
// Designed for AI agents, webhooks, and automated pipelines.
// If agentID is empty, writes to the primary agent.
func (c *Client) Ingest(ctx context.Context, agentID string, data map[string]any) (*WriteResponse, error) {
	path := "/v1/records/json"
	if agentID != "" {
		path = "/v1/agents/" + agentID + "/records/json"
	}
	var out WriteResponse
	if err := c.http.do(ctx, "POST", path, map[string]any{"data": data}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
