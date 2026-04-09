// Package noxy provides a Go SDK for the Noxy Decision Layer.
//
// Send encrypted, actionable decision payloads (tool proposals, approvals, next-step hints) to
// registered agent devices over gRPC.
package noxy

import (
	"context"
	"strings"

	"github.com/noxy-network/go-sdk/client"
	"github.com/noxy-network/go-sdk/internal/config"
	"github.com/noxy-network/go-sdk/internal/decisionoutcome"
	"github.com/noxy-network/go-sdk/internal/types"
)

// NoxyConfig holds configuration for the Noxy SDK client.
type NoxyConfig struct {
	// Endpoint is the Noxy relay gRPC endpoint (e.g. "https://relay.noxy.network:443").
	Endpoint string
	// AuthToken is the Bearer token for relay authentication.
	AuthToken string
	// DecisionTTLSeconds is the time-to-live for routed decisions in seconds.
	DecisionTTLSeconds uint32
}

// InitNoxyAgentClient initializes the Decision Layer SDK client.
func InitNoxyAgentClient(ctx context.Context, cfg NoxyConfig) (*client.NoxyAgentClient, error) {
	endpoint := strings.TrimPrefix(cfg.Endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimSuffix(endpoint, "/")

	internalCfg := &config.NoxyConfig{
		Endpoint:           endpoint,
		AuthToken:          cfg.AuthToken,
		DecisionTTLSeconds: cfg.DecisionTTLSeconds,
	}
	return client.NewNoxyAgentClient(ctx, internalCfg)
}

// Re-export types for convenience.
type (
	NoxyAgentClient                = client.NoxyAgentClient
	NoxyIdentityAddress            = types.NoxyIdentityAddress
	NoxyDeliveryOutcome            = types.NoxyDeliveryOutcome
	NoxyDeliveryStatus             = types.NoxyDeliveryStatus
	NoxyHumanDecisionOutcome       = types.NoxyHumanDecisionOutcome
	NoxyGetDecisionOutcomeResponse = types.NoxyGetDecisionOutcomeResponse
	NoxyGetQuotaResponse           = types.NoxyGetQuotaResponse
	NoxyQuotaStatus                = types.NoxyQuotaStatus
	WaitForDecisionOutcomeOptions  = decisionoutcome.WaitForDecisionOutcomeOptions
	SendDecisionAndWaitOptions     = decisionoutcome.SendDecisionAndWaitOptions
)

// NoxyHumanDecisionOutcome values (match proto DecisionOutcome).
const (
	NoxyHumanDecisionOutcomePending  = types.NoxyHumanDecisionOutcomePending
	NoxyHumanDecisionOutcomeApproved = types.NoxyHumanDecisionOutcomeApproved
	NoxyHumanDecisionOutcomeRejected = types.NoxyHumanDecisionOutcomeRejected
	NoxyHumanDecisionOutcomeExpired  = types.NoxyHumanDecisionOutcomeExpired
)

// Re-export decision outcome helpers.
var (
	ErrWaitForDecisionOutcomeTimeout = decisionoutcome.ErrWaitForDecisionOutcomeTimeout
	ErrSendDecisionNoDecisionID      = decisionoutcome.ErrSendDecisionNoDecisionID
)

// IsTerminalHumanOutcome reports whether the human has finalized (approved, rejected, or expired).
func IsTerminalHumanOutcome(o types.NoxyHumanDecisionOutcome) bool {
	return decisionoutcome.IsTerminalHumanOutcome(o)
}
