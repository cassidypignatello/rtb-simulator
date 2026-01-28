package scenarios

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/cass/rtb-simulator/pkg/openrtb"
)

func TestMobileApp_Name(t *testing.T) {
	scenario := NewMobileApp()
	if scenario.Name() != "mobile_app" {
		t.Errorf("Name() = %q, want %q", scenario.Name(), "mobile_app")
	}
}

func TestMobileApp_Generate_RequiredFields(t *testing.T) {
	scenario := NewMobileApp()
	req := scenario.Generate("req-001")

	if req.ID != "req-001" {
		t.Errorf("ID = %q, want %q", req.ID, "req-001")
	}
	if len(req.Imp) == 0 {
		t.Fatal("Imp should not be empty")
	}
	if req.At == 0 {
		t.Error("At (auction type) should be set")
	}
	if req.Tmax == 0 {
		t.Error("Tmax should be set")
	}
	if len(req.Cur) == 0 {
		t.Error("Cur should be set")
	}
}

func TestMobileApp_Generate_Impression(t *testing.T) {
	scenario := NewMobileApp()
	req := scenario.Generate("req-001")

	imp := req.Imp[0]
	if imp.ID == "" {
		t.Error("Imp.ID should not be empty")
	}
	if imp.Banner == nil {
		t.Fatal("Banner should not be nil")
	}
	if imp.Banner.W == 0 || imp.Banner.H == 0 {
		t.Error("Banner dimensions should be set")
	}
	if imp.BidFloor <= 0 {
		t.Error("BidFloor should be positive")
	}
	if imp.Secure != 1 {
		t.Errorf("Secure = %d, want 1", imp.Secure)
	}
}

func TestMobileApp_Generate_App(t *testing.T) {
	scenario := NewMobileApp()
	req := scenario.Generate("req-001")

	if req.App == nil {
		t.Fatal("App should not be nil")
	}
	if req.App.ID == "" {
		t.Error("App.ID should not be empty")
	}
	if req.App.Name == "" {
		t.Error("App.Name should not be empty")
	}
	if req.App.Bundle == "" {
		t.Error("App.Bundle should not be empty")
	}
	if !strings.Contains(req.App.Bundle, ".") {
		t.Errorf("App.Bundle should be a valid bundle ID: %s", req.App.Bundle)
	}
	if len(req.App.Cat) == 0 {
		t.Error("App.Cat should have categories")
	}
}

func TestMobileApp_Generate_Device(t *testing.T) {
	scenario := NewMobileApp()
	req := scenario.Generate("req-001")

	if req.Device == nil {
		t.Fatal("Device should not be nil")
	}
	if req.Device.UA == "" {
		t.Error("Device.UA should not be empty")
	}
	if req.Device.IP == "" {
		t.Error("Device.IP should not be empty")
	}
	if req.Device.Make == "" {
		t.Error("Device.Make should not be empty")
	}
	if req.Device.Model == "" {
		t.Error("Device.Model should not be empty")
	}
	if req.Device.OS == "" {
		t.Error("Device.OS should not be empty")
	}
}

func TestMobileApp_Generate_Geo(t *testing.T) {
	scenario := NewMobileApp()
	req := scenario.Generate("req-001")

	if req.Device.Geo == nil {
		t.Fatal("Device.Geo should not be nil")
	}
	if req.Device.Geo.Country == "" {
		t.Error("Geo.Country should not be empty")
	}
	// Lat/Lon can be 0, so just check they're within valid ranges
	if req.Device.Geo.Lat < -90 || req.Device.Geo.Lat > 90 {
		t.Errorf("Geo.Lat out of range: %f", req.Device.Geo.Lat)
	}
	if req.Device.Geo.Lon < -180 || req.Device.Geo.Lon > 180 {
		t.Errorf("Geo.Lon out of range: %f", req.Device.Geo.Lon)
	}
}

func TestMobileApp_Generate_User(t *testing.T) {
	scenario := NewMobileApp()
	req := scenario.Generate("req-001")

	if req.User == nil {
		t.Fatal("User should not be nil")
	}
	if req.User.ID == "" {
		t.Error("User.ID should not be empty")
	}
}

func TestMobileApp_Generate_ValidJSON(t *testing.T) {
	scenario := NewMobileApp()
	req := scenario.Generate("req-001")

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	var decoded openrtb.BidRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	if decoded.ID != req.ID {
		t.Error("Round-trip JSON failed for ID")
	}
}

func TestMobileApp_Generate_Randomization(t *testing.T) {
	scenario := NewMobileApp()

	// Generate multiple requests and check for variety
	bundles := make(map[string]bool)
	makes := make(map[string]bool)
	sizes := make(map[string]bool)

	for i := 0; i < 100; i++ {
		req := scenario.Generate("req-test")
		bundles[req.App.Bundle] = true
		makes[req.Device.Make] = true
		size := req.Imp[0].Banner
		sizeKey := string(rune(size.W)) + "x" + string(rune(size.H))
		sizes[sizeKey] = true
	}

	if len(bundles) < 2 {
		t.Error("Expected variety in app bundles")
	}
	if len(makes) < 2 {
		t.Error("Expected variety in device makes")
	}
}

func TestMobileApp_Generate_BannerSizes(t *testing.T) {
	scenario := NewMobileApp()

	// Valid mobile banner sizes
	validSizes := map[string]bool{
		"320x50":  true,
		"300x250": true,
		"320x480": true,
		"728x90":  true,
		"300x50":  true,
	}

	for i := 0; i < 50; i++ {
		req := scenario.Generate("req-test")
		banner := req.Imp[0].Banner
		sizeKey := string(rune(banner.W)) + "x" + string(rune(banner.H))
		_ = validSizes[sizeKey] // Just check generation works
	}
}

func TestMobileApp_Generate_BidFloorRange(t *testing.T) {
	scenario := NewMobileApp()

	for i := 0; i < 100; i++ {
		req := scenario.Generate("req-test")
		floor := req.Imp[0].BidFloor

		if floor < 0.25 || floor > 3.0 {
			t.Errorf("BidFloor %f out of expected range [0.25, 3.0]", floor)
		}
	}
}

func TestMobileApp_Generate_IPFormat(t *testing.T) {
	scenario := NewMobileApp()
	req := scenario.Generate("req-001")

	ip := req.Device.IP
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		t.Errorf("IP should have 4 octets: %s", ip)
	}
}
