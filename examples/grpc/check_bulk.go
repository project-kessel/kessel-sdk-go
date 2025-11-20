package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/project-kessel/kessel-sdk-go/kessel/inventory/v1beta2"
	v2 "github.com/project-kessel/kessel-sdk-go/kessel/rbac/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func checkBulk() {
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

	// Item 1: Check if bob can view widgets in workspace_123
	item1 := &v1beta2.CheckBulkRequestItem{
		Object:   v2.WorkspaceResource("workspace_123"),
		Relation: "view_widget",
		Subject:  v2.PrincipalSubject("bob", "redhat"),
	}

	// Item 2: Check if bob can use widgets in workspace_456
	item2 := &v1beta2.CheckBulkRequestItem{
		Object:   v2.WorkspaceResource("workspace_456"),
		Relation: "use_widget",
		Subject:  v2.PrincipalSubject("bob", "redhat"),
	}

	// Item 3: Check with invalid resource type to demonstrate error handling
	item3 := &v1beta2.CheckBulkRequestItem{
		Object: &v1beta2.ResourceReference{
			ResourceType: "not_a_valid_type",
			ResourceId:   "invalid_resource",
			Reporter: &v1beta2.ReporterReference{
				Type: "rbac",
			},
		},
		Relation: "view_widget",
		Subject:  v2.PrincipalSubject("alice", "redhat"),
	}

	checkBulkRequest := &v1beta2.CheckBulkRequest{
		Items: []*v1beta2.CheckBulkRequestItem{item1, item2, item3},
	}

	checkBulkResponse, err := inventoryClient.CheckBulk(ctx, checkBulkRequest)
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

	fmt.Println("CheckBulk response received successfully")
	fmt.Printf("Total pairs in response: %d\n\n", len(checkBulkResponse.Pairs))

	for idx, pair := range checkBulkResponse.Pairs {
		fmt.Printf("--- Result %d ---\n", idx+1)
		req := pair.Request
		fmt.Printf("Request: subject=%s relation=%s object=%s\n",
			req.Subject.Resource.ResourceId,
			req.Relation,
			req.Object.ResourceId,
		)

		if item := pair.GetItem(); item != nil {
			fmt.Printf("%v\n", item)
		} else if err := pair.GetError(); err != nil {
			fmt.Printf("Error: Code=%d, Message=%s\n", err.Code, err.Message)
		}
	}
}

func main() {
	checkBulk()
}
