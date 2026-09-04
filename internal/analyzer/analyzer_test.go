package analyzer

import (
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
	if r.Events == nil || r.Warnings == nil {
		t.Fatalf("nil slices events=%v warnings=%v", r.Events, r.Warnings)
	}
	if len(r.Events) != 0 || len(r.Warnings) != 0 {
		t.Fatalf("events=%d warnings=%d", len(r.Events), len(r.Warnings))
	}
}

func TestAnalyzeFixtures(t *testing.T) {
	r, e := Analyze(filepath.Join("..", "..", "testdata", "case"), Options{})
	if e != nil {
		t.Fatal(e)
	}
	if r.FilesScanned != 3 {
		t.Fatalf("files=%d", r.FilesScanned)
	}
	if r.Events == nil || r.Warnings == nil {
		t.Fatalf("nil slices events=%v warnings=%v", r.Events, r.Warnings)
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
	if len(r.Warnings) == 0 {
		t.Fatal("expected scanner overflow warning")
	}
	if len(r.Events) != 1 || r.Events[0].Message != "before" {
		t.Fatalf("expected to keep the pre-overflow event: %+v", r.Events)
	}
}
func TestAccessSeverity(t *testing.T) {
	e, ok := access(`127.0.0.1 - - [03/Sep/2026:10:00:03 +0200] "GET /fail HTTP/1.1" 503 12`)
	if !ok || e.Severity != Error || e.StatusCode != 503 {
		t.Fatalf("%+v %v", e, ok)
	}
}
func TestSignatureNormalizesIDs(t *testing.T) {
	a := Event{SourceType: "x", Severity: Error, Message: "failure request 12345"}
	b := Event{SourceType: "x", Severity: Error, Message: "failure request 67890"}
	if signature(a) != signature(b) {
		t.Error("volatile IDs should normalize")
	}
}
