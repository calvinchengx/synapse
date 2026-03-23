package server

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Recovery returns a middleware that recovers from panics.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[PANIC] %v", r)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "internal server error",
				})
			}
		}()
		c.Next()
	}
}

// RequestID injects an X-Request-ID header.
func RequestID() gin.HandlerFunc {
	counter := uint64(0)
	return func(c *gin.Context) {
		counter++
		id := fmt.Sprintf("req-%d-%d", time.Now().UnixNano(), counter)
		c.Set("requestID", id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

// Logger logs method, path, status, and latency.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Printf("[HTTP] %s %s %d %s",
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			time.Since(start),
		)
	}
}

// SecurityHeaders sets conservative security response headers.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self' 'unsafe-inline'; "+
				"style-src 'self' 'unsafe-inline'; "+
				"img-src 'self' data:;")
		c.Next()
	}
}

// CORS sets Access-Control-Allow-Origin based on allowed origins.
// Pass nil or empty slice to deny all cross-origin requests.
func CORS(allowedOrigins []string) gin.HandlerFunc {
	originSet := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		originSet[o] = true
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && originSet[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
			c.Header("Access-Control-Max-Age", "86400")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// HostCheck rejects requests whose Host header is not in allowedHosts.
// If allowedHosts is empty, all hosts are allowed (useful in tests).
func HostCheck(allowedHosts []string) gin.HandlerFunc {
	if len(allowedHosts) == 0 {
		return func(c *gin.Context) { c.Next() }
	}

	hostSet := make(map[string]bool, len(allowedHosts))
	for _, h := range allowedHosts {
		hostSet[strings.ToLower(h)] = true
	}

	return func(c *gin.Context) {
		host := strings.ToLower(c.Request.Host)
		// strip port if present
		if i := strings.LastIndex(host, ":"); i >= 0 {
			host = host[:i]
		}
		if !hostSet[host] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "host not allowed"})
			return
		}
		c.Next()
	}
}
