package analyzer

import (
	"fmt"
	"time"
)

type Severity string

const (
	Trace   Severity = "TRACE"
	Debug   Severity = "DEBUG"
	Info    Severity = "INFO"
	Warn    Severity = "WARN"
	Error   Severity = "ERROR"
	Fatal   Severity = "FATAL"
	Unknown Severity = "UNKNOWN"
)

type Event struct {
	Timestamp    time.Time `json:"timestamp"`
	HasTimestamp bool      `json:"hasTimestamp"`
	Severity     Severity  `json:"severity"`
	SourceType   string    `json:"sourceType"`
	Instance     string    `json:"instance"`
	File         string    `json:"file"`
	Line         int       `json:"line"`
	Message      string    `json:"message"`
	Signature    string    `json:"signature"`
	Occurrences  int       `json:"occurrences"`
	LastSeen     time.Time `json:"lastSeen"`
	StatusCode   int       `json:"statusCode,omitempty"`
}

// WarningCategory is a deterministic scan-failure class (FR-015).
type WarningCategory string

const (
	CategoryWalkError    WarningCategory = "walk-error"
	CategoryOpenError    WarningCategory = "open-error"
	CategoryScanOverflow WarningCategory = "scan-overflow"
	CategoryScanError    WarningCategory = "scan-error"
)

// Warning is one recoverable scan problem. Line/LineEnd are omitted when the
// failure is not attributable to a specific record range.
type Warning struct {
	File     string          `json:"file"`
	Category WarningCategory `json:"category"`
	Line     int             `json:"line,omitempty"`
	LineEnd  int             `json:"lineEnd,omitempty"`
	Message  string          `json:"message"`
}

func (w Warning) Location() string {
	if w.File == "" {
		if w.Line <= 0 {
			return ""
		}
		return fmt.Sprintf("%d", w.Line)
	}
	if w.Line <= 0 {
		return w.File
	}
	if w.LineEnd > w.Line {
		return fmt.Sprintf("%s:%d-%d", w.File, w.Line, w.LineEnd)
	}
	return fmt.Sprintf("%s:%d", w.File, w.Line)
}

func (w Warning) String() string {
	loc := w.Location()
	if loc == "" {
		return fmt.Sprintf("[%s] %s", w.Category, w.Message)
	}
	return fmt.Sprintf("%s [%s] %s", loc, w.Category, w.Message)
}

type Result struct {
	Root           string    `json:"root"`
	GeneratedAt    time.Time `json:"generatedAt"`
	FilesScanned   int       `json:"filesScanned"`
	FilesProcessed int       `json:"filesProcessed"`
	FilesSkipped   int       `json:"filesSkipped"`
	FilesFailed    int       `json:"filesFailed"`
	Events         []Event   `json:"events"`
	Warnings       []Warning `json:"warnings"`
}

func (r Result) SummaryLine() string {
	return fmt.Sprintf("%d events from %d files; processed=%d skipped=%d failed=%d; %d warnings",
		len(r.Events), r.FilesScanned, r.FilesProcessed, r.FilesSkipped, r.FilesFailed, len(r.Warnings))
}

func (r Result) CountsReconcile() bool {
	return r.FilesScanned == r.FilesProcessed+r.FilesFailed
}

type Options struct{ From, To *time.Time }
