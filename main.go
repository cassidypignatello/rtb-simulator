package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/cass/rtb-simulator/internal/config"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to configuration file")
	flag.Parse()

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

	log.Printf("Config loaded successfully")
}
