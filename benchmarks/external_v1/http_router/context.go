package httprouter

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Context wraps http.ResponseWriter and http.Request for a given request.
type Context struct {
	Writer     http.ResponseWriter
	Request    *http.Request
	Params     map[string]string
	StatusCode int
	handlers   []HandlerFunc
	index      int
}

// NewContext creates a new handler context.
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Writer:     w,
		Request:    r,
		Params:     make(map[string]string),
		StatusCode: http.StatusOK,
		index:      -1,
	}
}

// Next invokes the next middleware or handler in the chain.
func (c *Context) Next() {
	c.index++
	for c.index < len(c.handlers) {
		c.handlers[c.index](c)
		c.index++
	}
}

// Param returns a route parameter by key.
func (c *Context) Param(key string) string {
	return c.Params[key]
}

// Status sets the HTTP status code.
func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

// Header sets a response header.
func (c *Context) Header(key, value string) {
	c.Writer.Header().Set(key, value)
}

// JSON sends a JSON response.
// DEFECT: json.NewEncoder(c.Writer).Encode(data) return value is ignored without error checking.
func (c *Context) JSON(code int, data interface{}) {
	c.Header("Content-Type", "application/json")
	c.Status(code)
	json.NewEncoder(c.Writer).Encode(data)
}

// ExecuteWithTimeout executes a callback with a timeout context.
// DEFECT: context.WithTimeout creates a cancel function that is never called, causing a timer leak.
func (c *Context) ExecuteWithTimeout(d time.Duration, fn func(ctx context.Context) error) error {
	ctx, _ := context.WithTimeout(c.Request.Context(), d)
	return fn(ctx)
}

