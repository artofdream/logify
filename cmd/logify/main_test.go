package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNFR016StableFlagsExist(t *testing.T) {
	// NFR-016: documented public flags and defaults remain registered;
	// -h exits 0.
	fs := flag.NewFlagSet("logify", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var cli cliFlags
	registerCLI(fs, &cli)
	for _, name := range []string{"from", "to", "output", "redact", "redact-file"} {
		if fs.Lookup(name) == nil {
			t.Errorf("missing stable flag -%s", name)
		}
	}
	if got := fs.Lookup("output"); got == nil || got.DefValue != defaultOutput {
		t.Fatalf("output default=%v want %q", got, defaultOutput)
	}
	var stdout, stderr bytes.Buffer
	if code := run([]string{"-h"}, &stdout, &stderr); code != 0 {
		t.Fatalf("help exit=%d want 0 stderr=%q", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("help stdout=%q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "Usage:") {
		t.Fatalf("help stderr=%q", stderr.String())
	}
}

func TestNFR016VersionFlag(t *testing.T) {
	// NFR-016: -version / -V print the build version and exit 0 without a directory.
	for _, arg := range []string{"-version", "-V"} {
		t.Run(arg, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if code := run([]string{arg}, &stdout, &stderr); code != 0 {
				t.Fatalf("exit=%d want 0 stderr=%q", code, stderr.String())
			}
			want := "logify " + version
			if !strings.Contains(stdout.String(), want) {
				t.Fatalf("stdout=%q want substring %q", stdout.String(), want)
			}
		})
	}
}

func TestRunHelpExitsZero(t *testing.T) {
	// NFR-011: -h / -help print usage and are not a failure.
	for _, arg := range []string{"-h", "-help"} {
		t.Run(arg, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if code := run([]string{arg}, &stdout, &stderr); code != 0 {
				t.Fatalf("exit=%d want 0 stderr=%q", code, stderr.String())
			}
			if stdout.Len() != 0 {
				t.Fatalf("stdout=%q", stdout.String())
			}
			if !strings.Contains(stderr.String(), "Usage:") {
				t.Fatalf("stderr=%q", stderr.String())
			}
		})
	}
}

func TestRunUsageWhenDirectoryMissing(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run(nil, &stdout, &stderr); code != 2 {
		t.Fatalf("exit=%d want 2", code)
	}
	if !strings.Contains(stderr.String(), "Usage:") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}

func TestRunMissingRedactFile(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"-redact-file", filepath.Join(t.TempDir(), "missing.txt"), t.TempDir()}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("exit=%d want 1 stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "redact file") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}

func TestRunInvalidRedactRule(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"-redact", "(", t.TempDir()}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("exit=%d want 1 stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "redaction") && !strings.Contains(stderr.String(), "invalid") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}

func TestRunWarnsAndLeavesTextWithoutRedact(t *testing.T) {
	dir := writeSecretBundle(t)
	out := filepath.Join(t.TempDir(), "plain.html")
	var stdout, stderr bytes.Buffer
	code := run([]string{"-output", out, dir}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit=%d stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "copied log data") {
		t.Fatalf("missing sensitivity warning: %q", stderr.String())
	}
	if strings.Contains(stderr.String(), "applied") {
		t.Fatalf("unexpected redaction note: %q", stderr.String())
	}
	body := readFile(t, out)
	if !strings.Contains(body, "ops@example.com") {
		t.Fatal("default run redacted email without -redact")
	}
	if !strings.Contains(body, "Sensitive copy.") {
		t.Fatal("report missing sensitivity banner")
	}
	if !strings.Contains(body, `"enabled":false`) {
		t.Fatal("expected redaction.enabled false")
	}
}

func TestRunRedactFlagAndFile(t *testing.T) {
	dir := writeSecretBundle(t)
	rules := filepath.Join(t.TempDir(), "rules.txt")
	if err := os.WriteFile(rules, []byte("literal:fixed-token\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "redacted.html")
	var stdout, stderr bytes.Buffer
	code := run([]string{"-output", out, "-redact", "email", "-redact-file", rules, dir}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit=%d stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "copied log data") || !strings.Contains(stderr.String(), "applied 2 redaction rule") {
		t.Fatalf("stderr=%q", stderr.String())
	}
	body := readFile(t, out)
	if strings.Contains(body, "ops@example.com") || strings.Contains(body, "fixed-token") {
		t.Fatal("secrets still visible in report")
	}
	if !strings.Contains(body, "[REDACTED]") {
		t.Fatal("expected replacement token")
	}
	raw := mustEmbeddedJSON(t, body)
	var payload struct {
		Events []struct {
			Message string `json:"message"`
		} `json:"events"`
		Redaction struct {
			Enabled      bool `json:"enabled"`
			RuleCount    int  `json:"ruleCount"`
			Replacements int  `json:"replacements"`
		} `json:"redaction"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatal(err)
	}
	if !payload.Redaction.Enabled || payload.Redaction.RuleCount != 2 || payload.Redaction.Replacements < 2 {
		t.Fatalf("redaction meta %#v", payload.Redaction)
	}
	if len(payload.Events) == 0 || !strings.Contains(payload.Events[0].Message, "[REDACTED]") {
		t.Fatalf("events=%#v", payload.Events)
	}
}

func writeSecretBundle(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "node-a")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	line := "2026-09-09 10:00:00,123 ERROR leaked ops@example.com token=fixed-token\n"
	if err := os.WriteFile(filepath.Join(dir, "catalina.out"), []byte(line), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func mustEmbeddedJSON(t *testing.T, html string) []byte {
	t.Helper()
	i := strings.Index(html, "const REPORT =")
	if i < 0 {
		t.Fatal("embedded report JSON not found")
	}
	rest := strings.TrimSpace(html[i+len("const REPORT ="):])
	dec := json.NewDecoder(strings.NewReader(rest))
	var raw json.RawMessage
	if err := dec.Decode(&raw); err != nil {
		t.Fatal(err)
	}
	return raw
}
