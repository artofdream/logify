package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/artofdream/logify/internal/analyzer"
)

func TestContrastRatioKnownPair(t *testing.T) {
	// WCAG 2.2 example: black on white is 21:1.
	got, err := ContrastRatio("#000000", "#ffffff")
	if err != nil {
		t.Fatal(err)
	}
	if got < 20.9 || got > 21.1 {
		t.Fatalf("black/white contrast=%v want 21", got)
	}
	low, err := ContrastRatio("#2b3858", "#151e32")
	if err != nil {
		t.Fatal(err)
	}
	if low >= wcagAANonText {
		t.Fatalf("legacy --line on --panel unexpectedly passes: %v", low)
	}
}

func TestNFR013ReportCSSContrast(t *testing.T) {
	// NFR-013 AC3: token pairs used by the offline report meet WCAG 2.2 AA.
	css, err := os.ReadFile("page.css")
	if err != nil {
		t.Fatal(err)
	}
	if fails := checkContrastPairs(string(css)); len(fails) > 0 {
		t.Fatalf("NFR-013 contrast failures:\n  %s", strings.Join(fails, "\n  "))
	}
}

func TestNFR013GeneratedReportA11y(t *testing.T) {
	// NFR-013 AC1 / AC4: generated HTML controls have programmatic names;
	// embedded CSS contrast is re-checked on the written file. This is a
	// stdlib structural/contrast probe, not axe-core or an AT run.
	p := filepath.Join(t.TempDir(), "a11y.html")
	err := Write(p, analyzer.Result{
		Events: []analyzer.Event{{
			Message:    "disk full",
			Severity:   "ERROR",
			File:       "catalina.out",
			Instance:   "tomcat-a",
			SourceType: "tomcat-java",
			Line:       12,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	html := string(mustRead(t, p))
	if fails := checkLabeledControls(html); len(fails) > 0 {
		t.Fatalf("NFR-013 unlabeled controls in generated report:\n  %s", strings.Join(fails, "\n  "))
	}
	if !strings.Contains(html, `for="q"`) || !strings.Contains(html, `for="iq"`) || !strings.Contains(html, `for="ioverdue"`) {
		t.Fatal("generated report is missing explicit label for= bindings")
	}
	if !strings.Contains(html, "Severity: ") {
		t.Fatal("generated report JS is missing Severity: text")
	}
	if !strings.Contains(html, "ISSUE_PAGE_SIZE") || !strings.Contains(html, "issue-pager") {
		t.Fatal("generated report is missing NFR-021 issue-list windowing chrome")
	}
	if !strings.Contains(html, "--control-border") {
		t.Fatal("generated report CSS is missing AA control-border token")
	}
	styleStart := strings.Index(html, "<style>")
	styleEnd := strings.Index(html, "</style>")
	if styleStart < 0 || styleEnd < styleStart {
		t.Fatal("generated report missing embedded CSS")
	}
	if fails := checkContrastPairs(html[styleStart:styleEnd]); len(fails) > 0 {
		t.Fatalf("NFR-013 generated CSS contrast failures:\n  %s", strings.Join(fails, "\n  "))
	}
	if strings.Contains(html, "https://") || strings.Contains(html, "http://") {
		t.Fatal("generated report has a network URL")
	}
}

func TestNFR013PageJSLabeledFields(t *testing.T) {
	// Dynamic issue-card controls are created in page.js, not the static
	// HTML. Source contracts cover those labels without executing a browser.
	js := string(mustRead(t, "page.js"))
	for _, needle := range []string{
		"function labeledField(",
		"lab.setAttribute('for', control.id)",
		"labeledField('Title', title)",
		"labeledField('Workflow state', state)",
		"labeledField('New tag', tagInput)",
		"labeledField('Owner', ownerInput)",
		"labeledField('Notes', notes)",
		"dueLabel.setAttribute('for', dueInput.id)",
		"labeledField('Issue to link', linkSelect)",
		"'Severity: ' + sevText",
	} {
		if !strings.Contains(js, needle) {
			t.Fatalf("page.js missing NFR-013 contract %q", needle)
		}
	}
	if strings.Contains(js, "innerHTML") {
		t.Fatal("page.js must not assign innerHTML")
	}
}
