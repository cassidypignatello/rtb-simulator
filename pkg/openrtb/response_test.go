package openrtb

import (
	"encoding/json"
	"testing"
)

func TestBidResponse_JSON(t *testing.T) {
	resp := &BidResponse{
		ID: "resp-12345",
		SeatBid: []SeatBid{
			{
				Seat: "seat-001",
				Bid: []Bid{
					{
						ID:    "bid-1",
						ImpID: "imp-1",
						Price: 2.50,
						AdM:   "<div>ad markup</div>",
						CrID:  "creative-123",
						W:     320,
						H:     50,
					},
				},
			},
		},
		Cur: "USD",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded BidResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.ID != resp.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, resp.ID)
	}
	if len(decoded.SeatBid) != 1 {
		t.Fatalf("len(SeatBid) = %d, want 1", len(decoded.SeatBid))
	}
	if len(decoded.SeatBid[0].Bid) != 1 {
		t.Fatalf("len(Bid) = %d, want 1", len(decoded.SeatBid[0].Bid))
	}

	bid := decoded.SeatBid[0].Bid[0]
	if bid.Price != 2.50 {
		t.Errorf("Price = %f, want 2.50", bid.Price)
	}
	if bid.ImpID != "imp-1" {
		t.Errorf("ImpID = %q, want %q", bid.ImpID, "imp-1")
	}
}

func TestBidResponse_MultipleBids(t *testing.T) {
	resp := &BidResponse{
		ID: "resp-multi",
		SeatBid: []SeatBid{
			{
				Seat: "seat-1",
				Bid: []Bid{
					{ID: "bid-1", ImpID: "imp-1", Price: 1.50},
					{ID: "bid-2", ImpID: "imp-2", Price: 2.00},
				},
			},
			{
				Seat: "seat-2",
				Bid: []Bid{
					{ID: "bid-3", ImpID: "imp-1", Price: 1.75},
				},
			},
		},
		Cur: "USD",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded BidResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if len(decoded.SeatBid) != 2 {
		t.Fatalf("len(SeatBid) = %d, want 2", len(decoded.SeatBid))
	}

	totalBids := 0
	for _, sb := range decoded.SeatBid {
		totalBids += len(sb.Bid)
	}
	if totalBids != 3 {
		t.Errorf("total bids = %d, want 3", totalBids)
	}
}

func TestBidResponse_NoBid(t *testing.T) {
	// Empty seatbid array represents no-bid
	resp := &BidResponse{
		ID:      "resp-nobid",
		SeatBid: []SeatBid{},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded BidResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if len(decoded.SeatBid) != 0 {
		t.Errorf("len(SeatBid) = %d, want 0 for no-bid", len(decoded.SeatBid))
	}
}

func TestBidResponse_JSONFieldNames(t *testing.T) {
	resp := &BidResponse{
		ID: "test",
		SeatBid: []SeatBid{
			{
				Seat: "s1",
				Bid: []Bid{
					{
						ID:    "b1",
						ImpID: "i1",
						Price: 1.0,
						AdM:   "markup",
						CrID:  "cr1",
					},
				},
			},
		},
		Cur: "USD",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// Check that OpenRTB field names are used
	jsonStr := string(data)

	expectedFields := []string{`"id"`, `"seatbid"`, `"cur"`, `"impid"`, `"price"`, `"adm"`, `"crid"`}
	for _, field := range expectedFields {
		if !contains(jsonStr, field) {
			t.Errorf("expected field %s in JSON", field)
		}
	}
}

func TestBid_OptionalFields(t *testing.T) {
	// Bid with only required fields
	bid := Bid{
		ID:    "bid-1",
		ImpID: "imp-1",
		Price: 1.0,
	}

	data, err := json.Marshal(bid)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	// Optional fields should be omitted when empty
	if _, ok := m["adm"]; ok {
		t.Error("adm should be omitted when empty")
	}
	if _, ok := m["crid"]; ok {
		t.Error("crid should be omitted when empty")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
