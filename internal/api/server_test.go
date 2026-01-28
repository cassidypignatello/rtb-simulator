package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cass/rtb-simulator/internal/config"
	"github.com/cass/rtb-simulator/internal/stats"
)

// mockEngine implements EngineController for testing.
type mockEngine struct {
	running     bool
	startCalled bool
	stopCalled  bool
	startErr    error
}

func (m *mockEngine) Start() error {
	m.startCalled = true
	if m.startErr != nil {
		return m.startErr
	}
	m.running = true
	return nil
}

func (m *mockEngine) Stop() {
	m.stopCalled = true
	m.running = false
}

func (m *mockEngine) IsRunning() bool {
	return m.running
}

func TestServer_StartEndpoint(t *testing.T) {
	eng := &mockEngine{}
	collector := stats.New()
	cfg := &config.Config{}

	srv := New(eng, collector, cfg, WithAddr(":0"))
	handler := srv.Handler()

	req := httptest.NewRequest(http.MethodPost, "/start", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("POST /start status = %d, want %d", rec.Code, http.StatusOK)
	}

	if !eng.startCalled {
		t.Error("engine.Start() was not called")
	}

	var resp StatusResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Running {
		t.Error("response.Running = false, want true")
	}
}

func TestServer_StartEndpoint_AlreadyRunning(t *testing.T) {
	eng := &mockEngine{running: true}
	collector := stats.New()
	cfg := &config.Config{}

	srv := New(eng, collector, cfg)
	handler := srv.Handler()

	req := httptest.NewRequest(http.MethodPost, "/start", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("POST /start when running: status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func TestServer_StopEndpoint(t *testing.T) {
	eng := &mockEngine{running: true}
	collector := stats.New()
	cfg := &config.Config{}

	srv := New(eng, collector, cfg)
	handler := srv.Handler()

	req := httptest.NewRequest(http.MethodPost, "/stop", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("POST /stop status = %d, want %d", rec.Code, http.StatusOK)
	}

	if !eng.stopCalled {
		t.Error("engine.Stop() was not called")
	}

	var resp StatusResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Running {
		t.Error("response.Running = true after stop, want false")
	}
}

func TestServer_StatsEndpoint(t *testing.T) {
	eng := &mockEngine{}
	collector := stats.New()
	cfg := &config.Config{}

	srv := New(eng, collector, cfg)
	handler := srv.Handler()

	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /stats status = %d, want %d", rec.Code, http.StatusOK)
	}

	var snap stats.Snapshot
	if err := json.NewDecoder(rec.Body).Decode(&snap); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
}

func TestServer_ConfigEndpoint(t *testing.T) {
	eng := &mockEngine{}
	collector := stats.New()
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080},
		Simulation: config.SimulationConfig{
			RequestsPerSecond: 100,
			Scenario:          "mobile_app",
		},
	}

	srv := New(eng, collector, cfg)
	handler := srv.Handler()

	req := httptest.NewRequest(http.MethodGet, "/config", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /config status = %d, want %d", rec.Code, http.StatusOK)
	}

	var respCfg config.Config
	if err := json.NewDecoder(rec.Body).Decode(&respCfg); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if respCfg.Simulation.RequestsPerSecond != 100 {
		t.Errorf("config.Simulation.RequestsPerSecond = %d, want 100", respCfg.Simulation.RequestsPerSecond)
	}
}

func TestServer_StatusEndpoint(t *testing.T) {
	eng := &mockEngine{running: true}
	collector := stats.New()
	cfg := &config.Config{}

	srv := New(eng, collector, cfg)
	handler := srv.Handler()

	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /status status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp StatusResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Running {
		t.Error("response.Running = false, want true")
	}
}

func TestServer_HealthEndpoint(t *testing.T) {
	eng := &mockEngine{}
	collector := stats.New()
	cfg := &config.Config{}

	srv := New(eng, collector, cfg)
	handler := srv.Handler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /health status = %d, want %d", rec.Code, http.StatusOK)
	}

	body, _ := io.ReadAll(rec.Body)
	if string(body) != "ok" {
		t.Errorf("GET /health body = %q, want \"ok\"", string(body))
	}
}

func TestServer_MethodNotAllowed(t *testing.T) {
	eng := &mockEngine{}
	collector := stats.New()
	cfg := &config.Config{}

	srv := New(eng, collector, cfg)
	handler := srv.Handler()

	// GET on POST-only endpoint
	req := httptest.NewRequest(http.MethodGet, "/start", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /start status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestServer_NotFound(t *testing.T) {
	eng := &mockEngine{}
	collector := stats.New()
	cfg := &config.Config{}

	srv := New(eng, collector, cfg)
	handler := srv.Handler()

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("GET /nonexistent status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestServer_ListenAndServe(t *testing.T) {
	eng := &mockEngine{}
	collector := stats.New()
	cfg := &config.Config{}

	srv := New(eng, collector, cfg, WithAddr(":0"))

	// Start server in background
	go func() {
		_ = srv.ListenAndServe()
	}()

	// Give it time to start
	time.Sleep(50 * time.Millisecond)

	// Shutdown should work
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}
}

func TestServer_Options(t *testing.T) {
	eng := &mockEngine{}
	collector := stats.New()
	cfg := &config.Config{}

	srv := New(eng, collector, cfg,
		WithAddr(":9090"),
		WithReadTimeout(5*time.Second),
		WithWriteTimeout(10*time.Second),
	)

	if srv.server.Addr != ":9090" {
		t.Errorf("server.Addr = %q, want \":9090\"", srv.server.Addr)
	}
	if srv.server.ReadTimeout != 5*time.Second {
		t.Errorf("server.ReadTimeout = %v, want 5s", srv.server.ReadTimeout)
	}
	if srv.server.WriteTimeout != 10*time.Second {
		t.Errorf("server.WriteTimeout = %v, want 10s", srv.server.WriteTimeout)
	}
}
