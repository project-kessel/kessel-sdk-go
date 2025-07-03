package main

import (
	"context"
	"fmt"
	"log"

	"github.com/project-kessel/kessel-sdk-go/pkg/kessel/client"
	"github.com/project-kessel/kessel-sdk-go/pkg/kessel/config"
	"github.com/project-kessel/kessel-sdk-go/pkg/kessel/errors"
	"github.com/project-kessel/kessel-sdk-go/pkg/kessel/grpc"
	v1beta2 "github.com/project-kessel/kessel-sdk-go/pkg/kessel/inventory/v1beta2"
)

func main() {
	// Example gRPC configuration with OAuth2
	grpcConfig := config.NewGRPCConfig(
		config.WithGRPCEndpoint("127.0.0.1:9000"),
		config.WithGRPCInsecure(true),
		// Option 1: Direct token URL configuration
		//config.WithGRPCOAuth2(
		//	"svc-test",
		//	"h91qw8bPiDj9R6VSORsI5TYbceGU5PMH",
		//	"http://localhost:8085/realms/redhat-external/protocol/openid-connect/token",
		//),
		// Option 2: Issuer-based configuration with auto-discovery
		//config.WithGRPCOAuth2Issuer(
		//	"svc-test",
		//	"h91qw8bPiDj9R6VSORsI5TYbceGU5PMH",
		//	"http://localhost:8085/realms/redhat-external",
		//),
	)

	// Create gRPC client with OAuth2 support using builder pattern
	ctx := context.Background()

	// The SDK supports two OAuth2 configuration approaches:
	// 1. Direct token URL: Specify the exact OAuth2 token endpoint URL
	// 2. Issuer-based discovery: Provide the issuer URL and let the SDK discover
	//    the token endpoint via OpenID Connect Discovery (/.well-known/openid_configuration)

	// Option 1: Create builder from config
	builder := grpc.NewClientBuilderFromConfig(grpcConfig)

	// Option 2: Manual builder configuration (alternative approach)
	// builder := grpc.NewClientBuilder("localhost:9000").
	//     WithInsecure()

	grpcClient, err := client.NewGRPCClient(ctx, grpcConfig, builder)
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
		if closeErr := grpcClient.Close(); closeErr != nil {
			log.Printf("Failed to close gRPC client: %v", closeErr)
		}
	}()

	// Example request
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

	// Make the request and handle potential errors
	response, err := grpcClient.Check(ctx, checkRequest)
	if err != nil {
		fmt.Printf("Request failed with error: %v\n", err)
		return
	}

	// Process successful response
	fmt.Printf("Check response: %+v\n", response)

}
