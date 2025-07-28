package main

import (
	"context"
	"fmt"
	"log"

	"github.com/project-kessel/kessel-sdk-go/kessel/config"
	"github.com/project-kessel/kessel-sdk-go/kessel/errors"
	"github.com/project-kessel/kessel-sdk-go/kessel/grpc/auth"
	v1beta2 "github.com/project-kessel/kessel-sdk-go/kessel/inventory/v1beta2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ctx := context.Background()

	// The SDK supports two OAuth2 configuration approaches:
	// 1. Direct token URL: Specify the exact OAuth2 token endpoint URL
	// 2. Issuer-based discovery: Provide the issuer URL and let the SDK discover
	//    the token endpoint via OpenID Connect Discovery (/.well-known/openid_configuration)

	grpcConfig := &config.GRPCConfig{
		BaseConfig: config.BaseConfig{
			Endpoint:    "127.0.0.1:9000",
			Insecure:    true,
			EnableOauth: true,
			Oauth2: config.Oauth2{
				ClientID:     "svc-test",
				ClientSecret: "h91qw8bPiDj9R6VSORsI5TYbceGU5PMH",
				IssuerURL:    "http://localhost:8085/realms/redhat-external",
			},
		},
		MaxReceiveMessageSize: 8 * 1024 * 1024, // 8MB
		MaxSendMessageSize:    8 * 1024 * 1024, // 8MB
	}

	// Create OAuth2 token source
	tokenSource, err := auth.NewTokenSource(grpcConfig)
	if err != nil {
		if errors.IsTokenError(err) {
			log.Fatal("OAuth2 token configuration failed: ", err)
		} else {
			log.Fatal("Unknown auth error: ", err)
		}
	}

	// Using insecure credentials for local development
	var dialOpts []grpc.DialOption
	dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	dialOpts = append(dialOpts, grpc.WithPerRPCCredentials(tokenSource.GetInsecureGRPCCredentials()))

	dialOpts = append(dialOpts,
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(grpcConfig.MaxReceiveMessageSize)),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(grpcConfig.MaxSendMessageSize)),
	)

	conn, err := grpc.NewClient(grpcConfig.Endpoint, dialOpts...)
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
		if closeErr := conn.Close(); closeErr != nil {
			log.Printf("Failed to close gRPC client: %v", closeErr)
		}
	}()

	inventoryClient := v1beta2.NewKesselInventoryServiceClient(conn)

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
