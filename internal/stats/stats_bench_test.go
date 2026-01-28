package stats

import (
	"testing"
	"time"

	"github.com/cass/rtb-simulator/internal/auction"
	"github.com/cass/rtb-simulator/internal/dispatcher"
	"github.com/cass/rtb-simulator/pkg/openrtb"
)

func BenchmarkCollector_RecordAuction(b *testing.B) {
	c := New()

	outcome := auction.Outcome{
		RequestID:     "req-1",
		Winner:        &openrtb.Bid{ID: "bid-1", Price: 2.5},
		WinningDSP:    "dsp1",
		ClearingPrice: 2.5,
		AllBids: []auction.BidWithDSP{
			{Bid: openrtb.Bid{ID: "bid-1", Price: 2.5}, DSPName: "dsp1"},
			{Bid: openrtb.Bid{ID: "bid-2", Price: 1.5}, DSPName: "dsp2"},
			{Bid: openrtb.Bid{ID: "bid-3", Price: 1.0}, DSPName: "dsp3"},
		},
	}

	results := []dispatcher.Result{
		{DSPName: "dsp1", Latency: 10 * time.Millisecond},
		{DSPName: "dsp2", Latency: 15 * time.Millisecond},
		{DSPName: "dsp3", Latency: 12 * time.Millisecond},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		c.RecordAuction(outcome, results)
	}
}

func BenchmarkCollector_RecordAuction_Parallel(b *testing.B) {
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

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.RecordAuction(outcome, results)
		}
	})
}

func BenchmarkCollector_Snapshot(b *testing.B) {
	c := New()

	// Pre-populate with some data
	for i := 0; i < 100; i++ {
		outcome := auction.Outcome{
			RequestID:     "req",
			Winner:        &openrtb.Bid{ID: "bid", Price: 2.5},
			WinningDSP:    "dsp1",
			ClearingPrice: 2.5,
			AllBids: []auction.BidWithDSP{
				{Bid: openrtb.Bid{ID: "bid", Price: 2.5}, DSPName: "dsp1"},
			},
		}
		results := []dispatcher.Result{
			{DSPName: "dsp1", Latency: 10 * time.Millisecond},
			{DSPName: "dsp2", Latency: 15 * time.Millisecond},
			{DSPName: "dsp3", Latency: 12 * time.Millisecond},
		}
		c.RecordAuction(outcome, results)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = c.Snapshot()
	}
}
