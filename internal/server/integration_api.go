package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/calvinchengx/synapse/internal/integration"
)

// handleListIntegrations returns all loaded tool manifests with discovery status.
func (s *Server) handleListIntegrations(c *gin.Context) {
	results := make([]gin.H, 0, len(s.manifests))
	for _, m := range s.manifests {
		status := integration.Discover(m)
		results = append(results, gin.H{
			"id":          m.ID,
			"name":        m.Name,
			"description": m.Description,
			"binary":      m.Detection.Binary,
			"available":   status.BinaryFound,
			"data_found":  len(status.DataFilesFound),
			"data_missing": len(status.DataFilesMissing),
		})
	}
	c.JSON(http.StatusOK, gin.H{"integrations": results, "count": len(results)})
}

// handleGetIntegration returns detail for a single integration by ID.
func (s *Server) handleGetIntegration(c *gin.Context) {
	id := c.Param("id")
	for _, m := range s.manifests {
		if m.ID != id {
			continue
		}
		status := integration.Discover(m)
		c.JSON(http.StatusOK, gin.H{
			"id":          m.ID,
			"name":        m.Name,
			"description": m.Description,
			"detection":   m.Detection,
			"status":      status,
		})
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "integration not found"})
}
