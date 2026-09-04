package report

import (
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/artofdream/logify/internal/analyzer"
)

const (
	EvidenceIDPrefix = "evidence-v1-"
	IssueIDPrefix    = "issue-v1-"
	identityDomain   = "logify-evidence-v1"
)

// EvidenceID is a versioned, provenance-bound identity for one timeline row.
// FR-017 / NFR-017: independent of display order, message text, counts, and times.
func EvidenceID(e analyzer.Event) string {
	fields := []string{
		identityDomain,
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
	return EvidenceIDPrefix + hex.EncodeToString(h.Sum(nil))
}

// IssueID maps one evidence identity to one deterministic issue identity.
func IssueID(evidenceID string) string {
	return IssueIDPrefix + strings.TrimPrefix(evidenceID, EvidenceIDPrefix)
}
