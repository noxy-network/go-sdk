// Package noxy provides a Go SDK for the Noxy push notification network.
//
// Send encrypted push notifications to Web3 wallet addresses via the Noxy relay.
package noxy

import (
	"context"
	"strings"

	"github.com/noxy-network/go-sdk/client"
	"github.com/noxy-network/go-sdk/internal/config"
	"github.com/noxy-network/go-sdk/internal/types"
)

// NoxyConfig holds configuration for the Noxy SDK client.
type NoxyConfig struct {
	// Endpoint is the Noxy relay gRPC endpoint (e.g. "https://relay.noxy.network:443").
	Endpoint string
	// AuthToken is the Bearer token for relay authentication.
	AuthToken string
	// NotificationTTLSeconds is the time-to-live for notifications in seconds.
	NotificationTTLSeconds uint32
}

// InitNoxyClient initializes the SDK client. This is async because it establishes the gRPC connection.
func InitNoxyClient(ctx context.Context, cfg NoxyConfig) (*client.NoxyPushClient, error) {
	// Normalize endpoint: strip https:// or http://
	endpoint := strings.TrimPrefix(cfg.Endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimSuffix(endpoint, "/")

	internalCfg := &config.NoxyConfig{
		Endpoint:               endpoint,
		AuthToken:              cfg.AuthToken,
		NotificationTTLSeconds: cfg.NotificationTTLSeconds,
	}
	return client.NewNoxyPushClient(ctx, internalCfg)
}

// Re-export types for convenience.
type (
	NoxyPushClient        = client.NoxyPushClient
	NoxyIdentityAddress   = types.NoxyIdentityAddress
	NoxyPushResponse      = types.NoxyPushResponse
	NoxyPushDeliveryStatus = types.NoxyPushDeliveryStatus
	NoxyGetQuotaResponse  = types.NoxyGetQuotaResponse
	NoxyQuotaStatus       = types.NoxyQuotaStatus
)
