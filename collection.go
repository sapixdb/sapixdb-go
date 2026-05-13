package sapixdb

import (
	"context"
	"errors"
)

type queryBody struct {
	Collection string         `json:"collection"`
	Latest     bool           `json:"latest"`
	AsOf       string         `json:"as_of,omitempty"`
	Filter     map[string]any `json:"filter,omitempty"`
	Limit      *int           `json:"limit,omitempty"`
}

type queryResult struct {
	Results []NucleotideRecord `json:"results"`
}

// CollectionQuery is a time-scoped read-only view of a collection.
type CollectionQuery struct {
	http  *httpClient
	agent string
	name  string
	asOf  string
}

func (q *CollectionQuery) query(ctx context.Context, latest bool, filter map[string]any, limit int) ([]NucleotideRecord, error) {
	b := queryBody{Collection: q.name, Latest: latest, AsOf: q.asOf, Filter: filter}
	if limit > 0 {
		b.Limit = &limit
	}
	var out queryResult
	if err := q.http.do(ctx, "POST", "/v1/"+q.agent+"/strand/query", b, &out); err != nil {
		return nil, err
	}
	return out.Results, nil
}

// Latest returns the most recent version of every record as of the snapshot time.
func (q *CollectionQuery) Latest(ctx context.Context, filter map[string]any) ([]NucleotideRecord, error) {
	return q.query(ctx, true, filter, 0)
}

// All returns the full history as of the snapshot time.
func (q *CollectionQuery) All(ctx context.Context, filter map[string]any) ([]NucleotideRecord, error) {
	return q.query(ctx, false, filter, 0)
}

// Find filters records (latest version) as of the snapshot time.
func (q *CollectionQuery) Find(ctx context.Context, filter map[string]any, limit int) ([]NucleotideRecord, error) {
	return q.query(ctx, true, filter, limit)
}

// FindOne returns the first match or nil.
func (q *CollectionQuery) FindOne(ctx context.Context, filter map[string]any) (*NucleotideRecord, error) {
	results, err := q.query(ctx, true, filter, 1)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	return &results[0], nil
}

// CollectionClient provides read/write access to a named collection.
type CollectionClient struct {
	http  *httpClient
	agent string
	name  string
}

// AsOf scopes all reads to a specific point in time. Returns a CollectionQuery.
func (c *CollectionClient) AsOf(timestamp string) *CollectionQuery {
	return &CollectionQuery{http: c.http, agent: c.agent, name: c.name, asOf: timestamp}
}

// Write appends a new record. Nothing is ever overwritten.
func (c *CollectionClient) Write(ctx context.Context, data map[string]any) (*WriteResult, error) {
	body := map[string]any{"collection": c.name, "data": data}
	var out WriteResult
	if err := c.http.do(ctx, "POST", "/v1/"+c.agent+"/strand/write", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// WriteBatch appends multiple records sequentially.
func (c *CollectionClient) WriteBatch(ctx context.Context, records []map[string]any) ([]*WriteResult, error) {
	results := make([]*WriteResult, 0, len(records))
	for _, r := range records {
		wr, err := c.Write(ctx, r)
		if err != nil {
			return results, err
		}
		results = append(results, wr)
	}
	return results, nil
}

// Get fetches a record by ID. Returns *SapixNotFoundError if the record does not exist.
func (c *CollectionClient) Get(ctx context.Context, recordID string) (*NucleotideRecord, error) {
	var out NucleotideRecord
	err := c.http.do(ctx, "GET", "/v1/"+c.agent+"/strand/"+recordID, nil, &out)
	if err != nil {
		var sapixErr *SapixError
		if errors.As(err, &sapixErr) && sapixErr.Status == 404 {
			return nil, &SapixNotFoundError{RecordID: recordID}
		}
		return nil, err
	}
	return &out, nil
}

func (c *CollectionClient) query(ctx context.Context, latest bool, filter map[string]any, limit int) ([]NucleotideRecord, error) {
	b := queryBody{Collection: c.name, Latest: latest, Filter: filter}
	if limit > 0 {
		b.Limit = &limit
	}
	var out queryResult
	if err := c.http.do(ctx, "POST", "/v1/"+c.agent+"/strand/query", b, &out); err != nil {
		return nil, err
	}
	return out.Results, nil
}

// Latest returns the current (most recent) version of every record.
func (c *CollectionClient) Latest(ctx context.Context, filter map[string]any) ([]NucleotideRecord, error) {
	return c.query(ctx, true, filter, 0)
}

// History returns the full append-only history — every version ever written.
func (c *CollectionClient) History(ctx context.Context, filter map[string]any) ([]NucleotideRecord, error) {
	return c.query(ctx, false, filter, 0)
}

// Find filters records (latest version only).
func (c *CollectionClient) Find(ctx context.Context, filter map[string]any, limit int) ([]NucleotideRecord, error) {
	return c.query(ctx, true, filter, limit)
}

// FindOne returns the first matching record or nil.
func (c *CollectionClient) FindOne(ctx context.Context, filter map[string]any) (*NucleotideRecord, error) {
	results, err := c.query(ctx, true, filter, 1)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	return &results[0], nil
}
