package report

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
}

func TestWriteEmptySlicesAreJSONArrays(t *testing.T) {
	// FR-013 / FR-014: nil Events/Warnings must not marshal as JSON null or
	// `d.warnings.length` / `d.events.map` throw before render().
	p := filepath.Join(t.TempDir(), "empty.html")
	if err := Write(p, analyzer.Result{}); err != nil {
		t.Fatal(err)
	}
	assertEmbeddedArrays(t, p, 0, 0)
}

func TestWriteFixtureReportArraysAndRenderGuard(t *testing.T) {
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
	// Enough of the page script to reach render(): arrays support .map/.length/.filter.
	var payload struct {
		Events   []analyzer.Event `json:"events"`
		Warnings []string         `json:"warnings"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatal(err)
	}
	_ = len(payload.Events)
	_ = len(payload.Warnings)
	severities := make([]string, 0, len(payload.Events))
	for _, e := range payload.Events {
		severities = append(severities, string(e.Severity))
	}
	if len(severities) != 6 {
		t.Fatalf("render-equivalent map yielded %d severities", len(severities))
	}
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
		Events   []analyzer.Event `json:"events"`
		Warnings []string         `json:"warnings"`
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
	i := strings.Index(s, "const d=")
	j := strings.Index(s, ",$=x=>")
	if i < 0 || j < i {
		return nil, errors.New("embedded report JSON not found")
	}
	return []byte(s[i+len("const d=") : j]), nil
}
