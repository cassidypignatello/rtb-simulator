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

func TestBidResponse_IsNoBid(t *testing.T) {
	tests := []struct {
		name string
		resp BidResponse
		want bool
	}{
		{
			name: "empty seatbid",
			resp: BidResponse{ID: "1", SeatBid: []SeatBid{}},
			want: true,
		},
		{
			name: "nil seatbid",
			resp: BidResponse{ID: "1"},
			want: true,
		},
		{
			name: "seatbid with empty bids",
			resp: BidResponse{
				ID:      "1",
				SeatBid: []SeatBid{{Seat: "s1", Bid: []Bid{}}},
			},
			want: true,
		},
		{
			name: "has bids",
			resp: BidResponse{
				ID: "1",
				SeatBid: []SeatBid{
					{Seat: "s1", Bid: []Bid{{ID: "b1", Price: 1.0}}},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.resp.IsNoBid(); got != tt.want {
				t.Errorf("IsNoBid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBidResponse_AllBids(t *testing.T) {
	resp := BidResponse{
		ID: "1",
		SeatBid: []SeatBid{
			{Seat: "s1", Bid: []Bid{{ID: "b1"}, {ID: "b2"}}},
			{Seat: "s2", Bid: []Bid{{ID: "b3"}}},
		},
	}

	bids := resp.AllBids()
	if len(bids) != 3 {
		t.Errorf("AllBids() returned %d bids, want 3", len(bids))
	}
}

func TestBidResponse_HighestBid(t *testing.T) {
	tests := []struct {
		name      string
		resp      BidResponse
		wantID    string
		wantPrice float64
		wantNil   bool
	}{
		{
			name:    "no bids",
			resp:    BidResponse{ID: "1", SeatBid: []SeatBid{}},
			wantNil: true,
		},
		{
			name: "single bid",
			resp: BidResponse{
				ID:      "1",
				SeatBid: []SeatBid{{Bid: []Bid{{ID: "b1", Price: 1.5}}}},
			},
			wantID:    "b1",
			wantPrice: 1.5,
		},
		{
			name: "multiple bids",
			resp: BidResponse{
				ID: "1",
				SeatBid: []SeatBid{
					{Bid: []Bid{{ID: "b1", Price: 1.5}, {ID: "b2", Price: 2.5}}},
					{Bid: []Bid{{ID: "b3", Price: 2.0}}},
				},
			},
			wantID:    "b2",
			wantPrice: 2.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bid := tt.resp.HighestBid()
			if tt.wantNil {
				if bid != nil {
					t.Errorf("HighestBid() = %v, want nil", bid)
				}
				return
			}
			if bid == nil {
				t.Fatal("HighestBid() = nil, want non-nil")
			}
			if bid.ID != tt.wantID {
				t.Errorf("HighestBid().ID = %q, want %q", bid.ID, tt.wantID)
			}
			if bid.Price != tt.wantPrice {
				t.Errorf("HighestBid().Price = %f, want %f", bid.Price, tt.wantPrice)
			}
		})
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
