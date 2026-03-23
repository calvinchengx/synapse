package server

import (
	"context"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/calvinchengx/synapse/internal/integration"
	"github.com/calvinchengx/synapse/internal/rules"
)

// Config holds server configuration.
type Config struct {
	Host           string
	Port           int
	AllowedOrigins []string
	AllowedHosts   []string
	RateLimit      int // requests per second per IP
}

// Server wraps a Gin engine with all wired dependencies.
type Server struct {
	cfg        Config
	engine     *gin.Engine
	httpServer *http.Server
	rulesEng   *rules.Engine
	manifests  []integration.ToolManifest
	assets     fs.FS
}

// New creates a ready-to-use Server.
func New(cfg Config, rulesEng *rules.Engine, manifests []integration.ToolManifest, assets fs.FS) *Server {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	s := &Server{
		cfg:       cfg,
		engine:    r,
		rulesEng:  rulesEng,
		manifests: manifests,
		assets:    assets,
	}

	s.registerMiddleware()
	s.registerRoutes()

	return s
}

func (s *Server) registerMiddleware() {
	s.engine.Use(
		Recovery(),
		RequestID(),
		Logger(),
		SecurityHeaders(),
		CORS(s.cfg.AllowedOrigins),
		HostCheck(s.cfg.AllowedHosts),
	)
}

func (s *Server) registerRoutes() {
	api := s.engine.Group("/api")
	{
		api.GET("/status", s.handleStatus)
		api.GET("/doctor", s.handleDoctor)

		api.GET("/rules", s.handleListRules)
		api.GET("/rules/categories", s.handleCategories)
		api.GET("/rules/search", s.handleSearchRules)

		api.GET("/integrations", s.handleListIntegrations)
		api.GET("/integrations/:id", s.handleGetIntegration)
	}

	// SPA fallback — serve embedded dist or 404
	s.engine.NoRoute(s.handleSPA)
}

// Start listens on cfg.Host:cfg.Port and serves until ctx is cancelled.
func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}

	s.httpServer = &http.Server{
		Handler:           s.engine,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := s.httpServer.Serve(ln); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return s.httpServer.Shutdown(shutCtx)
	case err := <-errCh:
		return err
	}
}

// Engine exposes the Gin engine for testing.
func (s *Server) Engine() *gin.Engine { return s.engine }
