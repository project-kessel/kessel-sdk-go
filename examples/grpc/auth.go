package main

import (
	"context"
	"fmt"
	"github.com/project-kessel/kessel-sdk-go/kessel/inventory/v1beta2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	credentials "google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"

	"github.com/project-kessel/kessel-sdk-go/kessel/auth"
	kesselgrpc "github.com/project-kessel/kessel-sdk-go/kessel/grpc"
)

func main() {
	ctx := context.Background()

	discovered, err := auth.FetchOIDCDiscovery(auth.FetchOIDCDiscoveryOptions{
		IssuerUrl:  os.Getenv("AUTH_DISCOVERY_ISSUER_URL"),
		Context:    ctx, // Optionally specify a context - defaults to context.Background()
		HttpClient: nil, // Optionally specify an http client - defaults to http.DefaultClient
	})

	if err != nil {
		panic(err)
	}

	oauthCredentials := auth.MakeOAuth2ClientCredentials(os.Getenv("AUTH_CLIENT_ID"), os.Getenv("AUTH_CLIENT_SECRET"), discovered.TokenEndpoint)

	var dialOpts []grpc.DialOption
	dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(nil)))
	dialOpts = append(dialOpts, grpc.WithPerRPCCredentials(kesselgrpc.OAuth2CallCredentials(&oauthCredentials)))

	conn, err := grpc.NewClient(os.Getenv("KESSEL_ENDPOINT"), dialOpts...)
	if err != nil {
		log.Fatal("Failed to create gRPC client:", err)
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

	fmt.Println("Making basic gRPC request:")

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
