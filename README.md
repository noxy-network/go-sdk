# 📦 @noxy-network/go-sdk

SDK for **AI agent runtimes** integrating with the [Noxy](https://noxy.network) **Decision Layer**: send encrypted, **actionable** decision payloads (tool proposals, approvals, next-step hints) to registered agent devices over gRPC.

**Before you integrate:** Create your app at [noxy.network](https://noxy.network). When the app is created, you receive an **app id** and an **app token** (auth token). This Go SDK authenticates with the relay using the **app token** (`AuthToken` in `NoxyConfig`). The **app id** is used by client SDKs (browser, iOS, Android, Telegram bot), not as the bearer token here.

## Overview

Use this SDK to:

- **Route decisions** to devices bound to a Web3 identity (`0x…` address) — structured JSON you define (e.g. proposed tool calls, parameters, user-visible summaries).
- **Receive delivery outcomes** from the relay (`DELIVERED`, `QUEUED`, `NO_DEVICES`, etc.) plus a **`decision_id`** when the relay accepts the route.
- **Wait for human-in-the-loop resolution** — the wallet user **approves**, **rejects**, or the decision **expires**. The usual path is **`SendDecisionAndWaitForOutcome`** (route + poll in one step). Use `GetDecisionOutcome` / `WaitForDecisionOutcome` alone for finer control.
- **Query quota** for your agent application on the relay.
- **Resolve identity devices** so each device receives its own encrypted copy of the decision.

The wire API uses **`agent.proto`** (`noxy.agent.AgentService`): `RouteDecision`, `GetDecisionOutcome`, `GetQuota`, `GetIdentityDevices`.

Communication is **gRPC over TLS** with Bearer authentication. Payloads are **encrypted end-to-end** (Kyber + AES-256-GCM) per device before leaving your process; the relay sees ciphertext only.

## Architecture

The **encrypted path** covers **SDK → relay** and **relay → device**: decision content is ciphertext on both hops; the relay forwards without decrypting.

```
                      Ciphertext only (E2E)                  Ciphertext only (E2E)
┌──────────────────┐     gRPC (TLS)      ┌─────────────────┐     gRPC (TLS)         ┌──────────────────┐
│  AI agent /      │ ◄─────────────────► │  Noxy relay     │ ◄──────────────────► │  Agent device    │
│  orchestrator    │   RouteDecision     │  (Decision      │                      │  (human approves │
│  (this SDK)      │   GetDecisionOutcome│   Layer)        │                      │   or rejects)    │
│                  │   GetQuota          │   forwards only │                      │   decrypts       │
│                  │   GetIdentityDevices│                 │                      │                  │
└──────────────────┘                     └─────────────────┘                      └──────────────────┘
```

## Requirements

- Go **>= 1.21**
- **CGO** enabled (required for Kyber encryption)
- **PQClean** sources: the SDK uses PQClean ML-KEM 768 from [pq-wasm](https://github.com/noxy-network/pq-wasm) for interoperability with Node.js and Rust SDKs. Run `make build` which fetches pq-wasm and builds the library automatically.

## Installation

```bash
go get github.com/noxy-network/go-sdk
```

## Quick start

```go
package main

import (
	"context"
	"fmt"

	noxy "github.com/noxy-network/go-sdk"
)

func main() {
	ctx := context.Background()
	client, err := noxy.InitNoxyAgentClient(ctx, noxy.NoxyConfig{
		Endpoint:           "https://relay.noxy.network",
		AuthToken:          "your-api-token",
		DecisionTTLSeconds: 3600,
	})
	if err != nil {
		panic(err)
	}
	defer client.Close()

	resolution, err := client.SendDecisionAndWaitForOutcome(ctx,
		"0x...",
		map[string]interface{}{
			"kind":    "propose_tool_call",
			"tool":    "transfer_funds",
			"args":    map[string]interface{}{"to": "0x000000000000000000000000000000000000dEaD", "amountWei": "1" },
			"title": "[Go] Transfer 1 wei to the burn address",
			"summary": "[Go] The agent is requesting approval to send 1 wei to the burn address.",
		},
		nil,
	)
	if err != nil {
		panic(err)
	}

	if resolution.Outcome == noxy.NoxyHumanDecisionOutcomeApproved {
		fmt.Println("approved")
	}

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
| `Endpoint` | `string` | Yes | Relay gRPC endpoint (e.g. `https://relay.noxy.network`). Scheme is stripped; TLS is used. |
| `AuthToken` | `string` | Yes | Bearer token for relay auth (`Authorization` header). |
| `DecisionTTLSeconds` | `uint32` | Yes | TTL for routed decisions (seconds). |

## SendDecisionAndWaitOptions

Optional pointer argument to **`SendDecisionAndWaitForOutcome`**.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `InitialPollIntervalMs` | `*uint64` | No | Delay after the first poll before the next attempt (ms). Default `400`. |
| `MaxPollIntervalMs` | `*uint64` | No | Maximum delay between polls (ms). Default `30000`. |
| `MaxWaitMs` | `*uint64` | No | Total time budget for polling (ms). Default `900000` (15 minutes). Errors with `ErrWaitForDecisionOutcomeTimeout`. |
| `BackoffMultiplier` | `*float64` | No | Multiplier after each attempt. Default `1.6`. |

## API

### `InitNoxyAgentClient(ctx, config) (*NoxyAgentClient, error)`

Initializes the client (gRPC connection + Kyber).

### `NoxyAgentClient`

- **`SendDecision`** — route an encrypted decision to all devices for the identity.
- **`GetDecisionOutcome`** — single poll (`pending` + `outcome`).
- **`SendDecisionAndWaitForOutcome`** — `SendDecision` then `WaitForDecisionOutcome` using the first non-empty `decision_id`; polling uses the identity address as `identity_id`.
- **`WaitForDecisionOutcome`** — exponential backoff until terminal outcome, `pending == false`, or timeout.
- **`GetQuota`** — quota for the application.

### Errors

- `ErrWaitForDecisionOutcomeTimeout`
- `ErrSendDecisionNoDecisionID`

### Helpers

- `IsTerminalHumanOutcome(outcome)`

### Types

- **`NoxyDeliveryStatus`**, **`NoxyDeliveryOutcome`**
- **`NoxyHumanDecisionOutcome`**: `NoxyHumanDecisionOutcomePending` | `Approved` | `Rejected` | `Expired`
- **`NoxyQuotaStatus`**

## Encryption (summary)

1. Kyber768 encapsulation per device `pq_public_key`.
2. HKDF-SHA256 → AES-256-GCM key; random 12-byte nonce.
3. JSON payload encrypted; only `kyber_ct`, `nonce`, `ciphertext` cross the relay.

## License

MIT
