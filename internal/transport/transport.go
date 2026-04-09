package transport

import (
	"context"
	"strings"

	"github.com/noxy-network/go-sdk/grpc/noxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

// normalizeEndpoint strips https:// or http:// and trailing slashes.
func normalizeEndpoint(endpoint string) string {
	s := strings.TrimPrefix(endpoint, "https://")
	s = strings.TrimPrefix(s, "http://")
	return strings.TrimSuffix(s, "/")
}

// NewAgentServiceClient creates a gRPC client for AgentService. Pass auth via metadata on each RPC.
func NewAgentServiceClient(ctx context.Context, endpoint string) (noxy.AgentServiceClient, *grpc.ClientConn, error) {
	_ = ctx
	addr := normalizeEndpoint(endpoint)
	creds := credentials.NewTLS(nil)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, nil, err
	}
	client := noxy.NewAgentServiceClient(conn)
	return client, conn, nil
}

// AuthMetadata returns metadata with Bearer token for gRPC calls.
func AuthMetadata(authToken string) metadata.MD {
	return metadata.Pairs("authorization", "Bearer "+authToken)
}
