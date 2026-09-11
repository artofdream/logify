package redact

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/artofdream/logify/internal/analyzer"
)

func TestParseRuleForms(t *testing.T) {
	t.Parallel()
	cases := []struct {
		spec string
		in   string
		want string
	}{
		{spec: "literal:sekrit", in: "token=sekrit;", want: "token=" + Replacement + ";"},
		{spec: `regex:(?i)password=\S+`, in: "Password=hunter2 extra", want: Replacement + " extra"},
		{spec: "email", in: "reach ops@example.com please", want: "reach " + Replacement + " please"},
		{spec: "EMAIL", in: "a@b.co", want: Replacement},
		{spec: "ipv4", in: "client 10.1.2.3 connected", want: "client " + Replacement + " connected"},
		{spec: "uuid", in: "id=550e8400-e29b-41d4-a716-446655440000", want: "id=" + Replacement},
		{spec: "bearer", in: "Authorization: Bearer abc.def-ghi", want: "Authorization: " + Replacement},
		{spec: "jwt", in: "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJ4In0.sig", want: Replacement},
		{spec: `token-\d+`, in: "token-99", want: Replacement},
	}
	for _, tc := range cases {
		rule, err := ParseRule(tc.spec)
		if err != nil {
			t.Fatalf("ParseRule(%q): %v", tc.spec, err)
		}
		got, n := New([]Rule{rule}).Apply(tc.in)
		if got != tc.want || n != 1 {
			t.Fatalf("spec %q: got %q n=%d want %q n=1", tc.spec, got, n, tc.want)
		}
	}
}

func TestParseRuleRejectsEmptyMatchAndInvalid(t *testing.T) {
	t.Parallel()
	for _, spec := range []string{"", "   ", "literal:", "regex:", `a*`, "("} {
		if _, err := ParseRule(spec); err == nil {
			t.Fatalf("ParseRule(%q) succeeded, want error", spec)
		}
	}
}

func TestLoadFileSkipsCommentsAndBlanks(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "rules.txt")
	body := "# comment\n\nemail\nliteral:fixed-secret\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	rules, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 2 {
		t.Fatalf("rules=%d want 2", len(rules))
	}
	engine := New(rules)
	got, n := engine.Apply("fixed-secret mailed to ops@example.com")
	if n != 2 || !strings.Contains(got, Replacement) || strings.Contains(got, "ops@example.com") || strings.Contains(got, "fixed-secret") {
		t.Fatalf("got %q n=%d", got, n)
	}
}

func TestLoadFileReportsLineOnBadRule(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "bad.txt")
	if err := os.WriteFile(path, []byte("email\n(\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadFile(path)
	if err == nil || !strings.Contains(err.Error(), path+":2:") {
		t.Fatalf("err=%v", err)
	}
}

func TestCompileNilWhenNoRules(t *testing.T) {
	t.Parallel()
	e, err := Compile(nil, "")
	if err != nil || e != nil {
		t.Fatalf("engine=%v err=%v", e, err)
	}
}

func TestExampleRulesFileParses(t *testing.T) {
	rules, err := LoadFile(filepath.Join("..", "..", "testdata", "redact", "rules.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 6 {
		t.Fatalf("example rules=%d want 6", len(rules))
	}
}

func TestCompileMissingFile(t *testing.T) {
	t.Parallel()
	_, err := Compile(nil, filepath.Join(t.TempDir(), "missing.txt"))
	if err == nil {
		t.Fatal("expected error for missing redact file")
	}
}

func TestApplyResultRedactsVisibleFieldsOnly(t *testing.T) {
	t.Parallel()
	// NFR-006: messages and other operator-visible log-derived fields.
	rule, err := ParseRule("literal:secret")
	if err != nil {
		t.Fatal(err)
	}
	engine := New([]Rule{rule})
	in := analyzer.Result{
		Root: "/tmp/secret/bundle",
		Events: []analyzer.Event{{
			Message:    "password secret leaked",
			File:       "secret/app.log",
			Instance:   "secret-host",
			Signature:  "aabbccdd",
			Severity:   analyzer.Error,
			StatusCode: 500,
			Line:       9,
			Timestamp:  time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC),
		}},
		Warnings: []analyzer.Warning{{
			File:     "secret/app.log",
			Category: analyzer.CategoryScanOverflow,
			Message:  "overflow",
		}},
	}
	out, n := ApplyResult(in, engine)
	if n != 5 {
		t.Fatalf("replacements=%d want 5", n)
	}
	if out.Root != "/tmp/"+Replacement+"/bundle" {
		t.Fatalf("root=%q", out.Root)
	}
	if out.Warnings[0].File != Replacement+"/app.log" {
		t.Fatalf("warning=%+v", out.Warnings[0])
	}
	ev := out.Events[0]
	if ev.Message != "password "+Replacement+" leaked" {
		t.Fatalf("message=%q", ev.Message)
	}
	if ev.File != Replacement+"/app.log" {
		t.Fatalf("file=%q", ev.File)
	}
	if ev.Instance != Replacement+"-host" {
		t.Fatalf("instance=%q", ev.Instance)
	}
	if ev.Signature != "aabbccdd" || ev.Severity != analyzer.Error || ev.StatusCode != 500 || ev.Line != 9 {
		t.Fatalf("non-visible fields changed: %#v", ev)
	}
	if in.Events[0].Message != "password secret leaked" {
		t.Fatal("ApplyResult mutated the input event")
	}
}

func TestApplyResultRedactsCorrelationEvidence(t *testing.T) {
	t.Parallel()
	// FR-012 / NFR-006: correlation evidence is log-derived display text.
	rule, err := ParseRule("literal:10.2.3.4")
	if err != nil {
		t.Fatal(err)
	}
	in := analyzer.Result{
		Events: []analyzer.Event{{
			Message:    "client 10.2.3.4 timeout",
			ClientAddr: "10.2.3.4",
			Signature:  "keep",
		}},
		Correlations: []analyzer.Correlation{{
			ID:         "corr-v1-keep",
			Rule:       analyzer.RuleClientIPWindow,
			Kind:       analyzer.KindHeuristic,
			Confidence: analyzer.ConfidenceLow,
			Evidence:   "heuristic client 10.2.3.4 within 5s",
			Members:    []int{0},
		}},
	}
	out, n := ApplyResult(in, New([]Rule{rule}))
	if n != 3 {
		t.Fatalf("replacements=%d want 3", n)
	}
	if out.Events[0].Signature != "keep" || out.Correlations[0].ID != "corr-v1-keep" || out.Correlations[0].Rule != analyzer.RuleClientIPWindow {
		t.Fatalf("identity fields changed: %#v %#v", out.Events[0], out.Correlations[0])
	}
	if out.Events[0].ClientAddr != Replacement || !strings.Contains(out.Correlations[0].Evidence, Replacement) {
		t.Fatalf("event=%#v corr=%#v", out.Events[0], out.Correlations[0])
	}
	if in.Correlations[0].Evidence != "heuristic client 10.2.3.4 within 5s" {
		t.Fatal("ApplyResult mutated the input correlation")
	}
}

func TestApplyResultNoEngineLeavesInput(t *testing.T) {
	t.Parallel()
	in := analyzer.Result{Events: []analyzer.Event{{Message: "plain"}}}
	out, n := ApplyResult(in, nil)
	if n != 0 || out.Events[0].Message != "plain" {
		t.Fatalf("out=%#v n=%d", out, n)
	}
}
