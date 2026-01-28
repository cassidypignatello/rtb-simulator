package stats

import (
	"testing"
	"time"

	"github.com/cass/rtb-simulator/internal/auction"
	"github.com/cass/rtb-simulator/internal/dispatcher"
	"github.com/cass/rtb-simulator/pkg/openrtb"
)

func TestCollector_RecordAuction_WithWinner(t *testing.T) {
	c := New()

	outcome := auction.Outcome{
		RequestID:     "req-1",
		Winner:        &openrtb.Bid{ID: "bid-1", Price: 2.5},
		WinningDSP:    "dsp1",
		ClearingPrice: 2.5,
		AllBids: []auction.BidWithDSP{
			{Bid: openrtb.Bid{ID: "bid-1", Price: 2.5}, DSPName: "dsp1"},
			{Bid: openrtb.Bid{ID: "bid-2", Price: 1.5}, DSPName: "dsp2"},
		},
	}

	results := []dispatcher.Result{
		{DSPName: "dsp1", Latency: 10 * time.Millisecond},
		{DSPName: "dsp2", Latency: 15 * time.Millisecond},
	}

	c.RecordAuction(outcome, results)

	snapshot := c.Snapshot()

	if snapshot.TotalRequests != 1 {
		t.Errorf("expected 1 total request, got %d", snapshot.TotalRequests)
	}
	if snapshot.TotalBids != 2 {
		t.Errorf("expected 2 total bids, got %d", snapshot.TotalBids)
	}
	if snapshot.TotalWins != 1 {
		t.Errorf("expected 1 win, got %d", snapshot.TotalWins)
	}
	if snapshot.TotalNoBids != 0 {
		t.Errorf("expected 0 no-bids, got %d", snapshot.TotalNoBids)
	}
	if snapshot.TotalRevenue != 2.5 {
		t.Errorf("expected revenue 2.5, got %f", snapshot.TotalRevenue)
	}
}

func TestCollector_RecordAuction_NoBid(t *testing.T) {
	c := New()

	outcome := auction.Outcome{
		RequestID: "req-1",
		Winner:    nil,
		AllBids:   nil,
	}

	results := []dispatcher.Result{
		{DSPName: "dsp1", Latency: 10 * time.Millisecond},
	}

	c.RecordAuction(outcome, results)

	snapshot := c.Snapshot()

	if snapshot.TotalRequests != 1 {
		t.Errorf("expected 1 total request, got %d", snapshot.TotalRequests)
	}
	if snapshot.TotalNoBids != 1 {
		t.Errorf("expected 1 no-bid, got %d", snapshot.TotalNoBids)
	}
	if snapshot.TotalWins != 0 {
		t.Errorf("expected 0 wins, got %d", snapshot.TotalWins)
	}
}

func TestCollector_DSPStats(t *testing.T) {
	c := New()

	// First auction - dsp1 wins
	outcome1 := auction.Outcome{
		RequestID:     "req-1",
		Winner:        &openrtb.Bid{ID: "bid-1", Price: 2.5},
		WinningDSP:    "dsp1",
		ClearingPrice: 2.5,
		AllBids: []auction.BidWithDSP{
			{Bid: openrtb.Bid{ID: "bid-1", Price: 2.5}, DSPName: "dsp1"},
		},
	}
	results1 := []dispatcher.Result{
		{DSPName: "dsp1", Latency: 10 * time.Millisecond},
		{DSPName: "dsp2", Latency: 15 * time.Millisecond, Response: &openrtb.BidResponse{ID: "req-1"}},
	}
	c.RecordAuction(outcome1, results1)

	// Second auction - dsp2 wins
	outcome2 := auction.Outcome{
		RequestID:     "req-2",
		Winner:        &openrtb.Bid{ID: "bid-2", Price: 3.0},
		WinningDSP:    "dsp2",
		ClearingPrice: 3.0,
		AllBids: []auction.BidWithDSP{
			{Bid: openrtb.Bid{ID: "bid-2", Price: 3.0}, DSPName: "dsp2"},
		},
	}
	results2 := []dispatcher.Result{
		{DSPName: "dsp1", Latency: 12 * time.Millisecond, Response: &openrtb.BidResponse{ID: "req-2"}},
		{DSPName: "dsp2", Latency: 8 * time.Millisecond},
	}
	c.RecordAuction(outcome2, results2)

	snapshot := c.Snapshot()

	// Check DSP1 stats
	dsp1 := snapshot.DSPStats["dsp1"]
	if dsp1.Requests != 2 {
		t.Errorf("dsp1: expected 2 requests, got %d", dsp1.Requests)
	}
	if dsp1.Bids != 1 {
		t.Errorf("dsp1: expected 1 bid, got %d", dsp1.Bids)
	}
	if dsp1.Wins != 1 {
		t.Errorf("dsp1: expected 1 win, got %d", dsp1.Wins)
	}

	// Check DSP2 stats
	dsp2 := snapshot.DSPStats["dsp2"]
	if dsp2.Requests != 2 {
		t.Errorf("dsp2: expected 2 requests, got %d", dsp2.Requests)
	}
	if dsp2.Bids != 1 {
		t.Errorf("dsp2: expected 1 bid, got %d", dsp2.Bids)
	}
	if dsp2.Wins != 1 {
		t.Errorf("dsp2: expected 1 win, got %d", dsp2.Wins)
	}
}

func TestCollector_RecordError(t *testing.T) {
	c := New()

	results := []dispatcher.Result{
		{DSPName: "dsp1", Latency: 50 * time.Millisecond, Error: testError{}},
		{DSPName: "dsp2", Latency: 10 * time.Millisecond},
	}

	outcome := auction.Outcome{
		RequestID:     "req-1",
		Winner:        &openrtb.Bid{ID: "bid-1", Price: 2.0},
		WinningDSP:    "dsp2",
		ClearingPrice: 2.0,
		AllBids: []auction.BidWithDSP{
			{Bid: openrtb.Bid{ID: "bid-1", Price: 2.0}, DSPName: "dsp2"},
		},
	}

	c.RecordAuction(outcome, results)

	snapshot := c.Snapshot()

	if snapshot.TotalErrors != 1 {
		t.Errorf("expected 1 error, got %d", snapshot.TotalErrors)
	}

	dsp1 := snapshot.DSPStats["dsp1"]
	if dsp1.Errors != 1 {
		t.Errorf("dsp1: expected 1 error, got %d", dsp1.Errors)
	}
}

func TestCollector_Concurrency(t *testing.T) {
	c := New()

	done := make(chan struct{})

	// Writer goroutines
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				outcome := auction.Outcome{
					RequestID:     "req",
					Winner:        &openrtb.Bid{ID: "bid", Price: 1.0},
					WinningDSP:    "dsp",
					ClearingPrice: 1.0,
					AllBids: []auction.BidWithDSP{
						{Bid: openrtb.Bid{ID: "bid", Price: 1.0}, DSPName: "dsp"},
					},
				}
				results := []dispatcher.Result{
					{DSPName: "dsp", Latency: time.Millisecond},
				}
				c.RecordAuction(outcome, results)
			}
			done <- struct{}{}
		}()
	}

	// Reader goroutines
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = c.Snapshot()
			}
			done <- struct{}{}
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 15; i++ {
		<-done
	}

	snapshot := c.Snapshot()
	if snapshot.TotalRequests != 1000 {
		t.Errorf("expected 1000 requests, got %d", snapshot.TotalRequests)
	}
}

func TestCollector_Reset(t *testing.T) {
	c := New()

	outcome := auction.Outcome{
		RequestID:     "req-1",
		Winner:        &openrtb.Bid{ID: "bid-1", Price: 2.5},
		WinningDSP:    "dsp1",
		ClearingPrice: 2.5,
	}
	results := []dispatcher.Result{
		{DSPName: "dsp1", Latency: 10 * time.Millisecond},
	}

	c.RecordAuction(outcome, results)

	snapshot := c.Snapshot()
	if snapshot.TotalRequests != 1 {
		t.Errorf("expected 1 request before reset, got %d", snapshot.TotalRequests)
	}

	c.Reset()

	snapshot = c.Snapshot()
	if snapshot.TotalRequests != 0 {
		t.Errorf("expected 0 requests after reset, got %d", snapshot.TotalRequests)
	}
}

func TestCollector_AvgLatency(t *testing.T) {
	c := New()

	for i := 0; i < 5; i++ {
		outcome := auction.Outcome{RequestID: "req"}
		results := []dispatcher.Result{
			{DSPName: "dsp1", Latency: time.Duration(10*(i+1)) * time.Millisecond},
		}
		c.RecordAuction(outcome, results)
	}

	snapshot := c.Snapshot()
	dsp1 := snapshot.DSPStats["dsp1"]

	// Average of 10, 20, 30, 40, 50 = 30ms
	expectedAvg := 30 * time.Millisecond
	if dsp1.AvgLatency != expectedAvg {
		t.Errorf("expected avg latency %v, got %v", expectedAvg, dsp1.AvgLatency)
	}
}

type testError struct{}

func (testError) Error() string { return "test error" }
