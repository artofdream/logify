package analyzer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNFR009ScaleSmoke(t *testing.T) {
	// NFR-009 AC1 (shape) / AC4 (repeatable): generate a small multi-instance
	// storm fixture and analyze it. The 1 GiB run is TestNFR009Scale1GiB.
	dir := t.TempDir()
	const want = int64(2 << 20)
	meta, err := GenerateScaleBundle(dir, ScaleConfig{Bytes: want, Seed: DefaultScaleSeed})
	if err != nil {
		t.Fatal(err)
	}
	if meta.TotalBytes < want {
		t.Fatalf("fixture bytes=%d want >= %d", meta.TotalBytes, want)
	}
	if len(meta.Instances) < 2 {
		t.Fatalf("instances=%v", meta.Instances)
	}
	seen := map[string]bool{}
	for _, f := range meta.Files {
		if f.Bytes < 1 || f.Lines < 1 {
			t.Fatalf("empty generated file %+v", f)
		}
		seen[filepath.Dir(filepath.FromSlash(f.Rel))] = true
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(f.Rel))); err != nil {
			t.Fatal(err)
		}
	}
	if len(seen) < 2 {
		t.Fatalf("instance dirs=%v", seen)
	}

	again, err := EnsureScaleBundle(dir, ScaleConfig{Bytes: want, Seed: DefaultScaleSeed})
	if err != nil {
		t.Fatal(err)
	}
	if again.TotalBytes != meta.TotalBytes {
		t.Fatalf("EnsureScaleBundle rewrote fixture %d -> %d", meta.TotalBytes, again.TotalBytes)
	}

	run, err := RunScaleAnalysis(dir)
	if err != nil {
		t.Fatal(err)
	}
	if run.FilesScanned < 3 {
		t.Fatalf("files=%d", run.FilesScanned)
	}
	if run.Events < 8 || run.Events > 200 {
		t.Fatalf("unique events=%d; storm profile should stay bounded", run.Events)
	}
	if !run.Result.CountsReconcile() {
		t.Fatalf("counts do not reconcile: %+v", run.Result)
	}
	t.Logf("NFR-009 smoke: input=%d events=%d files=%d dur=%s heap=%d sys=%d rss_ok=%v rss=%d",
		run.InputBytes, run.Events, run.FilesScanned, run.Duration, run.HeapAlloc, run.SysBytes, run.PeakRSSOK, run.PeakRSSBytes)
}

func TestNFR009ScaleFixtureDeterministic(t *testing.T) {
	a := t.TempDir()
	b := t.TempDir()
	cfg := ScaleConfig{Bytes: 64 << 10, Seed: 42}
	ma, err := GenerateScaleBundle(a, cfg)
	if err != nil {
		t.Fatal(err)
	}
	mb, err := GenerateScaleBundle(b, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if ma.TotalBytes != mb.TotalBytes || len(ma.Files) != len(mb.Files) {
		t.Fatalf("meta mismatch A=%+v B=%+v", ma, mb)
	}
	for i := range ma.Files {
		if ma.Files[i] != mb.Files[i] {
			t.Fatalf("file[%d] A=%+v B=%+v", i, ma.Files[i], mb.Files[i])
		}
		da, err := os.ReadFile(filepath.Join(a, filepath.FromSlash(ma.Files[i].Rel)))
		if err != nil {
			t.Fatal(err)
		}
		db, err := os.ReadFile(filepath.Join(b, filepath.FromSlash(mb.Files[i].Rel)))
		if err != nil {
			t.Fatal(err)
		}
		if string(da) != string(db) {
			t.Fatalf("content mismatch %s", ma.Files[i].Rel)
		}
	}
}
