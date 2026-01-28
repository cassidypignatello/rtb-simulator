package auction

import (
	"testing"

	"github.com/cass/rtb-simulator/internal/dispatcher"
	"github.com/cass/rtb-simulator/pkg/openrtb"
)

func BenchmarkFirstPrice_Run_5Bids(b *testing.B) {
	auction := NewFirstPrice()

	results := make([]dispatcher.Result, 5)
	for i := range results {
		results[i] = dispatcher.Result{
			DSPName: "dsp",
			Response: &openrtb.BidResponse{
				ID: "req-1",
				SeatBid: []openrtb.SeatBid{{
					Bid: []openrtb.Bid{{
						ID:    "bid",
						ImpID: "imp-1",
						Price: float64(i) + 1.0,
					}},
				}},
			},
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		outcome := auction.Run("req-1", 0.5, results)
		if outcome.Winner == nil {
			b.Fatal("expected winner")
		}
	}
}

func BenchmarkFirstPrice_Run_20Bids(b *testing.B) {
	auction := NewFirstPrice()

	results := make([]dispatcher.Result, 20)
	for i := range results {
		results[i] = dispatcher.Result{
			DSPName: "dsp",
			Response: &openrtb.BidResponse{
				ID: "req-1",
				SeatBid: []openrtb.SeatBid{{
					Bid: []openrtb.Bid{{
						ID:    "bid",
						ImpID: "imp-1",
						Price: float64(i) + 1.0,
					}},
				}},
			},
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		outcome := auction.Run("req-1", 0.5, results)
		if outcome.Winner == nil {
			b.Fatal("expected winner")
		}
	}
}
