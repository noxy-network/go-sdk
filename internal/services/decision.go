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

// DecisionService encrypts and routes actionable decisions per device.
type DecisionService struct {
	kyber *kyber.KyberProvider
}

// NewDecisionService creates a new DecisionService.
func NewDecisionService(kyberProvider *kyber.KyberProvider) *DecisionService {
	return &DecisionService{kyber: kyberProvider}
}

// Send routes an actionable decision to all devices, encrypting per device.
// One UUID is generated per batch and set as DecisionId on each RouteDecision request.
func (s *DecisionService) Send(ctx context.Context, client noxy.AgentServiceClient, authToken string, devices []types.NoxyIdentityDevice, actionable interface{}, ttlSeconds uint32) ([]types.NoxyDeliveryOutcome, error) {
	plaintext, err := json.Marshal(actionable)
	if err != nil {
		return nil, fmt.Errorf("marshal decision: %w", err)
	}

	decisionID := uuid.New().String()
	results := make([]types.NoxyDeliveryOutcome, 0, len(devices))
	for _, device := range devices {
		kyberCt, sharedSecret, err := s.kyber.Encapsulate(device.PQPublicKey)
		if err != nil {
			return nil, fmt.Errorf("Kyber encapsulate: %w", err)
		}
		ciphertext, nonce, err := crypto.Encrypt(sharedSecret, plaintext)
		if err != nil {
			return nil, fmt.Errorf("AES-GCM encrypt: %w", err)
		}

		resp, err := s.sendToNetwork(ctx, client, authToken, ciphertext, ttlSeconds, device.DeviceID, kyberCt, nonce, decisionID)
		if err != nil {
			return nil, fmt.Errorf("gRPC RouteDecision: %w", err)
		}
		results = append(results, *resp)
	}
	return results, nil
}

func (s *DecisionService) sendToNetwork(ctx context.Context, client noxy.AgentServiceClient, authToken string, ciphertext []byte, ttlSeconds uint32, targetDeviceID string, kyberCt, nonce []byte, decisionID string) (*types.NoxyDeliveryOutcome, error) {
	req := &noxy.RouteDecisionRequest{
		RequestId:      uuid.New().String(),
		Ciphertext:     ciphertext,
		TtlSeconds:     ttlSeconds,
		TargetDeviceId: targetDeviceID,
		KyberCt:        kyberCt,
		Nonce:          nonce,
		DecisionId:     decisionID,
	}

	resp, err := retries.WithRetry(ctx, func() (*noxy.DeliveryOutcome, error) {
		ctx := metadata.NewOutgoingContext(ctx, transport.AuthMetadata(authToken))
		return client.RouteDecision(ctx, req)
	}, 3)
	if err != nil {
		return nil, err
	}

	outID := resp.DecisionId
	if outID == "" {
		outID = decisionID
	}
	return &types.NoxyDeliveryOutcome{
		Status:     types.NoxyDeliveryStatus(resp.Status),
		RequestID:  resp.RequestId,
		DecisionID: outID,
	}, nil
}
