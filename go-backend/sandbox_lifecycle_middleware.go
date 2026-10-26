package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func sandboxLifecycleMiddleware(lifecycle SandboxLifecycleController) gin.HandlerFunc {
	return func(c *gin.Context) {
		if lifecycle == nil {
			c.Next()
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), maxEvolutionDuration)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)

		release, err := lifecycle.Acquire(ctx)
		if err != nil {
			log.Printf("sandbox lifecycle admission failed: %v", err)
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "sandbox_unavailable",
			})
			return
		}

		c.Next()
		if err := release(); err != nil {
			log.Printf("sandbox lifecycle lease release failed: %v", err)
		}
	}
}
