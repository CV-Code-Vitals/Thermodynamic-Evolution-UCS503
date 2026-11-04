package httprouter

import (
	"strings"
	"sync"
)

// TrieNode represents a node in the route matching trie.
type TrieNode struct {
	path      string
	part      string
	children  map[string]*TrieNode
	isParam   bool
	isWild    bool
	handler   HandlerFunc
	mu        sync.RWMutex
}

// NewTrieNode creates an initialized TrieNode.
func NewTrieNode() *TrieNode {
	return &TrieNode{
		children: make(map[string]*TrieNode),
	}
}

// Insert inserts a route path and associates it with a handler.
func (n *TrieNode) Insert(pattern string, handler HandlerFunc) {
	n.mu.Lock()
	defer n.mu.Unlock()

	parts := parsePattern(pattern)
	curr := n

	for _, part := range parts {
		child, exists := curr.children[part]
		if !exists {
			isParam := strings.HasPrefix(part, ":")
			isWild := strings.HasPrefix(part, "*")
			child = &TrieNode{
				part:     part,
				children: make(map[string]*TrieNode),
				isParam:  isParam,
				isWild:   isWild,
			}
			curr.children[part] = child
		}
		curr = child
	}
	curr.path = pattern
	curr.handler = handler
}

// Search searches for a route matching the given path parts and extracts parameters.
func (n *TrieNode) Search(parts []string) (*TrieNode, map[string]string) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	curr := n
	params := make(map[string]string)

	for i, part := range parts {
		var next *TrieNode
		for p, child := range curr.children {
			if p == part {
				next = child
				break
			} else if child.isParam {
				next = child
				paramKey := strings.TrimPrefix(child.part, ":")
				params[paramKey] = part
				break
			} else if child.isWild {
				wildKey := strings.TrimPrefix(child.part, "*")
				params[wildKey] = strings.Join(parts[i:], "/")
				return child, params
			}
		}

		if next == nil {
			return nil, nil
		}
		curr = next
	}

	if curr.handler == nil {
		return nil, nil
	}
	return curr, params
}

func parsePattern(pattern string) []string {
	parts := strings.Split(pattern, "/")
	var res []string
	for _, p := range parts {
		if p != "" {
			res = append(res, p)
		}
	}
	return res
}

