package httprouter

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// Logger returns a middleware that logs request execution time.
func Logger() HandlerFunc {
	return func(c *Context) {
		start := time.Now()
		c.Next()
		log.Printf("[%s] %s - %d (%s)", c.Request.Method, c.Request.URL.Path, c.StatusCode, time.Since(start))
	}
}

// Recovery returns a middleware that catches panics and returns 500.
func Recovery() HandlerFunc {
	return func(c *Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic recovered: %v", r)
				c.Status(http.StatusInternalServerError)
				c.Writer.Write([]byte("Internal Server Error"))
			}
		}()
		c.Next()
	}
}

// UpstreamProbe calls an upstream URL to verify backend health.
// DEFECT: http.Get response body is not closed, leaking network connection / descriptor resources.
func UpstreamProbe(probeURL string) (int, error) {
	resp, err := http.Get(probeURL)
	if err != nil {
		return 0, fmt.Errorf("probe request failed: %w", err)
	}
	// Missing: defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, err
	}
	if len(body) == 0 {
		return resp.StatusCode, fmt.Errorf("empty probe response")
	}
	return resp.StatusCode, nil
}

