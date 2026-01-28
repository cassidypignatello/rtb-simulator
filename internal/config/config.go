package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Simulation SimulationConfig `yaml:"simulation"`
	Auction    AuctionConfig    `yaml:"auction"`
	DSPs       []DSPConfig      `yaml:"dsps"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type SimulationConfig struct {
	RequestsPerSecond int    `yaml:"requests_per_second"`
	Scenario          string `yaml:"scenario"`
}

type AuctionConfig struct {
	Type      string `yaml:"type"`
	TimeoutMS int    `yaml:"timeout_ms"`
}

type DSPConfig struct {
	Name     string `yaml:"name"`
	Endpoint string `yaml:"endpoint"`
	Enabled  bool   `yaml:"enabled"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	cfg.applyDefaults()

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return cfg, nil
}

func (c *Config) applyDefaults() {
	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}
	if c.Simulation.RequestsPerSecond == 0 {
		c.Simulation.RequestsPerSecond = 10
	}
	if c.Simulation.Scenario == "" {
		c.Simulation.Scenario = "mobile_app"
	}
	if c.Auction.Type == "" {
		c.Auction.Type = "first_price"
	}
	if c.Auction.TimeoutMS == 0 {
		c.Auction.TimeoutMS = 100
	}
}

func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return errors.New("server.port must be between 1 and 65535")
	}
	if c.Simulation.RequestsPerSecond <= 0 {
		return errors.New("simulation.requests_per_second must be positive")
	}
	if len(c.DSPs) == 0 {
		return errors.New("at least one DSP must be configured")
	}
	for i, dsp := range c.DSPs {
		if dsp.Endpoint == "" {
			return fmt.Errorf("dsps[%d].endpoint is required", i)
		}
	}
	return nil
}

func (c *Config) EnabledDSPs() []DSPConfig {
	var enabled []DSPConfig
	for _, dsp := range c.DSPs {
		if dsp.Enabled {
			enabled = append(enabled, dsp)
		}
	}
	return enabled
}
