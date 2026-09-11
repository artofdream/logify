package analyzer

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
	"os"
	"path/filepath"
	"strconv"
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
		// FR-016: rotation / gzip decorations reuse the same detect() rules.
		{"access.log.2026-09-03", "apache-access"},
		{"access.log.1.gz", "apache-access"},
		{"error.log.1.gz", "apache-error"},
		{"ssl_error.log.2", "apache-error"},
		{"catalina.out.1", "tomcat-java"},
		{"catalina.out.gz", "tomcat-java"},
		{"AccessControl.log.1", "tomcat-java"},
		{"application-error.log.1.gz", "tomcat-java"},
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

func TestLooksRotatedAndGzipNames(t *testing.T) {
	// FR-016 AC1/AC2: numeric/date rotation and .gz on supported kinds.
	want := map[string]bool{
		"catalina.out.1":                      true,
		"catalina.out.gz":                     true,
		"catalina.out.1.gz":                   true,
		"access.log.2026-09-03":               true,
		"access.log.20260903":                 true,
		"access.log-20260903":                 true,
		"access.log.1.gz":                     true,
		"error.log.1.gz":                      true,
		"error_log.1":                         true,
		"localhost_access_log.2026-09-03.txt": true,
		"ssl_error.log.2.gz":                  true,
		"something.log":                       true,
		"something.out":                       true,
		"readme.md":                           false,
		"notes.txt":                           false,
		"data.gz":                             false,
		"archive.tar.gz":                      false,
		"bundle.zip":                          false,
		"notes.out.backup":                    false,
	}
	for name, discover := range want {
		if got := looks(name); got != discover {
			t.Errorf("looks(%q)=%v want %v", name, got, discover)
		}
	}
}

func TestCanonicalLogNameStripsOneDecoration(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Catalina.Out.1", "catalina.out"},
		{"ERROR.LOG.1.GZ", "error.log"},
		{"access.log.2026-09-03", "access.log"},
		{"access.log.2026-09-03.txt", "access.log"},
		{"access.log-20260903", "access.log"},
		{"localhost_access_log.2026-09-03.txt", "localhost_access_log"},
		{"ssl_access.log.1", "ssl_access.log"},
		{"catalina.out", "catalina.out"},
		{"notes.gz", "notes"},
		{"archive.tar.gz", "archive.tar"},
	}
	for _, c := range cases {
		if got := canonicalLogName(c.in); got != c.want {
			t.Errorf("canonicalLogName(%q)=%q want %q", c.in, got, c.want)
		}
	}
}

func TestAnalyzeRotatedGzipFixtures(t *testing.T) {
	// FR-016 AC1–AC3: discover rotated/gzip names, stream .gz, FR-010 grouping.
	r, err := Analyze(filepath.Join("..", "..", "testdata", "rotated"), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Warnings) != 0 {
		t.Fatalf("warnings=%+v", r.Warnings)
	}
	assertCounts(t, r, 6, 6, 0, 0)
	if len(r.Events) != 9 {
		t.Fatalf("events=%d want 9: %#v", len(r.Events), eventBrief(r.Events))
	}

	var files []string
	for _, e := range r.Events {
		files = append(files, e.File)
	}
	joined := strings.Join(files, " ")
	for _, unexpected := range []string{"ignored.txt", "notes.gz", "archive.tar.gz"} {
		if strings.Contains(joined, unexpected) {
			t.Errorf("unsupported name %q was parsed", unexpected)
		}
	}

	type key struct{ inst, src, msg string }
	occ := map[key]int{}
	srcOK := 0
	for _, e := range r.Events {
		first := strings.Split(e.Message, "\n")[0]
		occ[key{e.Instance, e.SourceType, first}] += e.Occurrences
		switch {
		case strings.HasPrefix(e.File, "httpd-b/access."):
			if e.SourceType != "apache-access" {
				t.Errorf("access file %s typed %s", e.File, e.SourceType)
			} else {
				srcOK++
			}
		case strings.HasPrefix(e.File, "httpd-b/error."):
			if e.SourceType != "apache-error" {
				t.Errorf("error file %s typed %s", e.File, e.SourceType)
			} else {
				srcOK++
			}
		case strings.Contains(e.File, "catalina.out"):
			if e.SourceType != "tomcat-java" {
				t.Errorf("java file %s typed %s", e.File, e.SourceType)
			} else {
				srcOK++
			}
		}
	}
	if srcOK != len(r.Events) {
		t.Fatalf("source mapping failures among %d events", len(r.Events))
	}

	shared := "Request 12345 failed"
	if occ[key{"tomcat-a", "tomcat-java", shared}] != 2 {
		t.Fatalf("tomcat-a rotation duplicate occurrences=%d want 2: %+v", occ[key{"tomcat-a", "tomcat-java", shared}], occ)
	}
	if occ[key{"tomcat-b", "tomcat-java", shared}] != 1 {
		t.Fatalf("cross-instance events must not merge: %+v", occ)
	}
	if occ[key{"httpd-b", "apache-access", "GET /health HTTP/1.1 -> 200 (2 bytes)"}] != 2 {
		t.Fatalf("access rotation duplicate occurrences=%d: %+v", occ[key{"httpd-b", "apache-access", "GET /health HTTP/1.1 -> 200 (2 bytes)"}], occ)
	}
	if occ[key{"tomcat-a", "tomcat-java", "rotated file unique"}] != 1 {
		t.Fatal("distinct rotated event was dropped")
	}
	if occ[key{"tomcat-a", "tomcat-java", "gzip file unique"}] != 1 {
		t.Fatal("distinct gzip event was dropped")
	}
}

func TestGzipStreamParsesWithoutExtraction(t *testing.T) {
	// FR-016 AC2: gzip.Reader over the file; no gunzip sidecar written.
	dir := t.TempDir()
	body := "2026-09-03 10:00:00 INFO streamed gzip only\n"
	gzPath := filepath.Join(dir, "catalina.out.gz")
	writeGzipFile(t, gzPath, body)
	r, err := Analyze(dir, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Events) != 1 || r.Events[0].Message != "streamed gzip only" || r.Events[0].SourceType != "tomcat-java" {
		t.Fatalf("gzip events=%+v", r.Events)
	}
	assertCounts(t, r, 1, 1, 0, 0)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "catalina.out.gz" {
		t.Fatalf("gzip handling wrote extra files: %v", names(entries))
	}
}

func TestInvalidGzipIsScanErrorAndKeepsSiblings(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "ok.log"), []byte("2026-09-03 10:00:00 INFO visible\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "broken.log.gz"), []byte("not gzip"), 0644); err != nil {
		t.Fatal(err)
	}
	r, err := Analyze(dir, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Events) != 1 || r.Events[0].Message != "visible" {
		t.Fatalf("events=%+v", r.Events)
	}
	if len(r.Warnings) != 1 || r.Warnings[0].Category != CategoryScanError || r.Warnings[0].File != "broken.log.gz" {
		t.Fatalf("warnings=%+v", r.Warnings)
	}
	assertCounts(t, r, 2, 2, 0, 0)
}

func TestTruncatedGzipKeepsEarlierEvents(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write([]byte("2026-09-03 10:00:00 INFO before gzip cut\n2026-09-03 10:00:01 INFO after\n")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	raw := buf.Bytes()
	if len(raw) < 20 {
		t.Fatalf("gzip too small to truncate: %d", len(raw))
	}
	if err := os.WriteFile(filepath.Join(dir, "cut.out.gz"), raw[:len(raw)-8], 0644); err != nil {
		t.Fatal(err)
	}
	r, err := Analyze(dir, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Events) == 0 {
		t.Fatal("expected events parsed before gzip truncation")
	}
	if r.Events[0].Message != "before gzip cut" {
		t.Fatalf("first event=%q", r.Events[0].Message)
	}
	if len(r.Warnings) != 1 || r.Warnings[0].Category != CategoryScanError {
		t.Fatalf("warnings=%+v", r.Warnings)
	}
	assertCounts(t, r, 1, 1, 0, 0)
}

func writeGzipFile(t *testing.T, path, body string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw, err := gzip.NewWriterLevel(f, gzip.BestSpeed)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := zw.Write([]byte(body)); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

func eventBrief(events []Event) []string {
	out := make([]string, 0, len(events))
	for _, e := range events {
		out = append(out, e.Instance+"|"+e.SourceType+"|"+strings.Split(e.Message, "\n")[0]+"|occ="+strconv.Itoa(e.Occurrences))
	}
	return out
}

func names(entries []os.DirEntry) []string {
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, e.Name())
	}
	return out
}
