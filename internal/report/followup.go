package report

// Follow-up export schema (ADR-0002). The browser store in followup.js is the
// runtime implementation; these constants document the contract for tests.

const (
	FollowUpSchema        = "logify-follow-up-v1"
	FollowUpSchemaVersion = 1
	FollowUpMaxBytes      = 5 << 20
	FollowUpMaxIssues     = 10000
	FollowUpMaxTitle      = 200
	FollowUpMaxTags       = 50
	FollowUpMaxTagLength  = 64
	FollowUpMaxNotes      = 8000
	FollowUpMaxOwner      = 200
)

// WorkflowState is the FR-020 closed set.
var WorkflowStates = []string{"open", "investigating", "blocked", "resolved", "dismissed"}
