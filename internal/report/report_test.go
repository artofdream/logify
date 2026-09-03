package report

import (
	"github.com/artofdream/logify/internal/analyzer"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
