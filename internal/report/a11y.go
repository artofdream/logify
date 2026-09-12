package report

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// WCAG 2.2 AA thresholds used by the NFR-013 checker. These are contrast
// ratios, not a browser accessibility engine.
const (
	wcagAAText     = 4.5
	wcagAANonText  = 3.0
	hexColorBits   = 16
	srgbLinearCut  = 0.04045
	srgbLinearDiv  = 12.92
	srgbGammaOff   = 0.055
	srgbGammaScale = 1.055
	srgbGammaExp   = 2.4
	lumRed         = 0.2126
	lumGreen       = 0.7152
	lumBlue        = 0.0722
	contrastPad    = 0.05
)

var cssVarRe = regexp.MustCompile(`--([a-z0-9-]+)\s*:\s*(#[0-9A-Fa-f]{3,8})\b`)
var controlTagRe = regexp.MustCompile(`(?is)<(input|select|textarea|button)\b([^>]*)>(?:([^<]*)</button>)?`)
var attrRe = regexp.MustCompile(`(?i)\b([a-zA-Z:-]+)\s*=\s*"([^"]*)"`)
var labelForRe = regexp.MustCompile(`(?is)<label\b[^>]*\bfor="([^"]+)"`)

// ContrastRatio returns the WCAG contrast ratio of two #RGB/#RRGGBB colors.
func ContrastRatio(hex1, hex2 string) (float64, error) {
	c1, err := parseHexColor(hex1)
	if err != nil {
		return 0, err
	}
	c2, err := parseHexColor(hex2)
	if err != nil {
		return 0, err
	}
	l1 := relativeLuminance(c1[0], c1[1], c1[2])
	l2 := relativeLuminance(c2[0], c2[1], c2[2])
	lighter, darker := l1, l2
	if l2 > l1 {
		lighter, darker = l2, l1
	}
	return (lighter + contrastPad) / (darker + contrastPad), nil
}

func parseHexColor(hex string) ([3]float64, error) {
	var out [3]float64
	h := strings.TrimSpace(hex)
	h = strings.TrimPrefix(h, "#")
	if len(h) == 3 {
		h = string([]byte{h[0], h[0], h[1], h[1], h[2], h[2]})
	}
	if len(h) == 8 {
		h = h[:6]
	}
	if len(h) != 6 {
		return out, fmt.Errorf("invalid hex color %q", hex)
	}
	for i := 0; i < 3; i++ {
		v, err := strconv.ParseUint(h[i*2:i*2+2], hexColorBits, 8)
		if err != nil {
			return out, fmt.Errorf("invalid hex color %q: %w", hex, err)
		}
		out[i] = float64(v) / 255
	}
	return out, nil
}

func relativeLuminance(r, g, b float64) float64 {
	return lumRed*linearize(r) + lumGreen*linearize(g) + lumBlue*linearize(b)
}

func linearize(c float64) float64 {
	if c <= srgbLinearCut {
		return c / srgbLinearDiv
	}
	return math.Pow((c+srgbGammaOff)/srgbGammaScale, srgbGammaExp)
}

func parseCSSVariables(css string) map[string]string {
	out := map[string]string{}
	for _, m := range cssVarRe.FindAllStringSubmatch(css, -1) {
		out["--"+m[1]] = strings.ToLower(m[2])
	}
	return out
}

type contrastPair struct {
	name string
	fg   string
	bg   string
	min  float64
}

func reportContrastPairs() []contrastPair {
	return []contrastPair{
		{name: "body text", fg: "--text", bg: "--bg", min: wcagAAText},
		{name: "panel text", fg: "--text", bg: "--panel", min: wcagAAText},
		{name: "muted on page", fg: "--muted", bg: "--bg", min: wcagAAText},
		{name: "muted on panel", fg: "--muted", bg: "--panel", min: wcagAAText},
		{name: "success feedback", fg: "--success", bg: "--panel", min: wcagAAText},
		{name: "danger text", fg: "--danger", bg: "--panel", min: wcagAAText},
		{name: "flag text", fg: "--flag", bg: "--panel", min: wcagAAText},
		{name: "flag badge", fg: "--flag", bg: "--flag-bg", min: wcagAAText},
		{name: "accent text", fg: "--accent", bg: "--panel", min: wcagAAText},
		{name: "control border", fg: "--control-border", bg: "--panel", min: wcagAANonText},
		{name: "header text", fg: "--text", bg: "#17274c", min: wcagAAText},
		{name: "sensitivity banner", fg: "--text", bg: "#3a1218", min: wcagAAText},
		{name: "overdue badge", fg: "--danger", bg: "#3a1218", min: wcagAAText},
	}
}

func checkContrastPairs(css string) []string {
	vars := parseCSSVariables(css)
	resolve := func(token string) (string, bool) {
		if strings.HasPrefix(token, "#") {
			return token, true
		}
		v, ok := vars[token]
		return v, ok
	}
	var fails []string
	for _, p := range reportContrastPairs() {
		fg, okFG := resolve(p.fg)
		bg, okBG := resolve(p.bg)
		if !okFG || !okBG {
			fails = append(fails, fmt.Sprintf("%s: missing token %s or %s", p.name, p.fg, p.bg))
			continue
		}
		ratio, err := ContrastRatio(fg, bg)
		if err != nil {
			fails = append(fails, fmt.Sprintf("%s: %v", p.name, err))
			continue
		}
		if ratio+1e-9 < p.min {
			fails = append(fails, fmt.Sprintf("%s: %.2f < %.1f (%s on %s)", p.name, ratio, p.min, fg, bg))
		}
	}
	return fails
}

func htmlBeforeScript(html string) string {
	i := strings.Index(strings.ToLower(html), "<script")
	if i < 0 {
		return html
	}
	return html[:i]
}

func parseAttrs(raw string) map[string]string {
	out := map[string]string{}
	for _, m := range attrRe.FindAllStringSubmatch(raw, -1) {
		out[strings.ToLower(m[1])] = m[2]
	}
	return out
}

func labeledIDs(html string) map[string]bool {
	out := map[string]bool{}
	for _, m := range labelForRe.FindAllStringSubmatch(html, -1) {
		out[m[1]] = true
	}
	return out
}

func checkLabeledControls(html string) []string {
	body := htmlBeforeScript(html)
	labels := labeledIDs(body)
	var fails []string
	for _, m := range controlTagRe.FindAllStringSubmatch(body, -1) {
		tag := strings.ToLower(m[1])
		attrs := parseAttrs(m[2])
		if attrs["hidden"] != "" && accessibleName(attrs, labels) {
			continue
		}
		if tag == "button" {
			text := strings.TrimSpace(m[3])
			if text == "" && !accessibleName(attrs, labels) {
				fails = append(fails, "button missing accessible name: "+strings.TrimSpace(m[0]))
			}
			continue
		}
		if !accessibleName(attrs, labels) {
			id := attrs["id"]
			if id == "" {
				id = tag
			}
			fails = append(fails, tag+"#"+id+" missing label for= or aria-label")
		}
	}
	return fails
}

func accessibleName(attrs map[string]string, labels map[string]bool) bool {
	if attrs["aria-label"] != "" || attrs["aria-labelledby"] != "" {
		return true
	}
	if id := attrs["id"]; id != "" && labels[id] {
		return true
	}
	return false
}
