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
	indegree := make(map[string]int, len(g.Nodes))
	children := make(map[string][]string, len(g.Nodes))
	for id := range g.Nodes {
		indegree[id] = 0
	}
	for _, edge := range g.Edges {
		indegree[edge.To]++
		children[edge.From] = append(children[edge.From], edge.To)
	}
	queue := make([]string, 0, len(g.Nodes))
	for id, degree := range indegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}
	sort.Strings(queue)
	order := make([]string, 0, len(g.Nodes))
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		order = append(order, id)
		for _, child := range children[id] {
			indegree[child]--
			if indegree[child] == 0 {
				queue = append(queue, child)
				sort.Strings(queue)
			}
		}
	}
	if len(order) == len(g.Nodes) {
		return order, nil
	}
	cycle := make([]string, 0)
	for id, degree := range indegree {
		if degree > 0 {
			cycle = append(cycle, id)
		}
	}
	sort.Strings(cycle)
	return order, cycle
}
