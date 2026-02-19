package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/noxy-network/go-sdk/grpc/noxy"
	"github.com/noxy-network/go-sdk/internal/transport"
	"github.com/noxy-network/go-sdk/internal/types"
	"google.golang.org/grpc/metadata"
)

// IdentityService fetches identity devices from the relay.
type IdentityService struct{}

// NewIdentityService creates a new IdentityService.
func NewIdentityService() *IdentityService {
	return &IdentityService{}
}

// GetDevices returns all devices registered for the given identity address.
func (s *IdentityService) GetDevices(ctx context.Context, client noxy.PushServiceClient, authToken, identityID string) ([]types.NoxyIdentityDevice, error) {
	req := &noxy.GetIdentityDevicesRequest{
		RequestId:  uuid.New().String(),
		IdentityId: identityID,
	}
	ctx = metadata.NewOutgoingContext(ctx, transport.AuthMetadata(authToken))
	resp, err := client.GetIdentityDevices(ctx, req)
	if err != nil {
		return nil, err
	}
	devices := make([]types.NoxyIdentityDevice, len(resp.Devices))
	for i, d := range resp.Devices {
		devices[i] = types.NoxyIdentityDevice{
			DeviceID:    d.DeviceId,
			PublicKey:   d.PublicKey,
			PQPublicKey: d.PqPublicKey,
		}
	}
	return devices, nil
}
