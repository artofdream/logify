package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/artofdream/logify/internal/analyzer"
)

func TestWriteSelfContained(t *testing.T) {
	p := filepath.Join(t.TempDir(), "r.html")
	if e := Write(p, analyzer.Result{Events: []analyzer.Event{{Message: "</script><b>x</b>", EvidenceID: "evidence-v1-safe"}}}); e != nil {
		t.Fatal(e)
	}
	b, e := os.ReadFile(p)
	if e != nil {
		t.Fatal(e)
	}
	s := string(b)
	if !strings.Contains(s, "<!doctype html>") || strings.Contains(s, "</script><b>x</b>") {
		t.Fatal("invalid or unsafe report")
	}
	if strings.Contains(s, "https://") || strings.Contains(s, "http://") {
		t.Fatal("report has external dependency")
	}
}

// FR-017 / NFR-004 / NFR-019: the offline report exposes issue creation,
// versioned evidence identity, complete evidence, bounded editable titles, and
// DOM-safe rendering for untrusted evidence and operator-authored text.
func TestWriteIssueCreationFromEvidence(t *testing.T) {
	first := time.Date(2026, 9, 4, 8, 0, 0, 0, time.UTC)
	last := first.Add(15 * time.Minute)
	event := analyzer.Event{
		Timestamp:    first,
		HasTimestamp: true,
		Severity:     analyzer.Error,
		SourceType:   "tomcat-java",
		Instance:     `instance</script><svg onload="bad()">`,
		File:         `logs</script><img src=x onerror="bad()">.log`,
		Line:         42,
		Message:      `failure</script><script>bad()</script>`,
		Signature:    "0123456789abcdef",
		EvidenceID:   "evidence-v1-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Occurrences:  3,
		FirstSeen:    &first,
		LastSeen:     &last,
	}
	p := filepath.Join(t.TempDir(), "issues.html")
	if err := Write(p, analyzer.Result{Events: []analyzer.Event{event}}); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)

	for _, required := range []string{
		"Create issue",
		"issue-v1-",
		"evidence-v1-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"Evidence ID",
		"Signature",
		"First seen",
		"Last seen",
		"Occurrences",
		"input.maxLength=200",
		"window.addEventListener('beforeunload'",
		"Issues exist only in this open report tab",
		"role=\"status\" aria-live=\"polite\"",
	} {
		if !strings.Contains(s, required) {
			t.Errorf("report missing %q", required)
		}
	}
	for _, forbidden := range []string{
		`failure</script><script>bad()</script>`,
		`instance</script><svg onload="bad()">`,
		`.innerHTML`,
		"insertAdjacentHTML",
		"document.write",
		"localStorage",
	} {
		if strings.Contains(s, forbidden) {
			t.Errorf("unsafe or misleading report content %q", forbidden)
		}
	}
}
