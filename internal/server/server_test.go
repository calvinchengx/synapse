package server_test

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"

	"github.com/calvinchengx/synapse/internal/integration"
	"github.com/calvinchengx/synapse/internal/rules"
	"github.com/calvinchengx/synapse/internal/server"
)

func testRules() []rules.Rule {
	return []rules.Rule{
		{
			Path: "security.md",
			Type: rules.RuleTypeRule,
			Frontmatter: rules.Frontmatter{
				Name:        "security",
				Description: "Security guidelines",
				Keywords:    []string{"security", "auth"},
				AlwaysApply: true,
			},
		},
		{
			Path: "testing.md",
			Type: rules.RuleTypeRule,
			Frontmatter: rules.Frontmatter{
				Name:        "testing",
				Description: "Testing guidelines",
				Keywords:    []string{"test", "coverage"},
			},
		},
		{
			Path: "react.md",
			Type: rules.RuleTypeSkill,
			Frontmatter: rules.Frontmatter{
				Name:        "react",
				Description: "React skill",
				Keywords:    []string{"react", "frontend"},
			},
		},
	}
}

func testManifests() []integration.ToolManifest {
	return []integration.ToolManifest{
		{
			ID:          "rtk",
			Name:        "RTK",
			Description: "Roo Task Knowledge",
			Detection: integration.Detection{
				Binary: "rtk",
			},
		},
	}
}

func newTestServer() *server.Server {
	eng := rules.NewEngineFromRules(testRules())
	cfg := server.Config{
		Host: "127.0.0.1",
		Port: 0,
	}
	return server.New(cfg, eng, testManifests(), nil)
}

func do(t *testing.T, s *server.Server, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	s.Engine().ServeHTTP(w, req)
	return w
}

// --- status ---

func TestStatus(t *testing.T) {
	s := newTestServer()
	w := do(t, s, "GET", "/api/status")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["rules"] == nil {
		t.Error("missing rules field")
	}
}

func TestDoctor(t *testing.T) {
	s := newTestServer()
	w := do(t, s, "GET", "/api/doctor")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
}

// --- rules ---

func TestListRules(t *testing.T) {
	s := newTestServer()
	w := do(t, s, "GET", "/api/rules")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	count := int(body["count"].(float64))
	if count != 3 {
		t.Errorf("expected 3 rules, got %d", count)
	}
}

func TestListRulesByType(t *testing.T) {
	s := newTestServer()
	w := do(t, s, "GET", "/api/rules?type=skill")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var body map[string]any
	json.Unmarshal(w.Body.Bytes(), &body) //nolint
	if int(body["count"].(float64)) != 1 {
		t.Errorf("expected 1 skill, got %v", body["count"])
	}
}

func TestListCategories(t *testing.T) {
	s := newTestServer()
	w := do(t, s, "GET", "/api/rules/categories")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestSearchRules(t *testing.T) {
	s := newTestServer()
	w := do(t, s, "GET", "/api/rules/search?q=security")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var body map[string]any
	json.Unmarshal(w.Body.Bytes(), &body) //nolint
	if int(body["count"].(float64)) < 1 {
		t.Error("expected at least 1 search result")
	}
}

func TestSearchRulesMissingQ(t *testing.T) {
	s := newTestServer()
	w := do(t, s, "GET", "/api/rules/search")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// --- integrations ---

func TestListIntegrations(t *testing.T) {
	s := newTestServer()
	w := do(t, s, "GET", "/api/integrations")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var body map[string]any
	json.Unmarshal(w.Body.Bytes(), &body) //nolint
	if int(body["count"].(float64)) != 1 {
		t.Errorf("expected 1 integration, got %v", body["count"])
	}
}

func TestGetIntegration(t *testing.T) {
	s := newTestServer()
	w := do(t, s, "GET", "/api/integrations/rtk")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
}

func TestGetIntegrationNotFound(t *testing.T) {
	s := newTestServer()
	w := do(t, s, "GET", "/api/integrations/unknown")
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// --- SPA fallback ---

func TestSPANilAssets(t *testing.T) {
	s := newTestServer()
	w := do(t, s, "GET", "/some/spa/route")
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for nil assets, got %d", w.Code)
	}
}

// --- middleware ---

func TestSecurityHeaders(t *testing.T) {
	s := newTestServer()
	w := do(t, s, "GET", "/api/status")
	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("missing X-Content-Type-Options header")
	}
	if w.Header().Get("X-Frame-Options") != "DENY" {
		t.Error("missing X-Frame-Options header")
	}
}

func TestRequestIDHeader(t *testing.T) {
	s := newTestServer()
	w := do(t, s, "GET", "/api/status")
	if w.Header().Get("X-Request-ID") == "" {
		t.Error("missing X-Request-ID header")
	}
}

func TestCORSAllowedOrigin(t *testing.T) {
	eng := rules.NewEngineFromRules(testRules())
	cfg := server.Config{
		AllowedOrigins: []string{"http://localhost:5173"},
	}
	s := server.New(cfg, eng, nil, nil)

	req := httptest.NewRequest("GET", "/api/status", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	w := httptest.NewRecorder()
	s.Engine().ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Errorf("expected CORS header, got %q", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSPreflight(t *testing.T) {
	eng := rules.NewEngineFromRules(testRules())
	cfg := server.Config{
		AllowedOrigins: []string{"http://localhost:5173"},
	}
	s := server.New(cfg, eng, nil, nil)

	req := httptest.NewRequest("OPTIONS", "/api/status", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	w := httptest.NewRecorder()
	s.Engine().ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for OPTIONS, got %d", w.Code)
	}
}

func TestHostCheckBlocked(t *testing.T) {
	eng := rules.NewEngineFromRules(testRules())
	cfg := server.Config{
		AllowedHosts: []string{"localhost"},
	}
	s := server.New(cfg, eng, nil, nil)

	req := httptest.NewRequest("GET", "/api/status", nil)
	req.Host = "evil.example.com"
	w := httptest.NewRecorder()
	s.Engine().ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for disallowed host, got %d", w.Code)
	}
}

func TestPanicRecovery(t *testing.T) {
	eng := rules.NewEngineFromRules(testRules())
	cfg := server.Config{}
	s := server.New(cfg, eng, nil, nil)

	// Register a panicking route for test purposes
	s.Engine().GET("/test/panic", func(c *gin.Context) {
		panic("test panic")
	})

	w := do(t, s, "GET", "/test/panic")
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 after panic, got %d", w.Code)
	}
}

// --- SPA with real assets ---

func newServerWithAssets(t *testing.T) *server.Server {
	t.Helper()
	memFS := fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: []byte("<!doctype html><html><body>Synapse</body></html>"),
		},
		"assets/app.js": &fstest.MapFile{
			Data: []byte("console.log('hello')"),
		},
	}
	sub, err := fs.Sub(memFS, ".")
	if err != nil {
		t.Fatal(err)
	}
	eng := rules.NewEngineFromRules(testRules())
	return server.New(server.Config{}, eng, nil, sub)
}

func TestSPAServesIndexHTML(t *testing.T) {
	s := newServerWithAssets(t)
	w := do(t, s, "GET", "/")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct == "" {
		t.Error("expected Content-Type header")
	}
}

func TestSPAServesStaticAsset(t *testing.T) {
	s := newServerWithAssets(t)
	w := do(t, s, "GET", "/assets/app.js")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for JS asset, got %d", w.Code)
	}
}

func TestSPAFallsBackToIndex(t *testing.T) {
	s := newServerWithAssets(t)
	// A path that doesn't exist as a file should fall back to index.html
	w := do(t, s, "GET", "/some/client/route")
	if w.Code != http.StatusOK {
		t.Fatalf("expected SPA fallback 200, got %d", w.Code)
	}
}

func TestSPAMethodNotAllowed(t *testing.T) {
	s := newServerWithAssets(t)
	w := do(t, s, "POST", "/some/route")
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestDoctorAllFail(t *testing.T) {
	eng := rules.NewEngineFromRules(nil) // no rules
	s := server.New(server.Config{}, eng, nil, nil)
	w := do(t, s, "GET", "/api/doctor")
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when no rules loaded, got %d", w.Code)
	}
}
