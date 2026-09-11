// Package redact applies optional, operator-supplied rules to log-derived
// text before it is embedded in a report (NFR-006).
//
// Redaction is best-effort string replacement. It does not modify source
// bundles, does not invent secret-detection beyond the rules the operator
// supplied, and does not make a report safe to publish.
package redact

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/artofdream/logify/internal/analyzer"
)

// Replacement is the token substituted for every match.
const Replacement = "[REDACTED]"

// Named presets are documented shortcuts. They are conservative and incomplete;
// operators who need other identifiers should add explicit regex or literal rules.
var presets = map[string]string{
	"email":  `[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}`,
	"ipv4":   `\b(?:\d{1,3}\.){3}\d{1,3}\b`,
	"uuid":   `\b[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}\b`,
	"bearer": `(?i)\bBearer\s+[A-Za-z0-9\-._~+/]+=*`,
	"jwt":    `\beyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\b`,
}

// Rule is one compiled replacement.
type Rule struct {
	Spec    string
	Pattern *regexp.Regexp
}

// Engine applies rules in declaration order.
type Engine struct {
	rules []Rule
}

// New returns an engine. A nil or empty rule list is a no-op.
func New(rules []Rule) *Engine {
	return &Engine{rules: append([]Rule(nil), rules...)}
}

// RuleCount is the number of compiled rules.
func (e *Engine) RuleCount() int {
	if e == nil {
		return 0
	}
	return len(e.rules)
}

// ParseRule compiles one CLI or config-file specification.
//
// Forms:
//   - literal:<text>  exact substring (regexp-quoted)
//   - regex:<pattern> Go RE2 pattern
//   - <preset>        email, ipv4, uuid, bearer, or jwt (case-insensitive)
//   - <pattern>       treated as a Go RE2 pattern
//
// Patterns that match the empty string are rejected so replacement cannot loop.
func ParseRule(spec string) (Rule, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return Rule{}, fmt.Errorf("empty redaction rule")
	}
	pattern, err := patternFor(spec)
	if err != nil {
		return Rule{}, err
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return Rule{}, fmt.Errorf("invalid redaction rule %q: %w", spec, err)
	}
	if re.MatchString("") {
		return Rule{}, fmt.Errorf("redaction rule %q matches empty string", spec)
	}
	return Rule{Spec: spec, Pattern: re}, nil
}

func patternFor(spec string) (string, error) {
	switch {
	case strings.HasPrefix(spec, "literal:"):
		text := strings.TrimPrefix(spec, "literal:")
		if text == "" {
			return "", fmt.Errorf("redaction rule %q has empty literal", spec)
		}
		return regexp.QuoteMeta(text), nil
	case strings.HasPrefix(spec, "regex:"):
		pattern := strings.TrimPrefix(spec, "regex:")
		if pattern == "" {
			return "", fmt.Errorf("redaction rule %q has empty regex", spec)
		}
		return pattern, nil
	}
	if p, ok := presets[strings.ToLower(spec)]; ok {
		return p, nil
	}
	return spec, nil
}

// LoadFile reads one rule per line. Blank lines and trimmed lines starting
// with # are ignored.
func LoadFile(path string) ([]Rule, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("read redact file: %w", err)
	}
	defer f.Close()
	var rules []Rule
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	line := 0
	for sc.Scan() {
		line++
		text := strings.TrimSpace(sc.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		rule, err := ParseRule(text)
		if err != nil {
			return nil, fmt.Errorf("%s:%d: %w", path, line, err)
		}
		rules = append(rules, rule)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read redact file %s: %w", path, err)
	}
	return rules, nil
}

// Compile parses CLI specs then appends rules from an optional file.
func Compile(specs []string, file string) (*Engine, error) {
	var rules []Rule
	for _, spec := range specs {
		rule, err := ParseRule(spec)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	if file != "" {
		fromFile, err := LoadFile(file)
		if err != nil {
			return nil, err
		}
		rules = append(rules, fromFile...)
	}
	if len(rules) == 0 {
		return nil, nil
	}
	return New(rules), nil
}

// Apply replaces matches in s and returns the number of replacements.
func (e *Engine) Apply(s string) (string, int) {
	if e == nil || len(e.rules) == 0 || s == "" {
		return s, 0
	}
	n := 0
	out := s
	for _, rule := range e.rules {
		out = rule.Pattern.ReplaceAllStringFunc(out, func(string) string {
			n++
			return Replacement
		})
	}
	return out, n
}

// ApplyResult copies r and redacts operator-visible log-derived fields:
// Root, Warnings, each event Message, File, Instance, and ClientAddr, and
// each correlation Evidence string.
// Signatures, evidence identity inputs, timestamps, severity, status
// codes, correlation IDs, and rule/kind/confidence labels are left unchanged
// so the caller can compute identities first.
func ApplyResult(r analyzer.Result, e *Engine) (analyzer.Result, int) {
	if e == nil || e.RuleCount() == 0 {
		return r, 0
	}
	out := r
	n := 0
	out.Root, n = replaceCount(e, out.Root, n)
	if r.Warnings != nil {
		warnings := make([]analyzer.Warning, len(r.Warnings))
		for i, w := range r.Warnings {
			warnings[i] = w
			warnings[i].File, n = replaceCount(e, w.File, n)
			warnings[i].Message, n = replaceCount(e, w.Message, n)
		}
		out.Warnings = warnings
	}
	if r.Events != nil {
		events := make([]analyzer.Event, len(r.Events))
		copy(events, r.Events)
		for i := range events {
			events[i].Message, n = replaceCount(e, events[i].Message, n)
			events[i].File, n = replaceCount(e, events[i].File, n)
			events[i].Instance, n = replaceCount(e, events[i].Instance, n)
			events[i].ClientAddr, n = replaceCount(e, events[i].ClientAddr, n)
		}
		out.Events = events
	}
	if r.Correlations != nil {
		corrs := make([]analyzer.Correlation, len(r.Correlations))
		copy(corrs, r.Correlations)
		for i := range corrs {
			corrs[i].Evidence, n = replaceCount(e, corrs[i].Evidence, n)
		}
		out.Correlations = corrs
	}
	return out, n
}

func replaceCount(e *Engine, s string, n int) (string, int) {
	out, added := e.Apply(s)
	return out, n + added
}
