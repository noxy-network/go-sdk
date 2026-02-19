package config

// NoxyConfig holds configuration for the Noxy SDK client.
type NoxyConfig struct {
	// Endpoint is the Noxy relay gRPC endpoint (e.g. "https://relay.noxy.network:443").
	// Scheme is stripped; TLS is used by default.
	Endpoint string
	// AuthToken is the Bearer token for relay authentication.
	AuthToken string
	// NotificationTTLSeconds is the time-to-live for notifications in seconds.
	NotificationTTLSeconds uint32
}
