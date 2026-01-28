// Package dispatcher provides concurrent fan-out of bid requests to multiple DSP endpoints.
// It handles parallel HTTP requests with context cancellation and timeout support.
package dispatcher

import (
	"context"
	"time"

	"github.com/cass/rtb-simulator/internal/config"
	"github.com/cass/rtb-simulator/internal/httpclient"
	"github.com/cass/rtb-simulator/pkg/openrtb"
)

// Result represents the outcome of a bid request to a single DSP.
type Result struct {
	DSPName  string
	Response *openrtb.BidResponse
	Error    error
	Latency  time.Duration
}

// indexedResult pairs a result with its index for channel communication.
// Using a named struct avoids allocation overhead of anonymous structs.
type indexedResult struct {
	idx    int
	result Result
}

// Dispatcher sends bid requests to multiple DSPs concurrently.
type Dispatcher struct {
	client          *httpclient.Client
	dsps            []config.DSPConfig
	timeout         time.Duration
	maxConnsPerHost int
}

// Option configures the dispatcher.
type Option func(*Dispatcher)

// WithTimeout sets the request timeout.
func WithTimeout(d time.Duration) Option {
	return func(dp *Dispatcher) {
		dp.timeout = d
	}
}

// WithMaxConnsPerHost sets the maximum connections per host.
func WithMaxConnsPerHost(n int) Option {
	return func(dp *Dispatcher) {
		dp.maxConnsPerHost = n
	}
}

// New creates a new dispatcher for the given DSPs.
// The dsps slice should contain only enabled DSPs (use Config.EnabledDSPs()).
func New(dsps []config.DSPConfig, opts ...Option) *Dispatcher {
	d := &Dispatcher{
		dsps:            dsps,
		timeout:         100 * time.Millisecond,
		maxConnsPerHost: 100,
	}

	for _, opt := range opts {
		opt(d)
	}

	// Create client after all options are applied
	d.client = httpclient.New(
		httpclient.WithTimeout(d.timeout),
		httpclient.WithMaxConnsPerHost(d.maxConnsPerHost),
	)

	return d
}

// Dispatch sends a bid request to all configured DSPs concurrently
// and returns all results. Respects context cancellation.
func (d *Dispatcher) Dispatch(ctx context.Context, req *openrtb.BidRequest) []Result {
	if len(d.dsps) == 0 {
		return nil
	}

	results := make([]Result, len(d.dsps))
	resultCh := make(chan indexedResult, len(d.dsps))

	// Launch all requests
	for i, dsp := range d.dsps {
		go func(idx int, dspCfg config.DSPConfig) {
			resultCh <- indexedResult{idx, d.callDSP(ctx, dspCfg, req)}
		}(i, dsp)
	}

	// Collect results, respecting context cancellation
	received := 0
	for received < len(d.dsps) {
		select {
		case <-ctx.Done():
			// Context cancelled - fill remaining with errors
			for i := range results {
				if results[i].DSPName == "" {
					results[i] = Result{
						DSPName: d.dsps[i].Name,
						Error:   ctx.Err(),
					}
				}
			}
			return results
		case r := <-resultCh:
			results[r.idx] = r.result
			received++
		}
	}

	return results
}

// callDSP makes a single request to a DSP.
func (d *Dispatcher) callDSP(ctx context.Context, dsp config.DSPConfig, req *openrtb.BidRequest) Result {
	result := Result{DSPName: dsp.Name}

	// Check context before making request
	select {
	case <-ctx.Done():
		result.Error = ctx.Err()
		return result
	default:
	}

	start := time.Now()
	resp, err := d.client.Post(dsp.Endpoint, req)
	result.Latency = time.Since(start)

	if err != nil {
		// Check if context was cancelled during request
		select {
		case <-ctx.Done():
			result.Error = ctx.Err()
		default:
			result.Error = err
		}
		return result
	}

	result.Response = resp
	return result
}

// Close releases resources held by the dispatcher.
func (d *Dispatcher) Close() {
	if d.client != nil {
		d.client.Close()
	}
}

// AllBids extracts all valid bids from the results.
func AllBids(results []Result) []openrtb.Bid {
	var bids []openrtb.Bid
	for _, r := range results {
		if r.Error != nil || r.Response == nil {
			continue
		}
		bids = append(bids, r.Response.AllBids()...)
	}
	return bids
}
