package server

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/calvinchengx/synapse/internal/rules"
)

type ruleResponse struct {
	Path        string   `json:"path"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Keywords    []string `json:"keywords,omitempty"`
	AlwaysApply bool     `json:"always_apply"`
}

func toResponse(r rules.Rule) ruleResponse {
	name := r.Frontmatter.Name
	if name == "" {
		name = filepath.Base(r.Path)
	}
	return ruleResponse{
		Path:        r.Path,
		Name:        name,
		Type:        string(r.Type),
		Description: r.Frontmatter.Description,
		Keywords:    r.Frontmatter.Keywords,
		AlwaysApply: r.Frontmatter.AlwaysApply,
	}
}

// handleListRules returns all rules, optionally filtered by type query param.
func (s *Server) handleListRules(c *gin.Context) {
	ruleType := c.Query("type")

	var list []rules.Rule
	if ruleType != "" {
		list = s.rulesEng.ByType(rules.RuleType(ruleType))
	} else {
		list = s.rulesEng.All()
	}

	resp := make([]ruleResponse, 0, len(list))
	for _, r := range list {
		resp = append(resp, toResponse(r))
	}
	c.JSON(http.StatusOK, gin.H{"rules": resp, "count": len(resp)})
}

// handleCategories returns all known categories.
func (s *Server) handleCategories(c *gin.Context) {
	cats := s.rulesEng.Categories()
	c.JSON(http.StatusOK, gin.H{"categories": cats})
}

// handleSearchRules searches rules by query string.
func (s *Server) handleSearchRules(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q parameter required"})
		return
	}

	matches := s.rulesEng.Search(q)
	resp := make([]ruleResponse, 0, len(matches))
	for _, r := range matches {
		resp = append(resp, toResponse(r))
	}
	c.JSON(http.StatusOK, gin.H{"results": resp, "count": len(resp), "query": q})
}
