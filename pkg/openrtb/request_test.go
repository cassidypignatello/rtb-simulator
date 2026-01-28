package openrtb

import (
	"encoding/json"
	"testing"
)

func TestBidRequest_JSON(t *testing.T) {
	req := &BidRequest{
		ID: "req-12345",
		Imp: []Imp{
			{
				ID: "imp-1",
				Banner: &Banner{
					W:     320,
					H:     50,
					Btype: []int{1, 2},
				},
				BidFloor: 0.75,
				Secure:   1,
			},
		},
		App: &App{
			ID:     "app-001",
			Name:   "Test App",
			Bundle: "com.test.app",
			Cat:    []string{"IAB9-30"},
		},
		Device: &Device{
			UA:    "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X)",
			IP:    "192.168.1.1",
			Make:  "Apple",
			Model: "iPhone14,2",
			OS:    "iOS",
			OSV:   "16.0",
			Geo: &Geo{
				Lat:     37.7749,
				Lon:     -122.4194,
				Country: "USA",
			},
		},
		User: &User{
			ID: "user-abc123",
		},
		At:   1,
		Tmax: 100,
		Cur:  []string{"USD"},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded BidRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.ID != req.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, req.ID)
	}
	if len(decoded.Imp) != 1 {
		t.Fatalf("len(Imp) = %d, want 1", len(decoded.Imp))
	}
	if decoded.Imp[0].Banner.W != 320 {
		t.Errorf("Banner.W = %d, want 320", decoded.Imp[0].Banner.W)
	}
	if decoded.App.Bundle != "com.test.app" {
		t.Errorf("App.Bundle = %q, want %q", decoded.App.Bundle, "com.test.app")
	}
	if decoded.Device.Geo.Lat != 37.7749 {
		t.Errorf("Geo.Lat = %f, want 37.7749", decoded.Device.Geo.Lat)
	}
	if decoded.At != 1 {
		t.Errorf("At = %d, want 1", decoded.At)
	}
}

func TestBidRequest_MinimalJSON(t *testing.T) {
	req := &BidRequest{
		ID: "minimal-req",
		Imp: []Imp{
			{
				ID:       "imp-1",
				BidFloor: 0.5,
			},
		},
		At:   1,
		Tmax: 100,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// Verify omitempty works - no app, device, user fields
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Unmarshal to map error: %v", err)
	}

	if _, ok := m["app"]; ok {
		t.Error("app should be omitted when nil")
	}
	if _, ok := m["device"]; ok {
		t.Error("device should be omitted when nil")
	}
	if _, ok := m["user"]; ok {
		t.Error("user should be omitted when nil")
	}
}

func TestBidRequest_JSONFieldNames(t *testing.T) {
	req := &BidRequest{
		ID: "test",
		Imp: []Imp{
			{
				ID:       "imp-1",
				BidFloor: 1.5,
				Secure:   1,
			},
		},
		At:   2,
		Tmax: 150,
		Cur:  []string{"USD", "EUR"},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	// Check OpenRTB field names
	if _, ok := m["id"]; !ok {
		t.Error("expected 'id' field")
	}
	if _, ok := m["imp"]; !ok {
		t.Error("expected 'imp' field")
	}
	if _, ok := m["at"]; !ok {
		t.Error("expected 'at' field")
	}
	if _, ok := m["tmax"]; !ok {
		t.Error("expected 'tmax' field")
	}
	if _, ok := m["cur"]; !ok {
		t.Error("expected 'cur' field")
	}
}

func TestImp_BidFloorJSON(t *testing.T) {
	imp := Imp{
		ID:       "imp-1",
		BidFloor: 0.0,
	}

	data, err := json.Marshal(imp)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// bidfloor should be present even when 0
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if _, ok := m["bidfloor"]; !ok {
		t.Error("bidfloor should be present even when 0")
	}
}

func TestBanner_Sizes(t *testing.T) {
	banner := Banner{
		W:    300,
		H:    250,
		Wmax: 320,
		Hmax: 480,
	}

	data, err := json.Marshal(banner)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded Banner
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.W != 300 || decoded.H != 250 {
		t.Errorf("Size = %dx%d, want 300x250", decoded.W, decoded.H)
	}
	if decoded.Wmax != 320 || decoded.Hmax != 480 {
		t.Errorf("Max size = %dx%d, want 320x480", decoded.Wmax, decoded.Hmax)
	}
}
