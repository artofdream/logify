package analyzer

import (
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var javaStart = regexp.MustCompile(`^(?:\[)?(\d{4}-\d{2}-\d{2}[ T]\d{2}:\d{2}:\d{2}(?:[.,]\d+)?(?:Z|[+-]\d\d:?\d\d)?)(?:\])?\s+(?:\[([^]]+)\]\s+)?(?:(TRACE|DEBUG|INFO|WARN|WARNING|ERROR|FATAL|SEVERE)\b)?\s*(.*)$`)
var httpdError = regexp.MustCompile(`^\[([^]]+)\]\s+(?:\[([^]:]+)(?::([^]]+))?\]\s+)?(?:\[pid[^]]+\]\s+)?(?:\[client[^]]+\]\s+)?(.*)$`)
var httpdAccess = regexp.MustCompile(`^(\S+)\s+\S+\s+\S+\s+\[([^]]+)\]\s+"([^"]*)"\s+(\d{3})\s+(\S+)`)
var httpdClient = regexp.MustCompile(`\[client\s+([^]]+)\]`)
var volatile = regexp.MustCompile(`(?i)(?:0x[0-9a-f]+|\b\d{2,}\b|[0-9a-f]{8}-[0-9a-f-]{27,})`)
var javaException = regexp.MustCompile(`^(?:Caused by:\s+)?[\w.$]+(?:Exception|Error)(?::|$)`)

func Analyze(root string, o Options) (Result, error) {
	abs, e := filepath.Abs(root)
	if e != nil {
		return Result{}, e
	}
	i, e := os.Stat(abs)
	if e != nil {
		return Result{}, e
	}
	if !i.IsDir() {
		return Result{}, fmt.Errorf("%s is not a directory", abs)
	}
	r := Result{Root: abs, GeneratedAt: time.Now(), Events: []Event{}, Warnings: []Warning{}, Correlations: []Correlation{}}
	e = filepath.WalkDir(abs, func(p string, d os.DirEntry, we error) error {
		if we != nil {
			r.FilesSkipped++
			r.Warnings = append(r.Warnings, newWarning(relPath(abs, p), CategoryWalkError, 0, 0, we.Error()))
			return nil
		}
		if d.IsDir() || !looks(d.Name()) {
			return nil
		}
		r.FilesScanned++
		es, w := parseFile(abs, p)
		if w != nil {
			r.Warnings = append(r.Warnings, *w)
			if w.Category == CategoryOpenError {
				r.FilesFailed++
			} else {
				r.FilesProcessed++
			}
		} else {
			r.FilesProcessed++
		}
		for _, v := range es {
			if o.From != nil && (!v.HasTimestamp || v.Timestamp.Before(*o.From)) {
				continue
			}
			if o.To != nil && (!v.HasTimestamp || v.Timestamp.After(*o.To)) {
				continue
			}
			r.Events = append(r.Events, v)
		}
		return nil
	})
	if e != nil {
		return Result{}, e
	}
	r.Events = dedup(r.Events)
	sort.SliceStable(r.Events, func(i, j int) bool {
		if r.Events[i].HasTimestamp != r.Events[j].HasTimestamp {
			return r.Events[i].HasTimestamp
		}
		return r.Events[i].Timestamp.Before(r.Events[j].Timestamp)
	})
	r.Correlations = Correlate(r.Events)
	r.refreshObservability()
	return r, nil
}
func looks(n string) bool {
	n = canonicalLogName(n)
	return strings.HasSuffix(n, ".log") || strings.HasSuffix(n, ".out") || strings.Contains(n, "access_log") || strings.Contains(n, "error_log")
}

func parseFile(root, path string) ([]Event, *Warning) {
	rel := relPath(root, path)
	f, e := os.Open(path)
	if e != nil {
		w := newWarning(rel, CategoryOpenError, 0, 0, e.Error())
		return nil, &w
	}
	defer f.Close()
	in, closer, e := openLogStream(f, path)
	if e != nil {
		w := newWarning(rel, CategoryScanError, 0, 0, e.Error())
		return nil, &w
	}
	src := detect(path)
	inst := instance(rel)
	s := bufio.NewScanner(in)
	s.Buffer(make([]byte, 65536), 4*1024*1024)
	var out []Event
	var cur *Event
	line := 0
	flush := func() {
		if cur != nil {
			decorate(cur, src, inst, rel, cur.Line)
			out = append(out, *cur)
			cur = nil
		}
	}
	for s.Scan() {
		line++
		x := s.Text()
		if src == "apache-access" {
			if v, ok := access(x); ok {
				flush()
				decorate(&v, src, inst, rel, line)
				out = append(out, v)
			} else if strings.TrimSpace(x) != "" {
				flush()
				v := Event{Severity: infer(x), Message: x, ParseConfidence: ConfidenceLow}
				decorate(&v, src, inst, rel, line)
				out = append(out, v)
			}
			continue
		}
		v, ok := java(x)
		if !ok && src == "apache-error" {
			v, ok = apacheError(x)
		}
		if ok {
			flush()
			v.Line = line
			cur = &v
			continue
		}
		trim := strings.TrimSpace(x)
		if cur != nil && (strings.HasPrefix(x, "\t") || strings.HasPrefix(trim, "at ") || strings.HasPrefix(trim, "Caused by:") || strings.HasPrefix(trim, "... ") || javaException.MatchString(trim)) {
			cur.Message += "\n" + x
			continue
		}
		if trim != "" {
			flush()
			v = Event{Severity: infer(x), Message: x, Line: line, ParseConfidence: ConfidenceLow}
			cur = &v
		}
	}
	flush()
	scanErr := s.Err()
	if closer != nil {
		if ce := closer(); ce != nil && scanErr == nil {
			scanErr = ce
		}
	}
	return out, warningFromScan(rel, line, scanErr)
}

func relPath(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		rel = path
	}
	return filepath.ToSlash(rel)
}

func newWarning(file string, cat WarningCategory, line, lineEnd int, msg string) Warning {
	return Warning{File: file, Category: cat, Line: line, LineEnd: lineEnd, Message: msg}
}

func warningFromScan(file string, lastLine int, err error) *Warning {
	if err == nil {
		return nil
	}
	if errors.Is(err, bufio.ErrTooLong) {
		w := newWarning(file, CategoryScanOverflow, lastLine+1, 0, err.Error())
		return &w
	}
	w := newWarning(file, CategoryScanError, lastLine, 0, err.Error())
	return &w
}
func detect(p string) string {
	n := canonicalLogName(filepath.Base(p))
	if isApacheAccessName(n) {
		return "apache-access"
	}
	if isApacheErrorName(n) {
		return "apache-error"
	}
	return "tomcat-java"
}

// isApacheAccessName matches conventional Apache/Tomcat access names, not every
// filename that merely contains the substring "access" (e.g. AccessControl.log).
func isApacheAccessName(n string) bool {
	if n == "access.log" || n == "access_log" {
		return true
	}
	if strings.Contains(n, "access_log") {
		return true
	}
	if strings.Contains(n, "_access_") || strings.Contains(n, "-access-") {
		return true
	}
	return strings.HasSuffix(n, "_access.log") || strings.HasSuffix(n, "-access.log")
}

func isApacheErrorName(n string) bool {
	return n == "error.log" || n == "error_log" || n == "ssl_error.log" || n == "ssl_error_log"
}

// rotationSuffix is one trailing logrotate/Tomcat-style decoration after the
// supported basename: .N, .YYYY-MM-DD[T| -HH[:MM[:SS]]], .YYYYMMDD[HH[MM[SS]]],
// optional .txt (AccessLogValve), or logrotate dateext -YYYYMMDD.
var rotationSuffix = regexp.MustCompile(`(?i)(?:\.(?:\d{4}-\d{2}-\d{2}(?:[T-]\d{2}(?:[:.]?\d{2}(?::\d{2})?)?)?|\d{8}(?:\d{2,6})?|\d+)(?:\.txt)?|-\d{8})$`)

func canonicalLogName(n string) string {
	n = strings.ToLower(n)
	n = strings.TrimSuffix(n, ".gz")
	return rotationSuffix.ReplaceAllString(n, "")
}

func isGzipName(path string) bool {
	return strings.HasSuffix(strings.ToLower(filepath.Base(path)), ".gz")
}

// openLogStream returns a streaming reader for path. Gzip is decoded only when
// the basename ends in .gz; the file is never extracted to disk. closer is
// gzip.Reader.Close (checksum) or nil for plaintext.
func openLogStream(f *os.File, path string) (io.Reader, func() error, error) {
	if !isGzipName(path) {
		return f, nil, nil
	}
	zr, err := gzip.NewReader(f)
	if err != nil {
		return nil, nil, err
	}
	return zr, zr.Close, nil
}
func instance(rel string) string {
	p := strings.Split(filepath.ToSlash(rel), "/")
	if len(p) > 1 {
		return p[0]
	}
	return "root"
}
func decorate(e *Event, src, inst, file string, line int) {
	e.SourceType = src
	e.Instance = inst
	e.File = filepath.ToSlash(file)
	e.Line = line
	e.Signature = signature(*e)
	e.Occurrences = 1
	e.LastSeen = e.Timestamp
	if e.ParseConfidence == "" {
		e.ParseConfidence = ConfidenceHigh
	}
}
func java(s string) (Event, bool) {
	m := javaStart.FindStringSubmatch(s)
	if m == nil {
		return Event{}, false
	}
	t, ok := flexTime(m[1])
	if !ok {
		return Event{}, false
	}
	msg := strings.TrimSpace(m[4])
	if msg == "" {
		msg = strings.TrimSpace(s)
	}
	return Event{Timestamp: t, HasTimestamp: true, Severity: severity(m[3]), Message: msg}, true
}
func apacheError(s string) (Event, bool) {
	m := httpdError.FindStringSubmatch(s)
	if m == nil {
		return Event{}, false
	}
	var t time.Time
	var e error
	for _, l := range []string{"Mon Jan 02 15:04:05.000000 2006", "Mon Jan 02 15:04:05 2006"} {
		t, e = time.Parse(l, m[1])
		if e == nil {
			break
		}
	}
	if e != nil {
		return Event{}, false
	}
	level := m[3]
	if level == "" {
		level = m[2]
	}
	addr := ""
	if cm := httpdClient.FindStringSubmatch(s); cm != nil {
		addr = canonicalClientAddr(cm[1])
	}
	return Event{Timestamp: t, HasTimestamp: true, Severity: severity(level), Message: strings.TrimSpace(m[4]), ClientAddr: addr}, true
}
func access(s string) (Event, bool) {
	m := httpdAccess.FindStringSubmatch(s)
	if m == nil {
		return Event{}, false
	}
	t, e := time.Parse("02/Jan/2006:15:04:05 -0700", m[2])
	if e != nil {
		return Event{}, false
	}
	code, _ := strconv.Atoi(m[4])
	sev := Info
	if code >= 500 {
		sev = Error
	} else if code >= 400 {
		sev = Warn
	}
	return Event{Timestamp: t, HasTimestamp: true, Severity: sev, Message: m[3] + " -> " + m[4] + " (" + m[5] + " bytes)", StatusCode: code, ClientAddr: canonicalClientAddr(m[1])}, true
}

func canonicalClientAddr(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if ip := net.ParseIP(strings.Trim(raw, "[]")); ip != nil {
		return ip.String()
	}
	if host, _, err := net.SplitHostPort(raw); err == nil {
		if ip := net.ParseIP(host); ip != nil {
			return ip.String()
		}
	}
	return ""
}
func flexTime(s string) (time.Time, bool) {
	s = strings.Replace(s, ",", ".", 1)
	for _, l := range []string{"2006-01-02 15:04:05.999999999Z07:00", "2006-01-02T15:04:05.999999999Z07:00", "2006-01-02 15:04:05.999999999", "2006-01-02T15:04:05.999999999", "2006-01-02 15:04:05", "2006-01-02T15:04:05"} {
		if t, e := time.Parse(l, s); e == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
func severity(s string) Severity {
	switch strings.ToUpper(s) {
	case "TRACE":
		return Trace
	case "DEBUG":
		return Debug
	case "INFO", "NOTICE":
		return Info
	case "WARN", "WARNING":
		return Warn
	case "ERROR", "SEVERE":
		return Error
	case "FATAL", "EMERG", "ALERT", "CRIT":
		return Fatal
	}
	return Unknown
}
func infer(s string) Severity {
	u := strings.ToUpper(s)
	for _, x := range []struct {
		k string
		v Severity
	}{{"FATAL", Fatal}, {"SEVERE", Error}, {"ERROR", Error}, {"EXCEPTION", Error}, {"WARN", Warn}, {"INFO", Info}, {"DEBUG", Debug}} {
		if strings.Contains(u, x.k) {
			return x.v
		}
	}
	return Unknown
}
func signature(e Event) string {
	first := strings.Split(e.Message, "\n")[0]
	norm := strings.ToLower(strings.TrimSpace(volatile.ReplaceAllString(first, "#")))
	sum := sha256.Sum256([]byte(e.SourceType + "|" + string(e.Severity) + "|" + norm))
	return hex.EncodeToString(sum[:8])
}
func hintOf(e Event) occHint {
	return occHint{ts: e.Timestamp, has: e.HasTimestamp, msg: e.Message, addr: e.ClientAddr}
}

func dedup(in []Event) []Event {
	type key struct{ inst, sig string }
	idx := map[key]int{}
	out := make([]Event, 0, len(in))
	for _, e := range in {
		k := key{e.Instance, e.Signature}
		h := hintOf(e)
		if i, ok := idx[k]; ok {
			out[i].Occurrences++
			if e.HasTimestamp && e.Timestamp.After(out[i].LastSeen) {
				out[i].LastSeen = e.Timestamp
			}
			if e.ParseConfidence == ConfidenceLow {
				out[i].ParseConfidence = ConfidenceLow
			}
			out[i].occHints = append(out[i].occHints, h)
			continue
		}
		e.occHints = []occHint{h}
		idx[k] = len(out)
		out = append(out, e)
	}
	return out
}
