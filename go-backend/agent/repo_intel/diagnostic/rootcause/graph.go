package rootcause

import (
	"fmt"
	"sort"
)

// RootCauseGraph maintains the graph of findings and causal edges.
type RootCauseGraph struct {
	Nodes    map[string]*GraphNode
	Edges    []GraphEdge
	adjList  map[string][]GraphEdge
	revAdj   map[string][]GraphEdge
}

func NewRootCauseGraph() *RootCauseGraph {
	return &RootCauseGraph{
		Nodes:   make(map[string]*GraphNode),
		adjList: make(map[string][]GraphEdge),
		revAdj:  make(map[string][]GraphEdge),
	}
}

func (g *RootCauseGraph) AddNode(node *GraphNode) {
	g.Nodes[node.ID] = node
}

func (g *RootCauseGraph) AddEdge(edge GraphEdge) {
	g.Edges = append(g.Edges, edge)
	g.adjList[edge.SourceID] = append(g.adjList[edge.SourceID], edge)
	g.revAdj[edge.TargetID] = append(g.revAdj[edge.TargetID], edge)
}

// FindConnectedComponents partitions the graph into weakly connected clusters.
func (g *RootCauseGraph) FindConnectedComponents() [][]string {
	visited := make(map[string]bool)
	var components [][]string

	undirectedAdj := make(map[string][]string)
	for _, e := range g.Edges {
		undirectedAdj[e.SourceID] = append(undirectedAdj[e.SourceID], e.TargetID)
		undirectedAdj[e.TargetID] = append(undirectedAdj[e.TargetID], e.SourceID)
	}

	for nodeID := range g.Nodes {
		if !visited[nodeID] {
			var comp []string
			var queue []string
			queue = append(queue, nodeID)
			visited[nodeID] = true

			for len(queue) > 0 {
				curr := queue[0]
				queue = queue[1:]
				comp = append(comp, curr)

				for _, neighbor := range undirectedAdj[curr] {
					if !visited[neighbor] {
						visited[neighbor] = true
						queue = append(queue, neighbor)
					}
				}
			}

			sort.Strings(comp)
			components = append(components, comp)
		}
	}

	return components
}

// TopologicalSort returns nodes in dependency order or an error if a cycle exists.
func (g *RootCauseGraph) TopologicalSort() ([]string, error) {
	inDegree := make(map[string]int)
	for nodeID := range g.Nodes {
		inDegree[nodeID] = 0
	}
	for _, edge := range g.Edges {
		inDegree[edge.TargetID]++
	}

	var queue []string
	for nodeID, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, nodeID)
		}
	}
	sort.Strings(queue)

	var ordered []string
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		ordered = append(ordered, curr)

		for _, edge := range g.adjList[curr] {
			inDegree[edge.TargetID]--
			if inDegree[edge.TargetID] == 0 {
				queue = append(queue, edge.TargetID)
			}
		}
	}

	if len(ordered) != len(g.Nodes) {
		return ordered, fmt.Errorf("causal cycle detected in graph: %d of %d nodes sorted", len(ordered), len(g.Nodes))
	}

	return ordered, nil
}

