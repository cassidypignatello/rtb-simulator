# RTB Simulator Architecture

## Overview
Go-based OpenRTB 2.5 bid request simulator for educational purposes.

## Core Components

### 1. Configuration (`internal/config/`)
- YAML-based configuration
- DSP endpoints with enable/disable
- Simulation parameters (RPS, scenario, timeout)
- Auction settings (type, timeout_ms)

### 2. OpenRTB Types (`pkg/openrtb/`)
- `BidRequest` - Full OpenRTB 2.5 request
- `BidResponse` - Response with helper methods (IsNoBid, AllBids, HighestBid)
- Nested types: Imp, Banner, App, Device, Geo, User, SeatBid, Bid

### 3. Generator (`internal/generator/`)
- `Scenario` interface for pluggable request strategies
- `Generator` orchestrates ID generation and scenario execution
- Options pattern for configuration (WithTimeout, WithAuctionType)

### 4. Scenarios (`internal/generator/scenarios/`)
- `MobileApp` - Mobile app inventory simulation
- Data pools for realistic randomization
- Thread-safe with sync.Mutex on RNG

## Planned Components

### 5. Dispatcher (`internal/dispatcher/`)
- Fan-out pattern for concurrent DSP requests
- Timeout handling with context
- BidResult aggregation

### 6. Auction (`internal/auction/`)
- First-price auction logic
- Winner selection from collected bids
- Bid floor validation

### 7. Stats (`internal/stats/`)
- Thread-safe metrics collection
- Per-DSP statistics
- Recent auction history

### 8. Engine (`internal/engine/`)
- Simulation loop with ticker
- Start/Stop control
- Coordinates generator → dispatcher → auction → stats

### 9. API (`internal/api/`)
- HTTP endpoints: /start, /stop, /stats, /config
- Graceful shutdown support

## Data Flow
```
Config → Generator → Dispatcher → DSPs
                         ↓
                    Auction ← Responses
                         ↓
                      Stats
```

## Design Patterns Used
- **Interface abstraction** (Scenario)
- **Options pattern** (Generator configuration)
- **Fan-out/fan-in** (Dispatcher concurrency)
- **Dependency injection** (Component wiring in main)
