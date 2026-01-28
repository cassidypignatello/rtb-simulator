package engine

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cass/rtb-simulator/internal/auction"
	"github.com/cass/rtb-simulator/internal/dispatcher"
	"github.com/cass/rtb-simulator/internal/stats"
	"github.com/cass/rtb-simulator/pkg/openrtb"
)

// mockGenerator implements a simple generator for testing.
type mockGenerator struct {
	counter uint64
}

func (m *mockGenerator) Generate() *openrtb.BidRequest {
	id := atomic.AddUint64(&m.counter, 1)
	return &openrtb.BidRequest{
		ID: string(rune(id)),
		Imp: []openrtb.Imp{{
			ID:       "imp-1",
			BidFloor: 0.5,
		}},
	}
}

func (m *mockGenerator) ScenarioName() string {
	return "mock"
}

// mockDispatcher returns configurable results.
type mockDispatcher struct {
	results []dispatcher.Result
	calls   uint64
}

func (m *mockDispatcher) Dispatch(ctx context.Context, req *openrtb.BidRequest) []dispatcher.Result {
	atomic.AddUint64(&m.calls, 1)
	return m.results
}

func (m *mockDispatcher) Close() {}

func TestEngine_StartStop(t *testing.T) {
	gen := &mockGenerator{}
	disp := &mockDispatcher{
		results: []dispatcher.Result{
			{
				DSPName: "test-dsp",
				Response: &openrtb.BidResponse{
					ID: "resp-1",
					SeatBid: []openrtb.SeatBid{{
						Bid: []openrtb.Bid{{
							ID:    "bid-1",
							ImpID: "imp-1",
							Price: 1.0,
						}},
					}},
				},
				Latency: 10 * time.Millisecond,
			},
		},
	}
	auc := auction.NewFirstPrice()
	collector := stats.New()

	e := New(gen, disp, auc, collector, WithRPS(100))

	// Should start successfully
	err := e.Start()
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Should be running
	if !e.IsRunning() {
		t.Error("IsRunning() = false, want true")
	}

	// Starting again should error
	err = e.Start()
	if err != ErrAlreadyRunning {
		t.Errorf("Start() error = %v, want ErrAlreadyRunning", err)
	}

	// Let it run a bit
	time.Sleep(50 * time.Millisecond)

	// Stop should work
	e.Stop()

	// Should no longer be running
	if e.IsRunning() {
		t.Error("IsRunning() = true after Stop(), want false")
	}

	// Verify some requests were made
	if disp.calls == 0 {
		t.Error("dispatcher.Dispatch() was never called")
	}

	// Stats should show activity
	snap := collector.Snapshot()
	if snap.TotalRequests == 0 {
		t.Error("stats.TotalRequests = 0, expected > 0")
	}
}

func TestEngine_StopWithoutStart(t *testing.T) {
	gen := &mockGenerator{}
	disp := &mockDispatcher{}
	auc := auction.NewFirstPrice()
	collector := stats.New()

	e := New(gen, disp, auc, collector)

	// Stopping without starting should be safe
	e.Stop()

	if e.IsRunning() {
		t.Error("IsRunning() = true, want false")
	}
}

func TestEngine_RPS(t *testing.T) {
	gen := &mockGenerator{}
	disp := &mockDispatcher{
		results: []dispatcher.Result{
			{DSPName: "test", Response: &openrtb.BidResponse{ID: "1"}},
		},
	}
	auc := auction.NewFirstPrice()
	collector := stats.New()

	// Low RPS for testing
	e := New(gen, disp, auc, collector, WithRPS(10))

	_ = e.Start()
	time.Sleep(250 * time.Millisecond)
	e.Stop()

	// At 10 RPS over 250ms, expect ~2-3 requests
	calls := atomic.LoadUint64(&disp.calls)
	if calls < 1 || calls > 5 {
		t.Errorf("Dispatch calls = %d, expected 1-5 at 10 RPS over 250ms", calls)
	}
}

func TestEngine_GracefulShutdown(t *testing.T) {
	gen := &mockGenerator{}
	disp := &mockDispatcher{
		results: []dispatcher.Result{
			{DSPName: "test", Response: &openrtb.BidResponse{ID: "1"}},
		},
	}
	auc := auction.NewFirstPrice()
	collector := stats.New()

	e := New(gen, disp, auc, collector, WithRPS(1000))

	_ = e.Start()
	time.Sleep(10 * time.Millisecond)

	// Shutdown should complete without hanging
	done := make(chan struct{})
	go func() {
		e.Shutdown(context.Background())
		close(done)
	}()

	select {
	case <-done:
		// Good
	case <-time.After(2 * time.Second):
		t.Fatal("Shutdown() timed out")
	}

	if e.IsRunning() {
		t.Error("IsRunning() = true after Shutdown()")
	}
}

func TestEngine_Options(t *testing.T) {
	gen := &mockGenerator{}
	disp := &mockDispatcher{}
	auc := auction.NewFirstPrice()
	collector := stats.New()

	e := New(gen, disp, auc, collector,
		WithRPS(500),
		WithBidFloor(0.25),
	)

	if e.rps != 500 {
		t.Errorf("rps = %d, want 500", e.rps)
	}
	if e.bidFloor != 0.25 {
		t.Errorf("bidFloor = %f, want 0.25", e.bidFloor)
	}
}

func BenchmarkEngine_Tick(b *testing.B) {
	gen := &mockGenerator{}
	disp := &mockDispatcher{
		results: []dispatcher.Result{
			{
				DSPName: "test-dsp",
				Response: &openrtb.BidResponse{
					ID: "resp-1",
					SeatBid: []openrtb.SeatBid{{
						Bid: []openrtb.Bid{{ID: "bid-1", ImpID: "imp-1", Price: 1.0}},
					}},
				},
				Latency: time.Millisecond,
			},
		},
	}
	auc := auction.NewFirstPrice()
	collector := stats.New()

	e := New(gen, disp, auc, collector)

	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		e.tick(ctx)
	}
}
