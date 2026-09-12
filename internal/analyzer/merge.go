package analyzer

// maxExtraOccHints caps correlation hints kept after the lead row (NFR-009).
// FR-012 still sees later labeled IDs and client addresses on collapsed
// rows; hints without correlation value are dropped so a million repeats
// of the same signature do not retain a million message copies.
const maxExtraOccHints = 64

type mergeKey struct{ inst, sig string }

type eventMerger struct {
	idx    map[mergeKey]int
	out    []Event
	intern map[string]string
}

func newEventMerger() *eventMerger {
	return &eventMerger{
		idx:    map[mergeKey]int{},
		out:    []Event{},
		intern: map[string]string{},
	}
}

func (m *eventMerger) internStr(s string) string {
	if s == "" {
		return ""
	}
	if v, ok := m.intern[s]; ok {
		return v
	}
	m.intern[s] = s
	return s
}

func (m *eventMerger) add(e Event) {
	e.Instance = m.internStr(e.Instance)
	e.SourceType = m.internStr(e.SourceType)
	e.File = m.internStr(e.File)
	e.Signature = m.internStr(e.Signature)
	k := mergeKey{e.Instance, e.Signature}
	h := hintOf(e)
	if i, ok := m.idx[k]; ok {
		m.out[i].Occurrences++
		if e.HasTimestamp && e.Timestamp.After(m.out[i].LastSeen) {
			m.out[i].LastSeen = e.Timestamp
		}
		if e.ParseConfidence == ConfidenceLow {
			m.out[i].ParseConfidence = ConfidenceLow
			m.out[i].UnparsedOccurrences++
		}
		if len(m.out[i].occHints) < maxExtraOccHints && keepOccHint(m.out[i], h) {
			h.msg = m.internStr(h.msg)
			h.addr = m.internStr(h.addr)
			m.out[i].occHints = append(m.out[i].occHints, h)
		}
		return
	}
	e.occHints = nil
	if e.Occurrences < 1 {
		e.Occurrences = 1
	}
	m.idx[k] = len(m.out)
	m.out = append(m.out, e)
}

func keepOccHint(lead Event, h occHint) bool {
	if h.ts.Equal(lead.Timestamp) && h.has == lead.HasTimestamp && h.msg == lead.Message && h.addr == lead.ClientAddr {
		return false
	}
	if h.addr != "" && usableHeuristicAddr(h.addr) != "" {
		return true
	}
	if len(extractLabeledIDs(h.msg)) > 0 {
		return true
	}
	return len(messageIPs(h.msg)) > 0
}

func dedup(in []Event) []Event {
	m := newEventMerger()
	for _, e := range in {
		m.add(e)
	}
	return m.out
}
