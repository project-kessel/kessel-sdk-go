package main

import (
	"context"
	"fmt"
	"log"

	kesselclient "github.com/project-kessel/kessel-sdk-go/pkg/kessel/client"
	"github.com/project-kessel/kessel-sdk-go/pkg/kessel/config"
	v1beta2 "github.com/project-kessel/kessel-sdk-go/pkg/kessel/inventory/v1beta2"
)

func main() {
	//cfg := &config.Config{
	//	Endpoint:       "127.0.0.1:8000",
	//	Timeout:        30 * time.Second,
	//	UseHTTP:        true,  // Set to false for gRPC
	//	Insecure:       false, // For testing only
	//	EnableOIDCAuth: false,
	//}

	cfg := config.NewConfig(
		config.WithTLSInsecure(true),
		config.WithHttp(true),
		config.WithEndpoint("127.0.0.1:8000"))

	client, err := kesselclient.NewClient(context.Background(), cfg)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	//Use builder to create the gGPRC or HTTP client
	//grpcBuilder:= grpc.NewClientBuilder("127.0.0.1:9000").WithInsecure()
	//kesselclient.NewGRPCClient(context.Background(), cfg, grpcBuilder)

	defer func(client kesselclient.KesselClient) {
		err := client.Close()
		if err != nil {
			fmt.Printf("failed to close kessel client: %v", err)
		}
	}(client)

	res, err := client.Check(context.Background(), &v1beta2.CheckRequest{
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
	})
	if err != nil {
		log.Fatalf("Failed to get check: %v", err)
	}
	log.Println(res)
}
