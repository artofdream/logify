package report

import (
	_ "embed"
	"encoding/json"
	"html/template"
	"os"
	"time"

	"github.com/artofdream/logify/internal/analyzer"
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
	Root         string    `json:"root"`
	GeneratedAt  time.Time `json:"generatedAt"`
	FilesScanned int       `json:"filesScanned"`
	Events       []event   `json:"events"`
	Warnings     []string  `json:"warnings"`
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

func Write(path string, r analyzer.Result) error {
	raw, err := json.Marshal(buildPayload(r))
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

func buildPayload(r analyzer.Result) payload {
	events := r.Events
	if events == nil {
		events = []analyzer.Event{}
	}
	warnings := r.Warnings
	if warnings == nil {
		warnings = []string{}
	}
	out := make([]event, 0, len(events))
	for _, src := range events {
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
			EvidenceID:   EvidenceID(src),
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
		Root:         r.Root,
		GeneratedAt:  r.GeneratedAt,
		FilesScanned: r.FilesScanned,
		Events:       out,
		Warnings:     warnings,
	}
}
