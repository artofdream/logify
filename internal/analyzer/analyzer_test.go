package analyzer

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAnalyzeEmptyDirectorySlices(t *testing.T) {
	r, err := Analyze(t.TempDir(), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if r.Events == nil || r.Warnings == nil || r.Correlations == nil {
		t.Fatalf("nil slices events=%v warnings=%v correlations=%v", r.Events, r.Warnings, r.Correlations)
	}
	if len(r.Events) != 0 || len(r.Warnings) != 0 || len(r.Correlations) != 0 {
		t.Fatalf("events=%d warnings=%d correlations=%d", len(r.Events), len(r.Warnings), len(r.Correlations))
	}
	assertCounts(t, r, 0, 0, 0, 0)
}

func TestAnalyzeFixtures(t *testing.T) {
	r, e := Analyze(filepath.Join("..", "..", "testdata", "case"), Options{})
	if e != nil {
		t.Fatal(e)
	}
	assertCounts(t, r, 3, 3, 0, 0)
	if r.Events == nil || r.Warnings == nil || r.Correlations == nil {
		t.Fatalf("nil slices events=%v warnings=%v correlations=%v", r.Events, r.Warnings, r.Correlations)
	}
	if len(r.Correlations) != 0 {
		t.Fatalf("case fixture should stay ungrouped: %+v", r.Correlations)
	}
	if len(r.Events) != 6 {
		t.Fatalf("events=%d: %#v", len(r.Events), r.Events)
	}
	var stack, repeat, retainedMalformed, apacheSeverity bool
	for _, v := range r.Events {
		if strings.Contains(v.Message, "Caused by:") {
			stack = true
		}
		if v.Occurrences == 2 {
			repeat = true
		}
		if v.Message == "malformed access record" && v.SourceType == "apache-access" && !v.HasTimestamp {
			retainedMalformed = true
		}
		if v.Message == "backend connection failed" && v.SourceType == "apache-error" && v.Severity == Error {
			apacheSeverity = true
		}
		if v.Signature == "" {
			t.Error("missing signature")
		}
	}
	if !stack {
		t.Error("stack trace not joined")
	}
	if !repeat {
		t.Error("duplicate not aggregated")
	}
	if !retainedMalformed {
		t.Error("unparsed access line was not retained")
	}
	if !apacheSeverity {
		t.Error("facility-qualified Apache severity was not parsed")
	}
	for i := 1; i < len(r.Events); i++ {
		if r.Events[i-1].HasTimestamp && r.Events[i].HasTimestamp && r.Events[i].Timestamp.Before(r.Events[i-1].Timestamp) {
			t.Error("not chronological")
		}
	}
}
func TestDetectConventionalApacheErrorLog(t *testing.T) {
	if got := detect(filepath.Join("custom-instance", "error.log")); got != "apache-error" {
		t.Fatalf("source=%q", got)
	}
}

func TestDetectAccessNames(t *testing.T) {
	// FR-003 / conservative filename detection: conventional access names only.
	cases := []struct {
		path, want string
	}{
		{"access.log", "apache-access"},
		{"access_log", "apache-access"},
		{"localhost_access_log.2026-09-03.txt", "apache-access"},
		{"ssl_access.log", "apache-access"},
		{filepath.Join("httpd-b", "access.log"), "apache-access"},
		{"AccessControl.log", "tomcat-java"},
		{"application-error.log", "tomcat-java"},
		{"catalina.out", "tomcat-java"},
	}
	for _, c := range cases {
		if got := detect(c.path); got != c.want {
			t.Errorf("detect(%q)=%q want %q", c.path, got, c.want)
		}
	}
}

func TestFilterReadmePlus0200Window(t *testing.T) {
	// FR-011: README example bounds. Timezone-less Java/Apache-error stamps are
	// UTC, so 10:00:00.123Z is after 12:00+02:00 (10:00:00.000Z) and is excluded.
	from, err := time.Parse(time.RFC3339, "2026-09-03T08:00:00+02:00")
	if err != nil {
		t.Fatal(err)
	}
	to, err := time.Parse(time.RFC3339, "2026-09-03T12:00:00+02:00")
	if err != nil {
		t.Fatal(err)
	}
	r, err := Analyze(filepath.Join("..", "..", "testdata", "case"), Options{From: &from, To: &to})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Events) != 2 {
		t.Fatalf("events=%d want 2 (access only): %#v", len(r.Events), r.Events)
	}
	for _, e := range r.Events {
		if e.SourceType != "apache-access" || !e.HasTimestamp {
			t.Errorf("unexpected event src=%s hasTS=%v msg=%q", e.SourceType, e.HasTimestamp, e.Message)
		}
	}
}

func TestOverflowKeepsEarlierEvents(t *testing.T) {
	// NFR-008 / FR-015: a token-too-long line must not drop events already parsed.
	dir := t.TempDir()
	huge := strings.Repeat("A", 4*1024*1024+16)
	body := "2026-09-03 10:00:00 INFO before\n" + huge + "\n2026-09-03 10:00:01 INFO after\n"
	if err := os.WriteFile(filepath.Join(dir, "huge.log"), []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	r, err := Analyze(dir, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Warnings) != 1 {
		t.Fatalf("warnings=%d want 1: %+v", len(r.Warnings), r.Warnings)
	}
	w := r.Warnings[0]
	if w.Category != CategoryScanOverflow || w.File != "huge.log" || w.Line != 2 {
		t.Fatalf("overflow warning=%+v", w)
	}
	if !strings.Contains(w.Message, "token too long") {
		t.Fatalf("overflow message=%q", w.Message)
	}
	if len(r.Events) != 1 || r.Events[0].Message != "before" {
		t.Fatalf("expected to keep the pre-overflow event: %+v", r.Events)
	}
	assertCounts(t, r, 1, 1, 0, 0)
	if !r.CountsReconcile() {
		t.Fatalf("counts do not reconcile: scanned=%d processed=%d failed=%d", r.FilesScanned, r.FilesProcessed, r.FilesFailed)
	}
}

func TestOpenErrorKeepsOtherFiles(t *testing.T) {
	// FR-015 AC1/AC2/AC4/AC5: unreadable supported file warns and fails; siblings stay.
	dir := t.TempDir()
	ok := filepath.Join(dir, "ok.log")
	locked := filepath.Join(dir, "locked.log")
	if err := os.WriteFile(ok, []byte("2026-09-03 10:00:00 INFO visible\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(locked, []byte("2026-09-03 10:00:00 INFO secret\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(locked, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(locked, 0644) })
	if f, err := os.Open(locked); err == nil {
		f.Close()
		t.Skip("process can open a 000-mode file; cannot probe open-error")
	}
	r, err := Analyze(dir, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Events) != 1 || r.Events[0].Message != "visible" {
		t.Fatalf("expected only the readable file's event: %+v", r.Events)
	}
	if len(r.Warnings) != 1 || r.Warnings[0].Category != CategoryOpenError || r.Warnings[0].File != "locked.log" {
		t.Fatalf("open warning=%+v", r.Warnings)
	}
	if r.Warnings[0].Line != 0 || r.Warnings[0].Message == "" {
		t.Fatalf("open warning should have a message and no line: %+v", r.Warnings[0])
	}
	assertCounts(t, r, 2, 1, 0, 1)
}

func TestWalkErrorSkippedAndOtherFilesKept(t *testing.T) {
	// FR-015 AC5: an unreadable directory is skipped; known siblings still process.
	dir := t.TempDir()
	blocked := filepath.Join(dir, "blocked")
	if err := os.Mkdir(blocked, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(blocked, "hidden.log"), []byte("2026-09-03 10:00:00 INFO hidden\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ok.log"), []byte("2026-09-03 10:00:00 INFO visible\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(blocked, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(blocked, 0755) })
	if entries, err := os.ReadDir(blocked); err == nil {
		_ = entries
		t.Skip("process can read a 000-mode directory; cannot probe walk-error")
	}
	r, err := Analyze(dir, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Events) != 1 || r.Events[0].Message != "visible" {
		t.Fatalf("expected only the reachable file's event: %+v", r.Events)
	}
	if len(r.Warnings) != 1 || r.Warnings[0].Category != CategoryWalkError {
		t.Fatalf("walk warning=%+v", r.Warnings)
	}
	if r.Warnings[0].File != "blocked" {
		t.Fatalf("walk warning file=%q", r.Warnings[0].File)
	}
	assertCounts(t, r, 1, 1, 1, 0)
}

func TestWarningFromScanCategories(t *testing.T) {
	if warningFromScan("x.log", 3, nil) != nil {
		t.Fatal("nil scan error should not warn")
	}
	ov := warningFromScan("x.log", 3, bufio.ErrTooLong)
	if ov == nil || ov.Category != CategoryScanOverflow || ov.Line != 4 || ov.File != "x.log" {
		t.Fatalf("overflow classify=%+v", ov)
	}
	other := warningFromScan("x.log", 3, errors.New("read interrupted"))
	if other == nil || other.Category != CategoryScanError || other.Line != 3 {
		t.Fatalf("scan-error classify=%+v", other)
	}
}

func TestWarningStringAndSummary(t *testing.T) {
	w := Warning{File: "a.log", Category: CategoryScanOverflow, Line: 2, Message: "token too long"}
	if got, want := w.String(), "a.log:2 [scan-overflow] token too long"; got != want {
		t.Fatalf("String()=%q want %q", got, want)
	}
	ranged := Warning{File: "a.log", Category: CategoryScanError, Line: 4, LineEnd: 7, Message: "read"}
	if got, want := ranged.Location(), "a.log:4-7"; got != want {
		t.Fatalf("Location()=%q want %q", got, want)
	}
	r := Result{
		FilesScanned:   3,
		FilesProcessed: 2,
		FilesSkipped:   1,
		FilesFailed:    1,
		Events:         []Event{{Message: "x"}, {Message: "y"}},
		Warnings:       []Warning{w},
	}
	if !r.CountsReconcile() {
		t.Fatal("expected scanned = processed + failed")
	}
	want := "2 events from 3 files; processed=2 skipped=1 failed=1; 1 warnings"
	if r.SummaryLine() != want {
		t.Fatalf("SummaryLine()=%q want %q", r.SummaryLine(), want)
	}
}

func assertCounts(t *testing.T, r Result, scanned, processed, skipped, failed int) {
	t.Helper()
	if r.FilesScanned != scanned || r.FilesProcessed != processed || r.FilesSkipped != skipped || r.FilesFailed != failed {
		t.Fatalf("counts scanned=%d processed=%d skipped=%d failed=%d want %d/%d/%d/%d",
			r.FilesScanned, r.FilesProcessed, r.FilesSkipped, r.FilesFailed, scanned, processed, skipped, failed)
	}
	if !r.CountsReconcile() {
		t.Fatalf("filesScanned=%d != processed+failed=%d", r.FilesScanned, r.FilesProcessed+r.FilesFailed)
	}
}
func TestAccessSeverity(t *testing.T) {
	e, ok := access(`127.0.0.1 - - [03/Sep/2026:10:00:03 +0200] "GET /fail HTTP/1.1" 503 12`)
	if !ok || e.Severity != Error || e.StatusCode != 503 || e.ClientAddr != "127.0.0.1" {
		t.Fatalf("%+v %v", e, ok)
	}
}

func TestApacheErrorRetainsClientAddr(t *testing.T) {
	e, ok := apacheError(`[Thu Sep 03 10:00:01.050000 2026] [proxy:error] [pid 7] [client 198.51.100.10:51234] requestId=abc-req-001 backend connection failed`)
	if !ok || e.ClientAddr != "198.51.100.10" || !strings.Contains(e.Message, "requestId=abc-req-001") {
		t.Fatalf("%+v %v", e, ok)
	}
	if strings.Contains(e.Message, "[client") {
		t.Fatalf("client token leaked into message: %q", e.Message)
	}
}
func TestSignatureNormalizesIDs(t *testing.T) {
	a := Event{SourceType: "x", Severity: Error, Message: "failure request 12345"}
	b := Event{SourceType: "x", Severity: Error, Message: "failure request 67890"}
	if signature(a) != signature(b) {
		t.Error("volatile IDs should normalize")
	}
}
