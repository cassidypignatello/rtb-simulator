# Go Performance Best Practices for RTB Simulator

## Research Date: 2025-01-28

## Summary
Comprehensive research on high-throughput HTTP request simulation patterns for OpenRTB bid request generation.

## Key Recommendations

### 1. HTTP Client: fasthttp
- 10-22x faster than net/http in benchmarks
- Zero allocations per request (vs 35+ for net/http)
- Built-in connection pooling and reuse

### 2. JSON Serialization: sonic (recommended) or json-iterator
- Sonic: 3-10x faster than stdlib, JIT-based
- json-iterator: 2-3x faster, drop-in replacement

### 3. Object Pooling: sync.Pool
- Use for BidRequest/BidResponse objects
- Per-P local caches reduce contention
- Reset objects before Put()

### 4. Random Numbers: math/rand/v2
- Per-OS-thread ChaCha8 for globals (thread-safe)
- PCG for seeded/deterministic needs

### 5. Connection Pooling
- MaxIdleConnsPerHost: 100+ (default is 2!)
- MaxIdleConns: 100+
- MaxConnsPerHost: 100+

## Implementation Status
- [ ] Integrate fasthttp client
- [ ] Add JSON library selection
- [ ] Implement sync.Pool for requests
- [ ] Configure connection pooling
- [ ] Benchmark current vs optimized
