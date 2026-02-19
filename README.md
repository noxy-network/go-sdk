# @noxy-network/go-sdk

Backend SDK for Go servers to integrate with the [Noxy](https://noxy.network) push notification network. Send encrypted push notifications to Web3 wallet addresses via the Noxy relay infrastructure.

## Overview

This SDK enables server-side applications to:

- **Send push notifications** to users by their Web3 wallet address (EVM `0x` format)
- **Query quota usage** for your application's relay allocation
- **Resolve identity devices** to deliver notifications to all registered devices

Communication with the Noxy relay is performed over **gRPC** using Protocol Buffers. All notifications are **encrypted end-to-end** on the backend before transmission; decryption occurs only on the recipient's Noxy device. The SDK uses **post-quantum encryption** (Kyber768) to protect payloads against future quantum attacks.

## Architecture

```
┌─────────────────┐     gRPC (TLS)      ┌─────────────────┐     E2E Encrypted     ┌─────────────────┐
│  Your Backend   │ ◄─────────────────► │  Noxy Relay     │ ◄──────────────────► │  Noxy Device    │
│  (this SDK)     │   PushNotification  │                 │   Ciphertext only    │  (decrypts)      │
│                 │   GetQuota          │                 │                      │                 │
│                 │   GetIdentityDevices│                 │                      │                 │
└─────────────────┘                     └─────────────────┘                      └─────────────────┘
```

- **Encryption**: Kyber768 (post-quantum KEM) + AES-256-GCM. Each notification is encrypted per-device using the device's post-quantum public key.
- **Transport**: gRPC over TLS with Bearer token authentication.
- **Relay**: The Noxy relay forwards ciphertext only; it cannot decrypt notification payloads.

## Requirements

- Go **>= 1.21**
- **CGO** enabled (required for Kyber encryption)
- **PQClean** sources: the SDK uses PQClean ML-KEM 768 from [pq-wasm](https://github.com/noxy-network/pq-wasm) for interoperability with Node.js and Rust SDKs. Run `make build` which fetches pq-wasm and builds the library automatically.

## Installation

```bash
go get github.com/noxy-network/go-sdk
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"

	"github.com/noxy-network/go-sdk/noxy"
)

func main() {
	ctx := context.Background()
	client, err := noxy.InitNoxyClient(ctx, noxy.NoxyConfig{
		Endpoint:               "https://relay.noxy.network:443",
		AuthToken:              "your-api-token",
		NotificationTTLSeconds: 3600,
	})
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// Send a push notification to a wallet address
	results, err := client.SendPush(ctx, "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1", map[string]interface{}{
		"title": "New message",
		"body":  "You have a new notification",
		"data":  map[string]interface{}{"action": "open_chat", "id": "123"},
	})
	if err != nil {
		panic(err)
	}

	// Check quota usage
	quota, err := client.GetQuota(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%d remaining\n", quota.QuotaRemaining)
}
```

## Configuration

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `Endpoint` | `string` | Yes | Noxy relay gRPC endpoint (e.g. `https://relay.noxy.network:443` or `localhost:4433`). Scheme is stripped; TLS is used by default. |
| `AuthToken` | `string` | Yes | Bearer token for relay authentication. Sent in the `Authorization` header on every request. |
| `NotificationTTLSeconds` | `uint32` | Yes | Time-to-live for notifications in seconds. |

## API Reference

### `InitNoxyClient(ctx, config) (*NoxyPushClient, error)`

Initializes the SDK client. This is asynchronous because it establishes the gRPC connection.

### `NoxyPushClient`

#### `SendPush(ctx, identityAddress, pushNotification) ([]NoxyPushResponse, error)`

Sends a push notification to all devices registered for the given Web3 identity address.

- **`identityAddress`**: EVM address in `0x` format (e.g. `0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1`)
- **`pushNotification`**: Any JSON-serializable value (e.g. `map[string]interface{}`). Encrypted before transmission.
- **Returns**: Slice of `NoxyPushResponse` per device, with `Status` and `RequestID`.

#### `GetQuota(ctx) (*NoxyGetQuotaResponse, error)`

Returns quota usage for your application.

- **Returns**: `NoxyGetQuotaResponse` with `RequestID`, `AppName`, `QuotaTotal`, `QuotaRemaining`, `Status`.

### Types

- **`NoxyPushDeliveryStatus`**: `Delivered` (0) | `Queued` (1) | `NoDevices` (2) | `Rejected` (3) | `Error` (4)
- **`NoxyQuotaStatus`**: `QuotaActive` (0) | `QuotaSuspended` (1) | `QuotaDeleted` (2)

## Encryption Details

1. **Key encapsulation**: For each device, the SDK encapsulates a shared secret using the device's Kyber768 post-quantum public key (`pq_public_key`).
2. **Key derivation**: The shared secret is expanded via HKDF-SHA256 to a 256-bit AES key.
3. **Payload encryption**: The notification payload (JSON) is encrypted with AES-256-GCM. The ciphertext includes the GCM auth tag appended for integrity verification.
4. **Transmission**: Only `kyber_ct`, `nonce`, and `ciphertext` are sent to the relay. The relay cannot decrypt; only the target device (with its secret key) can decrypt.

## Development

```bash
# Generate proto and build (includes PQClean library for Kyber)
make build

# Or step by step:
make pqclean-lib   # Fetches pq-wasm from GitHub and builds PQClean ML-KEM 768
make proto         # Generate gRPC code
CGO_ENABLED=1 go build ./...

# Test
go test ./...

# Run example (requires NOXY_AUTH_TOKEN)
go run ./examples/basic
```

**Note:** The build fetches [pq-wasm](https://github.com/noxy-network/pq-wasm) from GitHub (including the PQClean submodule). The Kyber provider uses PQClean ML-KEM 768 for full interoperability with the Node.js WASM and Rust pqcrypto-kyber implementations.

## License

MIT
