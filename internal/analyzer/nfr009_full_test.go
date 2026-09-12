package analyzer_test

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/artofdream/logify/internal/analyzer"
	"github.com/artofdream/logify/internal/report"
)

const nfr009ResultPrefix = "NFR009_RESULT="

type nfr009ChildResult struct {
	DurationNS   int64  `json:"duration_ns"`
	PeakRSSBytes uint64 `json:"peak_rss_bytes"`
	PeakRSSOK    bool   `json:"peak_rss_ok"`
	HeapAlloc    uint64 `json:"heap_alloc_bytes"`
	SysBytes     uint64 `json:"memstats_sys_bytes"`
	Events       int    `json:"events"`
	FilesScanned int    `json:"files_scanned"`
	InputBytes   int64  `json:"input_bytes"`
	ReportBytes  int64  `json:"report_bytes"`
	Instances    int    `json:"instances"`
}

func TestNFR009Scale1GiB(t *testing.T) {
	// NFR-009: full 1 GiB analysis in a fresh process so VmHWM excludes
	// fixture generation. Skipped in ordinary `go test ./...`.
	if os.Getenv("LOGIFY_NFR009_CHILD") == "analyze" {
		runNFR009Child(t)
		return
	}
	if os.Getenv("LOGIFY_NFR009_FULL") != "1" {
		t.Skip("set LOGIFY_NFR009_FULL=1 to run the 1 GiB NFR-009 bench (manual/nightly)")
	}

	dir := os.Getenv("LOGIFY_NFR009_DIR")
	if dir == "" {
		dir = analyzer.DefaultScaleDir()
	}
	t.Logf("generating/reusing fixture in %s", dir)
	meta, err := analyzer.EnsureScaleBundle(dir, analyzer.ScaleConfig{Bytes: analyzer.ScaleGiB, Seed: analyzer.DefaultScaleSeed})
	if err != nil {
		t.Fatal(err)
	}
	if meta.TotalBytes < analyzer.ScaleGiB {
		t.Fatalf("AC1 fixture bytes=%d want >= 1 GiB", meta.TotalBytes)
	}
	if len(meta.Instances) < 2 {
		t.Fatalf("AC1 instances=%v", meta.Instances)
	}

	cmd := exec.Command(os.Args[0], "-test.run", "^TestNFR009Scale1GiB$", "-test.v")
	cmd.Env = append(os.Environ(),
		"LOGIFY_NFR009_CHILD=analyze",
		"LOGIFY_NFR009_DIR="+dir,
		"LOGIFY_NFR009_FULL=1",
	)
	out, err := cmd.CombinedOutput()
	t.Logf("child output:\n%s", out)
	if err != nil {
		t.Fatalf("child analyze: %v", err)
	}
	got, ok := parseNFR009Result(string(out))
	if !ok {
		t.Fatal("child did not print NFR009_RESULT")
	}
	t.Logf("NFR-009 1GiB: input=%d events=%d files=%d instances=%d dur=%s rss=%d (ok=%v) heap=%d sys=%d report=%d",
		got.InputBytes, got.Events, got.FilesScanned, got.Instances,
		time.Duration(got.DurationNS), got.PeakRSSBytes, got.PeakRSSOK, got.HeapAlloc, got.SysBytes, got.ReportBytes)

	if got.Events < 8 {
		t.Fatalf("too few unique events: %d", got.Events)
	}
	if got.PeakRSSOK {
		if got.PeakRSSBytes >= analyzer.ScaleMemoryBudget {
			t.Errorf("AC2 MISS: peak RSS %d >= 512 MiB (storm fixture unique events=%d)", got.PeakRSSBytes, got.Events)
		} else {
			t.Logf("AC2 PASS: peak RSS %d < 512 MiB", got.PeakRSSBytes)
		}
	} else {
		t.Logf("AC2 unknown: peak RSS not available on this OS (heap_alloc=%d sys=%d)", got.HeapAlloc, got.SysBytes)
	}
	if time.Duration(got.DurationNS) > analyzer.ScaleTimeBudget {
		t.Errorf("AC3 MISS: analysis %s > 5m", time.Duration(got.DurationNS))
	} else {
		t.Logf("AC3 PASS: analysis %s <= 5m", time.Duration(got.DurationNS))
	}
}

func runNFR009Child(t *testing.T) {
	t.Helper()
	dir := os.Getenv("LOGIFY_NFR009_DIR")
	if dir == "" {
		t.Fatal("LOGIFY_NFR009_DIR required in child")
	}
	run, err := analyzer.RunScaleAnalysis(dir)
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(os.TempDir(), "logify-nfr009-report.html")
	if err := report.Write(out, run.Result, report.Options{}); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(out)
	if err != nil {
		t.Fatal(err)
	}
	instances := map[string]bool{}
	for _, e := range run.Result.Events {
		instances[e.Instance] = true
	}
	payload := nfr009ChildResult{
		DurationNS:   run.Duration.Nanoseconds(),
		PeakRSSBytes: run.PeakRSSBytes,
		PeakRSSOK:    run.PeakRSSOK,
		HeapAlloc:    run.HeapAlloc,
		SysBytes:     run.SysBytes,
		Events:       run.Events,
		FilesScanned: run.FilesScanned,
		InputBytes:   run.InputBytes,
		ReportBytes:  info.Size(),
		Instances:    len(instances),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(nfr009ResultPrefix + string(raw))
}

func parseNFR009Result(out string) (nfr009ChildResult, bool) {
	var got nfr009ChildResult
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, nfr009ResultPrefix) {
			continue
		}
		if json.Unmarshal([]byte(strings.TrimPrefix(line, nfr009ResultPrefix)), &got) != nil {
			return nfr009ChildResult{}, false
		}
		return got, true
	}
	return nfr009ChildResult{}, false
}
