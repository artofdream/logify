package analyzer

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestAnalyzeFixtures(t *testing.T) {
	r, e := Analyze(filepath.Join("..", "..", "testdata", "case"), Options{})
	if e != nil {
		t.Fatal(e)
	}
	if r.FilesScanned != 3 {
		t.Fatalf("files=%d", r.FilesScanned)
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
