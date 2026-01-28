package generator

import (
	"testing"

	"github.com/cass/rtb-simulator/internal/generator/scenarios"
)

// BenchmarkGenerator_Generate benchmarks the full generation pipeline.
func BenchmarkGenerator_Generate(b *testing.B) {
	scenario := scenarios.NewMobileApp()
	gen := New(scenario)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = gen.Generate()
	}
}

// BenchmarkGenerator_Generate_Parallel benchmarks concurrent generation.
func BenchmarkGenerator_Generate_Parallel(b *testing.B) {
	scenario := scenarios.NewMobileApp()
	gen := New(scenario)
	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = gen.Generate()
		}
	})
}

// BenchmarkNextID benchmarks request ID generation.
func BenchmarkNextID(b *testing.B) {
	scenario := scenarios.NewMobileApp()
	gen := New(scenario)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = gen.nextID()
	}
}
