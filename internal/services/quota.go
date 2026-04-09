package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/noxy-network/go-sdk/grpc/noxy"
	"github.com/noxy-network/go-sdk/internal/transport"
	"github.com/noxy-network/go-sdk/internal/types"
	"google.golang.org/grpc/metadata"
)

// QuotaService fetches quota usage from the relay.
type QuotaService struct{}

// NewQuotaService creates a new QuotaService.
func NewQuotaService() *QuotaService {
	return &QuotaService{}
}

// Get returns quota usage for the application.
func (s *QuotaService) Get(ctx context.Context, client noxy.AgentServiceClient, authToken string) (*types.NoxyGetQuotaResponse, error) {
	req := &noxy.GetQuotaRequest{
		RequestId: uuid.New().String(),
	}
	ctx = metadata.NewOutgoingContext(ctx, transport.AuthMetadata(authToken))
	resp, err := client.GetQuota(ctx, req)
	if err != nil {
		return nil, err
	}
	return &types.NoxyGetQuotaResponse{
		RequestID:      resp.RequestId,
		AppName:        resp.AppName,
		QuotaTotal:     resp.QuotaTotal,
		QuotaRemaining: resp.QuotaRemaining,
		Status:         types.NoxyQuotaStatus(resp.Status),
	}, nil
}
