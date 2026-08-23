package analysis

import (
	"fmt"
	"sort"
)

type Node struct {
	ArtifactID string `json:"artifact_id"`
	Digest     string `json:"digest"`
}

type Edge struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Contract string `json:"contract"`
	Implicit bool   `json:"implicit"`
	Evidence string `json:"evidence"`
}

type Graph struct {
	Nodes map[string]Node `json:"nodes"`
	Edges []Edge          `json:"edges"`
}

func NewGraph(nodes []Node, edges []Edge) (Graph, error) {
	graph := Graph{Nodes: make(map[string]Node, len(nodes)), Edges: append([]Edge(nil), edges...)}
	for _, node := range nodes {
		if node.ArtifactID == "" || node.Digest == "" {
			return Graph{}, fmt.Errorf("graph node identity and digest are required")
		}
		if _, exists := graph.Nodes[node.ArtifactID]; exists {
			return Graph{}, fmt.Errorf("duplicate graph node %s", node.ArtifactID)
		}
		graph.Nodes[node.ArtifactID] = node
	}
	for _, edge := range edges {
		if _, ok := graph.Nodes[edge.From]; !ok {
			return Graph{}, fmt.Errorf("unknown edge source %s", edge.From)
		}
		if _, ok := graph.Nodes[edge.To]; !ok {
			return Graph{}, fmt.Errorf("unknown edge target %s", edge.To)
		}
	}
	return graph, nil
}

// TopologicalOrder returns a deterministic execution order for the graph's
// nodes. When the graph contains a cycle, no valid order exists: the returned
// order is empty and the second result lists the artifact ids that
// participate in the detected cycle, so callers can surface them as evidence.
func (g Graph) TopologicalOrder() ([]string, []string) {
	children := make(map[string][]string, len(g.Nodes))
	for _, edge := range g.Edges {
		children[edge.From] = append(children[edge.From], edge.To)
	}
	for id := range children {
		sort.Strings(children[id])
	}
	ids := make([]string, 0, len(g.Nodes))
	for id := range g.Nodes {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	const (
		white = 0 // unvisited
		gray  = 1 // on the recursion stack
		black = 2 // fully processed
	)
	color := make(map[string]int, len(g.Nodes))
	order := make([]string, 0, len(g.Nodes))
	var cycle []string
	var stack []string
	var visit func(string)
	visit = func(id string) {
		color[id] = gray
		stack = append(stack, id)
		for _, child := range children[id] {
			switch color[child] {
			case black:
			case gray:
				// A back edge closes a cycle. Recover the participating nodes by
				// walking the current recursion stack from the first occurrence
				// of the back edge's target up to the node that reached it.
				for i := 0; i < len(stack); i++ {
					if stack[i] == child {
						cycle = append(cycle, stack[i:]...)
						break
					}
				}
			default:
				visit(child)
			}
		}
		stack = stack[:len(stack)-1]
		color[id] = black
		order = append(order, id)
	}
	for _, id := range ids {
		if color[id] == white {
			visit(id)
		}
	}
	if len(cycle) > 0 {
		// A cyclic graph has no meaningful execution order; keep only the
		// participating nodes so downstream tooling cannot mistake a partial
		// ordering for a trustworthy check sequence.
		sort.Strings(cycle)
		return nil, cycle
	}
	for left, right := 0, len(order)-1; left < right; left, right = left+1, right-1 {
		order[left], order[right] = order[right], order[left]
	}
	return order, nil
}
