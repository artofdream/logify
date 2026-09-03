package main

import (
	"flag"
	"fmt"
	"github.com/artofdream/logify/internal/analyzer"
	"github.com/artofdream/logify/internal/report"
	"os"
	"time"
)

func main() {
	out := flag.String("output", "logify-report.html", "output HTML file")
	fromS := flag.String("from", "", "RFC3339 lower time bound")
	toS := flag.String("to", "", "RFC3339 upper time bound")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] DIRECTORY\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(2)
	}
	from, e := pt(*fromS)
	if e != nil {
		fatal("invalid -from: %v", e)
	}
	to, e := pt(*toS)
	if e != nil {
		fatal("invalid -to: %v", e)
	}
	r, e := analyzer.Analyze(flag.Arg(0), analyzer.Options{From: from, To: to})
	if e != nil {
		fatal("analyze: %v", e)
	}
	if e = report.Write(*out, r); e != nil {
		fatal("write report: %v", e)
	}
	fmt.Printf("Wrote %s (%d events from %d files, %d warnings)\n", *out, len(r.Events), r.FilesScanned, len(r.Warnings))
}
func pt(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, e := time.Parse(time.RFC3339, s)
	return &t, e
}
func fatal(f string, a ...any) { fmt.Fprintf(os.Stderr, "logify: "+f+"\n", a...); os.Exit(1) }
