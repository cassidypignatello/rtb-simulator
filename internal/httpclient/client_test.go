package httpclient

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cass/rtb-simulator/pkg/openrtb"
)

func TestClient_Post_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"req-1","seatbid":[{"bid":[{"id":"bid-1","impid":"imp-1","price":2.5}]}]}`))
	}))
	defer server.Close()

	client := New(WithTimeout(5 * time.Second))
	defer client.Close()

	req := &openrtb.BidRequest{ID: "req-1"}
	resp, err := client.Post(server.URL, req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "req-1" {
		t.Errorf("expected ID req-1, got %s", resp.ID)
	}
	if len(resp.SeatBid) != 1 {
		t.Errorf("expected 1 seatbid, got %d", len(resp.SeatBid))
	}
	if resp.SeatBid[0].Bid[0].Price != 2.5 {
		t.Errorf("expected price 2.5, got %f", resp.SeatBid[0].Bid[0].Price)
	}
}

func TestClient_Post_NoBid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := New(WithTimeout(5 * time.Second))
	defer client.Close()

	req := &openrtb.BidRequest{ID: "req-1"}
	resp, err := client.Post(server.URL, req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.IsNoBid() {
		t.Error("expected no bid response")
	}
}

func TestClient_Post_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(WithTimeout(50 * time.Millisecond))
	defer client.Close()

	req := &openrtb.BidRequest{ID: "req-1"}
	_, err := client.Post(server.URL, req)

	if err == nil {
		t.Error("expected timeout error")
	}
	if !IsTimeout(err) {
		t.Errorf("expected timeout error, got: %v", err)
	}
}

func TestClient_Post_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := New(WithTimeout(5 * time.Second))
	defer client.Close()

	req := &openrtb.BidRequest{ID: "req-1"}
	_, err := client.Post(server.URL, req)

	if err == nil {
		t.Error("expected error for 500 response")
	}
}

func TestClient_Post_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid json}`))
	}))
	defer server.Close()

	client := New(WithTimeout(5 * time.Second))
	defer client.Close()

	req := &openrtb.BidRequest{ID: "req-1"}
	_, err := client.Post(server.URL, req)

	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestClient_Post_ConnectionRefused(t *testing.T) {
	client := New(WithTimeout(1 * time.Second))
	defer client.Close()

	req := &openrtb.BidRequest{ID: "req-1"}
	_, err := client.Post("http://localhost:59999", req)

	if err == nil {
		t.Error("expected connection error")
	}
}

func TestClientOptions(t *testing.T) {
	client := New(
		WithTimeout(2*time.Second),
		WithMaxConnsPerHost(50),
		WithMaxIdleConns(100),
	)
	defer client.Close()

	if client.timeout != 2*time.Second {
		t.Errorf("expected timeout 2s, got %v", client.timeout)
	}
	if client.maxConnsPerHost != 50 {
		t.Errorf("expected maxConnsPerHost 50, got %d", client.maxConnsPerHost)
	}
	if client.maxIdleConns != 100 {
		t.Errorf("expected maxIdleConns 100, got %d", client.maxIdleConns)
	}
}
