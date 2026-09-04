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
)

func TestWriteSelfContained(t *testing.T) {
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
}

func TestWriteEmptySlicesAreJSONArrays(t *testing.T) {
	// FR-013 / FR-014: nil Events/Warnings must not marshal as JSON null.
	p := filepath.Join(t.TempDir(), "empty.html")
	if err := Write(p, analyzer.Result{}); err != nil {
		t.Fatal(err)
	}
	assertEmbeddedArrays(t, p, 0, 0)
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
	raw := assertEmbeddedArrays(t, p, 6, 0)
	var payload struct {
		Events []struct {
			EvidenceID  string `json:"evidenceId"`
			Signature   string `json:"signature"`
			Instance    string `json:"instance"`
			File        string `json:"file"`
			Line        int    `json:"line"`
			Occurrences int    `json:"occurrences"`
		} `json:"events"`
		Warnings []string `json:"warnings"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Events == nil || payload.Warnings == nil {
		t.Fatal("unmarshaled nil slice")
	}
	for i, ev := range payload.Events {
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
	for _, name := range []string{"followup.js", "page.js", "followup_node_test.js"} {
		cmd := exec.Command(node, "--check", name)
		cmd.Dir = "."
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%s: %v\n%s", name, err, out)
		}
	}
}

func TestFollowUpStoreCreateExportImport(t *testing.T) {
	// FR-017 / FR-018 / FR-019 / FR-020 / FR-022 / FR-023: executed page-script
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

func requireNode(t *testing.T) string {
	t.Helper()
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal("node not on PATH; JS report/follow-up tests cannot run")
	}
	return node
}

func assertEmbeddedArrays(t *testing.T, path string, wantEvents, wantWarnings int) []byte {
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
	for _, key := range []string{"events", "warnings"} {
		v := fields[key]
		if bytes.Equal(v, []byte("null")) || !bytes.HasPrefix(bytes.TrimSpace(v), []byte("[")) {
			t.Fatalf("%s marshaled as %s, want a JSON array", key, v)
		}
	}
	var payload struct {
		Events   []json.RawMessage `json:"events"`
		Warnings []string          `json:"warnings"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Events == nil || payload.Warnings == nil {
		t.Fatalf("unmarshaled nil slice events=%v warnings=%v", payload.Events, payload.Warnings)
	}
	if len(payload.Events) != wantEvents || len(payload.Warnings) != wantWarnings {
		t.Fatalf("events=%d want %d; warnings=%d want %d", len(payload.Events), wantEvents, len(payload.Warnings), wantWarnings)
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
