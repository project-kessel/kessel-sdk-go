package main

import (
	"context"
	"fmt"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"

	"github.com/project-kessel/kessel-sdk-go/kessel/inventory/v1beta2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func deleteResource() {
	ctx := context.Background()

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

	deleteResourceRequest := &v1beta2.DeleteResourceRequest{
		Reference: &v1beta2.ResourceReference{
			ResourceType: "host",
			ResourceId:   "854589f0-3be7-4cad-8bcd-45e18f33cb81",
			Reporter: &v1beta2.ReporterReference{
				Type: "hbi",
			},
		},
	}

	fmt.Println("Making delete resource request:")

	response, err := inventoryClient.DeleteResource(ctx, deleteResourceRequest)
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
	fmt.Printf("Delete resource response: %+v\n", response)
}

func main() { deleteResource() }
