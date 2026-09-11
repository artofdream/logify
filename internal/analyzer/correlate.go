package analyzer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"regexp"
	"sort"
	"strings"
	"time"
)

// clientIPWindow is the inclusive heuristic bound for client-ip-window (ADR-0008).
const clientIPWindow = 5 * time.Second

var labeledID = regexp.MustCompile(`(?i)(?:^|[^A-Za-z0-9-])((?:x-)?(?:correlation[-_]?id|request[-_]?id|trace[-_]?id|req(?:uest)?id|corr[-_]?id|traceid))\s*[:=]\s*["']?([A-Za-z0-9][A-Za-z0-9._:-]{7,127})`)
var traceparent = regexp.MustCompile(`(?i)\btraceparent\s*[:=]\s*["']?([0-9a-f]{2}-([0-9a-f]{32})-[0-9a-f]{16}-[0-9a-f]{2})`)
var ipv4Literal = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)

var discardedIDValues = map[string]bool{
	"undefined": true,
	"null":      true,
	"none":      true,
	"unknown":   true,
	"requestid": true,
	"traceid":   true,
}

// Correlate builds deterministic groups from already-normalized events (FR-012).
// It does not modify events. An empty result is a non-nil slice.
func Correlate(events []Event) []Correlation {
	if len(events) == 0 {
		return []Correlation{}
	}
	var groups []Correlation
	groups = append(groups, exactRequestIDGroups(events)...)
	groups = append(groups, heuristicClientIPGroups(events)...)
	sort.SliceStable(groups, func(i, j int) bool {
		if groups[i].Kind != groups[j].Kind {
			return groups[i].Kind == KindExact
		}
		if groups[i].Rule != groups[j].Rule {
			return groups[i].Rule < groups[j].Rule
		}
		if groups[i].Evidence != groups[j].Evidence {
			return groups[i].Evidence < groups[j].Evidence
		}
		return firstMember(groups[i]) < firstMember(groups[j])
	})
	return groups
}

func firstMember(c Correlation) int {
	if len(c.Members) == 0 {
		return 1 << 30
	}
	return c.Members[0]
}

func exactRequestIDGroups(events []Event) []Correlation {
	buckets := map[string][]int{}
	display := map[string]string{}
	for i, e := range events {
		for _, id := range extractLabeledIDs(e.Message) {
			key := id.canon + "\x00" + id.value
			buckets[key] = append(buckets[key], i)
			if display[key] == "" {
				display[key] = id.canon + "=" + id.display
			}
		}
	}
	keys := make([]string, 0, len(buckets))
	for k := range buckets {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var out []Correlation
	for _, key := range keys {
		members := uniqueSortedInts(buckets[key])
		if len(members) < 2 {
			continue
		}
		out = append(out, Correlation{
			ID:         correlationID(RuleSharedRequestID, key, members, events),
			Rule:       RuleSharedRequestID,
			Kind:       KindExact,
			Confidence: ConfidenceHigh,
			Evidence:   "exact identifier " + display[key],
			Members:    members,
		})
	}
	return out
}

func heuristicClientIPGroups(events []Event) []Correlation {
	accessByIP := map[string][]int{}
	tomcatByIP := map[string][]int{}
	for i, e := range events {
		if !e.HasTimestamp {
			continue
		}
		if e.SourceType == "apache-access" {
			if ip := usableHeuristicAddr(e.ClientAddr); ip != "" {
				accessByIP[ip] = append(accessByIP[ip], i)
			}
			continue
		}
		if e.SourceType == "tomcat-java" {
			for _, ip := range messageIPs(e.Message) {
				tomcatByIP[ip] = append(tomcatByIP[ip], i)
			}
		}
	}
	ips := make([]string, 0, len(accessByIP))
	for ip := range accessByIP {
		if len(tomcatByIP[ip]) == 0 {
			continue
		}
		ips = append(ips, ip)
	}
	sort.Strings(ips)
	u := newUF(len(events))
	linked := make([]bool, len(events))
	for _, ip := range ips {
		for _, a := range accessByIP[ip] {
			for _, t := range tomcatByIP[ip] {
				if withinWindow(events[a], events[t], clientIPWindow) {
					u.union(a, t)
					linked[a] = true
					linked[t] = true
				}
			}
		}
	}
	components := map[int][]int{}
	for i := range events {
		if !linked[i] {
			continue
		}
		r := u.find(i)
		components[r] = append(components[r], i)
	}
	roots := make([]int, 0, len(components))
	for r := range components {
		roots = append(roots, r)
	}
	sort.Ints(roots)
	var out []Correlation
	for _, r := range roots {
		members := append([]int(nil), components[r]...)
		sort.Ints(members)
		ip := sharedHeuristicIP(members, events)
		if ip == "" || !componentHasBothSides(members, events) {
			continue
		}
		if len(members) < 2 {
			continue
		}
		out = append(out, Correlation{
			ID:         correlationID(RuleClientIPWindow, ip, members, events),
			Rule:       RuleClientIPWindow,
			Kind:       KindHeuristic,
			Confidence: ConfidenceLow,
			Evidence:   "heuristic client " + ip + " within 5s across apache-access and tomcat-java",
			Members:    members,
		})
	}
	return out
}

func sharedHeuristicIP(members []int, events []Event) string {
	seen := map[string]int{}
	for _, i := range members {
		e := events[i]
		if e.SourceType == "apache-access" {
			if ip := usableHeuristicAddr(e.ClientAddr); ip != "" {
				seen[ip]++
			}
		}
		if e.SourceType == "tomcat-java" {
			for _, ip := range messageIPs(e.Message) {
				seen[ip]++
			}
		}
	}
	var best string
	for ip := range seen {
		if best == "" || ip < best {
			best = ip
		}
	}
	return best
}

func componentHasBothSides(members []int, events []Event) bool {
	var access, tomcat bool
	for _, i := range members {
		switch events[i].SourceType {
		case "apache-access":
			access = true
		case "tomcat-java":
			tomcat = true
		}
	}
	return access && tomcat
}

type labeledRef struct {
	canon   string
	value   string
	display string
}

func extractLabeledIDs(msg string) []labeledRef {
	if msg == "" {
		return nil
	}
	seen := map[string]bool{}
	var out []labeledRef
	add := func(canon, raw string) {
		if canon == "" || !usableIDValue(raw) {
			return
		}
		val := strings.ToLower(raw)
		key := canon + "\x00" + val
		if seen[key] {
			return
		}
		seen[key] = true
		out = append(out, labeledRef{canon: canon, value: val, display: raw})
	}
	for _, m := range labeledID.FindAllStringSubmatch(msg, -1) {
		add(canonLabel(m[1]), m[2])
	}
	for _, m := range traceparent.FindAllStringSubmatch(msg, -1) {
		add("trace-id", m[2])
	}
	return out
}

func canonLabel(raw string) string {
	s := strings.ToLower(raw)
	s = strings.ReplaceAll(s, "_", "")
	s = strings.ReplaceAll(s, "-", "")
	switch s {
	case "requestid", "reqid", "xrequestid":
		return "request-id"
	case "correlationid", "corrid", "xcorrelationid":
		return "correlation-id"
	case "traceid", "xtraceid", "traceparent":
		return "trace-id"
	default:
		return ""
	}
}

func usableIDValue(v string) bool {
	if len(v) < 8 || len(v) > 128 {
		return false
	}
	return !discardedIDValues[strings.ToLower(v)]
}

func uniqueSortedInts(in []int) []int {
	seen := map[int]bool{}
	var out []int
	for _, i := range in {
		if seen[i] {
			continue
		}
		seen[i] = true
		out = append(out, i)
	}
	sort.Ints(out)
	return out
}

func messageIPs(msg string) []string {
	seen := map[string]bool{}
	var out []string
	for _, raw := range ipv4Literal.FindAllString(msg, -1) {
		ip := usableHeuristicAddr(canonicalClientAddr(raw))
		if ip == "" || seen[ip] {
			continue
		}
		seen[ip] = true
		out = append(out, ip)
	}
	sort.Strings(out)
	return out
}

func usableHeuristicAddr(addr string) string {
	ip := net.ParseIP(addr)
	if ip == nil || ip.IsLoopback() || ip.IsUnspecified() || ip.IsMulticast() || ip.IsLinkLocalUnicast() {
		return ""
	}
	return ip.String()
}

func withinWindow(a, b Event, window time.Duration) bool {
	if !a.HasTimestamp || !b.HasTimestamp {
		return false
	}
	d := a.Timestamp.Sub(b.Timestamp)
	if d < 0 {
		d = -d
	}
	return d <= window
}

func correlationID(rule, key string, members []int, events []Event) string {
	h := sha256.New()
	_, _ = fmt.Fprintf(h, "logify-corr-v1\x00%s\x00%s", rule, key)
	for _, i := range members {
		e := events[i]
		_, _ = fmt.Fprintf(h, "\x00%s\x00%s\x00%s\x00%d", e.Signature, e.Instance, e.File, e.Line)
	}
	return "corr-v1-" + hex.EncodeToString(h.Sum(nil))[:16]
}

type uf struct{ p []int }

func newUF(n int) *uf {
	p := make([]int, n)
	for i := range p {
		p[i] = i
	}
	return &uf{p: p}
}

func (u *uf) find(x int) int {
	for u.p[x] != x {
		u.p[x] = u.p[u.p[x]]
		x = u.p[x]
	}
	return x
}

func (u *uf) union(a, b int) {
	ra, rb := u.find(a), u.find(b)
	if ra != rb {
		u.p[rb] = ra
	}
}
