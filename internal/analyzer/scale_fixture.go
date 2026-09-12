package analyzer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

// Scale fixture sizes. The 1 GiB corpus is generated on demand and must not
// be committed (NFR-009 AC1).
const (
	ScaleGiB         int64 = 1 << 30
	ScaleSmokeBytes  int64 = 8 << 20
	DefaultScaleSeed int64 = 9
)

const (
	ScaleProfileStorm = "storm"
	scaleMetaName     = ".logify-nfr009-meta.json"
)

// ScaleConfig describes a generated multi-instance support-bundle fixture.
type ScaleConfig struct {
	Bytes   int64
	Seed    int64
	Profile string
}

func (c ScaleConfig) normalized() ScaleConfig {
	if c.Bytes <= 0 {
		c.Bytes = ScaleSmokeBytes
	}
	if c.Seed == 0 {
		c.Seed = DefaultScaleSeed
	}
	if c.Profile == "" {
		c.Profile = ScaleProfileStorm
	}
	return c
}

// ScaleFile is one generated log in the fixture.
type ScaleFile struct {
	Rel   string `json:"rel"`
	Bytes int64  `json:"bytes"`
	Lines int64  `json:"lines"`
}

// ScaleMeta is written beside the generated logs so a later run can reuse them.
type ScaleMeta struct {
	Profile    string      `json:"profile"`
	Bytes      int64       `json:"bytes"`
	Seed       int64       `json:"seed"`
	TotalBytes int64       `json:"totalBytes"`
	Instances  []string    `json:"instances"`
	Files      []ScaleFile `json:"files"`
}

type scaleTarget struct {
	rel    string
	weight int
	kind   string
}

var stormTargets = []scaleTarget{
	{rel: "node-a/catalina.out", weight: 40, kind: "java"},
	{rel: "node-b/catalina.out", weight: 25, kind: "java"},
	{rel: "httpd-a/access.log", weight: 20, kind: "access"},
	{rel: "httpd-b/access.log", weight: 10, kind: "access"},
	{rel: "httpd-b/error.log", weight: 5, kind: "error"},
}

var stormAccessPaths = []string{
	"/health",
	"/api/v1/items",
	"/api/v1/checkout",
	"/inventory",
	"/static/app.js",
}

var stormStatuses = []int{200, 200, 200, 404, 503}

var stormJavaMsgs = []string{
	"Checkout failed",
	"Inventory timeout",
	"downstream connection reset",
	"cache miss overflow",
	"payment gateway timeout",
	"session store unavailable",
	"jdbc pool exhausted",
	"heartbeat",
}

var stormErrorMsgs = []string{
	"backend connection failed",
	"proxy: error reading status",
	"AH01144: No protocol handler",
	"client denied by server configuration",
}

// GenerateScaleBundle writes a deterministic multi-instance log tree under dir.
// It streams records to disk and does not keep the corpus in memory.
func GenerateScaleBundle(dir string, cfg ScaleConfig) (ScaleMeta, error) {
	cfg = cfg.normalized()
	if cfg.Profile != ScaleProfileStorm {
		return ScaleMeta{}, fmt.Errorf("unsupported scale profile %q", cfg.Profile)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return ScaleMeta{}, err
	}
	if err := removeScaleOutputs(dir); err != nil {
		return ScaleMeta{}, err
	}

	totalWeight := 0
	for _, t := range stormTargets {
		totalWeight += t.weight
	}

	rng := rand.New(rand.NewSource(cfg.Seed))
	meta := ScaleMeta{
		Profile:   cfg.Profile,
		Bytes:     cfg.Bytes,
		Seed:      cfg.Seed,
		Instances: []string{"node-a", "node-b", "httpd-a", "httpd-b"},
	}

	var assigned int64
	for i, t := range stormTargets {
		target := cfg.Bytes * int64(t.weight) / int64(totalWeight)
		if i == len(stormTargets)-1 {
			target = cfg.Bytes - assigned
		}
		if target < 1 {
			target = 1
		}
		path := filepath.Join(dir, filepath.FromSlash(t.rel))
		n, lines, err := writeScaleFile(path, target, t.kind, rng)
		if err != nil {
			return ScaleMeta{}, err
		}
		assigned += n
		meta.Files = append(meta.Files, ScaleFile{Rel: t.rel, Bytes: n, Lines: lines})
	}
	meta.TotalBytes = assigned
	if err := writeScaleMeta(dir, meta); err != nil {
		return ScaleMeta{}, err
	}
	return meta, nil
}

// EnsureScaleBundle reuses a matching generated tree or writes a new one.
func EnsureScaleBundle(dir string, cfg ScaleConfig) (ScaleMeta, error) {
	cfg = cfg.normalized()
	if meta, ok := readScaleMeta(dir); ok && meta.Profile == cfg.Profile && meta.Bytes == cfg.Bytes && meta.Seed == cfg.Seed && meta.TotalBytes >= cfg.Bytes {
		if scaleFilesExist(dir, meta) {
			return meta, nil
		}
	}
	return GenerateScaleBundle(dir, cfg)
}

func writeScaleFile(path string, target int64, kind string, rng *rand.Rand) (int64, int64, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return 0, 0, err
	}
	f, err := os.Create(path)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()
	w := bufio.NewWriterSize(f, 64*1024)
	var written, lines, seq int64
	start := time.Date(2026, 9, 3, 10, 0, 0, 0, time.UTC)
	for written < target {
		ts := start.Add(time.Duration(seq) * time.Millisecond)
		n, err := writeScaleRecord(w, kind, ts, seq, rng)
		if err != nil {
			return written, lines, err
		}
		written += int64(n)
		lines++
		seq++
	}
	if err := w.Flush(); err != nil {
		return written, lines, err
	}
	return written, lines, nil
}

func writeScaleRecord(w *bufio.Writer, kind string, ts time.Time, seq int64, rng *rand.Rand) (int, error) {
	switch kind {
	case "java":
		return writeJavaRecord(w, ts, seq, rng)
	case "access":
		return writeAccessRecord(w, ts, seq, rng)
	case "error":
		return writeErrorRecord(w, ts, seq, rng)
	default:
		return 0, fmt.Errorf("unknown scale record kind %q", kind)
	}
}

func writeJavaRecord(w *bufio.Writer, ts time.Time, seq int64, rng *rand.Rand) (int, error) {
	msg := stormJavaMsgs[int(seq)%len(stormJavaMsgs)]
	req := fmt.Sprintf("abc-req-%02d", int(seq)%16)
	ip := fmt.Sprintf("10.2.0.%d", 1+int(seq)%16)
	var n int
	var err error
	if msg == "heartbeat" {
		n, err = fmt.Fprintf(w, "%s INFO heartbeat\n", ts.Format("2006-01-02 15:04:05,000"))
		return n, err
	}
	n, err = fmt.Fprintf(w, "%s ERROR %s requestId=%s client %s\n", ts.Format("2006-01-02 15:04:05,000"), msg, req, ip)
	if err != nil {
		return n, err
	}
	if rng.Intn(40) == 0 {
		extra, e := fmt.Fprintf(w, "java.io.IOException: %s\n\tat com.example.App.run(App.java:42)\nCaused by: java.net.SocketException: broken pipe\n", msg)
		n += extra
		return n, e
	}
	return n, nil
}

func writeAccessRecord(w *bufio.Writer, ts time.Time, seq int64, rng *rand.Rand) (int, error) {
	ip := fmt.Sprintf("10.2.0.%d", 1+int(seq)%16)
	path := stormAccessPaths[int(seq)%len(stormAccessPaths)]
	status := stormStatuses[int(seq)%len(stormStatuses)]
	bytes := 32 + rng.Intn(256)
	if status >= 400 {
		path = path + "?requestId=" + fmt.Sprintf("abc-req-%02d", int(seq)%16)
	}
	return fmt.Fprintf(w, "%s - - [%s] \"GET %s HTTP/1.1\" %d %d\n",
		ip, ts.Format("02/Jan/2006:15:04:05 -0700"), path, status, bytes)
}

func writeErrorRecord(w *bufio.Writer, ts time.Time, seq int64, rng *rand.Rand) (int, error) {
	msg := stormErrorMsgs[int(seq)%len(stormErrorMsgs)]
	pid := 100 + int(seq)%50
	_ = rng
	return fmt.Fprintf(w, "[%s] [proxy:error] [pid %d] %s\n",
		ts.Format("Mon Jan 02 15:04:05.000000 2006"), pid, msg)
}

func writeScaleMeta(dir string, meta ScaleMeta) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(filepath.Join(dir, scaleMetaName), data, 0o644)
}

func readScaleMeta(dir string) (ScaleMeta, bool) {
	data, err := os.ReadFile(filepath.Join(dir, scaleMetaName))
	if err != nil {
		return ScaleMeta{}, false
	}
	var meta ScaleMeta
	if json.Unmarshal(data, &meta) != nil {
		return ScaleMeta{}, false
	}
	return meta, true
}

func scaleFilesExist(dir string, meta ScaleMeta) bool {
	for _, f := range meta.Files {
		info, err := os.Stat(filepath.Join(dir, filepath.FromSlash(f.Rel)))
		if err != nil || info.Size() < 1 {
			return false
		}
	}
	return true
}

func removeScaleOutputs(dir string) error {
	for _, t := range stormTargets {
		_ = os.Remove(filepath.Join(dir, filepath.FromSlash(t.rel)))
	}
	_ = os.Remove(filepath.Join(dir, scaleMetaName))
	return nil
}

// DefaultScaleDir is the on-disk location for the manual 1 GiB fixture.
func DefaultScaleDir() string {
	return filepath.Join(os.TempDir(), "logify-nfr009")
}
