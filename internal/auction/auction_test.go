package auction

import (
	"testing"

	"github.com/cass/rtb-simulator/internal/dispatcher"
	"github.com/cass/rtb-simulator/pkg/openrtb"
)

func TestFirstPriceAuction_Run_SingleBid(t *testing.T) {
	auction := NewFirstPrice()

	results := []dispatcher.Result{
		{
			DSPName: "dsp1",
			Response: &openrtb.BidResponse{
				ID:      "req-1",
				SeatBid: []openrtb.SeatBid{{Bid: []openrtb.Bid{{ID: "bid-1", ImpID: "imp-1", Price: 2.5}}}},
			},
		},
	}

	outcome := auction.Run("req-1", 0.5, results)

	if outcome.Winner == nil {
		t.Fatal("expected a winner")
	}
	if outcome.Winner.Price != 2.5 {
		t.Errorf("expected winning price 2.5, got %f", outcome.Winner.Price)
	}
	if outcome.WinningDSP != "dsp1" {
		t.Errorf("expected winning DSP dsp1, got %s", outcome.WinningDSP)
	}
	if outcome.ClearingPrice != 2.5 {
		t.Errorf("expected clearing price 2.5, got %f", outcome.ClearingPrice)
	}
}

func TestFirstPriceAuction_Run_MultipleBids(t *testing.T) {
	auction := NewFirstPrice()

	results := []dispatcher.Result{
		{
			DSPName: "dsp1",
			Response: &openrtb.BidResponse{
				ID:      "req-1",
				SeatBid: []openrtb.SeatBid{{Bid: []openrtb.Bid{{ID: "bid-1", ImpID: "imp-1", Price: 2.0}}}},
			},
		},
		{
			DSPName: "dsp2",
			Response: &openrtb.BidResponse{
				ID:      "req-1",
				SeatBid: []openrtb.SeatBid{{Bid: []openrtb.Bid{{ID: "bid-2", ImpID: "imp-1", Price: 3.5}}}},
			},
		},
		{
			DSPName: "dsp3",
			Response: &openrtb.BidResponse{
				ID:      "req-1",
				SeatBid: []openrtb.SeatBid{{Bid: []openrtb.Bid{{ID: "bid-3", ImpID: "imp-1", Price: 1.5}}}},
			},
		},
	}

	outcome := auction.Run("req-1", 0.5, results)

	if outcome.Winner == nil {
		t.Fatal("expected a winner")
	}
	if outcome.Winner.ID != "bid-2" {
		t.Errorf("expected winner bid-2, got %s", outcome.Winner.ID)
	}
	if outcome.WinningDSP != "dsp2" {
		t.Errorf("expected winning DSP dsp2, got %s", outcome.WinningDSP)
	}
	if outcome.ClearingPrice != 3.5 {
		t.Errorf("expected clearing price 3.5, got %f", outcome.ClearingPrice)
	}
	if len(outcome.AllBids) != 3 {
		t.Errorf("expected 3 total bids, got %d", len(outcome.AllBids))
	}
}

func TestFirstPriceAuction_Run_NoBids(t *testing.T) {
	auction := NewFirstPrice()

	results := []dispatcher.Result{
		{
			DSPName:  "dsp1",
			Response: &openrtb.BidResponse{ID: "req-1"},
		},
	}

	outcome := auction.Run("req-1", 0.5, results)

	if outcome.Winner != nil {
		t.Error("expected no winner for no bids")
	}
	if outcome.WinningDSP != "" {
		t.Error("expected empty winning DSP")
	}
}

func TestFirstPriceAuction_Run_AllBelowFloor(t *testing.T) {
	auction := NewFirstPrice()

	results := []dispatcher.Result{
		{
			DSPName: "dsp1",
			Response: &openrtb.BidResponse{
				ID:      "req-1",
				SeatBid: []openrtb.SeatBid{{Bid: []openrtb.Bid{{ID: "bid-1", ImpID: "imp-1", Price: 0.3}}}},
			},
		},
		{
			DSPName: "dsp2",
			Response: &openrtb.BidResponse{
				ID:      "req-1",
				SeatBid: []openrtb.SeatBid{{Bid: []openrtb.Bid{{ID: "bid-2", ImpID: "imp-1", Price: 0.4}}}},
			},
		},
	}

	outcome := auction.Run("req-1", 0.5, results)

	if outcome.Winner != nil {
		t.Error("expected no winner when all bids below floor")
	}
	if len(outcome.AllBids) != 0 {
		t.Errorf("expected 0 eligible bids, got %d", len(outcome.AllBids))
	}
}

func TestFirstPriceAuction_Run_SomeAboveFloor(t *testing.T) {
	auction := NewFirstPrice()

	results := []dispatcher.Result{
		{
			DSPName: "dsp1",
			Response: &openrtb.BidResponse{
				ID:      "req-1",
				SeatBid: []openrtb.SeatBid{{Bid: []openrtb.Bid{{ID: "bid-1", ImpID: "imp-1", Price: 0.3}}}},
			},
		},
		{
			DSPName: "dsp2",
			Response: &openrtb.BidResponse{
				ID:      "req-1",
				SeatBid: []openrtb.SeatBid{{Bid: []openrtb.Bid{{ID: "bid-2", ImpID: "imp-1", Price: 1.0}}}},
			},
		},
	}

	outcome := auction.Run("req-1", 0.5, results)

	if outcome.Winner == nil {
		t.Fatal("expected a winner")
	}
	if outcome.Winner.ID != "bid-2" {
		t.Errorf("expected winner bid-2, got %s", outcome.Winner.ID)
	}
	if len(outcome.AllBids) != 1 {
		t.Errorf("expected 1 eligible bid, got %d", len(outcome.AllBids))
	}
}

func TestFirstPriceAuction_Run_MultipleBidsFromOneDSP(t *testing.T) {
	auction := NewFirstPrice()

	results := []dispatcher.Result{
		{
			DSPName: "dsp1",
			Response: &openrtb.BidResponse{
				ID: "req-1",
				SeatBid: []openrtb.SeatBid{{
					Bid: []openrtb.Bid{
						{ID: "bid-1", ImpID: "imp-1", Price: 2.0},
						{ID: "bid-2", ImpID: "imp-1", Price: 3.0},
					},
				}},
			},
		},
	}

	outcome := auction.Run("req-1", 0.5, results)

	if outcome.Winner == nil {
		t.Fatal("expected a winner")
	}
	if outcome.Winner.ID != "bid-2" {
		t.Errorf("expected winner bid-2 (highest), got %s", outcome.Winner.ID)
	}
	if len(outcome.AllBids) != 2 {
		t.Errorf("expected 2 total bids, got %d", len(outcome.AllBids))
	}
}

func TestFirstPriceAuction_Run_WithErrors(t *testing.T) {
	auction := NewFirstPrice()

	results := []dispatcher.Result{
		{
			DSPName: "dsp1",
			Error:   context.DeadlineExceeded,
		},
		{
			DSPName: "dsp2",
			Response: &openrtb.BidResponse{
				ID:      "req-1",
				SeatBid: []openrtb.SeatBid{{Bid: []openrtb.Bid{{ID: "bid-1", ImpID: "imp-1", Price: 2.0}}}},
			},
		},
	}

	outcome := auction.Run("req-1", 0.5, results)

	if outcome.Winner == nil {
		t.Fatal("expected a winner")
	}
	if outcome.WinningDSP != "dsp2" {
		t.Errorf("expected winning DSP dsp2, got %s", outcome.WinningDSP)
	}
}

func TestFirstPriceAuction_Run_ZeroFloor(t *testing.T) {
	auction := NewFirstPrice()

	results := []dispatcher.Result{
		{
			DSPName: "dsp1",
			Response: &openrtb.BidResponse{
				ID:      "req-1",
				SeatBid: []openrtb.SeatBid{{Bid: []openrtb.Bid{{ID: "bid-1", ImpID: "imp-1", Price: 0.01}}}},
			},
		},
	}

	outcome := auction.Run("req-1", 0, results)

	if outcome.Winner == nil {
		t.Fatal("expected a winner with zero floor")
	}
}

var context = struct{ DeadlineExceeded error }{DeadlineExceeded: nil}

func init() {
	context.DeadlineExceeded = testError{}
}

type testError struct{}

func (testError) Error() string { return "deadline exceeded" }
