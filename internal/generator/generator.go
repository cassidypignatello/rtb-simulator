package generator

import (
	"fmt"
	"sync/atomic"

	"github.com/cass/rtb-simulator/pkg/openrtb"
)

// Generator creates OpenRTB bid requests using a configured scenario.
type Generator struct {
	scenario    Scenario
	counter     uint64
	timeout     int
	auctionType int
}

// Option configures the generator.
type Option func(*Generator)

// WithTimeout sets the Tmax value for generated requests.
func WithTimeout(ms int) Option {
	return func(g *Generator) {
		g.timeout = ms
	}
}

// WithAuctionType sets the auction type for generated requests.
func WithAuctionType(at int) Option {
	return func(g *Generator) {
		g.auctionType = at
	}
}

// New creates a new generator with the given scenario and options.
func New(scenario Scenario, opts ...Option) *Generator {
	g := &Generator{
		scenario:    scenario,
		timeout:     100, // default 100ms
		auctionType: openrtb.AuctionFirstPrice,
	}

	for _, opt := range opts {
		opt(g)
	}

	return g
}

// Generate creates a new bid request.
func (g *Generator) Generate() *openrtb.BidRequest {
	id := g.nextID()
	req := g.scenario.Generate(id)

	// Apply generator-level overrides
	if g.timeout > 0 {
		req.Tmax = g.timeout
	}
	if g.auctionType > 0 {
		req.At = g.auctionType
	}

	return req
}

// ScenarioName returns the name of the current scenario.
func (g *Generator) ScenarioName() string {
	return g.scenario.Name()
}

// nextID generates a unique request ID.
func (g *Generator) nextID() string {
	n := atomic.AddUint64(&g.counter, 1)
	return fmt.Sprintf("req-%08d", n)
}
