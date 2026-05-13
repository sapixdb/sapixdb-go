package sapixdb

import "context"

// TraverseOptions controls how the graph is walked.
type TraverseOptions struct {
	Depth     int    // default 1
	Direction string // "outbound" | "inbound" | "both" — default "outbound"
}

// GraphClient provides graph relationship operations.
type GraphClient struct {
	http  *httpClient
	agent string
}

// Relate creates a typed directed edge between two records. Shorthand for AddEdge.
func (g *GraphClient) Relate(ctx context.Context, src, dst, edgeType string, weight float64) error {
	return g.AddEdge(ctx, src, dst, edgeType, weight)
}

// AddEdge creates a directed edge with a given weight.
func (g *GraphClient) AddEdge(ctx context.Context, src, dst, edgeType string, weight float64) error {
	body := map[string]any{"src": src, "dst": dst, "edge_type": edgeType, "weight": weight}
	return g.http.do(ctx, "POST", "/v1/"+g.agent+"/graph/edge", body, nil)
}

// RemoveEdge deletes an edge.
func (g *GraphClient) RemoveEdge(ctx context.Context, src, dst, edgeType string) error {
	body := map[string]any{"src": src, "dst": dst, "edge_type": edgeType}
	return g.http.do(ctx, "DELETE", "/v1/"+g.agent+"/graph/edge", body, nil)
}

// Traverse walks the graph from fromID using the given options.
func (g *GraphClient) Traverse(ctx context.Context, fromID string, opts TraverseOptions) (*TraverseResult, error) {
	depth := opts.Depth
	if depth == 0 {
		depth = 1
	}
	dir := opts.Direction
	if dir == "" {
		dir = "outbound"
	}
	body := map[string]any{"from": fromID, "depth": depth, "direction": dir}
	var out TraverseResult
	if err := g.http.do(ctx, "POST", "/v1/"+g.agent+"/graph/traverse", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Neighbors returns direct neighbours (depth=1 shortcut).
func (g *GraphClient) Neighbors(ctx context.Context, nodeID, direction string) ([]NucleotideRecord, error) {
	if direction == "" {
		direction = "outbound"
	}
	result, err := g.Traverse(ctx, nodeID, TraverseOptions{Depth: 1, Direction: direction})
	if err != nil {
		return nil, err
	}
	return result.Nodes, nil
}

// Edges returns all outbound edges from a node.
func (g *GraphClient) Edges(ctx context.Context, nodeID string) ([]GraphEdge, error) {
	var out struct {
		Edges []GraphEdge `json:"edges"`
	}
	if err := g.http.do(ctx, "GET", "/v1/"+g.agent+"/graph/edges/"+nodeID, nil, &out); err != nil {
		return nil, err
	}
	return out.Edges, nil
}
