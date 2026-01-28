package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cass/rtb-simulator/internal/api"
	"github.com/cass/rtb-simulator/internal/auction"
	"github.com/cass/rtb-simulator/internal/config"
	"github.com/cass/rtb-simulator/internal/dispatcher"
	"github.com/cass/rtb-simulator/internal/engine"
	"github.com/cass/rtb-simulator/internal/generator"
	"github.com/cass/rtb-simulator/internal/generator/scenarios"
	"github.com/cass/rtb-simulator/internal/stats"
	"github.com/cass/rtb-simulator/pkg/openrtb"
)

// mockDSP creates a test DSP server that responds to bid requests.
type mockDSP struct {
	server       *httptest.Server
	requestCount uint64
	bidPrice     float64
	delay        time.Duration
}

func newMockDSP(bidPrice float64, delay time.Duration) *mockDSP {
	m := &mockDSP{
		bidPrice: bidPrice,
		delay:    delay,
	}

	m.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint64(&m.requestCount, 1)

		if m.delay > 0 {
			time.Sleep(m.delay)
		}

		var req openrtb.BidRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Return a bid if price > 0, otherwise no-bid
		resp := openrtb.BidResponse{ID: req.ID}

		if m.bidPrice > 0 && len(req.Imp) > 0 {
			resp.SeatBid = []openrtb.SeatBid{{
				Bid: []openrtb.Bid{{
					ID:    "bid-" + req.ID,
					ImpID: req.Imp[0].ID,
					Price: m.bidPrice,
				}},
			}}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))

	return m
}

func (m *mockDSP) URL() string {
	return m.server.URL
}

func (m *mockDSP) Requests() uint64 {
	return atomic.LoadUint64(&m.requestCount)
}

func (m *mockDSP) Close() {
	m.server.Close()
}

func TestIntegration_FullSimulationCycle(t *testing.T) {
	// Create mock DSPs
	dsp1 := newMockDSP(1.50, 0)
	defer dsp1.Close()

	dsp2 := newMockDSP(2.00, 0)
	defer dsp2.Close()

	// Configure the system
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 0},
		Simulation: config.SimulationConfig{
			RequestsPerSecond: 50,
			Scenario:          "mobile_app",
		},
		Auction: config.AuctionConfig{
			Type:      "first_price",
			TimeoutMS: 100,
		},
		DSPs: []config.DSPConfig{
			{Name: "dsp-1", Endpoint: dsp1.URL(), Enabled: true},
			{Name: "dsp-2", Endpoint: dsp2.URL(), Enabled: true},
		},
	}

	// Initialize components
	scenario := scenarios.NewMobileApp()
	gen := generator.New(scenario, generator.WithTimeout(cfg.Auction.TimeoutMS))
	disp := dispatcher.New(cfg.DSPs, dispatcher.WithTimeout(100*time.Millisecond))
	defer disp.Close()

	auc := auction.NewFirstPrice()
	collector := stats.New()

	eng := engine.New(gen, disp, auc, collector, engine.WithRPS(cfg.Simulation.RequestsPerSecond))

	// Start simulation
	if err := eng.Start(); err != nil {
		t.Fatalf("failed to start engine: %v", err)
	}

	// Let it run for a bit
	time.Sleep(200 * time.Millisecond)

	// Verify both DSPs received requests
	if dsp1.Requests() == 0 {
		t.Error("dsp-1 received no requests")
	}
	if dsp2.Requests() == 0 {
		t.Error("dsp-2 received no requests")
	}

	// Stop simulation
	eng.Stop()

	// Verify stats were collected
	snap := collector.Snapshot()
	if snap.TotalRequests == 0 {
		t.Error("TotalRequests = 0, expected > 0")
	}
	if snap.TotalBids == 0 {
		t.Error("TotalBids = 0, expected > 0")
	}
	if snap.TotalWins == 0 {
		t.Error("TotalWins = 0, expected > 0")
	}
	if snap.TotalRevenue <= 0 {
		t.Error("TotalRevenue <= 0, expected > 0")
	}

	// Verify per-DSP stats
	if len(snap.DSPStats) != 2 {
		t.Errorf("DSP stats count = %d, expected 2", len(snap.DSPStats))
	}

	t.Logf("Integration test results:")
	t.Logf("  DSP-1 requests: %d", dsp1.Requests())
	t.Logf("  DSP-2 requests: %d", dsp2.Requests())
	t.Logf("  Total requests: %d", snap.TotalRequests)
	t.Logf("  Total bids: %d", snap.TotalBids)
	t.Logf("  Total wins: %d", snap.TotalWins)
	t.Logf("  Total revenue: $%.4f", snap.TotalRevenue)
}

func TestIntegration_APIControlFlow(t *testing.T) {
	// Create mock DSP
	dsp := newMockDSP(1.00, 0)
	defer dsp.Close()

	cfg := &config.Config{
		Server: config.ServerConfig{Port: 0},
		Simulation: config.SimulationConfig{
			RequestsPerSecond: 100,
			Scenario:          "mobile_app",
		},
		Auction: config.AuctionConfig{
			Type:      "first_price",
			TimeoutMS: 100,
		},
		DSPs: []config.DSPConfig{
			{Name: "test-dsp", Endpoint: dsp.URL(), Enabled: true},
		},
	}

	// Initialize components
	scenario := scenarios.NewMobileApp()
	gen := generator.New(scenario)
	disp := dispatcher.New(cfg.DSPs)
	defer disp.Close()

	auc := auction.NewFirstPrice()
	collector := stats.New()

	eng := engine.New(gen, disp, auc, collector, engine.WithRPS(100))

	// Create API server
	srv := api.New(eng, collector, cfg)
	handler := srv.Handler()

	// Test initial status - should not be running
	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var status api.StatusResponse
	_ = json.NewDecoder(rec.Body).Decode(&status)
	if status.Running {
		t.Error("initial status should be not running")
	}

	// Start via API
	req = httptest.NewRequest(http.MethodPost, "/start", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("POST /start status = %d, want 200", rec.Code)
	}

	// Let it run
	time.Sleep(100 * time.Millisecond)

	// Check stats via API
	req = httptest.NewRequest(http.MethodGet, "/stats", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var snap stats.Snapshot
	_ = json.NewDecoder(rec.Body).Decode(&snap)
	if snap.TotalRequests == 0 {
		t.Error("stats should show requests after running")
	}

	// Stop via API
	req = httptest.NewRequest(http.MethodPost, "/stop", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("POST /stop status = %d, want 200", rec.Code)
	}

	// Verify stopped
	req = httptest.NewRequest(http.MethodGet, "/status", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	_ = json.NewDecoder(rec.Body).Decode(&status)
	if status.Running {
		t.Error("status should be stopped after /stop")
	}

	t.Logf("API control flow test passed with %d requests", snap.TotalRequests)
}

func TestIntegration_DSPTimeout(t *testing.T) {
	// Create a slow DSP that will timeout
	slowDSP := newMockDSP(5.00, 500*time.Millisecond)
	defer slowDSP.Close()

	// Create a fast DSP
	fastDSP := newMockDSP(1.00, 0)
	defer fastDSP.Close()

	cfg := &config.Config{
		DSPs: []config.DSPConfig{
			{Name: "slow-dsp", Endpoint: slowDSP.URL(), Enabled: true},
			{Name: "fast-dsp", Endpoint: fastDSP.URL(), Enabled: true},
		},
	}

	scenario := scenarios.NewMobileApp()
	gen := generator.New(scenario)
	// Very short timeout - slow DSP should timeout
	disp := dispatcher.New(cfg.DSPs, dispatcher.WithTimeout(50*time.Millisecond))
	defer disp.Close()

	auc := auction.NewFirstPrice()
	collector := stats.New()

	eng := engine.New(gen, disp, auc, collector, engine.WithRPS(20))

	_ = eng.Start()
	time.Sleep(200 * time.Millisecond)
	eng.Stop()

	snap := collector.Snapshot()

	// Fast DSP should have won most auctions since slow one times out
	fastStats, hasFast := snap.DSPStats["fast-dsp"]
	if !hasFast {
		t.Fatal("missing stats for fast-dsp")
	}

	slowStats, hasSlow := snap.DSPStats["slow-dsp"]
	if !hasSlow {
		t.Fatal("missing stats for slow-dsp")
	}

	// Slow DSP should have more errors (timeouts)
	if slowStats.Errors == 0 {
		t.Error("slow-dsp should have timeout errors")
	}

	t.Logf("Timeout test:")
	t.Logf("  Fast DSP - requests: %d, bids: %d, errors: %d", fastStats.Requests, fastStats.Bids, fastStats.Errors)
	t.Logf("  Slow DSP - requests: %d, bids: %d, errors: %d", slowStats.Requests, slowStats.Bids, slowStats.Errors)
}

func TestIntegration_NoBidResponse(t *testing.T) {
	// DSP that returns no bids (price = 0)
	noBidDSP := newMockDSP(0, 0)
	defer noBidDSP.Close()

	cfg := &config.Config{
		DSPs: []config.DSPConfig{
			{Name: "nobid-dsp", Endpoint: noBidDSP.URL(), Enabled: true},
		},
	}

	scenario := scenarios.NewMobileApp()
	gen := generator.New(scenario)
	disp := dispatcher.New(cfg.DSPs)
	defer disp.Close()

	auc := auction.NewFirstPrice()
	collector := stats.New()

	eng := engine.New(gen, disp, auc, collector, engine.WithRPS(50))

	_ = eng.Start()
	time.Sleep(150 * time.Millisecond)
	eng.Stop()

	snap := collector.Snapshot()

	// All requests should result in no-bids
	if snap.TotalRequests == 0 {
		t.Fatal("expected some requests")
	}
	if snap.TotalWins != 0 {
		t.Errorf("TotalWins = %d, expected 0 with no-bid DSP", snap.TotalWins)
	}
	if snap.TotalNoBids == 0 {
		t.Error("TotalNoBids = 0, expected > 0")
	}

	t.Logf("No-bid test: %d requests, %d no-bids", snap.TotalRequests, snap.TotalNoBids)
}

func TestIntegration_GracefulShutdown(t *testing.T) {
	dsp := newMockDSP(1.00, 0)
	defer dsp.Close()

	cfg := &config.Config{
		DSPs: []config.DSPConfig{
			{Name: "test-dsp", Endpoint: dsp.URL(), Enabled: true},
		},
	}

	scenario := scenarios.NewMobileApp()
	gen := generator.New(scenario)
	disp := dispatcher.New(cfg.DSPs)
	defer disp.Close()

	auc := auction.NewFirstPrice()
	collector := stats.New()

	eng := engine.New(gen, disp, auc, collector, engine.WithRPS(1000))

	_ = eng.Start()
	time.Sleep(50 * time.Millisecond)

	// Graceful shutdown should complete within timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	shutdownDone := make(chan error, 1)
	go func() {
		shutdownDone <- eng.Shutdown(ctx)
	}()

	select {
	case err := <-shutdownDone:
		if err != nil {
			t.Errorf("Shutdown error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Shutdown timed out")
	}

	if eng.IsRunning() {
		t.Error("engine should not be running after shutdown")
	}
}

func BenchmarkIntegration_FullPipeline(b *testing.B) {
	dsp := newMockDSP(1.00, 0)
	defer dsp.Close()

	cfg := &config.Config{
		DSPs: []config.DSPConfig{
			{Name: "bench-dsp", Endpoint: dsp.URL(), Enabled: true},
		},
	}

	scenario := scenarios.NewMobileApp()
	gen := generator.New(scenario)
	disp := dispatcher.New(cfg.DSPs)
	defer disp.Close()

	auc := auction.NewFirstPrice()
	collector := stats.New()

	// Note: Engine not started, manually executing pipeline for benchmark
	_ = engine.New(gen, disp, auc, collector)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Simulate one full tick
		req := gen.Generate()
		results := disp.Dispatch(context.Background(), req)
		outcome := auc.Run(req.ID, 0.01, results)
		collector.RecordAuction(outcome, results)
	}
}
