package generator

import "github.com/cass/rtb-simulator/pkg/openrtb"

// Scenario defines the interface for bid request generation strategies.
type Scenario interface {
	// Name returns the scenario identifier.
	Name() string

	// Generate creates a new bid request with scenario-specific data.
	// The requestID is provided by the generator for tracking.
	Generate(requestID string) *openrtb.BidRequest
}
