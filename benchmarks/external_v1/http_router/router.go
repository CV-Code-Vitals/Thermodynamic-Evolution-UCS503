package httprouter

import (
	"net/http"
	"strings"
	"sync"
)

// HandlerFunc defines the request handler signature.
type HandlerFunc func(*Context)

// Engine is the core router engine implementing http.Handler.
type Engine struct {
	trees       map[string]*TrieNode
	middlewares []HandlerFunc
	mu          sync.RWMutex
}

// New creates an initialized Engine.
func New() *Engine {
	return &Engine{
		trees: make(map[string]*TrieNode),
	}
}

// Use adds global middlewares to the router engine.
func (e *Engine) Use(middlewares ...HandlerFunc) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.middlewares = append(e.middlewares, middlewares...)
}

// Handle registers a new request handle with the given path and method.
func (e *Engine) Handle(method, path string, handler HandlerFunc) {
	e.mu.Lock()
	defer e.mu.Unlock()

	method = strings.ToUpper(method)
	root, ok := e.trees[method]
	if !ok {
		root = NewTrieNode()
		e.trees[method] = root
	}
	root.Insert(path, handler)
}

// GET is a shortcut for Handle("GET", path, handler).
func (e *Engine) GET(path string, handler HandlerFunc) {
	e.Handle(http.MethodGet, path, handler)
}

// POST is a shortcut for Handle("POST", path, handler).
func (e *Engine) POST(path string, handler HandlerFunc) {
	e.Handle(http.MethodPost, path, handler)
}

// ServeHTTP conforms to the http.Handler interface.
func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := NewContext(w, r)

	e.mu.RLock()
	c.handlers = append(c.handlers, e.middlewares...)
	root, ok := e.trees[r.Method]
	e.mu.RUnlock()

	if ok {
		parts := parsePattern(r.URL.Path)
		node, params := root.Search(parts)
		if node != nil && node.handler != nil {
			c.Params = params
			c.handlers = append(c.handlers, node.handler)
			c.Next()
			return
		}
	}

	c.Status(http.StatusNotFound)
	w.Write([]byte("404 page not found\n"))
}

