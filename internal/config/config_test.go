package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_ValidConfig(t *testing.T) {
	content := `
server:
  port: 8080

simulation:
  requests_per_second: 10
  scenario: "mobile_app"

auction:
  type: "first_price"
  timeout_ms: 100

dsps:
  - name: "test-dsp"
    endpoint: "http://localhost:9000/bid"
    enabled: true
`
	path := createTempConfig(t, content)
	defer os.Remove(path)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want 8080", cfg.Server.Port)
	}
	if cfg.Simulation.RequestsPerSecond != 10 {
		t.Errorf("Simulation.RequestsPerSecond = %d, want 10", cfg.Simulation.RequestsPerSecond)
	}
	if cfg.Simulation.Scenario != "mobile_app" {
		t.Errorf("Simulation.Scenario = %q, want %q", cfg.Simulation.Scenario, "mobile_app")
	}
	if cfg.Auction.Type != "first_price" {
		t.Errorf("Auction.Type = %q, want %q", cfg.Auction.Type, "first_price")
	}
	if cfg.Auction.TimeoutMS != 100 {
		t.Errorf("Auction.TimeoutMS = %d, want 100", cfg.Auction.TimeoutMS)
	}
	if len(cfg.DSPs) != 1 {
		t.Fatalf("len(DSPs) = %d, want 1", len(cfg.DSPs))
	}
	if cfg.DSPs[0].Name != "test-dsp" {
		t.Errorf("DSPs[0].Name = %q, want %q", cfg.DSPs[0].Name, "test-dsp")
	}
	if !cfg.DSPs[0].Enabled {
		t.Error("DSPs[0].Enabled = false, want true")
	}
}

func TestLoad_Defaults(t *testing.T) {
	content := `
dsps:
  - name: "minimal-dsp"
    endpoint: "http://localhost:9000/bid"
`
	path := createTempConfig(t, content)
	defer os.Remove(path)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want default 8080", cfg.Server.Port)
	}
	if cfg.Simulation.RequestsPerSecond != 10 {
		t.Errorf("Simulation.RequestsPerSecond = %d, want default 10", cfg.Simulation.RequestsPerSecond)
	}
	if cfg.Auction.TimeoutMS != 100 {
		t.Errorf("Auction.TimeoutMS = %d, want default 100", cfg.Auction.TimeoutMS)
	}
	if cfg.Auction.Type != "first_price" {
		t.Errorf("Auction.Type = %q, want default %q", cfg.Auction.Type, "first_price")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	if err == nil {
		t.Error("Load() expected error for nonexistent file")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	content := `
server:
  port: "not a number"
`
	path := createTempConfig(t, content)
	defer os.Remove(path)

	_, err := Load(path)
	if err == nil {
		t.Error("Load() expected error for invalid YAML")
	}
}

func TestLoad_MultipleDSPs(t *testing.T) {
	content := `
dsps:
  - name: "dsp-1"
    endpoint: "http://localhost:9000/bid"
    enabled: true
  - name: "dsp-2"
    endpoint: "http://localhost:9001/bid"
    enabled: false
  - name: "dsp-3"
    endpoint: "http://localhost:9002/bid"
    enabled: true
`
	path := createTempConfig(t, content)
	defer os.Remove(path)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(cfg.DSPs) != 3 {
		t.Fatalf("len(DSPs) = %d, want 3", len(cfg.DSPs))
	}

	enabled := cfg.EnabledDSPs()
	if len(enabled) != 2 {
		t.Errorf("len(EnabledDSPs()) = %d, want 2", len(enabled))
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: Config{
				Server:     ServerConfig{Port: 8080},
				Simulation: SimulationConfig{RequestsPerSecond: 10, Scenario: "mobile_app"},
				Auction:    AuctionConfig{Type: "first_price", TimeoutMS: 100},
				DSPs:       []DSPConfig{{Name: "dsp", Endpoint: "http://localhost/bid", Enabled: true}},
			},
			wantErr: false,
		},
		{
			name: "no DSPs",
			cfg: Config{
				Server:     ServerConfig{Port: 8080},
				Simulation: SimulationConfig{RequestsPerSecond: 10},
				Auction:    AuctionConfig{Type: "first_price", TimeoutMS: 100},
				DSPs:       []DSPConfig{},
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			cfg: Config{
				Server:     ServerConfig{Port: 0},
				Simulation: SimulationConfig{RequestsPerSecond: 10},
				Auction:    AuctionConfig{Type: "first_price", TimeoutMS: 100},
				DSPs:       []DSPConfig{{Name: "dsp", Endpoint: "http://localhost/bid"}},
			},
			wantErr: true,
		},
		{
			name: "invalid RPS",
			cfg: Config{
				Server:     ServerConfig{Port: 8080},
				Simulation: SimulationConfig{RequestsPerSecond: 0},
				Auction:    AuctionConfig{Type: "first_price", TimeoutMS: 100},
				DSPs:       []DSPConfig{{Name: "dsp", Endpoint: "http://localhost/bid"}},
			},
			wantErr: true,
		},
		{
			name: "DSP missing endpoint",
			cfg: Config{
				Server:     ServerConfig{Port: 8080},
				Simulation: SimulationConfig{RequestsPerSecond: 10},
				Auction:    AuctionConfig{Type: "first_price", TimeoutMS: 100},
				DSPs:       []DSPConfig{{Name: "dsp", Endpoint: ""}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func createTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}
	return path
}
