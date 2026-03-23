package server

import (
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// handleSPA serves static assets from the embedded dist FS.
// For paths that don't match a real file it falls back to index.html
// so the Svelte router can handle client-side navigation.
func (s *Server) handleSPA(c *gin.Context) {
	if s.assets == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	// Only serve GET/HEAD for SPA
	if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
		return
	}

	path := strings.TrimPrefix(c.Request.URL.Path, "/")
	if path == "" {
		path = "index.html"
	}

	data, err := fs.ReadFile(s.assets, path)
	if err != nil {
		// Fallback to index.html for SPA routing
		data, err = fs.ReadFile(s.assets, "index.html")
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		path = "index.html"
	}

	ext := filepath.Ext(path)
	ct := mime.TypeByExtension(ext)
	if ct == "" {
		ct = "application/octet-stream"
	}

	c.Data(http.StatusOK, ct, data)
}
