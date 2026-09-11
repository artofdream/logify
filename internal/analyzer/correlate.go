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
	groups := make([]Correlation, 0)
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

func eventHints(e Event) []occHint {
	if len(e.occHints) > 0 {
		return e.occHints
	}
	return []occHint{{ts: e.Timestamp, has: e.HasTimestamp, msg: e.Message, addr: e.ClientAddr}}
}

func exactRequestIDGroups(events []Event) []Correlation {
	buckets := map[string][]int{}
	display := map[string]string{}
	for i, e := range events {
		for _, h := range eventHints(e) {
			for _, id := range extractLabeledIDs(h.msg) {
				key := id.canon + "\x00" + id.value
				buckets[key] = append(buckets[key], i)
				if display[key] == "" {
					display[key] = id.canon + "=" + id.display
				}
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
	// One group per in-window access/Tomcat pair and IP. Do not union pairs
	// across time or across distinct addresses (ADR-0008).
	type pair struct {
		a, t int
		ip   string
	}
	seen := map[pair]bool{}
	var pairs []pair
	for i, e := range events {
		if e.SourceType != "apache-access" {
			continue
		}
		for _, ah := range eventHints(e) {
			if !ah.has {
				continue
			}
			ip := usableHeuristicAddr(ah.addr)
			if ip == "" {
				continue
			}
			for j, te := range events {
				if te.SourceType != "tomcat-java" {
					continue
				}
				for _, th := range eventHints(te) {
					if !th.has || !withinHintWindow(ah, th) || !hintHasIP(th, ip) {
						continue
					}
					p := pair{a: i, t: j, ip: ip}
					if seen[p] {
						continue
					}
					seen[p] = true
					pairs = append(pairs, p)
				}
			}
		}
	}
	sort.SliceStable(pairs, func(i, j int) bool {
		if pairs[i].ip != pairs[j].ip {
			return pairs[i].ip < pairs[j].ip
		}
		if pairs[i].a != pairs[j].a {
			return pairs[i].a < pairs[j].a
		}
		return pairs[i].t < pairs[j].t
	})
	var out []Correlation
	for _, p := range pairs {
		members := []int{p.a, p.t}
		sort.Ints(members)
		out = append(out, Correlation{
			ID:         correlationID(RuleClientIPWindow, p.ip, members, events),
			Rule:       RuleClientIPWindow,
			Kind:       KindHeuristic,
			Confidence: ConfidenceLow,
			Evidence:   "heuristic client " + p.ip + " within 5s across apache-access and tomcat-java",
			Members:    members,
		})
	}
	return out
}

func hintHasIP(h occHint, ip string) bool {
	for _, got := range messageIPs(h.msg) {
		if got == ip {
			return true
		}
	}
	return false
}

func withinHintWindow(a, b occHint) bool {
	if !a.has || !b.has {
		return false
	}
	d := a.ts.Sub(b.ts)
	if d < 0 {
		d = -d
	}
	return d <= clientIPWindow
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
