package sapixdb

// WriteResponse is returned by every write operation.
type WriteResponse struct {
	RecordID    string `json:"record_id"`
	ContentHash string `json:"content_hash"`
	ChainHead   string `json:"chain_head"`
}

// RecordView is a single strand record returned by reads and queries.
type RecordView struct {
	RecordID      string `json:"record_id"`
	ContentHash   string `json:"content_hash"`
	ParentHash    string `json:"parent_hash"`
	TimestampHLC  uint64 `json:"timestamp_hlc"`
	TimestampMS   int64  `json:"timestamp_ms"`
	PayloadB64    string `json:"payload_b64"`
	Payload       any    `json:"payload,omitempty"`
	Flags         int    `json:"flags"`
	SchemaVersion int    `json:"schema_version"`
}

// ChainHeadResponse is returned by the strand head endpoints.
type ChainHeadResponse struct {
	ChainHead   string `json:"chain_head"`
	RecordCount uint64 `json:"record_count"`
}

// GraphEdge is a directed relationship between two agents.
type GraphEdge struct {
	Src          string  `json:"src"`
	Dst          string  `json:"dst"`
	EdgeType     string  `json:"edge_type"`
	TimestampHLC uint64  `json:"timestamp_hlc"`
	Weight       float64 `json:"weight"`
}

// TraversalEdge is an edge returned by GET /v1/graph/traverse.
type TraversalEdge struct {
	Src          string  `json:"src"`
	Dst          string  `json:"dst"`
	EdgeType     string  `json:"edge_type"`
	Weight       float32 `json:"weight"`
	AcrossPeer   bool    `json:"across_peer"`
	PeerEndpoint *string `json:"peer_endpoint,omitempty"`
}

// TraversalResult is returned by GET /v1/graph/traverse.
type TraversalResult struct {
	StartAgent    string          `json:"start_agent"`
	AgentsVisited []string        `json:"agents_visited"`
	Edges         []TraversalEdge `json:"edges"`
	DepthReached  int             `json:"depth_reached"`
}

// HealthResponse is returned by GET /v1/health.
type HealthResponse struct {
	Status      string `json:"status"`
	AgentID     string `json:"agent_id"`
	RecordCount uint64 `json:"record_count"`
	ChainHead   string `json:"chain_head"`
}

// QueryResult is the response from POST /v1/query and POST /v1/agents/:id/query.
type QueryResult struct {
	Records []RecordView `json:"records"`
	Total   int          `json:"total"`
}

// SemanticQueryResult is the response from POST /v1/query/semantic.
type SemanticQueryResult struct {
	Source  string       `json:"source"` // "cache" | "heuristic" | "model"
	Plan    any          `json:"plan"`
	Records []RecordView `json:"records"`
}

// AgentListResponse is returned by GET /v1/agents.
type AgentListResponse struct {
	Agents []string `json:"agents"`
	Total  int      `json:"total"`
}

// CreateAgentResponse is returned by POST /v1/agents.
type CreateAgentResponse struct {
	AgentID         string `json:"agent_id"`
	PublicKeyHex    string `json:"public_key_hex"`
	GenesisRecordID string `json:"genesis_record_id"`
	ChainHeadHex    string `json:"chain_head_hex"`
	Zone            string `json:"zone"`
}

// AgentStatusResponse is returned by GET /v1/agents/:id/status.
type AgentStatusResponse struct {
	AgentID     string `json:"agent_id"`
	RecordCount uint64 `json:"record_count"`
	ChainHead   string `json:"chain_head"`
}
