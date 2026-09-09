package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/artofdream/logify/internal/analyzer"
	"github.com/artofdream/logify/internal/redact"
	"github.com/artofdream/logify/internal/report"
)

const sensitivityWarning = "logify: the HTML report contains copied log data (messages, paths, and host identifiers); treat it as sensitive as the source bundle"

type stringList []string

func (s *stringList) String() string { return strings.Join(*s, ", ") }
func (s *stringList) Set(v string) error {
	*s = append(*s, v)
	return nil
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("logify", flag.ContinueOnError)
	fs.SetOutput(stderr)
	out := fs.String("output", "logify-report.html", "output HTML file")
	fromS := fs.String("from", "", "RFC3339 lower time bound")
	toS := fs.String("to", "", "RFC3339 upper time bound")
	var redactSpecs stringList
	fs.Var(&redactSpecs, "redact", "optional redaction rule (repeatable): named preset (email, ipv4, uuid, bearer, jwt), regex:<pattern>, literal:<text>, or a bare regex")
	redactFile := fs.String("redact-file", "", "optional file of redaction rules (one per line; # comments and blanks ignored)")
	fs.Usage = func() {
		fmt.Fprintf(stderr, "Usage: %s [flags] DIRECTORY\n", os.Args[0])
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		fs.Usage()
		return 2
	}
	from, err := pt(*fromS)
	if err != nil {
		fmt.Fprintf(stderr, "logify: invalid -from: %v\n", err)
		return 1
	}
	to, err := pt(*toS)
	if err != nil {
		fmt.Fprintf(stderr, "logify: invalid -to: %v\n", err)
		return 1
	}
	engine, err := redact.Compile(redactSpecs, *redactFile)
	if err != nil {
		fmt.Fprintf(stderr, "logify: %v\n", err)
		return 1
	}
	result, err := analyzer.Analyze(fs.Arg(0), analyzer.Options{From: from, To: to})
	if err != nil {
		fmt.Fprintf(stderr, "logify: analyze: %v\n", err)
		return 1
	}
	if err = report.Write(*out, result, report.Options{Redact: engine}); err != nil {
		fmt.Fprintf(stderr, "logify: write report: %v\n", err)
		return 1
	}
	fmt.Fprintln(stderr, sensitivityWarning)
	if engine != nil {
		fmt.Fprintf(stderr, "logify: applied %d redaction rule(s); redaction is best-effort and does not guarantee secrets are gone\n", engine.RuleCount())
	}
	fmt.Fprintf(stdout, "Wrote %s (%s)\n", *out, result.SummaryLine())
	for _, w := range result.Warnings {
		fmt.Fprintf(stdout, "  %s\n", w.String())
	}
	return 0
}

func pt(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, e := time.Parse(time.RFC3339, s)
	return &t, e
}
