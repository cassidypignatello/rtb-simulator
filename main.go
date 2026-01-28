package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cass/rtb-simulator/internal/api"
	"github.com/cass/rtb-simulator/internal/auction"
	"github.com/cass/rtb-simulator/internal/config"
	"github.com/cass/rtb-simulator/internal/dispatcher"
	"github.com/cass/rtb-simulator/internal/engine"
	"github.com/cass/rtb-simulator/internal/generator"
	"github.com/cass/rtb-simulator/internal/generator/scenarios"
	"github.com/cass/rtb-simulator/internal/stats"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to configuration file")
	autoStart := flag.Bool("auto-start", false, "automatically start simulation on startup")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	log.Printf("RTB Simulator starting...")
	log.Printf("  Server port: %d", cfg.Server.Port)
	log.Printf("  Requests/sec: %d", cfg.Simulation.RequestsPerSecond)
	log.Printf("  Scenario: %s", cfg.Simulation.Scenario)
	log.Printf("  Auction type: %s", cfg.Auction.Type)
	log.Printf("  Timeout: %dms", cfg.Auction.TimeoutMS)
	log.Printf("  DSPs configured: %d (%d enabled)", len(cfg.DSPs), len(cfg.EnabledDSPs()))

	for _, dsp := range cfg.DSPs {
		status := "disabled"
		if dsp.Enabled {
			status = "enabled"
		}
		log.Printf("    - %s: %s [%s]", dsp.Name, dsp.Endpoint, status)
	}

	// Initialize components
	scenario := createScenario(cfg.Simulation.Scenario)
	gen := generator.New(scenario,
		generator.WithTimeout(cfg.Auction.TimeoutMS),
	)

	disp := dispatcher.New(cfg.EnabledDSPs(),
		dispatcher.WithTimeout(time.Duration(cfg.Auction.TimeoutMS)*time.Millisecond),
	)
	defer disp.Close()

	auc := auction.NewFirstPrice()
	collector := stats.New()

	eng := engine.New(gen, disp, auc, collector,
		engine.WithRPS(cfg.Simulation.RequestsPerSecond),
	)

	// Create API server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := api.New(eng, collector, cfg,
		api.WithAddr(addr),
	)

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Start API server
	go func() {
		log.Printf("API server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Server error: %v", err)
		}
	}()

	// Auto-start simulation if requested
	if *autoStart {
		log.Printf("Auto-starting simulation...")
		if err := eng.Start(); err != nil {
			log.Printf("Failed to start simulation: %v", err)
		} else {
			log.Printf("Simulation started")
		}
	} else {
		log.Printf("Simulation ready. POST /start to begin.")
	}

	// Wait for shutdown signal
	sig := <-shutdown
	log.Printf("Received signal %v, shutting down...", sig)

	// Stop simulation if running
	if eng.IsRunning() {
		log.Printf("Stopping simulation...")
		eng.Stop()
	}

	// Shutdown API server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	// Print final stats
	snap := collector.Snapshot()
	log.Printf("Final statistics:")
	log.Printf("  Total requests: %d", snap.TotalRequests)
	log.Printf("  Total bids: %d", snap.TotalBids)
	log.Printf("  Total wins: %d", snap.TotalWins)
	log.Printf("  Total no-bids: %d", snap.TotalNoBids)
	log.Printf("  Total errors: %d", snap.TotalErrors)
	log.Printf("  Total revenue: $%.4f", snap.TotalRevenue)

	log.Printf("Shutdown complete")
}

// createScenario returns the appropriate scenario based on name.
func createScenario(name string) generator.Scenario {
	switch name {
	case "mobile_app":
		return scenarios.NewMobileApp()
	default:
		log.Printf("Unknown scenario %q, defaulting to mobile_app", name)
		return scenarios.NewMobileApp()
	}
}
