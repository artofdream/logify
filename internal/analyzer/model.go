package analyzer

import "time"

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
	EvidenceID   string    `json:"evidenceId"`
	Occurrences  int       `json:"occurrences"`
	FirstSeen    *time.Time `json:"firstSeen,omitempty"`
	LastSeen     *time.Time `json:"lastSeen,omitempty"`
	StatusCode   int       `json:"statusCode,omitempty"`
}
type Result struct {
	Root         string    `json:"root"`
	GeneratedAt  time.Time `json:"generatedAt"`
	FilesScanned int       `json:"filesScanned"`
	Events       []Event   `json:"events"`
	Warnings     []string  `json:"warnings"`
}
type Options struct{ From, To *time.Time }
