package scenarios

import (
	"testing"
)

// BenchmarkMobileApp_Generate benchmarks single-threaded request generation.
func BenchmarkMobileApp_Generate(b *testing.B) {
	scenario := NewMobileApp()
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = scenario.Generate("req-00000001")
	}
}

// BenchmarkMobileApp_Generate_Parallel benchmarks concurrent request generation.
// This is the critical benchmark - it should scale with CPU cores after optimization.
func BenchmarkMobileApp_Generate_Parallel(b *testing.B) {
	scenario := NewMobileApp()
	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = scenario.Generate("req-00000001")
		}
	})
}

// BenchmarkRandomIP benchmarks IP address generation.
func BenchmarkRandomIP(b *testing.B) {
	scenario := NewMobileApp()
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = scenario.randomIP()
	}
}

// BenchmarkRandomUserID benchmarks user ID generation.
func BenchmarkRandomUserID(b *testing.B) {
	scenario := NewMobileApp()
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = scenario.randomUserID()
	}
}

// BenchmarkRandomApp benchmarks app info generation.
func BenchmarkRandomApp(b *testing.B) {
	scenario := NewMobileApp()
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = scenario.randomApp()
	}
}

// BenchmarkRandomDevice benchmarks device info generation.
func BenchmarkRandomDevice(b *testing.B) {
	scenario := NewMobileApp()
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = scenario.randomDevice()
	}
}
