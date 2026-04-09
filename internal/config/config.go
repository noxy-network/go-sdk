package config

// NoxyConfig holds configuration for the Noxy Decision Layer SDK client.
type NoxyConfig struct {
	// Endpoint is the Noxy relay gRPC endpoint (e.g. "https://relay.noxy.network:443").
	// Scheme is stripped; TLS is used by default.
	Endpoint string
	// AuthToken is the Bearer token for relay authentication.
	AuthToken string
	// DecisionTTLSeconds is the time-to-live for routed decisions in seconds.
	DecisionTTLSeconds uint32
}
