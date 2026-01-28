package httpclient

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cass/rtb-simulator/pkg/openrtb"
)

func BenchmarkClient_Post(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"req-1","seatbid":[{"bid":[{"id":"bid-1","impid":"imp-1","price":2.5}]}]}`))
	}))
	defer server.Close()

	client := New(WithTimeout(5 * time.Second))
	defer client.Close()

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

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := client.Post(server.URL, req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkClient_Post_Parallel(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"req-1","seatbid":[{"bid":[{"id":"bid-1","impid":"imp-1","price":2.5}]}]}`))
	}))
	defer server.Close()

	client := New(WithTimeout(5 * time.Second))
	defer client.Close()

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

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := client.Post(server.URL, req)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
