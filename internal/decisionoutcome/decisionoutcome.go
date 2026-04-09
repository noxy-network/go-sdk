// Package decisionoutcome holds polling options and helpers for human-in-the-loop decisions.
package decisionoutcome

import (
	"errors"

	"github.com/noxy-network/go-sdk/internal/types"
)

// Errors returned by WaitForDecisionOutcome / SendDecisionAndWaitForOutcome.
var (
	ErrWaitForDecisionOutcomeTimeout = errors.New("wait for decision outcome exceeded max_wait_ms")
	ErrSendDecisionNoDecisionID      = errors.New("send decision returned no decision_id to poll; check delivery statuses")
)

// SendDecisionAndWaitOptions configures polling for SendDecisionAndWaitForOutcome (no decision_id / identity_id).
type SendDecisionAndWaitOptions struct {
	InitialPollIntervalMs *uint64
	MaxPollIntervalMs     *uint64
	MaxWaitMs             *uint64
	BackoffMultiplier     *float64
}

// WaitForDecisionOutcomeOptions configures exponential-backoff polling for GetDecisionOutcome.
type WaitForDecisionOutcomeOptions struct {
	DecisionID            string
	IdentityID            string
	InitialPollIntervalMs *uint64
	MaxPollIntervalMs     *uint64
	MaxWaitMs             *uint64
	BackoffMultiplier     *float64
}

// IsTerminalHumanOutcome reports whether the human has finalized (approved, rejected, or expired).
func IsTerminalHumanOutcome(o types.NoxyHumanDecisionOutcome) bool {
	switch o {
	case types.NoxyHumanDecisionOutcomeApproved,
		types.NoxyHumanDecisionOutcomeRejected,
		types.NoxyHumanDecisionOutcomeExpired:
		return true
	default:
		return false
	}
}
