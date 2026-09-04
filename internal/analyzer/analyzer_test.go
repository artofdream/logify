package analyzer

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
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
		if !strings.HasPrefix(v.EvidenceID, "evidence-v1-") {
			t.Errorf("missing versioned evidence ID: %q", v.EvidenceID)
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

// FR-017 / NFR-017: evidence identity is stable, versioned, and independent of
// mutable observation details or timeline display order.
func TestEvidenceIDStableAndProvenanceBound(t *testing.T) {
	base := Event{
		Signature: "cafebabe",
		Instance:  "tomcat-a",
		File:      filepath.Join("tomcat-a", "catalina.out"),
		Line:      17,
	}
	got := evidenceID(base)
	if got != evidenceID(base) || !strings.HasPrefix(got, "evidence-v1-") || len(got) != len("evidence-v1-")+64 {
		t.Fatalf("unstable or malformed evidence ID %q", got)
	}

	changedDisplayDetails := base
	changedDisplayDetails.Message = "a different display message"
	changedDisplayDetails.Occurrences = 99
	changedDisplayDetails.Timestamp = time.Date(2026, 9, 4, 1, 2, 3, 0, time.UTC)
	if evidenceID(changedDisplayDetails) != got {
		t.Fatal("display/count/time details changed evidence identity")
	}

	for name, mutate := range map[string]func(*Event){
		"signature": func(e *Event) { e.Signature = "deadbeef" },
		"instance":  func(e *Event) { e.Instance = "tomcat-b" },
		"file":      func(e *Event) { e.File = "tomcat-a/localhost.log" },
		"line":      func(e *Event) { e.Line++ },
	} {
		t.Run(name, func(t *testing.T) {
			changed := base
			mutate(&changed)
			if evidenceID(changed) == got {
				t.Fatalf("%s did not change evidence identity", name)
			}
		})
	}
}

// FR-010 / FR-017: a deduplicated group's evidence retains its actual minimum
// and maximum observation times even when discovery order is not chronological.
func TestDedupTracksFirstAndLastSeen(t *testing.T) {
	later := time.Date(2026, 9, 4, 12, 0, 0, 0, time.UTC)
	earlier := later.Add(-2 * time.Hour)
	a := Event{Instance: "a", Signature: "same", Occurrences: 1, FirstSeen: &later, LastSeen: &later}
	b := Event{Instance: "a", Signature: "same", Occurrences: 1, FirstSeen: &earlier, LastSeen: &earlier}

	got := dedup([]Event{a, b})
	if len(got) != 1 || got[0].Occurrences != 2 {
		t.Fatalf("dedup result: %#v", got)
	}
	if got[0].FirstSeen == nil || !got[0].FirstSeen.Equal(earlier) {
		t.Fatalf("first seen = %v, want %v", got[0].FirstSeen, earlier)
	}
	if got[0].LastSeen == nil || !got[0].LastSeen.Equal(later) {
		t.Fatalf("last seen = %v, want %v", got[0].LastSeen, later)
	}
}
