package report

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/artofdream/logify/internal/analyzer"
	"github.com/artofdream/logify/internal/redact"
)

func TestWriteSelfContained(t *testing.T) {
	// NFR-003 / NFR-004 / NFR-012 / NFR-014: offline self-contained HTML, untrusted
	// log text escaped, CSS max-width containers, compiler-backed report probe.
	p := filepath.Join(t.TempDir(), "r.html")
	if e := Write(p, analyzer.Result{Events: []analyzer.Event{{Message: "</script><b>x</b>"}}}); e != nil {
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
	if !strings.Contains(s, FollowUpSchema) || !strings.Contains(s, "Issue queue") {
		t.Fatal("report is missing the issue work queue")
	}
	if !strings.Contains(s, "Clear due date") || !strings.Contains(s, "Overdue only") {
		t.Fatal("report is missing FR-021 follow-up detail editors")
	}
	if !strings.Contains(s, "Sensitive copy.") || !strings.Contains(s, "embeds parsed log text") {
		t.Fatal("report is missing the NFR-006 sensitivity banner")
	}
	if !strings.Contains(s, "issue-workflow-hint") || !strings.Contains(s, "issue-summary") {
		t.Fatal("report is missing NFR-021 issue-workflow usability chrome")
	}
	if !strings.Contains(s, `aria-live="polite"`) || !strings.Contains(s, `aria-atomic="true"`) {
		t.Fatal("report is missing live status semantics")
	}
	if !strings.Contains(s, "Link to existing issue") || !strings.Contains(s, "Recurring evidence review") {
		t.Fatal("report is missing FR-024 merge/review chrome")
	}
	if !strings.Contains(s, "Correlation groups") || !strings.Contains(s, "Exact identifier") {
		t.Fatal("report is missing FR-012 correlation chrome")
	}
	if !strings.Contains(s, "documented groups with a named rule") {
		t.Fatal("legend still claims correlations are unavailable")
	}
	if !strings.Contains(s, "unparsed records") || !strings.Contains(s, "Unparsed record (low parse confidence)") {
		t.Fatal("report is missing NFR-028 unparsed-record chrome")
	}
	if !strings.Contains(s, "max-width: 1400px") {
		t.Fatal("report CSS is missing NFR-012 max-width container")
	}
}

func TestWriteEmptySlicesAreJSONArrays(t *testing.T) {
	// FR-013 / FR-014: nil Events/Warnings must not marshal as JSON null.
	p := filepath.Join(t.TempDir(), "empty.html")
	if err := Write(p, analyzer.Result{}); err != nil {
		t.Fatal(err)
	}
	assertEmbeddedArrays(t, p, 0, 0, 0)
}

func TestWriteFixtureReportArraysAndEvidence(t *testing.T) {
	r, err := analyzer.Analyze(filepath.Join("..", "..", "testdata", "case"), analyzer.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Warnings) != 0 {
		t.Fatalf("fixture warnings=%d", len(r.Warnings))
	}
	p := filepath.Join(t.TempDir(), "fixture.html")
	if err := Write(p, r); err != nil {
		t.Fatal(err)
	}
	raw := assertEmbeddedArrays(t, p, 6, 0, 0)
	var payload struct {
		UnparsedRecords            int `json:"unparsedRecords"`
		HighConfidenceCorrelations int `json:"highConfidenceCorrelations"`
		LowConfidenceCorrelations  int `json:"lowConfidenceCorrelations"`
		Events                     []struct {
			EvidenceID      string `json:"evidenceId"`
			Signature       string `json:"signature"`
			Instance        string `json:"instance"`
			File            string `json:"file"`
			Line            int    `json:"line"`
			Occurrences     int    `json:"occurrences"`
			ParseConfidence string `json:"parseConfidence"`
		} `json:"events"`
		Warnings []analyzer.Warning `json:"warnings"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Events == nil || payload.Warnings == nil {
		t.Fatal("unmarshaled nil slice")
	}
	if payload.UnparsedRecords != 1 || payload.HighConfidenceCorrelations != 0 || payload.LowConfidenceCorrelations != 0 {
		t.Fatalf("fixture observability counts=%+v", payload)
	}
	var sawUnparsed bool
	for i, ev := range payload.Events {
		if ev.ParseConfidence == "low" {
			sawUnparsed = true
		} else if ev.ParseConfidence != "high" {
			t.Fatalf("event %d parseConfidence=%q", i, ev.ParseConfidence)
		}
		if !strings.HasPrefix(ev.EvidenceID, EvidenceIDPrefix) {
			t.Fatalf("event %d missing evidenceId: %q", i, ev.EvidenceID)
		}
		want := EvidenceID(analyzer.Event{
			Signature: ev.Signature,
			Instance:  ev.Instance,
			File:      ev.File,
			Line:      ev.Line,
		})
		if ev.EvidenceID != want {
			t.Fatalf("event %d evidenceId=%q want %q", i, ev.EvidenceID, want)
		}
	}
	if !sawUnparsed {
		t.Fatal("fixture report missing a low-confidence unparsed event")
	}
}

func TestWriteRedactsAfterEvidenceID(t *testing.T) {
	// NFR-006: optional redaction is applied to operator-visible fields before
	// HTML embedding; evidence IDs stay bound to the original provenance.
	rule, err := redact.ParseRule("literal:secret")
	if err != nil {
		t.Fatal(err)
	}
	src := analyzer.Event{
		Message:   "token secret",
		File:      "secret.log",
		Instance:  "secret-node",
		Signature: "deadbeef",
		Line:      4,
	}
	wantID := EvidenceID(src)
	p := filepath.Join(t.TempDir(), "redacted.html")
	err = Write(p, analyzer.Result{
		Root:   "/tmp/secret",
		Events: []analyzer.Event{src},
		Warnings: []analyzer.Warning{{
			File:     "other.log",
			Category: analyzer.CategoryScanOverflow,
			Message:  "secret overflow",
		}},
	}, Options{Redact: redact.New([]redact.Rule{rule})})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := embeddedJSON(mustRead(t, p))
	if err != nil {
		t.Fatal(err)
	}
	var payload struct {
		Root   string `json:"root"`
		Events []struct {
			Message    string `json:"message"`
			File       string `json:"file"`
			Instance   string `json:"instance"`
			EvidenceID string `json:"evidenceId"`
		} `json:"events"`
		Warnings  []analyzer.Warning `json:"warnings"`
		Redaction struct {
			Enabled      bool `json:"enabled"`
			RuleCount    int  `json:"ruleCount"`
			Replacements int  `json:"replacements"`
		} `json:"redaction"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Root != "/tmp/"+redact.Replacement {
		t.Fatalf("root=%q", payload.Root)
	}
	if payload.Warnings[0].Message != redact.Replacement+" overflow" {
		t.Fatalf("warning=%+v", payload.Warnings[0])
	}
	ev := payload.Events[0]
	if ev.Message != "token "+redact.Replacement || ev.File != redact.Replacement+".log" || ev.Instance != redact.Replacement+"-node" {
		t.Fatalf("event %#v", ev)
	}
	if ev.EvidenceID != wantID {
		t.Fatalf("evidenceId=%q want %q", ev.EvidenceID, wantID)
	}
	if !payload.Redaction.Enabled || payload.Redaction.RuleCount != 1 || payload.Redaction.Replacements != 5 {
		t.Fatalf("redaction %#v", payload.Redaction)
	}
	html := string(mustRead(t, p))
	if strings.Contains(html, "token secret") {
		t.Fatal("unredacted secret leaked into HTML")
	}
}

func TestWriteDefaultDoesNotRedact(t *testing.T) {
	p := filepath.Join(t.TempDir(), "plain.html")
	if err := Write(p, analyzer.Result{Events: []analyzer.Event{{Message: "ops@example.com"}}}); err != nil {
		t.Fatal(err)
	}
	raw, err := embeddedJSON(mustRead(t, p))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(raw, []byte("ops@example.com")) {
		t.Fatal("default Write applied redaction")
	}
	var payload struct {
		Redaction struct {
			Enabled bool `json:"enabled"`
		} `json:"redaction"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Redaction.Enabled {
		t.Fatal("redaction.enabled true by default")
	}
}

func TestWriteEscapesOperatorAndLogText(t *testing.T) {
	p := filepath.Join(t.TempDir(), "xss.html")
	err := Write(p, analyzer.Result{
		Root: `<img src=x>`,
		Events: []analyzer.Event{{
			Message:  `<script>alert(1)</script>`,
			File:     `a".log`,
			Instance: `i<>`,
			Line:     3,
		}},
		GeneratedAt: time.Date(2026, 9, 4, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if strings.Contains(s, "<script>alert(1)</script>") {
		t.Fatal("log script leaked into HTML")
	}
	if !strings.Contains(s, `\u003cscript\u003e`) && !strings.Contains(s, `\u003c`) {
		t.Fatal("expected JSON-escaped log text")
	}
}

func TestPageScriptsAreSyntacticallyValid(t *testing.T) {
	node := requireNode(t)
	for _, name := range []string{"followup.js", "page.js", "followup_node_test.js", "nfr021_a11y_test.js", "nfr021_filter_probe.js"} {
		cmd := exec.Command(node, "--check", name)
		cmd.Dir = "."
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%s: %v\n%s", name, err, out)
		}
	}
}

func TestFollowUpStoreCreateExportImport(t *testing.T) {
	// FR-017 / FR-018 / FR-019 / FR-020 / FR-021 / FR-022 / FR-023 /
	// NFR-018 / NFR-019 / NFR-020: executed page-script
	// store, not only embedded JSON shape. The prior happy-path crash was missed
	// because CI never ran the report JavaScript.
	node := requireNode(t)
	cmd := exec.Command(node, "followup_node_test.js")
	cmd.Dir = "."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("follow-up JS tests: %v\n%s", err, out)
	}
}

func TestWriteStructuredWarningsAndCounts(t *testing.T) {
	// FR-015: report embeds structured warnings and reconciled input counts.
	p := filepath.Join(t.TempDir(), "warn.html")
	err := Write(p, analyzer.Result{
		FilesScanned:   2,
		FilesProcessed: 1,
		FilesSkipped:   1,
		FilesFailed:    1,
		Events:         []analyzer.Event{{Message: "kept"}},
		Warnings: []analyzer.Warning{{
			File:     `locked".log`,
			Category: analyzer.CategoryOpenError,
			Message:  `<script>alert(1)</script>`,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	raw := assertEmbeddedArrays(t, p, 1, 1, 0)
	var payload struct {
		FilesScanned   int                `json:"filesScanned"`
		FilesProcessed int                `json:"filesProcessed"`
		FilesSkipped   int                `json:"filesSkipped"`
		FilesFailed    int                `json:"filesFailed"`
		Warnings       []analyzer.Warning `json:"warnings"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.FilesScanned != 2 || payload.FilesProcessed != 1 || payload.FilesSkipped != 1 || payload.FilesFailed != 1 {
		t.Fatalf("embedded counts=%+v", payload)
	}
	if payload.Warnings[0].Category != analyzer.CategoryOpenError || payload.Warnings[0].File != `locked".log` {
		t.Fatalf("embedded warning=%+v", payload.Warnings[0])
	}
	html, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	s := string(html)
	if strings.Contains(s, "<script>alert(1)</script>") {
		t.Fatal("warning message leaked into HTML")
	}
	if !strings.Contains(s, "Scan warnings") {
		t.Fatal("report is missing the scan warnings heading")
	}
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestNFR021IssueWorkflowA11yContracts(t *testing.T) {
	// NFR-021 AC1–AC3: source-contract audit, not a browser AT or axe run.
	node := requireNode(t)
	cmd := exec.Command(node, "nfr021_a11y_test.js")
	cmd.Dir = "."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("NFR-021 a11y contracts: %v\n%s", err, out)
	}
}

func TestNFR021IssueFilterProbe(t *testing.T) {
	// NFR-021 AC4: 10,000-issue store.filter probe. Fails only on the CI guard,
	// not on the 100ms interactive target (see the printed JSON + research note).
	node := requireNode(t)
	cmd := exec.Command(node, "nfr021_filter_probe.js")
	cmd.Dir = "."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("NFR-021 filter probe: %v\n%s", err, out)
	}
	t.Logf("%s", out)
}

func requireNode(t *testing.T) string {
	t.Helper()
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal("node not on PATH; JS report/follow-up tests cannot run")
	}
	return node
}

func TestWriteCorrelateGroupsExactAndHeuristic(t *testing.T) {
	// FR-012 AC2–AC5: groups are embedded with rule/kind/confidence/evidence;
	// events remain individually listed.
	r, err := analyzer.Analyze(filepath.Join("..", "..", "testdata", "correlate"), analyzer.Options{})
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(t.TempDir(), "corr.html")
	if err := Write(p, r); err != nil {
		t.Fatal(err)
	}
	raw := assertEmbeddedArrays(t, p, 14, 0, 2)
	var payload struct {
		HighConfidenceCorrelations int `json:"highConfidenceCorrelations"`
		LowConfidenceCorrelations  int `json:"lowConfidenceCorrelations"`
		Events                     []struct {
			EvidenceID string `json:"evidenceId"`
			Message    string `json:"message"`
		} `json:"events"`
		Correlations []struct {
			ID         string   `json:"id"`
			Rule       string   `json:"rule"`
			Kind       string   `json:"kind"`
			Confidence string   `json:"confidence"`
			Evidence   string   `json:"evidence"`
			Members    []string `json:"members"`
		} `json:"correlations"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Correlations == nil {
		t.Fatal("correlations unmarshaled nil")
	}
	var exact, heuristic int
	ids := map[string]bool{}
	for _, ev := range payload.Events {
		ids[ev.EvidenceID] = true
	}
	for _, c := range payload.Correlations {
		if !strings.HasPrefix(c.ID, "corr-v1-") || c.Evidence == "" || len(c.Members) < 2 {
			t.Fatalf("incomplete group %+v", c)
		}
		for _, m := range c.Members {
			if !ids[m] {
				t.Fatalf("member %q is not an event evidence id", m)
			}
		}
		switch {
		case c.Rule == analyzer.RuleSharedRequestID && c.Kind == "exact" && c.Confidence == "high":
			exact++
		case c.Rule == analyzer.RuleClientIPWindow && c.Kind == "heuristic" && c.Confidence == "low":
			heuristic++
		default:
			t.Fatalf("unexpected group %+v", c)
		}
	}
	if exact != 1 || heuristic != 1 {
		t.Fatalf("exact=%d heuristic=%d", exact, heuristic)
	}
	if payload.HighConfidenceCorrelations != 1 || payload.LowConfidenceCorrelations != 1 {
		t.Fatalf("embedded correlation counts high=%d low=%d", payload.HighConfidenceCorrelations, payload.LowConfidenceCorrelations)
	}
	html := string(mustRead(t, p))
	if !strings.Contains(html, "Exact identifier (confidence:") || !strings.Contains(html, "Heuristic (confidence:") {
		t.Fatal("report JS is missing distinguishable kind labels")
	}
	if !strings.Contains(html, "shared-request-id") || !strings.Contains(html, "client-ip-window") {
		t.Fatal("report is missing documented rule ids")
	}
}

func TestWriteRedactsCorrelationEvidence(t *testing.T) {
	rule, err := redact.ParseRule("literal:abc-req-001")
	if err != nil {
		t.Fatal(err)
	}
	src := analyzer.Event{Message: "requestId=abc-req-001", File: "a.out", Instance: "t", Signature: "aa", Line: 1}
	p := filepath.Join(t.TempDir(), "corr-redact.html")
	err = Write(p, analyzer.Result{
		Events: []analyzer.Event{src},
		Correlations: []analyzer.Correlation{{
			ID:         "corr-v1-demo",
			Rule:       analyzer.RuleSharedRequestID,
			Kind:       analyzer.KindExact,
			Confidence: analyzer.ConfidenceHigh,
			Evidence:   "exact identifier request-id=abc-req-001",
			Members:    []int{0},
		}},
	}, Options{Redact: redact.New([]redact.Rule{rule})})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := embeddedJSON(mustRead(t, p))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(raw, []byte("abc-req-001")) {
		t.Fatal("unredacted correlation identifier leaked into embedded JSON")
	}
	if !bytes.Contains(raw, []byte("corr-v1-demo")) || !bytes.Contains(raw, []byte(redact.Replacement)) {
		t.Fatal("expected redacted evidence and stable correlation id")
	}
}

func assertEmbeddedArrays(t *testing.T, path string, wantEvents, wantWarnings, wantCorrelations int) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := embeddedJSON(b)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatalf("embedded JSON: %v", err)
	}
	for _, key := range []string{"events", "warnings", "correlations"} {
		v := fields[key]
		if bytes.Equal(v, []byte("null")) || !bytes.HasPrefix(bytes.TrimSpace(v), []byte("[")) {
			t.Fatalf("%s marshaled as %s, want a JSON array", key, v)
		}
	}
	var payload struct {
		Events       []json.RawMessage  `json:"events"`
		Warnings     []analyzer.Warning `json:"warnings"`
		Correlations []json.RawMessage  `json:"correlations"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Events == nil || payload.Warnings == nil || payload.Correlations == nil {
		t.Fatalf("unmarshaled nil slice events=%v warnings=%v correlations=%v", payload.Events, payload.Warnings, payload.Correlations)
	}
	if len(payload.Events) != wantEvents || len(payload.Warnings) != wantWarnings || len(payload.Correlations) != wantCorrelations {
		t.Fatalf("events=%d want %d; warnings=%d want %d; correlations=%d want %d",
			len(payload.Events), wantEvents, len(payload.Warnings), wantWarnings, len(payload.Correlations), wantCorrelations)
	}
	return raw
}

func embeddedJSON(html []byte) ([]byte, error) {
	s := string(html)
	i := strings.Index(s, "const REPORT =")
	if i < 0 {
		return nil, errors.New("embedded report JSON not found")
	}
	rest := strings.TrimSpace(s[i+len("const REPORT ="):])
	dec := json.NewDecoder(strings.NewReader(rest))
	var raw json.RawMessage
	if err := dec.Decode(&raw); err != nil {
		return nil, err
	}
	return raw, nil
}
