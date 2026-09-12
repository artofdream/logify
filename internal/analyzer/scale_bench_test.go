package analyzer

import (
	"os"
	"testing"
)

func BenchmarkNFR009ScaleSmoke(b *testing.B) {
	// NFR-009 AC4: small repeatable bench used in CI. Not the 1 GiB AC2/AC3 run.
	dir := b.TempDir()
	meta, err := GenerateScaleBundle(dir, ScaleConfig{Bytes: ScaleSmokeBytes, Seed: DefaultScaleSeed})
	if err != nil {
		b.Fatal(err)
	}
	b.ReportMetric(float64(meta.TotalBytes), "fixture_bytes")
	b.ReportMetric(float64(len(meta.Instances)), "instances")
	b.ResetTimer()
	var events int
	for i := 0; i < b.N; i++ {
		r, err := Analyze(dir, Options{})
		if err != nil {
			b.Fatal(err)
		}
		events = len(r.Events)
	}
	b.ReportMetric(float64(events), "unique_events")
}

func BenchmarkNFR009Scale1GiB(b *testing.B) {
	// NFR-009 AC1-AC4: full fixture. Manual/nightly — not run by go test ./...
	if os.Getenv("LOGIFY_NFR009_FULL") != "1" {
		b.Skip("set LOGIFY_NFR009_FULL=1 for the 1 GiB benchmark")
	}
	dir := os.Getenv("LOGIFY_NFR009_DIR")
	if dir == "" {
		dir = DefaultScaleDir()
	}
	meta, err := EnsureScaleBundle(dir, ScaleConfig{Bytes: ScaleGiB, Seed: DefaultScaleSeed})
	if err != nil {
		b.Fatal(err)
	}
	if meta.TotalBytes < ScaleGiB {
		b.Fatalf("fixture bytes=%d want >= %d", meta.TotalBytes, ScaleGiB)
	}
	b.ReportMetric(float64(meta.TotalBytes), "fixture_bytes")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Analyze(dir, Options{}); err != nil {
			b.Fatal(err)
		}
	}
}
