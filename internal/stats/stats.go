package stats

import (
	"sync"
	"time"

	"github.com/cass/rtb-simulator/internal/auction"
	"github.com/cass/rtb-simulator/internal/dispatcher"
)

// Collector aggregates auction statistics in a thread-safe manner.
type Collector struct {
	mu sync.RWMutex

	totalRequests uint64
	totalBids     uint64
	totalWins     uint64
	totalNoBids   uint64
	totalErrors   uint64
	totalRevenue  float64

	dspStats map[string]*dspStatsInternal
}

// dspStatsInternal holds per-DSP statistics (internal mutable version).
type dspStatsInternal struct {
	requests     uint64
	bids         uint64
	wins         uint64
	noBids       uint64
	errors       uint64
	totalLatency time.Duration
}

// New creates a new statistics collector.
func New() *Collector {
	return &Collector{
		dspStats: make(map[string]*dspStatsInternal),
	}
}

// RecordAuction records the outcome of a single auction.
func (c *Collector) RecordAuction(outcome auction.Outcome, results []dispatcher.Result) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.totalRequests++
	c.totalBids += uint64(len(outcome.AllBids))

	if outcome.Winner != nil {
		c.totalWins++
		c.totalRevenue += outcome.ClearingPrice
	} else {
		c.totalNoBids++
	}

	// Track per-DSP stats from results
	for _, r := range results {
		dsp := c.getOrCreateDSP(r.DSPName)
		dsp.requests++
		dsp.totalLatency += r.Latency

		if r.Error != nil {
			dsp.errors++
			c.totalErrors++
		} else if r.Response != nil && r.Response.IsNoBid() {
			dsp.noBids++
		}
	}

	// Track bids per DSP from outcome
	bidsByDSP := make(map[string]int)
	for _, b := range outcome.AllBids {
		bidsByDSP[b.DSPName]++
	}
	for dspName, count := range bidsByDSP {
		dsp := c.getOrCreateDSP(dspName)
		dsp.bids += uint64(count)
	}

	// Track wins per DSP
	if outcome.Winner != nil && outcome.WinningDSP != "" {
		dsp := c.getOrCreateDSP(outcome.WinningDSP)
		dsp.wins++
	}
}

// getOrCreateDSP returns the DSP stats, creating it if necessary.
// Must be called with mu held.
func (c *Collector) getOrCreateDSP(name string) *dspStatsInternal {
	if dsp, ok := c.dspStats[name]; ok {
		return dsp
	}
	dsp := &dspStatsInternal{}
	c.dspStats[name] = dsp
	return dsp
}

// Snapshot returns a point-in-time copy of all statistics.
func (c *Collector) Snapshot() Snapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()

	snap := Snapshot{
		TotalRequests: c.totalRequests,
		TotalBids:     c.totalBids,
		TotalWins:     c.totalWins,
		TotalNoBids:   c.totalNoBids,
		TotalErrors:   c.totalErrors,
		TotalRevenue:  c.totalRevenue,
		DSPStats:      make(map[string]DSPStats, len(c.dspStats)),
	}

	for name, internal := range c.dspStats {
		var avgLatency time.Duration
		if internal.requests > 0 {
			avgLatency = internal.totalLatency / time.Duration(internal.requests)
		}

		snap.DSPStats[name] = DSPStats{
			Requests:   internal.requests,
			Bids:       internal.bids,
			Wins:       internal.wins,
			NoBids:     internal.noBids,
			Errors:     internal.errors,
			AvgLatency: avgLatency,
		}
	}

	return snap
}

// Reset clears all statistics.
func (c *Collector) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.totalRequests = 0
	c.totalBids = 0
	c.totalWins = 0
	c.totalNoBids = 0
	c.totalErrors = 0
	c.totalRevenue = 0
	c.dspStats = make(map[string]*dspStatsInternal)
}

// Snapshot represents a point-in-time copy of statistics.
type Snapshot struct {
	TotalRequests uint64
	TotalBids     uint64
	TotalWins     uint64
	TotalNoBids   uint64
	TotalErrors   uint64
	TotalRevenue  float64
	DSPStats      map[string]DSPStats
}

// DSPStats holds per-DSP statistics.
type DSPStats struct {
	Requests   uint64
	Bids       uint64
	Wins       uint64
	NoBids     uint64
	Errors     uint64
	AvgLatency time.Duration
}
