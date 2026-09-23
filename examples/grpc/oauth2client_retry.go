package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/project-kessel/kessel-sdk-go/kessel/inventory/v1beta2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	_ "github.com/joho/godotenv/autoload"

	"github.com/project-kessel/kessel-sdk-go/kessel/auth"
)

func oauth2clientRetry() {

	ctx := context.Background()
	discovered, err := auth.FetchOIDCDiscovery(ctx, os.Getenv("AUTH_DISCOVERY_ISSUER_URL"), auth.FetchOIDCDiscoveryOptions{
		HttpClient: nil, // Optionally specify an http client - defaults to http.DefaultClient
	})

	if err != nil {
		panic(err)
	}

	// Configure custom retry options for the OIDC token endpoint.
	// Omitted fields (BaseDelay, MaxDelay, Jitter) keep their defaults
	// (0.5s, 2.0s, full jitter).
	retryOpts := auth.RetryOptions{
		MaxRetries: 5,
	}

	oauthCredentials := auth.NewOAuth2ClientCredentials(
		os.Getenv("AUTH_CLIENT_ID"),
		os.Getenv("AUTH_CLIENT_SECRET"),
		discovered.TokenEndpoint,
		retryOpts,
	)
	inventoryClient, conn, err := v1beta2.NewClientBuilder(os.Getenv("KESSEL_ENDPOINT")).
		OAuth2ClientAuthenticated(&oauthCredentials, nil).
		Build()
	if err != nil {
		log.Fatal("Failed to create gRPC client:", err)
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			log.Printf("Failed to close gRPC client: %v", closeErr)
		}
	}()

	// To fully customize retry behavior:
	//
	//   retryOpts := auth.RetryOptions{
	//       MaxRetries: 5,       // Retry up to 5 times after initial request
	//       BaseDelay:  1.0,     // Initial backoff delay: 1 second
	//       MaxDelay:   10.0,    // Maximum backoff cap: 10 seconds
	//       Jitter:     auth.JitterNone, // Use exact computed delay (no randomization)
	//   }
	//
	// To disable retries entirely:
	//
	//   retryOpts := auth.RetryOptions{MaxRetries: 0}

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

	fmt.Println("Making gRPC request with retry-enabled auth:")

	response, err := inventoryClient.Check(ctx, checkRequest)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.Unavailable:
				log.Fatal("Service unavailable: ", err)
			case codes.PermissionDenied:
				log.Fatal("Permission denied: ", err)
			default:
				log.Fatal("gRPC connection error: ", err)
			}
		} else {
			log.Fatal("Unknown error: ", err)
		}
	}
	fmt.Printf("Check response: %+v\n", response)
}

func main() { oauth2clientRetry() }
