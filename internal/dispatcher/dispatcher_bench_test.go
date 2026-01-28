package dispatcher

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cass/rtb-simulator/internal/config"
	"github.com/cass/rtb-simulator/pkg/openrtb"
)

func BenchmarkDispatcher_Dispatch_3DSPs(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	req := &openrtb.BidRequest{
		ID:   "req-1",
		Tmax: 100,
		At:   1,
		Imp: []openrtb.Imp{{
			ID:       "imp-1",
			BidFloor: 0.5,
			Banner:   &openrtb.Banner{W: 320, H: 50},
		}},
	}

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		results := d.Dispatch(ctx, req)
		if len(results) != 3 {
			b.Fatalf("expected 3 results, got %d", len(results))
		}
	}
}

func BenchmarkDispatcher_Dispatch_10DSPs(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"req-1","seatbid":[{"bid":[{"id":"bid-1","impid":"imp-1","price":2.5}]}]}`))
	}))
	defer server.Close()

	dsps := make([]config.DSPConfig, 10)
	for i := range dsps {
		dsps[i] = config.DSPConfig{
			Name:     "dsp",
			Endpoint: server.URL,
			Enabled:  true,
		}
	}

	d := New(dsps, WithTimeout(5*time.Second))

	req := &openrtb.BidRequest{
		ID:   "req-1",
		Tmax: 100,
		At:   1,
		Imp: []openrtb.Imp{{
			ID:       "imp-1",
			BidFloor: 0.5,
			Banner:   &openrtb.Banner{W: 320, H: 50},
		}},
	}

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		results := d.Dispatch(ctx, req)
		if len(results) != 10 {
			b.Fatalf("expected 10 results, got %d", len(results))
		}
	}
}
