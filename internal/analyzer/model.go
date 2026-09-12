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
	// ParseConfidence is high when a format-specific parser matched the
	// record and low when the line was retained as unrecognized (FR-008).
	ParseConfidence Confidence `json:"parseConfidence,omitempty"`
	// UnparsedOccurrences is how many collapsed rows were unrecognized.
	// It is the unparsed-line count after FR-010 merge; Occurrences may be larger.
	UnparsedOccurrences int `json:"unparsedOccurrences,omitempty"`
	// ClientAddr is the parsed HTTPD remote address when known. It is used by
	// correlation (FR-012) and is not part of signature or evidence identity.
	ClientAddr string `json:"clientAddr,omitempty"`
	// occHints are this row plus occurrences collapsed by FR-010 so
	// correlation can still see later labeled IDs and client addresses.
	occHints []occHint
}

// occHint is one pre-dedup occurrence used only by correlation. It is not
// serialized into the report payload.
type occHint struct {
	ts   time.Time
	has  bool
	msg  string
	addr string
}

type CorrelationKind string

const (
	KindExact     CorrelationKind = "exact"
	KindHeuristic CorrelationKind = "heuristic"
)

type Confidence string

const (
	ConfidenceHigh Confidence = "high"
	ConfidenceLow  Confidence = "low"
)

const (
	RuleSharedRequestID = "shared-request-id"
	RuleClientIPWindow  = "client-ip-window"
)

// Correlation is one inferred group. Members are indexes into Result.Events.
type Correlation struct {
	ID         string          `json:"id"`
	Rule       string          `json:"rule"`
	Kind       CorrelationKind `json:"kind"`
	Confidence Confidence      `json:"confidence"`
	Evidence   string          `json:"evidence"`
	Members    []int           `json:"members"`
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
	Root                       string        `json:"root"`
	GeneratedAt                time.Time     `json:"generatedAt"`
	FilesScanned               int           `json:"filesScanned"`
	FilesProcessed             int           `json:"filesProcessed"`
	FilesSkipped               int           `json:"filesSkipped"`
	FilesFailed                int           `json:"filesFailed"`
	UnparsedRecords            int           `json:"unparsedRecords"`
	HighConfidenceCorrelations int           `json:"highConfidenceCorrelations"`
	LowConfidenceCorrelations  int           `json:"lowConfidenceCorrelations"`
	Events                     []Event       `json:"events"`
	Warnings                   []Warning     `json:"warnings"`
	Correlations               []Correlation `json:"correlations"`
}

func (r Result) SummaryLine() string {
	unparsed, high, low := r.Observability()
	return fmt.Sprintf("%d events from %d files; processed=%d skipped=%d failed=%d; %d warnings; unparsed=%d; correlations=%d (high=%d low=%d)",
		len(r.Events), r.FilesScanned, r.FilesProcessed, r.FilesSkipped, r.FilesFailed, len(r.Warnings),
		unparsed, len(r.Correlations), high, low)
}

// Observability counts retained unrecognized records (by occurrence) and
// correlation groups by confidence. Empty parseConfidence is not unparsed.
func (r Result) Observability() (unparsed, highCorr, lowCorr int) {
	for _, e := range r.Events {
		if e.UnparsedOccurrences > 0 {
			unparsed += e.UnparsedOccurrences
			continue
		}
		if e.ParseConfidence == ConfidenceLow {
			unparsed++
		}
	}
	for _, c := range r.Correlations {
		switch c.Confidence {
		case ConfidenceHigh:
			highCorr++
		case ConfidenceLow:
			lowCorr++
		}
	}
	return unparsed, highCorr, lowCorr
}

func (r *Result) refreshObservability() {
	r.UnparsedRecords, r.HighConfidenceCorrelations, r.LowConfidenceCorrelations = r.Observability()
}

func (r Result) CountsReconcile() bool {
	return r.FilesScanned == r.FilesProcessed+r.FilesFailed
}

type Options struct{ From, To *time.Time }
