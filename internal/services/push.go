package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/noxy-network/go-sdk/grpc/noxy"
	"github.com/noxy-network/go-sdk/internal/crypto"
	"github.com/noxy-network/go-sdk/internal/kyber"
	"github.com/noxy-network/go-sdk/internal/retries"
	"github.com/noxy-network/go-sdk/internal/transport"
	"github.com/noxy-network/go-sdk/internal/types"
	"google.golang.org/grpc/metadata"
)

// PushService sends encrypted push notifications.
type PushService struct {
	kyber *kyber.KyberProvider
}

// NewPushService creates a new PushService.
func NewPushService(kyberProvider *kyber.KyberProvider) *PushService {
	return &PushService{kyber: kyberProvider}
}

// Send sends a push notification to all devices, encrypting per device.
func (s *PushService) Send(ctx context.Context, client noxy.PushServiceClient, authToken string, devices []types.NoxyIdentityDevice, pushNotification interface{}, ttlSeconds uint32) ([]types.NoxyPushResponse, error) {
	plaintext, err := json.Marshal(pushNotification)
	if err != nil {
		return nil, fmt.Errorf("marshal notification: %w", err)
	}

	results := make([]types.NoxyPushResponse, 0, len(devices))
	for _, device := range devices {
		kyberCt, sharedSecret, err := s.kyber.Encapsulate(device.PQPublicKey)
		if err != nil {
			return nil, fmt.Errorf("Kyber encapsulate: %w", err)
		}
		ciphertext, nonce, err := crypto.Encrypt(sharedSecret, plaintext)
		if err != nil {
			return nil, fmt.Errorf("AES-GCM encrypt: %w", err)
		}

		resp, err := s.sendToNetwork(ctx, client, authToken, ciphertext, ttlSeconds, device.DeviceID, kyberCt, nonce)
		if err != nil {
			return nil, fmt.Errorf("gRPC PushNotification: %w", err)
		}
		results = append(results, *resp)
	}
	return results, nil
}

func (s *PushService) sendToNetwork(ctx context.Context, client noxy.PushServiceClient, authToken string, ciphertext []byte, ttlSeconds uint32, targetDeviceID string, kyberCt, nonce []byte) (*types.NoxyPushResponse, error) {
	req := &noxy.PushNotificationRequest{
		RequestId:       uuid.New().String(),
		Ciphertext:      ciphertext,
		TtlSeconds:      ttlSeconds,
		TargetDeviceId:  targetDeviceID,
		KyberCt:         kyberCt,
		Nonce:           nonce,
	}

	resp, err := retries.WithRetry(ctx, func() (*noxy.PushResponse, error) {
		ctx := metadata.NewOutgoingContext(ctx, transport.AuthMetadata(authToken))
		return client.PushNotification(ctx, req)
	}, 3)
	if err != nil {
		return nil, err
	}

	return &types.NoxyPushResponse{
		Status:    types.NoxyPushDeliveryStatus(resp.Status),
		RequestID: resp.RequestId,
	}, nil
}
