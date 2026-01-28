package generator

import (
	"testing"

	"github.com/cass/rtb-simulator/pkg/openrtb"
)

// mockScenario implements Scenario for testing
type mockScenario struct {
	name string
}

func (m *mockScenario) Name() string {
	return m.name
}

func (m *mockScenario) Generate(requestID string) *openrtb.BidRequest {
	return &openrtb.BidRequest{
		ID: requestID,
		Imp: []openrtb.Imp{
			{
				ID: "imp-1",
				Banner: &openrtb.Banner{
					W: 320,
					H: 50,
				},
				BidFloor: 1.0,
			},
		},
		At:   openrtb.AuctionFirstPrice,
		Tmax: 100,
		Cur:  []string{"USD"},
	}
}

func TestNewGenerator(t *testing.T) {
	scenario := &mockScenario{name: "test-scenario"}
	gen := New(scenario)

	if gen == nil {
		t.Fatal("New() returned nil")
	}
	if gen.scenario != scenario {
		t.Error("scenario not set correctly")
	}
}

func TestGenerator_Generate(t *testing.T) {
	scenario := &mockScenario{name: "test-scenario"}
	gen := New(scenario)

	req := gen.Generate()

	if req == nil {
		t.Fatal("Generate() returned nil")
	}
	if req.ID == "" {
		t.Error("request ID should not be empty")
	}
	if len(req.Imp) == 0 {
		t.Error("request should have impressions")
	}
}

func TestGenerator_UniqueIDs(t *testing.T) {
	scenario := &mockScenario{name: "test-scenario"}
	gen := New(scenario)

	ids := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		req := gen.Generate()
		if ids[req.ID] {
			t.Errorf("duplicate request ID: %s", req.ID)
		}
		ids[req.ID] = true
	}
}

func TestGenerator_IDFormat(t *testing.T) {
	scenario := &mockScenario{name: "test-scenario"}
	gen := New(scenario)

	req := gen.Generate()

	// ID should start with "req-"
	if len(req.ID) < 4 || req.ID[:4] != "req-" {
		t.Errorf("ID format unexpected: %s", req.ID)
	}
}

func TestGenerator_ScenarioName(t *testing.T) {
	scenario := &mockScenario{name: "mobile_app"}
	gen := New(scenario)

	if gen.ScenarioName() != "mobile_app" {
		t.Errorf("ScenarioName() = %q, want %q", gen.ScenarioName(), "mobile_app")
	}
}

func TestGenerator_WithTimeout(t *testing.T) {
	scenario := &mockScenario{name: "test-scenario"}
	gen := New(scenario, WithTimeout(150))

	req := gen.Generate()

	if req.Tmax != 150 {
		t.Errorf("Tmax = %d, want 150", req.Tmax)
	}
}

func TestGenerator_WithAuctionType(t *testing.T) {
	scenario := &mockScenario{name: "test-scenario"}
	gen := New(scenario, WithAuctionType(openrtb.AuctionSecondPrice))

	req := gen.Generate()

	if req.At != openrtb.AuctionSecondPrice {
		t.Errorf("At = %d, want %d", req.At, openrtb.AuctionSecondPrice)
	}
}

func TestGenerator_ConcurrentSafety(t *testing.T) {
	scenario := &mockScenario{name: "test-scenario"}
	gen := New(scenario)

	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				req := gen.Generate()
				if req.ID == "" {
					t.Error("empty ID in concurrent generation")
				}
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
