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
	// Example configuration
	cfg := &config.Config{
		Endpoint:       "localhost:9000",
		Insecure:       true,
		EnableOIDCAuth: false,
	}

	// Create gRPC client
	ctx := context.Background()
	builder := grpc.NewClientBuilder(cfg.Endpoint).WithInsecure()

	grpcClient, err := client.NewGRPCClient(ctx, cfg, builder)
	if err != nil {
		// Example of checking for specific error types using sentinel errors
		if errors.IsConnectionError(err) {
			log.Fatal("Failed to establish connection:", err)
		} else if errors.IsHTTPClientError(err) {
			log.Fatal("HTTP client creation failed:", err)
		} else {
			log.Fatal("Unknown error:", err)
		}
	}
	defer func() {
		if closeErr := grpcClient.Close(); closeErr != nil {
			log.Printf("Failed to close gRPC client: %v", closeErr)
		}
	}()

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
		// Handle different types of errors appropriately
		switch {
		case errors.IsTokenError(err):
			fmt.Println("Authentication issue - token retrieval failed:", err)
			// Could implement retry logic or re-authentication here
		case errors.IsConnectionError(err):
			fmt.Println("Network connectivity issue:", err)
			// Could implement retry with exponential backoff
		default:
			fmt.Println("Request failed with error:", err)
		}
		return
	}

	// Process successful response
	fmt.Printf("Check response: %+v\n", response)

	// Additional example showing how to handle token cache errors
	if err := demonstrateTokenCacheHandling(); err != nil {
		if errors.IsTokenCacheError(err) {
			fmt.Println("Token cache miss - this is expected behavior:", err)
		}
	}
}

func demonstrateTokenCacheHandling() error {
	// This would typically be called within token client logic
	// Here we simulate a cache miss scenario
	return errors.NewTokenCacheError("token not found in cache")
}
