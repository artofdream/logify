package report

import (
	"strings"
	"testing"
	"time"

	"github.com/artofdream/logify/internal/analyzer"
)

// FR-017 / NFR-017: evidence identity is stable, versioned, and independent of
// display order, message text, counts, and observation times.
func TestEvidenceIDStableAndProvenanceBound(t *testing.T) {
	base := analyzer.Event{
		Signature: "cafebabe",
		Instance:  "tomcat-a",
		File:      "tomcat-a/catalina.out",
		Line:      17,
	}
	got := EvidenceID(base)
	if got != EvidenceID(base) || !strings.HasPrefix(got, EvidenceIDPrefix) {
		t.Fatalf("unstable or malformed evidence ID %q", got)
	}
	if want := IssueIDPrefix + strings.TrimPrefix(got, EvidenceIDPrefix); IssueID(got) != want {
		t.Fatalf("issue id %q", IssueID(got))
	}

	changed := base
	changed.Message = "</script><b>x</b>"
	changed.Occurrences = 99
	changed.Timestamp = time.Date(2026, 9, 4, 1, 2, 3, 0, time.UTC)
	changed.LastSeen = changed.Timestamp
	if EvidenceID(changed) != got {
		t.Fatal("display/count/time details changed evidence identity")
	}

	for name, mutate := range map[string]func(*analyzer.Event){
		"signature": func(e *analyzer.Event) { e.Signature = "deadbeef" },
		"instance":  func(e *analyzer.Event) { e.Instance = "tomcat-b" },
		"file":      func(e *analyzer.Event) { e.File = "tomcat-a/localhost.log" },
		"line":      func(e *analyzer.Event) { e.Line++ },
	} {
		t.Run(name, func(t *testing.T) {
			next := base
			mutate(&next)
			if EvidenceID(next) == got {
				t.Fatalf("%s did not change evidence identity", name)
			}
		})
	}
}
