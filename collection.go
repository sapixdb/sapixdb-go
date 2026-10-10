package sapixdb

import (
	"context"
	"fmt"
)

// CollectionClient provides read/write access to a named agent (collection).
// In SapixDB, every collection IS an agent — the agent-as-table paradigm.
// Operations route to /v1/agents/{name}/... endpoints.
type CollectionClient struct {
	http *httpClient
	name string
}

// CollectionQuery is a read-only view of a collection scoped to a specific
// HLC timestamp. Obtain one via CollectionClient.AsOf.
type CollectionQuery struct {
	http         *httpClient
	name         string
	timestampHLC uint64
}

// Find returns records matching filter at the scoped timestamp (limit capped at 1000).
func (q *CollectionQuery) Find(ctx context.Context, filter map[string]any, limit int) (*QueryResult, error) {
	if limit <= 0 {
		limit = 20
	}
	body := map[string]any{
		"type":          "as_of",
		"timestamp_hlc": q.timestampHLC,
		"limit":         limit,
	}
	if len(filter) > 0 {
		f := buildFilterBody(filter)
		body = map[string]any{
			"type":   "scan",
			"limit":  limit,
			"filter": f,
		}
	}
	var out QueryResult
	if err := q.http.do(ctx, "POST", "/v1/agents/"+q.name+"/query", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FindOne returns the first record matching filter at the scoped timestamp, or nil.
func (q *CollectionQuery) FindOne(ctx context.Context, filter map[string]any) (*RecordView, error) {
	res, err := q.Find(ctx, filter, 1)
	if err != nil {
		return nil, err
	}
	if len(res.Records) == 0 {
		return nil, nil
	}
	return &res.Records[0], nil
}

// buildFilterBody converts a {field: value} map to a SaQL WhereFilter.
// Only the first key is used.
func buildFilterBody(filter map[string]any) map[string]any {
	for field, value := range filter {
		return map[string]any{"field": field, "op": "eq", "value": value}
	}
	return nil
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

// Find returns records matching a simple {field: value} equality filter.
// Uses a server-side index when available; falls back to a full strand scan.
// limit defaults to 20 when <= 0.
func (c *CollectionClient) Find(ctx context.Context, filter map[string]any, limit int) (*QueryResult, error) {
	if limit <= 0 {
		limit = 20
	}
	var body map[string]any
	if len(filter) > 0 {
		body = map[string]any{
			"type":   "scan",
			"limit":  limit,
			"filter": buildFilterBody(filter),
		}
	} else {
		body = map[string]any{"type": "latest", "limit": limit}
	}
	var out QueryResult
	if err := c.http.do(ctx, "POST", "/v1/agents/"+c.name+"/query", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FindOne returns the first record matching filter, or nil when nothing matches.
func (c *CollectionClient) FindOne(ctx context.Context, filter map[string]any) (*RecordView, error) {
	res, err := c.Find(ctx, filter, 1)
	if err != nil {
		return nil, err
	}
	if len(res.Records) == 0 {
		return nil, nil
	}
	return &res.Records[0], nil
}

// AsOf returns a CollectionQuery scoped to a specific HLC timestamp.
// Pass a RecordView.TimestampHLC value to replay the collection at that instant.
func (c *CollectionClient) AsOf(timestampHLC uint64) *CollectionQuery {
	return &CollectionQuery{http: c.http, name: c.name, timestampHLC: timestampHLC}
}

// History returns records in chronological order (oldest first).
// limit defaults to 20 when <= 0; offset is zero-based.
func (c *CollectionClient) History(ctx context.Context, limit, offset int) (*QueryResult, error) {
	if limit <= 0 {
		limit = 20
	}
	path := fmt.Sprintf("/v1/agents/%s/strand/records?limit=%d&offset=%d", c.name, limit, offset)
	var out QueryResult
	if err := c.http.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete soft-deletes a record by content hash. Requires an admin-scoped
// client. Writes a tombstone — the record is excluded from Current/CountCurrent
// going forward, but remains readable via ListDeleted for audit and can be
// brought back with Restore.
func (c *CollectionClient) Delete(ctx context.Context, contentHash string) (*SoftDeleteResponse, error) {
	var out SoftDeleteResponse
	if err := c.http.do(ctx, "DELETE", "/v1/agents/"+c.name+"/records/"+contentHash, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListDeleted returns every soft-deleted record for this collection, newest
// first, with the original payload joined in where still readable.
func (c *CollectionClient) ListDeleted(ctx context.Context) ([]DeletedRecord, error) {
	var out deletedRecordsResponse
	if err := c.http.do(ctx, "GET", "/v1/agents/"+c.name+"/records/deleted", nil, &out); err != nil {
		return nil, err
	}
	return out.Records, nil
}

// Restore re-appends a record's original payload as a brand-new active
// record — not an in-place undelete. Works on any valid content hash,
// regardless of whether it was ever soft-deleted.
func (c *CollectionClient) Restore(ctx context.Context, contentHash string) (*RestoreResponse, error) {
	var out RestoreResponse
	if err := c.http.do(ctx, "POST", "/v1/agents/"+c.name+"/records/"+contentHash+"/restore", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Current returns the latest record per distinct value of key — the
// current-state view. Tombstoned records (including ones soft-deleted via
// Delete) are excluded.
func (c *CollectionClient) Current(ctx context.Context, key string, limit int) (*QueryResult, error) {
	if limit <= 0 {
		limit = 20
	}
	return c.Query(ctx, map[string]any{"type": "current", "key": key, "limit": limit})
}

// CountCurrent counts records in the current-state view (collapsed +
// tombstone-filtered, same semantics as Current).
func (c *CollectionClient) CountCurrent(ctx context.Context, key string) (int, error) {
	var out struct {
		Count int `json:"count"`
	}
	body := map[string]any{"type": "count_current", "key": key}
	if err := c.http.do(ctx, "POST", "/v1/agents/"+c.name+"/query", body, &out); err != nil {
		return 0, err
	}
	return out.Count, nil
}
