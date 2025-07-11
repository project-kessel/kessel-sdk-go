package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/project-kessel/kessel-sdk-go/kessel/config"
	"github.com/project-kessel/kessel-sdk-go/kessel/errors"
	v1beta2 "github.com/project-kessel/kessel-sdk-go/kessel/inventory/v1beta2"
)

func main() {
	ctx := context.Background()

	// The SDK supports two OAuth2 configuration approaches:
	// 1. Direct token URL: Specify the exact OAuth2 token endpoint URL
	// 2. Issuer-based discovery: Provide the issuer URL and let the SDK discover
	//    the token endpoint via OpenID Connect Discovery (/.well-known/openid_configuration)

	// Create HTTP configuration with OAuth2
	httpConfig := config.NewHTTPConfig(
		config.WithHTTPEndpoint("http://127.0.0.1:8000"),
		config.WithHTTPInsecure(true),
		config.WithHTTPTimeout(30*time.Second),
		config.WithHTTPUserAgent("kessel-http-example"),
		config.WithHTTPOAuth2Issuer("svc-test", "h91qw8bPiDj9R6VSORsI5TYbceGU5PMH", "http://localhost:8085/realms/redhat-external"),
	)

	// Create the inventory HTTP client directly
	inventoryClient, err := v1beta2.NewInventoryHTTPClient(httpConfig)
	if err != nil {
		// Example of checking for specific error types using sentinel errors
		if errors.IsConnectionError(err) {
			log.Fatal("Failed to establish HTTP connection:", err)
		} else if errors.IsTokenError(err) {
			log.Fatal("OAuth2 token configuration failed:", err)
		} else {
			log.Fatal("HTTP client configuration failed:", err)
		}
	}

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
	tokenOpts, err := inventoryClient.GetTokenHTTPOption()
	if err != nil {
		fmt.Printf("Failed to get token HTTP options: %v\n", err)
		return
	}

	fmt.Println("Using standard HTTP client approach:")
	response, err := inventoryClient.KesselInventoryService.Check(ctx, checkRequest, tokenOpts...)
	if err != nil {
		fmt.Printf("HTTP request failed with error: %v\n", err)
		return
	}
	fmt.Printf("HTTP Check response: %+v\n", response)

}
