package server

import (
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
)

// handleStatus returns a lightweight health/status payload.
func (s *Server) handleStatus(c *gin.Context) {
	status := gin.H{
		"version":    "dev",
		"rules":      s.rulesEng.Count(),
		"categories": len(s.rulesEng.Categories()),
		"go_version": runtime.Version(),
	}
	c.JSON(http.StatusOK, status)
}

// handleDoctor runs the diagnostics checks and reports them.
func (s *Server) handleDoctor(c *gin.Context) {
	type check struct {
		Name   string `json:"name"`
		OK     bool   `json:"ok"`
		Detail string `json:"detail,omitempty"`
	}

	checks := []check{
		{
			Name:   "rules_loaded",
			OK:     s.rulesEng.Count() > 0,
			Detail: "rules engine has at least one rule",
		},
		{
			Name:   "integrations_loaded",
			OK:     len(s.manifests) > 0,
			Detail: "at least one integration manifest loaded",
		},
	}

	allOK := true
	for _, ch := range checks {
		if !ch.OK {
			allOK = false
		}
	}

	status := http.StatusOK
	if !allOK {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, gin.H{
		"ok":     allOK,
		"checks": checks,
	})
}
