package harness_test

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// NFR-015 / NFR-022 / NFR-023 / NFR-024 / NFR-025 / NFR-026 / NFR-027 / NFR-028:
// mechanical outer-harness probes. Status words are claims; these tests fail
// when critical docs links, ledger mirrors, Implemented probes, or open
// work-item ownership fields are missing or contradictory.

var (
	reqHeading  = regexp.MustCompile(`^### ((?:FR|NFR)-\d{3})\b`)
	reqStatus   = regexp.MustCompile(`(?m)^\- \*\*Status:\*\* (\w+)`)
	mdLink      = regexp.MustCompile(`!?\[(?:[^\]]*)\]\(([^)]+)\)`)
	reqID       = regexp.MustCompile(`\b((?:FR|NFR)-\d{3})\b`)
	mirrorRow   = regexp.MustCompile(`^\|\s*((?:FR|NFR)-\d{3})\s*\|\s*(\w+)\s*\|`)
	layerRow    = regexp.MustCompile(`^\|\s*(Guides|Sensors|Loop|Memory|Permissions|Observability)\s*\|`)
	frontID     = regexp.MustCompile(`(?m)^id:\s*(\S+)`)
	frontType   = regexp.MustCompile(`(?m)^type:\s*(\S+)`)
	frontStatus = regexp.MustCompile(`(?m)^status:\s*(\S+)`)
	frontOwner  = regexp.MustCompile(`(?m)^owner:\s*(.+)$`)
	frontLease  = regexp.MustCompile(`(?m)^lease_expires:`)
	frontReqs   = regexp.MustCompile(`(?m)^requirements:\s*\[([^\]]*)\]`)
	nfr028FM    = regexp.MustCompile(`(?m)^nfr028_status:\s*(\w+)`)
)

func TestNFR028DocsLinksAndLedger(t *testing.T) {
	root := repoRoot(t)
	reqs := parseRequirements(t, root)
	if _, ok := reqs["NFR-028"]; !ok {
		t.Fatal("NFR-028 missing from requirements")
	}

	var broken []string
	for _, rel := range criticalDocs(t, root) {
		broken = append(broken, checkDocLinks(t, root, rel)...)
	}
	if len(broken) > 0 {
		t.Fatalf("broken critical docs links:\n  %s", strings.Join(broken, "\n  "))
	}

	adoption := readRepoFile(t, root, "docs/framework-adoption.md")
	gotNFR028 := nfr028FM.FindStringSubmatch(adoption)
	if gotNFR028 == nil {
		t.Fatal("docs/framework-adoption.md missing nfr028_status frontmatter")
	}
	if gotNFR028[1] != reqs["NFR-028"] {
		t.Fatalf("adoption nfr028_status=%q contradicts requirements status=%q", gotNFR028[1], reqs["NFR-028"])
	}

	layers := map[string]bool{}
	for _, line := range strings.Split(adoption, "\n") {
		if m := layerRow.FindStringSubmatch(line); m != nil {
			cols := splitTableRow(line)
			if len(cols) < 5 {
				t.Fatalf("ledger row for %s is incomplete: %s", m[1], line)
			}
			status := cols[2]
			switch status {
			case "Adopted", "Adopted on `main`", "Adopted locally", "Partial", "Unknown", "Proposed":
			default:
				t.Fatalf("ledger %s has unknown status %q", m[1], status)
			}
			if strings.TrimSpace(cols[3]) == "" || strings.TrimSpace(cols[4]) == "" {
				t.Fatalf("ledger %s is missing probe/evidence or gap text", m[1])
			}
			layers[m[1]] = true
		}
	}
	for _, name := range []string{"Guides", "Sensors", "Loop", "Memory", "Permissions", "Observability"} {
		if !layers[name] {
			t.Fatalf("adoption ledger missing layer %s", name)
		}
	}

	mirrored := map[string]string{}
	inMirror := false
	for _, line := range strings.Split(adoption, "\n") {
		if strings.Contains(line, "Requirement status mirror") {
			inMirror = true
			continue
		}
		if inMirror && strings.HasPrefix(line, "## ") && !strings.Contains(line, "Requirement status mirror") {
			break
		}
		if m := mirrorRow.FindStringSubmatch(line); m != nil {
			mirrored[m[1]] = m[2]
		}
	}
	if mirrored["NFR-028"] != reqs["NFR-028"] {
		t.Fatalf("status mirror NFR-028=%q want %q", mirrored["NFR-028"], reqs["NFR-028"])
	}
	for id, status := range reqs {
		if status != "Partial" && status != "Proposed" && status != "Deferred" {
			continue
		}
		got, ok := mirrored[id]
		if !ok {
			t.Errorf("Partial/Proposed/Deferred %s %s is missing from the adoption status mirror", id, status)
			continue
		}
		if got != status {
			t.Errorf("status mirror %s=%q contradicts requirements %q", id, got, status)
		}
	}
	for id, status := range mirrored {
		want, ok := reqs[id]
		if !ok {
			t.Errorf("status mirror lists unknown %s", id)
			continue
		}
		if status != want {
			t.Errorf("status mirror %s=%q contradicts requirements %q", id, status, want)
		}
	}
}

func TestNFR028ImplementedClaimsHaveNamedProbes(t *testing.T) {
	root := repoRoot(t)
	reqs := parseRequirements(t, root)
	hits := scanProbeMentions(t, root)
	var missing []string
	for id, status := range reqs {
		if status != "Implemented" {
			continue
		}
		if len(hits[id]) == 0 {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("Implemented requirements without a named test/CI/JS probe (Unknown discipline): %s", strings.Join(missing, ", "))
	}
}

func TestNFR028OpenWorkItemsHaveOwnershipFields(t *testing.T) {
	root := repoRoot(t)
	dir := filepath.Join(root, "docs", "collaboration", "work-items")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	open := map[string]bool{"proposed": true, "active": true, "blocked": true, "review": true}
	var problems []string
	for _, e := range entries {
		if e.IsDir() || e.Name() == "TEMPLATE.md" || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		body := readRepoFile(t, root, filepath.ToSlash(filepath.Join("docs", "collaboration", "work-items", e.Name())))
		st := firstSub(frontStatus, body)
		if !open[st] {
			continue
		}
		id := firstSub(frontID, body)
		if id == "" || id+".md" != e.Name() && strings.ToLower(id)+".md" != strings.ToLower(e.Name()) {
			// id is WI-...; filename is WI-....md
			if id+".md" != e.Name() {
				problems = append(problems, e.Name()+": id "+id+" does not match filename")
			}
		}
		if firstSub(frontType, body) != "work-item" {
			problems = append(problems, e.Name()+": type must be work-item")
		}
		owner := strings.TrimSpace(firstSub(frontOwner, body))
		if owner == "" || strings.Contains(owner, "|") {
			problems = append(problems, e.Name()+": owner is missing or still a template placeholder")
		}
		if !frontLease.MatchString(body) {
			problems = append(problems, e.Name()+": lease_expires field is missing")
		}
		scope, hasScope := frontmatterList(body, "scope")
		if !hasScope || len(scope) == 0 {
			problems = append(problems, e.Name()+": scope must list at least one path")
		}
		reqMatch := frontReqs.FindStringSubmatch(body)
		if reqMatch == nil {
			problems = append(problems, e.Name()+": requirements field is missing")
		} else if (st == "active" || st == "blocked") && !reqID.MatchString(reqMatch[1]) {
			problems = append(problems, e.Name()+": active/blocked work items must list at least one FR/NFR id")
		}
	}
	owners := readRepoFile(t, root, ".github/CODEOWNERS")
	if !strings.Contains(owners, "@artofdream") {
		problems = append(problems, ".github/CODEOWNERS must name @artofdream")
	}
	if len(problems) > 0 {
		t.Fatalf("permissions/ownership probe failed:\n  %s", strings.Join(problems, "\n  "))
	}
}

func TestFrontmatterListReadsOnlyNamedKey(t *testing.T) {
	// NFR-028: scope: [] must not pass because depends_on lists an id.
	body := "---\n" +
		"id: WI-demo\n" +
		"scope: []\n" +
		"depends_on:\n" +
		"  - WI-other\n" +
		"supersedes:\n" +
		"  - WI-old\n" +
		"---\n"
	scope, has := frontmatterList(body, "scope")
	if !has || len(scope) != 0 {
		t.Fatalf("empty scope: has=%v items=%v", has, scope)
	}
	deps, hasDeps := frontmatterList(body, "depends_on")
	if !hasDeps || len(deps) != 1 || deps[0] != "WI-other" {
		t.Fatalf("depends_on=%v has=%v", deps, hasDeps)
	}
	block := "---\n" +
		"scope:\n" +
		"  - docs/a.md\n" +
		"  - internal/b.go\n" +
		"depends_on:\n" +
		"  - WI-other\n" +
		"---\n"
	paths, ok := frontmatterList(block, "scope")
	if !ok || len(paths) != 2 || paths[0] != "docs/a.md" || paths[1] != "internal/b.go" {
		t.Fatalf("block scope=%v ok=%v", paths, ok)
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Clean(filepath.Join(wd, "..", ".."))
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("repo root from %s: %v", wd, err)
	}
	return root
}

func parseRequirements(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	for _, rel := range []string{
		"docs/requirements/functional-requirements.md",
		"docs/requirements/non-functional-requirements.md",
	} {
		body := readRepoFile(t, root, rel)
		var cur string
		for _, line := range strings.Split(body, "\n") {
			if m := reqHeading.FindStringSubmatch(line); m != nil {
				cur = m[1]
				continue
			}
			if cur == "" {
				continue
			}
			if m := reqStatus.FindStringSubmatch(line); m != nil {
				out[cur] = m[1]
				cur = ""
			}
		}
	}
	if len(out) < 20 {
		t.Fatalf("parsed only %d requirements", len(out))
	}
	return out
}

func criticalDocs(t *testing.T, root string) []string {
	t.Helper()
	var out []string
	for _, rel := range []string{"README.md", "AGENTS.md", "CLAUDE.md"} {
		out = append(out, rel)
	}
	err := filepath.WalkDir(filepath.Join(root, "docs"), func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		out = append(out, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func checkDocLinks(t *testing.T, root, rel string) []string {
	t.Helper()
	body := stripCode(readRepoFile(t, root, rel))
	dir := filepath.Dir(filepath.Join(root, filepath.FromSlash(rel)))
	var broken []string
	for _, m := range mdLink.FindAllStringSubmatch(body, -1) {
		target := strings.TrimSpace(m[1])
		if i := strings.IndexAny(target, " \t"); i >= 0 {
			target = target[:i]
		}
		target = strings.Trim(target, `"'`)
		if hash := strings.Index(target, "#"); hash >= 0 {
			target = target[:hash]
		}
		if target == "" || strings.Contains(target, "://") || strings.HasPrefix(target, "mailto:") {
			continue
		}
		var dest string
		if strings.HasPrefix(target, "/") {
			dest = filepath.Join(root, filepath.FromSlash(strings.TrimPrefix(target, "/")))
		} else {
			dest = filepath.Join(dir, filepath.FromSlash(target))
		}
		clean, err := filepath.Abs(dest)
		if err != nil {
			broken = append(broken, rel+" -> "+target+" (abs: "+err.Error()+")")
			continue
		}
		absRoot, err := filepath.Abs(root)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(clean, absRoot+string(os.PathSeparator)) && clean != absRoot {
			broken = append(broken, rel+" -> "+target+" escapes the repository")
			continue
		}
		if _, err := os.Stat(clean); err != nil {
			broken = append(broken, rel+" -> "+target)
		}
	}
	return broken
}

func scanProbeMentions(t *testing.T, root string) map[string][]string {
	t.Helper()
	hits := map[string][]string{}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		base := d.Name()
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		slash := filepath.ToSlash(rel)
		keep := strings.HasSuffix(base, "_test.go") ||
			strings.HasSuffix(base, "_test.js") ||
			strings.Contains(base, "probe.js") ||
			(strings.HasPrefix(slash, ".github/workflows/") && strings.HasSuffix(base, ".yml"))
		if !keep {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, id := range reqID.FindAllString(string(data), -1) {
			hits[id] = append(hits[id], slash)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return hits
}

func stripCode(s string) string {
	var b strings.Builder
	fence := false
	sc := bufio.NewScanner(strings.NewReader(s))
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		line := sc.Text()
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "```") {
			fence = !fence
			continue
		}
		if fence {
			continue
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String()
}

func splitTableRow(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")
	parts := strings.Split(line, "|")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, strings.TrimSpace(p))
	}
	return out
}

func frontmatterList(body, key string) ([]string, bool) {
	fm, ok := yamlFrontmatter(body)
	if !ok {
		return nil, false
	}
	lines := strings.Split(fm, "\n")
	prefix := key + ":"
	for i, line := range lines {
		line = strings.TrimRight(line, "\r")
		if line != prefix && !strings.HasPrefix(line, prefix+" ") {
			continue
		}
		rest := strings.TrimSpace(strings.TrimPrefix(line, prefix))
		if strings.HasPrefix(rest, "[") && strings.HasSuffix(rest, "]") {
			inner := strings.TrimSpace(rest[1 : len(rest)-1])
			if inner == "" {
				return nil, true
			}
			var out []string
			for _, p := range strings.Split(inner, ",") {
				p = strings.TrimSpace(strings.Trim(p, `"'`))
				if p != "" {
					out = append(out, p)
				}
			}
			return out, true
		}
		var out []string
		for _, nxt := range lines[i+1:] {
			trim := strings.TrimRight(nxt, "\r")
			if strings.HasPrefix(trim, "  - ") {
				item := strings.TrimSpace(strings.TrimPrefix(trim, "  - "))
				if item != "" && item != "[]" {
					out = append(out, item)
				}
				continue
			}
			if trim == "" || strings.HasPrefix(trim, " ") || strings.HasPrefix(trim, "\t") {
				continue
			}
			break
		}
		return out, true
	}
	return nil, false
}

func yamlFrontmatter(body string) (string, bool) {
	if !strings.HasPrefix(body, "---") {
		return "", false
	}
	rest := strings.TrimPrefix(body, "---")
	rest = strings.TrimPrefix(rest, "\n")
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return "", false
	}
	return rest[:end], true
}

func firstSub(re *regexp.Regexp, s string) string {
	m := re.FindStringSubmatch(s)
	if m == nil {
		return ""
	}
	if len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return strings.TrimSpace(m[0])
}

func readRepoFile(t *testing.T, root, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
