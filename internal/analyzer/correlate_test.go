package analyzer

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCorrelateFixtureExactVersusHeuristic(t *testing.T) {
	// FR-012 AC1/AC4/AC5: labeled IDs are exact/high; IP+window is heuristic/low.
	r, err := Analyze(filepath.Join("..", "..", "testdata", "correlate"), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if r.Correlations == nil {
		t.Fatal("correlations slice is nil")
	}
	if len(r.Correlations) != 2 {
		t.Fatalf("groups=%d want 2: %+v", len(r.Correlations), r.Correlations)
	}
	var exact, heuristic *Correlation
	for i := range r.Correlations {
		c := &r.Correlations[i]
		switch c.Rule {
		case RuleSharedRequestID:
			exact = c
		case RuleClientIPWindow:
			heuristic = c
		default:
			t.Fatalf("unexpected rule %q", c.Rule)
		}
	}
	if exact == nil || heuristic == nil {
		t.Fatalf("missing rule: exact=%v heuristic=%v", exact, heuristic)
	}
	if exact.Kind != KindExact || exact.Confidence != ConfidenceHigh {
		t.Fatalf("exact labels=%+v", exact)
	}
	if !strings.Contains(exact.Evidence, "request-id=abc-req-001") {
		t.Fatalf("exact evidence=%q", exact.Evidence)
	}
	if heuristic.Kind != KindHeuristic || heuristic.Confidence != ConfidenceLow {
		t.Fatalf("heuristic labels=%+v", heuristic)
	}
	if !strings.Contains(heuristic.Evidence, "10.2.3.4") || !strings.Contains(heuristic.Evidence, "heuristic") {
		t.Fatalf("heuristic evidence=%q", heuristic.Evidence)
	}
	exactMsgs := memberMessages(r.Events, exact.Members)
	assertContains(t, exactMsgs, "checkout failed")
	assertContains(t, exactMsgs, "retry scheduled")
	assertContains(t, exactMsgs, "GET /api?requestId=abc-req-001")
	assertContains(t, exactMsgs, "backend connection failed")
	if len(exact.Members) != 4 {
		t.Fatalf("exact members=%d msgs=%v", len(exact.Members), exactMsgs)
	}
	heuristicMsgs := memberMessages(r.Events, heuristic.Members)
	assertContains(t, heuristicMsgs, "GET /inventory")
	assertContains(t, heuristicMsgs, "inventory timeout")
	if len(heuristic.Members) != 2 {
		t.Fatalf("heuristic members=%d msgs=%v", len(heuristic.Members), heuristicMsgs)
	}

	ungrouped := ungroupedMessages(r)
	for _, needle := range []string{
		"zzz-alone",
		"heartbeat",
		"stale after window",
		"unlabeled number",
		"GET /other",
		"GET /local",
		"local only",
		"leftover without ids",
	} {
		assertContains(t, ungrouped, needle)
	}
	if !strings.Contains(strings.Join(exactMsgs, "\n"), "Caused by:") {
		t.Fatal("exact checkout event lost its intra-event cause chain")
	}
	for _, e := range r.Events {
		if e.File == "" || e.Line <= 0 {
			t.Fatalf("event lost provenance: %+v", e)
		}
	}
	again, err := Analyze(filepath.Join("..", "..", "testdata", "correlate"), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if again.Correlations[0].ID != r.Correlations[0].ID || again.Correlations[1].ID != r.Correlations[1].ID {
		t.Fatalf("correlation ids are not deterministic:\n%q\n%q", r.Correlations[0].ID, again.Correlations[0].ID)
	}
	if !strings.HasPrefix(exact.ID, "corr-v1-") || !strings.HasPrefix(heuristic.ID, "corr-v1-") {
		t.Fatalf("ids=%q %q", exact.ID, heuristic.ID)
	}
}

func TestCorrelateCaseFixtureHasNoGroups(t *testing.T) {
	// testdata/case uses unlabeled Request N and loopback 127.0.0.1 (ADR-0008).
	r, err := Analyze(filepath.Join("..", "..", "testdata", "case"), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if r.Correlations == nil {
		t.Fatal("nil correlations")
	}
	if len(r.Correlations) != 0 {
		t.Fatalf("case fixture groups=%+v", r.Correlations)
	}
}

func TestCorrelatePrefersNoLinkWithoutEvidence(t *testing.T) {
	now := time.Date(2026, 9, 3, 10, 0, 0, 0, time.UTC)
	events := []Event{
		{HasTimestamp: true, Timestamp: now, SourceType: "apache-access", Message: "GET /a -> 200", ClientAddr: "10.1.1.1", Signature: "a", Instance: "h", File: "a.log", Line: 1},
		{HasTimestamp: true, Timestamp: now.Add(time.Second), SourceType: "apache-access", Message: "GET /b -> 200", ClientAddr: "10.1.1.1", Signature: "b", Instance: "h", File: "a.log", Line: 2},
		{HasTimestamp: true, Timestamp: now, SourceType: "tomcat-java", Message: "Request 12345678 failed", Signature: "c", Instance: "t", File: "c.out", Line: 1},
		{HasTimestamp: true, Timestamp: now, SourceType: "tomcat-java", Message: "requestId=null ignored", Signature: "d", Instance: "t", File: "c.out", Line: 2},
		{HasTimestamp: true, Timestamp: now, SourceType: "tomcat-java", Message: "id=550e8400-e29b-41d4-a716-446655440000 unlabeled uuid", Signature: "e", Instance: "t", File: "c.out", Line: 3},
		{HasTimestamp: true, Timestamp: now.Add(time.Second), SourceType: "tomcat-java", Message: "loop client 127.0.0.1", Signature: "f", Instance: "t", File: "c.out", Line: 4},
		{HasTimestamp: true, Timestamp: now.Add(time.Second), SourceType: "apache-access", Message: "GET /local -> 200", ClientAddr: "127.0.0.1", Signature: "g", Instance: "h", File: "a.log", Line: 3},
	}
	got := Correlate(events)
	if len(got) != 0 {
		t.Fatalf("expected no groups, got %+v", got)
	}
}

func TestCorrelateTraceparentAndRequestID(t *testing.T) {
	now := time.Date(2026, 9, 3, 10, 0, 0, 0, time.UTC)
	events := []Event{
		{HasTimestamp: true, Timestamp: now, SourceType: "tomcat-java", Message: "traceparent=00-0af7651916cd43dd8448eb211c80319c-00f067aa0ba902b7-01 start", Signature: "a", Instance: "t", File: "a.out", Line: 1},
		{HasTimestamp: true, Timestamp: now.Add(time.Second), SourceType: "apache-access", Message: "GET /x X-Trace-Id=0af7651916cd43dd8448eb211c80319c -> 500", Signature: "b", Instance: "h", File: "a.log", Line: 1},
	}
	got := Correlate(events)
	if len(got) != 1 || got[0].Rule != RuleSharedRequestID || got[0].Kind != KindExact {
		t.Fatalf("got %+v", got)
	}
	if !strings.Contains(got[0].Evidence, "trace-id=") {
		t.Fatalf("evidence=%q", got[0].Evidence)
	}
	if len(got[0].Members) != 2 {
		t.Fatalf("members=%v", got[0].Members)
	}
}

func TestCorrelateDoesNotHideEvents(t *testing.T) {
	r, err := Analyze(filepath.Join("..", "..", "testdata", "correlate"), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Events) != 14 {
		t.Fatalf("events=%d want 14 (each record stays inspectable)", len(r.Events))
	}
}

func memberMessages(events []Event, members []int) []string {
	out := make([]string, 0, len(members))
	for _, i := range members {
		out = append(out, events[i].Message)
	}
	return out
}

func ungroupedMessages(r Result) []string {
	in := map[int]bool{}
	for _, c := range r.Correlations {
		for _, i := range c.Members {
			in[i] = true
		}
	}
	var out []string
	for i, e := range r.Events {
		if !in[i] {
			out = append(out, e.Message)
		}
	}
	return out
}

func assertContains(t *testing.T, msgs []string, needle string) {
	t.Helper()
	for _, m := range msgs {
		if strings.Contains(m, needle) {
			return
		}
	}
	t.Fatalf("missing %q in %#v", needle, msgs)
}
