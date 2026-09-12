package analyzer

import (
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	// ScaleMemoryBudget is NFR-009 AC2.
	ScaleMemoryBudget = 512 << 20
	// ScaleTimeBudget is NFR-009 AC3.
	ScaleTimeBudget = 5 * time.Minute
)

// ScaleRun is one Analyze pass over a generated fixture.
type ScaleRun struct {
	Duration     time.Duration `json:"duration_ns"`
	PeakRSSBytes uint64        `json:"peak_rss_bytes"`
	PeakRSSOK    bool          `json:"peak_rss_ok"`
	HeapAlloc    uint64        `json:"heap_alloc_bytes"`
	SysBytes     uint64        `json:"memstats_sys_bytes"`
	Events       int           `json:"events"`
	FilesScanned int           `json:"files_scanned"`
	InputBytes   int64         `json:"input_bytes"`
	Result       Result        `json:"-"`
}

// PeakRSSBytes reports the process high-water RSS from /proc on Linux.
func PeakRSSBytes() (uint64, bool) {
	data, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return 0, false
	}
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "VmHWM:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return 0, false
		}
		v, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return 0, false
		}
		return v * 1024, true
	}
	return 0, false
}

// RunScaleAnalysis runs Analyze on dir and records duration plus memory.
func RunScaleAnalysis(dir string) (ScaleRun, error) {
	var input int64
	if meta, ok := readScaleMeta(dir); ok {
		input = meta.TotalBytes
	}
	start := time.Now()
	res, err := Analyze(dir, Options{})
	run := ScaleRun{
		Duration:     time.Since(start),
		InputBytes:   input,
		Result:       res,
		Events:       len(res.Events),
		FilesScanned: res.FilesScanned,
	}
	if err != nil {
		return run, err
	}
	run.PeakRSSBytes, run.PeakRSSOK = PeakRSSBytes()
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	run.HeapAlloc = ms.HeapAlloc
	run.SysBytes = ms.Sys
	return run, nil
}
