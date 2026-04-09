package client

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/noxy-network/go-sdk/grpc/noxy"
	"github.com/noxy-network/go-sdk/internal/config"
	"github.com/noxy-network/go-sdk/internal/decisionoutcome"
	"github.com/noxy-network/go-sdk/internal/kyber"
	"github.com/noxy-network/go-sdk/internal/services"
	"github.com/noxy-network/go-sdk/internal/transport"
	"github.com/noxy-network/go-sdk/internal/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// NoxyAgentClient is the main SDK client for the Noxy Decision Layer.
type NoxyAgentClient struct {
	config   *config.NoxyConfig
	conn     *grpc.ClientConn
	grpc     noxy.AgentServiceClient
	identity *services.IdentityService
	decision *services.DecisionService
	quota    *services.QuotaService
}

// NewNoxyAgentClient creates and initializes a NoxyAgentClient.
func NewNoxyAgentClient(ctx context.Context, cfg *config.NoxyConfig) (*NoxyAgentClient, error) {
	grpcClient, conn, err := transport.NewAgentServiceClient(ctx, cfg.Endpoint)
	if err != nil {
		return nil, err
	}

	kyberProvider := kyber.NewKyberProvider()
	return &NoxyAgentClient{
		config:   cfg,
		conn:     conn,
		grpc:     grpcClient,
		identity: services.NewIdentityService(),
		decision: services.NewDecisionService(kyberProvider),
		quota:    services.NewQuotaService(),
	}, nil
}

// SendDecision routes an encrypted actionable decision to all devices for the identity.
// One client-generated decision UUID is used for the entire fan-out batch (same id on each RouteDecision).
func (c *NoxyAgentClient) SendDecision(ctx context.Context, identityAddress types.NoxyIdentityAddress, actionable interface{}) ([]types.NoxyDeliveryOutcome, error) {
	devices, err := c.identity.GetDevices(ctx, c.grpc, c.config.AuthToken, identityAddress)
	if err != nil {
		return nil, err
	}
	return c.decision.Send(ctx, c.grpc, c.config.AuthToken, devices, actionable, c.config.DecisionTTLSeconds)
}

// GetDecisionOutcome performs a single poll for human-in-the-loop resolution.
func (c *NoxyAgentClient) GetDecisionOutcome(ctx context.Context, decisionID, identityID string) (*types.NoxyGetDecisionOutcomeResponse, error) {
	req := &noxy.GetDecisionOutcomeRequest{
		RequestId:  uuid.New().String(),
		DecisionId: decisionID,
		IdentityId: identityID,
	}
	ctx = metadata.NewOutgoingContext(ctx, transport.AuthMetadata(c.config.AuthToken))
	resp, err := c.grpc.GetDecisionOutcome(ctx, req)
	if err != nil {
		return nil, err
	}
	return &types.NoxyGetDecisionOutcomeResponse{
		RequestID: resp.RequestId,
		Pending:   resp.Pending,
		Outcome:   types.NoxyHumanDecisionOutcome(resp.Outcome),
	}, nil
}

// WaitForDecisionOutcome polls GetDecisionOutcome with exponential backoff until terminal outcome or pending is false.
func (c *NoxyAgentClient) WaitForDecisionOutcome(ctx context.Context, opts decisionoutcome.WaitForDecisionOutcomeOptions) (*types.NoxyGetDecisionOutcomeResponse, error) {
	initialMs := uint64(400)
	if opts.InitialPollIntervalMs != nil {
		initialMs = *opts.InitialPollIntervalMs
	}
	maxPollMs := uint64(30000)
	if opts.MaxPollIntervalMs != nil {
		maxPollMs = *opts.MaxPollIntervalMs
	}
	maxWaitMs := uint64(900000)
	if opts.MaxWaitMs != nil {
		maxWaitMs = *opts.MaxWaitMs
	}
	backoff := 1.6
	if opts.BackoffMultiplier != nil {
		backoff = *opts.BackoffMultiplier
	}

	started := time.Now()
	intervalMs := float64(initialMs)
	maxPollF := float64(maxPollMs)
	maxWait := time.Duration(maxWaitMs) * time.Millisecond

	for {
		if time.Since(started) > maxWait {
			return nil, decisionoutcome.ErrWaitForDecisionOutcomeTimeout
		}

		raw, err := c.GetDecisionOutcome(ctx, opts.DecisionID, opts.IdentityID)
		if err != nil {
			return nil, err
		}
		if decisionoutcome.IsTerminalHumanOutcome(raw.Outcome) {
			return raw, nil
		}
		if !raw.Pending {
			return raw, nil
		}

		sleep := time.Duration(intervalMs) * time.Millisecond
		if sleep > time.Duration(maxPollMs)*time.Millisecond {
			sleep = time.Duration(maxPollMs) * time.Millisecond
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(sleep):
		}

		intervalMs = intervalMs * backoff
		if intervalMs > maxPollF {
			intervalMs = maxPollF
		}
	}
}

// SendDecisionAndWaitForOutcome runs SendDecision then WaitForDecisionOutcome using the first delivery with a non-empty decision_id.
func (c *NoxyAgentClient) SendDecisionAndWaitForOutcome(ctx context.Context, identityAddress types.NoxyIdentityAddress, actionable interface{}, opts *decisionoutcome.SendDecisionAndWaitOptions) (*types.NoxyGetDecisionOutcomeResponse, error) {
	deliveries, err := c.SendDecision(ctx, identityAddress, actionable)
	if err != nil {
		return nil, err
	}
	var decisionID string
	for _, d := range deliveries {
		if d.DecisionID != "" {
			decisionID = d.DecisionID
			break
		}
	}
	if decisionID == "" {
		return nil, decisionoutcome.ErrSendDecisionNoDecisionID
	}

	waitOpts := decisionoutcome.WaitForDecisionOutcomeOptions{
		DecisionID:            decisionID,
		IdentityID:            identityAddress,
		InitialPollIntervalMs: nil,
		MaxPollIntervalMs:     nil,
		MaxWaitMs:             nil,
		BackoffMultiplier:     nil,
	}
	if opts != nil {
		waitOpts.InitialPollIntervalMs = opts.InitialPollIntervalMs
		waitOpts.MaxPollIntervalMs = opts.MaxPollIntervalMs
		waitOpts.MaxWaitMs = opts.MaxWaitMs
		waitOpts.BackoffMultiplier = opts.BackoffMultiplier
	}
	return c.WaitForDecisionOutcome(ctx, waitOpts)
}

// GetQuota returns quota usage for your application.
func (c *NoxyAgentClient) GetQuota(ctx context.Context) (*types.NoxyGetQuotaResponse, error) {
	return c.quota.Get(ctx, c.grpc, c.config.AuthToken)
}

// Close closes the gRPC connection.
func (c *NoxyAgentClient) Close() error {
	return c.conn.Close()
}
