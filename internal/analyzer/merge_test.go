package analyzer

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestNFR009OccHintsBoundedWithoutCorrelationValue(t *testing.T) {
	// NFR-009: repeats of one signature must not keep a hint per occurrence.
	now := time.Date(2026, 9, 3, 10, 0, 0, 0, time.UTC)
	var in []Event
	for i := 0; i < 5000; i++ {
		in = append(in, Event{
			HasTimestamp: true,
			Timestamp:    now.Add(time.Duration(i) * time.Millisecond),
			Severity:     Error,
			SourceType:   "tomcat-java",
			Instance:     "node-a",
			File:         "node-a/catalina.out",
			Line:         i + 1,
			Message:      "heartbeat",
			Signature:    "same",
			Occurrences:  1,
		})
	}
	out := dedup(in)
	if len(out) != 1 {
		t.Fatalf("events=%d want 1", len(out))
	}
	if out[0].Occurrences != 5000 {
		t.Fatalf("occurrences=%d want 5000", out[0].Occurrences)
	}
	if len(out[0].occHints) != 0 {
		t.Fatalf("occHints=%d want 0 (no correlation value)", len(out[0].occHints))
	}
}

func TestNFR009OccHintsCapAndKeepLaterRequestIDs(t *testing.T) {
	now := time.Date(2026, 9, 3, 10, 0, 0, 0, time.UTC)
	lead := Event{
		HasTimestamp: true,
		Timestamp:    now,
		Severity:     Error,
		SourceType:   "tomcat-java",
		Instance:     "node-a",
		File:         "node-a/catalina.out",
		Line:         1,
		Message:      "failed requestId=abc-req-0001",
		Signature:    "same",
		Occurrences:  1,
	}
	in := []Event{lead}
	for i := 0; i < maxExtraOccHints+20; i++ {
		in = append(in, Event{
			HasTimestamp: true,
			Timestamp:    now.Add(time.Duration(i+1) * time.Second),
			Severity:     Error,
			SourceType:   "tomcat-java",
			Instance:     "node-a",
			File:         "node-a/catalina.out",
			Line:         i + 2,
			Message:      "failed requestId=abc-req-" + strings.Repeat("x", 8) + strconv.Itoa(i),
			Signature:    "same",
			Occurrences:  1,
		})
	}
	out := dedup(in)
	if len(out) != 1 {
		t.Fatalf("events=%d want 1", len(out))
	}
	if len(out[0].occHints) != maxExtraOccHints {
		t.Fatalf("occHints=%d want cap %d", len(out[0].occHints), maxExtraOccHints)
	}
	if !strings.Contains(out[0].occHints[0].msg, "requestId=") {
		t.Fatalf("first extra hint lost labeled id: %q", out[0].occHints[0].msg)
	}
}
