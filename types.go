package sapixdb

// WriteResult is returned by every write operation.
type WriteResult struct {
	ID         string `json:"id"`
	Hash       string `json:"hash"`
	PrevHash   string `json:"prev_hash,omitempty"`
	Timestamp  string `json:"timestamp"`
	Collection string `json:"collection"`
}

// NucleotideRecord is a single stored record.
type NucleotideRecord struct {
	ID         string         `json:"id"`
	Data       map[string]any `json:"data"`
	Timestamp  string         `json:"timestamp"`
	Hash       string         `json:"hash"`
	PrevHash   string         `json:"prev_hash,omitempty"`
	Collection string         `json:"collection"`
}

// GraphEdge is a directed relationship between two records.
type GraphEdge struct {
	Src          string  `json:"src"`
	Dst          string  `json:"dst"`
	EdgeType     string  `json:"edge_type"`
	Weight       float64 `json:"weight"`
	TimestampHLC *int64  `json:"timestamp_hlc,omitempty"`
}

// TraverseResult is the result of a graph traversal.
type TraverseResult struct {
	Nodes []NucleotideRecord `json:"nodes"`
	Edges []GraphEdge        `json:"edges"`
}

// HealthResponse is returned by the health endpoint.
type HealthResponse struct {
	Status string `json:"status"`
	Agent  string `json:"agent"`
}
