package sapixdb

import "context"

// CollectionClient provides read/write access to a named agent (collection).
// In SapixDB, every collection IS an agent — the agent-as-table paradigm.
// Operations route to /v1/agents/{name}/... endpoints.
type CollectionClient struct {
	http *httpClient
	name string
}

// Write appends a JSON record to the agent's strand. Nothing is ever overwritten.
func (c *CollectionClient) Write(ctx context.Context, data map[string]any) (*WriteResponse, error) {
	var out WriteResponse
	if err := c.http.do(ctx, "POST", "/v1/agents/"+c.name+"/records/json", map[string]any{"data": data}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// WriteBatch appends multiple records sequentially.
func (c *CollectionClient) WriteBatch(ctx context.Context, records []map[string]any) ([]*WriteResponse, error) {
	results := make([]*WriteResponse, 0, len(records))
	for _, r := range records {
		wr, err := c.Write(ctx, r)
		if err != nil {
			return results, err
		}
		results = append(results, wr)
	}
	return results, nil
}

// Get fetches a record by content hash.
func (c *CollectionClient) Get(ctx context.Context, contentHash string) (*RecordView, error) {
	var out RecordView
	if err := c.http.do(ctx, "GET", "/v1/agents/"+c.name+"/records/"+contentHash, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Latest returns the most recent N records from the agent.
func (c *CollectionClient) Latest(ctx context.Context, limit int) (*QueryResult, error) {
	if limit <= 0 {
		limit = 20
	}
	return c.Query(ctx, map[string]any{"type": "latest", "limit": limit})
}

// Query executes a structured SaQL query against the agent.
// body must have at least a "type" key, e.g.:
//
//	{"type": "latest", "limit": 10}
//	{"type": "hash",   "content_hash": "e7f2..."}
//	{"type": "time_range", "from_ts": ..., "to_ts": 18446744073709551615}
//	{"type": "chain_head"}
func (c *CollectionClient) Query(ctx context.Context, body map[string]any) (*QueryResult, error) {
	var out QueryResult
	if err := c.http.do(ctx, "POST", "/v1/agents/"+c.name+"/query", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Head returns the current chain head and record count for the agent.
func (c *CollectionClient) Head(ctx context.Context) (*ChainHeadResponse, error) {
	var out ChainHeadResponse
	if err := c.http.do(ctx, "GET", "/v1/agents/"+c.name+"/strand/head", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Status returns the agent status (record count + chain head).
func (c *CollectionClient) Status(ctx context.Context) (*AgentStatusResponse, error) {
	var out AgentStatusResponse
	if err := c.http.do(ctx, "GET", "/v1/agents/"+c.name+"/status", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
