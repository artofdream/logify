package report

import (
	_ "embed"
	"encoding/json"
	"html/template"
	"os"
	"time"

	"github.com/artofdream/logify/internal/analyzer"
	"github.com/artofdream/logify/internal/redact"
)

//go:embed page.html
var pageHTML string

//go:embed page.css
var pageCSS string

//go:embed followup.js
var followUpJS string

//go:embed page.js
var pageJS string

type payload struct {
	Root           string             `json:"root"`
	GeneratedAt    time.Time          `json:"generatedAt"`
	FilesScanned   int                `json:"filesScanned"`
	FilesProcessed int                `json:"filesProcessed"`
	FilesSkipped   int                `json:"filesSkipped"`
	FilesFailed    int                `json:"filesFailed"`
	Events         []event            `json:"events"`
	Warnings       []analyzer.Warning `json:"warnings"`
	Redaction      redactionInfo      `json:"redaction"`
}

type redactionInfo struct {
	Enabled      bool `json:"enabled"`
	RuleCount    int  `json:"ruleCount"`
	Replacements int  `json:"replacements"`
}

// Options controls report embedding. Zero value preserves historical Write
// behavior except for the always-on sensitivity banner in the HTML shell.
type Options struct {
	Redact *redact.Engine
}

type event struct {
	Timestamp    time.Time         `json:"timestamp"`
	HasTimestamp bool              `json:"hasTimestamp"`
	Severity     analyzer.Severity `json:"severity"`
	SourceType   string            `json:"sourceType"`
	Instance     string            `json:"instance"`
	File         string            `json:"file"`
	Line         int               `json:"line"`
	Message      string            `json:"message"`
	Signature    string            `json:"signature"`
	EvidenceID   string            `json:"evidenceId"`
	Occurrences  int               `json:"occurrences"`
	FirstSeen    *time.Time        `json:"firstSeen,omitempty"`
	LastSeen     *time.Time        `json:"lastSeen,omitempty"`
	StatusCode   int               `json:"statusCode,omitempty"`
}

type pageView struct {
	Data     template.JS
	CSS      template.CSS
	FollowUp template.JS
	Page     template.JS
}

func Write(path string, r analyzer.Result, opt ...Options) error {
	var o Options
	if len(opt) > 0 {
		o = opt[0]
	}
	raw, err := json.Marshal(buildPayload(r, o))
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	tmpl := template.Must(template.New("page").Parse(pageHTML))
	return tmpl.Execute(f, pageView{
		Data:     template.JS(raw),
		CSS:      template.CSS(pageCSS),
		FollowUp: template.JS(followUpJS),
		Page:     template.JS(pageJS),
	})
}

func buildPayload(r analyzer.Result, opt Options) payload {
	events := r.Events
	if events == nil {
		events = []analyzer.Event{}
	}
	ids := make([]string, len(events))
	for i, src := range events {
		ids[i] = EvidenceID(src)
	}
	redacted, replacements := redact.ApplyResult(analyzer.Result{
		Root:           r.Root,
		GeneratedAt:    r.GeneratedAt,
		FilesScanned:   r.FilesScanned,
		FilesProcessed: r.FilesProcessed,
		FilesSkipped:   r.FilesSkipped,
		FilesFailed:    r.FilesFailed,
		Events:         events,
		Warnings:       r.Warnings,
	}, opt.Redact)
	events = redacted.Events
	warnings := redacted.Warnings
	if warnings == nil {
		warnings = []analyzer.Warning{}
	}
	info := redactionInfo{
		Enabled:      opt.Redact != nil && opt.Redact.RuleCount() > 0,
		RuleCount:    opt.Redact.RuleCount(),
		Replacements: replacements,
	}
	out := make([]event, 0, len(events))
	for i, src := range events {
		item := event{
			Timestamp:    src.Timestamp,
			HasTimestamp: src.HasTimestamp,
			Severity:     src.Severity,
			SourceType:   src.SourceType,
			Instance:     src.Instance,
			File:         src.File,
			Line:         src.Line,
			Message:      src.Message,
			Signature:    src.Signature,
			EvidenceID:   ids[i],
			Occurrences:  src.Occurrences,
			StatusCode:   src.StatusCode,
		}
		if src.HasTimestamp {
			first := src.Timestamp
			last := src.LastSeen
			item.FirstSeen = &first
			item.LastSeen = &last
		}
		out = append(out, item)
	}
	return payload{
		Root:           redacted.Root,
		GeneratedAt:    r.GeneratedAt,
		FilesScanned:   r.FilesScanned,
		FilesProcessed: r.FilesProcessed,
		FilesSkipped:   r.FilesSkipped,
		FilesFailed:    r.FilesFailed,
		Events:         out,
		Warnings:       warnings,
		Redaction:      info,
	}
}
