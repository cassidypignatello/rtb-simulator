package engine

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/cass/rtb-simulator/internal/auction"
	"github.com/cass/rtb-simulator/internal/dispatcher"
	"github.com/cass/rtb-simulator/internal/stats"
	"github.com/cass/rtb-simulator/pkg/openrtb"
)

var (
	ErrAlreadyRunning = errors.New("engine is already running")
	ErrNotRunning     = errors.New("engine is not running")
)

// Generator defines the interface for bid request generation.
type Generator interface {
	Generate() *openrtb.BidRequest
	ScenarioName() string
}

// Dispatcher defines the interface for dispatching requests to DSPs.
type Dispatcher interface {
	Dispatch(ctx context.Context, req *openrtb.BidRequest) []dispatcher.Result
	Close()
}

// Engine orchestrates the RTB simulation loop.
type Engine struct {
	generator  Generator
	dispatcher Dispatcher
	auction    auction.Auction
	stats      *stats.Collector

	rps      int
	bidFloor float64

	mu       sync.RWMutex
	running  bool
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// Option configures the engine.
type Option func(*Engine)

// WithRPS sets the requests per second rate.
func WithRPS(rps int) Option {
	return func(e *Engine) {
		e.rps = rps
	}
}

// WithBidFloor sets the minimum bid floor for auctions.
func WithBidFloor(floor float64) Option {
	return func(e *Engine) {
		e.bidFloor = floor
	}
}

// New creates a new simulation engine.
func New(gen Generator, disp Dispatcher, auc auction.Auction, stats *stats.Collector, opts ...Option) *Engine {
	e := &Engine{
		generator:  gen,
		dispatcher: disp,
		auction:    auc,
		stats:      stats,
		rps:        100,      // default 100 RPS
		bidFloor:   0.01,     // default $0.01 floor
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

// Start begins the simulation loop.
func (e *Engine) Start() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.running {
		return ErrAlreadyRunning
	}

	ctx, cancel := context.WithCancel(context.Background())
	e.cancel = cancel
	e.running = true

	e.wg.Add(1)
	go e.loop(ctx)

	return nil
}

// Stop halts the simulation loop.
func (e *Engine) Stop() {
	e.mu.Lock()
	cancel := e.cancel
	e.mu.Unlock()

	if cancel != nil {
		cancel()
	}

	e.wg.Wait()

	e.mu.Lock()
	e.running = false
	e.cancel = nil
	e.mu.Unlock()
}

// Shutdown gracefully stops the engine with context timeout.
func (e *Engine) Shutdown(ctx context.Context) error {
	e.mu.Lock()
	cancel := e.cancel
	e.mu.Unlock()

	if cancel != nil {
		cancel()
	}

	done := make(chan struct{})
	go func() {
		e.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		e.mu.Lock()
		e.running = false
		e.cancel = nil
		e.mu.Unlock()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// IsRunning returns whether the engine is currently running.
func (e *Engine) IsRunning() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.running
}

// loop runs the main simulation loop.
func (e *Engine) loop(ctx context.Context) {
	defer e.wg.Done()

	interval := time.Second / time.Duration(e.rps)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.tick(ctx)
		}
	}
}

// tick performs a single simulation cycle.
func (e *Engine) tick(ctx context.Context) {
	// Generate request
	req := e.generator.Generate()

	// Get bid floor from first impression if available
	bidFloor := e.bidFloor
	if len(req.Imp) > 0 && req.Imp[0].BidFloor > 0 {
		bidFloor = req.Imp[0].BidFloor
	}

	// Dispatch to DSPs
	results := e.dispatcher.Dispatch(ctx, req)

	// Run auction
	outcome := e.auction.Run(req.ID, bidFloor, results)

	// Record stats
	e.stats.RecordAuction(outcome, results)
}
