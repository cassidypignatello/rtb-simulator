// Package api provides the HTTP API server for controlling the RTB simulator.
// It exposes endpoints for starting/stopping simulation, retrieving stats, and health checks.
package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/cass/rtb-simulator/internal/config"
	"github.com/cass/rtb-simulator/internal/stats"
)

// EngineController defines the interface for controlling the simulation engine.
type EngineController interface {
	Start() error
	Stop()
	IsRunning() bool
}

// StatusResponse represents the engine status response.
type StatusResponse struct {
	Running bool   `json:"running"`
	Message string `json:"message,omitempty"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error string `json:"error"`
}

// Server handles HTTP API requests for the RTB simulator.
type Server struct {
	engine    EngineController
	stats     *stats.Collector
	config    *config.Config
	server    *http.Server
	mux       *http.ServeMux
}

// Option configures the server.
type Option func(*Server)

// WithAddr sets the server address.
func WithAddr(addr string) Option {
	return func(s *Server) {
		s.server.Addr = addr
	}
}

// WithReadTimeout sets the read timeout.
func WithReadTimeout(d time.Duration) Option {
	return func(s *Server) {
		s.server.ReadTimeout = d
	}
}

// WithWriteTimeout sets the write timeout.
func WithWriteTimeout(d time.Duration) Option {
	return func(s *Server) {
		s.server.WriteTimeout = d
	}
}

// New creates a new API server.
func New(engine EngineController, stats *stats.Collector, cfg *config.Config, opts ...Option) *Server {
	s := &Server{
		engine: engine,
		stats:  stats,
		config: cfg,
		mux:    http.NewServeMux(),
		server: &http.Server{
			Addr:         ":8080",
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(s)
	}

	s.setupRoutes()
	s.server.Handler = s.mux

	return s
}

// setupRoutes registers all API routes.
func (s *Server) setupRoutes() {
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/status", s.handleStatus)
	s.mux.HandleFunc("/start", s.handleStart)
	s.mux.HandleFunc("/stop", s.handleStop)
	s.mux.HandleFunc("/stats", s.handleStats)
	s.mux.HandleFunc("/config", s.handleConfig)
}

// Handler returns the HTTP handler for testing.
func (s *Server) Handler() http.Handler {
	return s.mux
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe() error {
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// handleHealth returns a simple health check response.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// handleStatus returns the current engine status.
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp := StatusResponse{Running: s.engine.IsRunning()}
	s.writeJSON(w, http.StatusOK, resp)
}

// handleStart starts the simulation engine.
func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.engine.IsRunning() {
		s.writeJSON(w, http.StatusConflict, ErrorResponse{Error: "engine is already running"})
		return
	}

	if err := s.engine.Start(); err != nil {
		s.writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	resp := StatusResponse{Running: true, Message: "simulation started"}
	s.writeJSON(w, http.StatusOK, resp)
}

// handleStop stops the simulation engine.
func (s *Server) handleStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.engine.Stop()

	resp := StatusResponse{Running: false, Message: "simulation stopped"}
	s.writeJSON(w, http.StatusOK, resp)
}

// handleStats returns the current statistics snapshot.
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	snap := s.stats.Snapshot()
	s.writeJSON(w, http.StatusOK, snap)
}

// handleConfig returns the current configuration.
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.writeJSON(w, http.StatusOK, s.config)
}

// writeJSON writes a JSON response.
func (s *Server) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("failed to encode JSON response: %v", err)
	}
}
