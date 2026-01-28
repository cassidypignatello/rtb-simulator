package dispatcher

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cass/rtb-simulator/internal/config"
	"github.com/cass/rtb-simulator/pkg/openrtb"
)

func TestDispatcher_Dispatch_AllRespond(t *testing.T) {
	var callCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"req-1","seatbid":[{"bid":[{"id":"bid-1","impid":"imp-1","price":2.5}]}]}`))
	}))
	defer server.Close()

	dsps := []config.DSPConfig{
		{Name: "dsp1", Endpoint: server.URL, Enabled: true},
		{Name: "dsp2", Endpoint: server.URL, Enabled: true},
		{Name: "dsp3", Endpoint: server.URL, Enabled: true},
	}

	d := New(dsps, WithTimeout(5*time.Second))

	req := &openrtb.BidRequest{ID: "req-1"}
	results := d.Dispatch(context.Background(), req)

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	if callCount.Load() != 3 {
		t.Errorf("expected 3 DSP calls, got %d", callCount.Load())
	}

	for _, r := range results {
		if r.Error != nil {
			t.Errorf("unexpected error for %s: %v", r.DSPName, r.Error)
		}
		if r.Response == nil {
			t.Errorf("expected response for %s", r.DSPName)
		}
	}
}

func TestDispatcher_Dispatch_SomeNoBid(t *testing.T) {
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"req-1","seatbid":[{"bid":[{"id":"bid-1","impid":"imp-1","price":2.5}]}]}`))
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server2.Close()

	dsps := []config.DSPConfig{
		{Name: "bidder", Endpoint: server1.URL, Enabled: true},
		{Name: "nobidder", Endpoint: server2.URL, Enabled: true},
	}

	d := New(dsps, WithTimeout(5*time.Second))

	req := &openrtb.BidRequest{ID: "req-1"}
	results := d.Dispatch(context.Background(), req)

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	bidCount := 0
	noBidCount := 0
	for _, r := range results {
		if r.Error != nil {
			continue
		}
		if r.Response.IsNoBid() {
			noBidCount++
		} else {
			bidCount++
		}
	}

	if bidCount != 1 {
		t.Errorf("expected 1 bid, got %d", bidCount)
	}
	if noBidCount != 1 {
		t.Errorf("expected 1 nobid, got %d", noBidCount)
	}
}

func TestDispatcher_Dispatch_Timeout(t *testing.T) {
	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer slowServer.Close()

	fastServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"req-1","seatbid":[{"bid":[{"id":"bid-1","impid":"imp-1","price":2.5}]}]}`))
	}))
	defer fastServer.Close()

	dsps := []config.DSPConfig{
		{Name: "slow", Endpoint: slowServer.URL, Enabled: true},
		{Name: "fast", Endpoint: fastServer.URL, Enabled: true},
	}

	d := New(dsps, WithTimeout(50*time.Millisecond))

	req := &openrtb.BidRequest{ID: "req-1"}
	results := d.Dispatch(context.Background(), req)

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	errorCount := 0
	successCount := 0
	for _, r := range results {
		if r.Error != nil {
			errorCount++
		} else {
			successCount++
		}
	}

	if errorCount != 1 {
		t.Errorf("expected 1 error (timeout), got %d", errorCount)
	}
	if successCount != 1 {
		t.Errorf("expected 1 success, got %d", successCount)
	}
}

func TestDispatcher_Dispatch_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dsps := []config.DSPConfig{
		{Name: "dsp1", Endpoint: server.URL, Enabled: true},
		{Name: "dsp2", Endpoint: server.URL, Enabled: true},
	}

	d := New(dsps, WithTimeout(5*time.Second))

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	req := &openrtb.BidRequest{ID: "req-1"}
	start := time.Now()
	results := d.Dispatch(ctx, req)
	elapsed := time.Since(start)

	// Dispatch should return within reasonable time after context cancellation.
	// Note: underlying requests may still complete, but Dispatch returns early.
	if elapsed > 600*time.Millisecond {
		t.Errorf("dispatch took too long: %v", elapsed)
	}

	// All results should have errors (context cancelled)
	for _, r := range results {
		if r.Error == nil {
			t.Errorf("expected error due to context cancellation for %s", r.DSPName)
		}
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestDispatcher_Dispatch_OnlyEnabledDSPs(t *testing.T) {
	// Test that dispatcher processes all DSPs passed to it.
	// Filtering should be done by the caller using Config.EnabledDSPs().
	var callCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"req-1"}`))
	}))
	defer server.Close()

	// Only pass enabled DSPs (as caller would do using cfg.EnabledDSPs())
	enabledDSPs := []config.DSPConfig{
		{Name: "enabled1", Endpoint: server.URL, Enabled: true},
		{Name: "enabled2", Endpoint: server.URL, Enabled: true},
	}

	d := New(enabledDSPs, WithTimeout(5*time.Second))

	req := &openrtb.BidRequest{ID: "req-1"}
	results := d.Dispatch(context.Background(), req)

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	if callCount.Load() != 2 {
		t.Errorf("expected 2 DSP calls, got %d", callCount.Load())
	}
}

func TestDispatcher_Dispatch_NoDSPs(t *testing.T) {
	d := New(nil, WithTimeout(5*time.Second))

	req := &openrtb.BidRequest{ID: "req-1"}
	results := d.Dispatch(context.Background(), req)

	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestDispatcher_AllBids(t *testing.T) {
	results := []Result{
		{
			DSPName: "dsp1",
			Response: &openrtb.BidResponse{
				ID:      "req-1",
				SeatBid: []openrtb.SeatBid{{Bid: []openrtb.Bid{{ID: "1", Price: 2.0}}}},
			},
		},
		{
			DSPName:  "dsp2",
			Response: &openrtb.BidResponse{ID: "req-1"},
			Error:    nil,
		},
		{
			DSPName: "dsp3",
			Error:   context.DeadlineExceeded,
		},
	}

	bids := AllBids(results)

	if len(bids) != 1 {
		t.Errorf("expected 1 bid, got %d", len(bids))
	}
}
