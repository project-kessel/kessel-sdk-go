package main

import (
	"context"
	"fmt"
	"log"

	"github.com/project-kessel/kessel-sdk-go/kessel/errors"
	v1beta2 "github.com/project-kessel/kessel-sdk-go/kessel/inventory/v1beta2"
)

func main() {
	ctx := context.Background()

	// The SDK supports two OAuth2 configuration approaches:
	// 1. Direct token URL: Specify the exact OAuth2 token endpoint URL
	// 2. Issuer-based discovery: Provide the issuer URL and let the SDK discover
	//    the token endpoint via OpenID Connect Discovery (/.well-known/openid_configuration)

	inventoryClient, err := v1beta2.NewInventoryGRPCClientBuilder().
		WithEndpoint("127.0.0.1:9000").
		WithInsecure(true).
		WithOAuth2Issuer("svc-test", "h91qw8bPiDj9R6VSORsI5TYbceGU5PMH", "http://localhost:8085/realms/redhat-external").
		WithMaxReceiveMessageSize(8 * 1024 * 1024). // 8MB
		WithMaxSendMessageSize(8 * 1024 * 1024).    // 8MB
		Build()

	if err != nil {
		// Example of checking for specific error types using sentinel errors
		if errors.IsConnectionError(err) {
			log.Fatal("Failed to establish connection:", err)
		} else if errors.IsTokenError(err) {
			log.Fatal("OAuth2 token configuration failed:", err)
		} else {
			log.Fatal("Unknown error:", err)
		}
	}
	defer func() {
		if closeErr := inventoryClient.Close(); closeErr != nil {
			log.Printf("Failed to close gRPC client: %v", closeErr)
		}
	}()

	// Example request using the external API types
	checkRequest := &v1beta2.CheckRequest{
		Object: &v1beta2.ResourceReference{
			ResourceType: "host",
			ResourceId:   "1213",
			Reporter: &v1beta2.ReporterReference{
				Type: "HBI",
			},
		},
		Relation: "member",
		Subject: &v1beta2.SubjectReference{
			Resource: &v1beta2.ResourceReference{
				ResourceType: "user",
				ResourceId:   "tim",
			},
		},
	}

	// OAuth2 tokens are automatically injected - no manual token management needed!
	fmt.Println("Making request with automatic OAuth2 token injection:")

	response, err := inventoryClient.Check(ctx, checkRequest)
	if err != nil {
		fmt.Printf("Request failed with error: %v\n", err)
		return
	}
	fmt.Printf("Check response: %+v\n", response)
}
