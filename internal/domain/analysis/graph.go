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
	visited := make(map[string]bool, len(g.Nodes))
	active := make(map[string]bool, len(g.Nodes))
	order := make([]string, 0, len(g.Nodes))
	var visit func(string)
	visit = func(id string) {
		if visited[id] {
			return
		}
		if active[id] {
			return
		}
		active[id] = true
		for _, child := range children[id] {
			visit(child)
		}
		active[id] = false
		visited[id] = true
		order = append(order, id)
	}
	for _, id := range ids {
		visit(id)
	}
	for left, right := 0, len(order)-1; left < right; left, right = left+1, right-1 {
		order[left], order[right] = order[right], order[left]
	}
	return order, nil
}
