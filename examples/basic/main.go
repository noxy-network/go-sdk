// Basic example: route a decision and check quota.
//
// Run with: go run ./examples/basic
//
// Required environment variables:
//	NOXY_APP_TOKEN=your-api-token
//	NOXY_TARGET_ADDRESS=0x... (Web3 identity to route the decision to)
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	noxy "github.com/noxy-network/go-sdk"
)

const relayEndpoint = "https://relay.noxy.network"

func main() {
	authToken := os.Getenv("NOXY_APP_TOKEN")
	if authToken == "" {
		fmt.Fprintf(os.Stderr, "NOXY_APP_TOKEN is required. Set it to your API token.\n")
		os.Exit(1)
	}
	identityAddress := os.Getenv("NOXY_TARGET_ADDRESS")
	if identityAddress == "" {
		fmt.Fprintf(os.Stderr, "NOXY_TARGET_ADDRESS is required. Set it to the Web3 identity (0x...).\n")
		os.Exit(1)
	}

	cfg := noxy.NoxyConfig{
		Endpoint:           relayEndpoint,
		AuthToken:          authToken,
		DecisionTTLSeconds: 3600,
	}

	ctx := context.Background()
	client, err := noxy.InitNoxyAgentClient(ctx, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "InitNoxyAgentClient: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	actionable := map[string]interface{}{
		"kind":    "propose_tool_call",
		"tool":    "transfer_funds",
		"args":    map[string]interface{}{"to": "0x000000000000000000000000000000000000dEaD", "amountWei": "1" },
		"title": "[Go] Transfer 1 wei to the burn address",
		"summary": "[Go] The agent is requesting approval to send 1 wei to the burn address.",
	}

	fmt.Printf("Routing decision to %s...\n", identityAddress)
	resolution, err := client.SendDecisionAndWaitForOutcome(ctx, identityAddress, actionable, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "SendDecisionAndWaitForOutcome: %v\n", err)
		os.Exit(1)
	}
	resJSON, _ := json.MarshalIndent(resolution, "", "  ")
	fmt.Printf("Resolution: %s\n", string(resJSON))

	if resolution.Outcome == noxy.NoxyHumanDecisionOutcomeApproved {
		fmt.Println("User approved — continue agent loop.")
	}
	if resolution.Outcome == noxy.NoxyHumanDecisionOutcomeRejected {
		fmt.Println("User rejected — do not proceed with the agent loop.")
	}

	quota, err := client.GetQuota(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "GetQuota: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Quota: %d / %d remaining (status: %d)\n", quota.QuotaRemaining, quota.QuotaTotal, quota.Status)
}
