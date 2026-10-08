// WARNING: This example demonstrates INSECURE configuration for local development only.
// DO NOT USE IN PRODUCTION. The .Insecure() option disables TLS certificate validation
// and encryption, leaving traffic vulnerable to interception and tampering.
// For production use, see authenticated.go or oauth2client_authenticated.go examples.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/project-kessel/kessel-sdk-go/kessel/inventory/v1beta2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func insecure() {
	ctx := context.Background()
	// WARNING: .Insecure() disables TLS - local development only
	inventoryClient, conn, err := v1beta2.NewClientBuilder(os.Getenv("KESSEL_ENDPOINT")).
		Insecure().
		Build()
	if err != nil {
		log.Fatal("Failed to create gRPC client:", err)
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
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

func main() { insecure() }
