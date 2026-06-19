package sapixdb

import (
	"context"
	"fmt"
)

// TraverseOptions controls how the agent graph is walked.
type TraverseOptions struct {
	// Depth is the maximum hop count (default 1, max 3).
	Depth int
	// EdgeType filters traversal to edges of this type only.
	EdgeType string
}

// GraphClient provides graph relationship operations.
type GraphClient struct {
	http  *httpClient
	agent string
}

// AddEdge creates a directed edge between two agents or records.
// src and dst are agent IDs or content hashes.
func (g *GraphClient) AddEdge(ctx context.Context, src, dst, edgeType string, weight float64) error {
	body := map[string]any{"src": src, "dst": dst, "edge_type": edgeType, "weight": weight}
	return g.http.do(ctx, "POST", "/v1/graph/edges", body, nil)
}

// Relate is a shorthand for AddEdge with weight 1.0.
func (g *GraphClient) Relate(ctx context.Context, src, dst, edgeType string) error {
	return g.AddEdge(ctx, src, dst, edgeType, 1.0)
}

// RemoveEdge deletes a directed edge.
func (g *GraphClient) RemoveEdge(ctx context.Context, src, edgeType, dst string) error {
	return g.http.do(ctx, "DELETE", fmt.Sprintf("/v1/graph/edges/%s/%s/%s", src, edgeType, dst), nil, nil)
}

// Edges returns all outbound edges from an agent.
func (g *GraphClient) Edges(ctx context.Context, agentID string) ([]GraphEdge, error) {
	var out []GraphEdge
	if err := g.http.do(ctx, "GET", "/v1/graph/edges/"+agentID, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// InboundEdges returns all inbound edges to an agent.
func (g *GraphClient) InboundEdges(ctx context.Context, agentID string) ([]GraphEdge, error) {
	var out []GraphEdge
	if err := g.http.do(ctx, "GET", "/v1/graph/edges/"+agentID+"?direction=in", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Traverse walks the agent graph from agentID up to opts.Depth hops.
func (g *GraphClient) Traverse(ctx context.Context, agentID string, opts TraverseOptions) (*TraversalResult, error) {
	depth := opts.Depth
	if depth == 0 {
		depth = 1
	}
	path := fmt.Sprintf("/v1/graph/traverse?agent_id=%s&depth=%d", agentID, depth)
	if opts.EdgeType != "" {
		path += "&edge_type=" + opts.EdgeType
	}
	var out TraversalResult
	if err := g.http.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AddRecordRef creates a cross-agent record reference edge.
func (g *GraphClient) AddRecordRef(ctx context.Context, srcAgent, srcHash, edgeType, dstAgent, dstHash string) error {
	body := map[string]any{
		"src_agent":        srcAgent,
		"src_content_hash": srcHash,
		"edge_type":        edgeType,
		"dst_agent":        dstAgent,
		"dst_content_hash": dstHash,
	}
	return g.http.do(ctx, "POST", "/v1/graph/refs", body, nil)
}
