package client

import (
	"context"

	"github.com/noxy-network/go-sdk/grpc/noxy"
	"github.com/noxy-network/go-sdk/internal/config"
	"github.com/noxy-network/go-sdk/internal/kyber"
	"github.com/noxy-network/go-sdk/internal/services"
	"github.com/noxy-network/go-sdk/internal/transport"
	"github.com/noxy-network/go-sdk/internal/types"
	"google.golang.org/grpc"
)

// NoxyPushClient is the main SDK client for sending push notifications.
type NoxyPushClient struct {
	config   *config.NoxyConfig
	conn     *grpc.ClientConn
	grpc     noxy.PushServiceClient
	identity *services.IdentityService
	push     *services.PushService
	quota    *services.QuotaService
}

// NewNoxyPushClient creates and initializes a NoxyPushClient.
func NewNoxyPushClient(ctx context.Context, cfg *config.NoxyConfig) (*NoxyPushClient, error) {
	grpcClient, conn, err := transport.NewPushServiceClient(ctx, cfg.Endpoint, cfg.AuthToken)
	if err != nil {
		return nil, err
	}

	kyberProvider := kyber.NewKyberProvider()
	return &NoxyPushClient{
		config:   cfg,
		conn:     conn,
		grpc:     grpcClient,
		identity: services.NewIdentityService(),
		push:     services.NewPushService(kyberProvider),
		quota:    services.NewQuotaService(),
	}, nil
}

// SendPush sends a push notification to all devices registered for the given Web3 identity address.
func (c *NoxyPushClient) SendPush(ctx context.Context, identityAddress types.NoxyIdentityAddress, pushNotification interface{}) ([]types.NoxyPushResponse, error) {
	devices, err := c.identity.GetDevices(ctx, c.grpc, c.config.AuthToken, identityAddress)
	if err != nil {
		return nil, err
	}
	return c.push.Send(ctx, c.grpc, c.config.AuthToken, devices, pushNotification, c.config.NotificationTTLSeconds)
}

// GetQuota returns quota usage for your application.
func (c *NoxyPushClient) GetQuota(ctx context.Context) (*types.NoxyGetQuotaResponse, error) {
	return c.quota.Get(ctx, c.grpc, c.config.AuthToken)
}

// Close closes the gRPC connection.
func (c *NoxyPushClient) Close() error {
	return c.conn.Close()
}
