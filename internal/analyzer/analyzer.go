package analyzer

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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
	r := Result{Root: abs, GeneratedAt: time.Now()}
	e = filepath.WalkDir(abs, func(p string, d os.DirEntry, we error) error {
		if we != nil {
			r.Warnings = append(r.Warnings, we.Error())
			return nil
		}
		if d.IsDir() || !looks(d.Name()) {
			return nil
		}
		r.FilesScanned++
		es, x := parseFile(abs, p)
		if x != nil {
			r.Warnings = append(r.Warnings, fmt.Sprintf("%s: %v", p, x))
			return nil
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
	return r, nil
}
func looks(n string) bool {
	n = strings.ToLower(n)
	return strings.HasSuffix(n, ".log") || strings.HasSuffix(n, ".out") || strings.Contains(n, "access_log") || strings.Contains(n, "error_log")
}
func parseFile(root, path string) ([]Event, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	rel, _ := filepath.Rel(root, path)
	src := detect(path)
	inst := instance(rel)
	s := bufio.NewScanner(f)
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
				v := Event{Severity: infer(x), Message: x}
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
			v = Event{Severity: infer(x), Message: x, Line: line}
			cur = &v
		}
	}
	flush()
	return out, s.Err()
}
func detect(p string) string {
	n := strings.ToLower(filepath.Base(p))
	if strings.Contains(n, "access") {
		return "apache-access"
	}
	if n == "error.log" || n == "error_log" || n == "ssl_error.log" || n == "ssl_error_log" {
		return "apache-error"
	}
	return "tomcat-java"
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
	e.EvidenceID = evidenceID(*e)
	e.Occurrences = 1
	if e.HasTimestamp {
		firstSeen := e.Timestamp
		lastSeen := e.Timestamp
		e.FirstSeen = &firstSeen
		e.LastSeen = &lastSeen
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
	return Event{Timestamp: t, HasTimestamp: true, Severity: severity(level), Message: strings.TrimSpace(m[4])}, true
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
	return Event{Timestamp: t, HasTimestamp: true, Severity: sev, Message: m[3] + " -> " + m[4] + " (" + m[5] + " bytes)", StatusCode: code}, true
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

func evidenceID(e Event) string {
	fields := []string{
		"logify-evidence-v1",
		e.Signature,
		e.Instance,
		filepath.ToSlash(e.File),
		strconv.Itoa(e.Line),
	}
	h := sha256.New()
	for i, field := range fields {
		if i > 0 {
			h.Write([]byte{0})
		}
		h.Write([]byte(field))
	}
	return "evidence-v1-" + hex.EncodeToString(h.Sum(nil))
}

func dedup(in []Event) []Event {
	type key struct{ inst, sig string }
	idx := map[key]int{}
	out := make([]Event, 0, len(in))
	for _, e := range in {
		k := key{e.Instance, e.Signature}
		if i, ok := idx[k]; ok {
			out[i].Occurrences++
			if e.FirstSeen != nil && (out[i].FirstSeen == nil || e.FirstSeen.Before(*out[i].FirstSeen)) {
				firstSeen := *e.FirstSeen
				out[i].FirstSeen = &firstSeen
			}
			if e.LastSeen != nil && (out[i].LastSeen == nil || e.LastSeen.After(*out[i].LastSeen)) {
				lastSeen := *e.LastSeen
				out[i].LastSeen = &lastSeen
			}
			continue
		}
		idx[k] = len(out)
		out = append(out, e)
	}
	return out
}
