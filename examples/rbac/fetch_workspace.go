package main

import (
	"context"
	_ "github.com/joho/godotenv/autoload"
	"github.com/project-kessel/kessel-sdk-go/kessel/auth"
	"log"
	"net/http"
	"os"
)

import "github.com/project-kessel/kessel-sdk-go/kessel/rbac/v2"

func fetchWorkspace() {
	ctx := context.Background()

	discovered, err := auth.FetchOIDCDiscovery(ctx, os.Getenv("AUTH_DISCOVERY_ISSUER_URL"), auth.FetchOIDCDiscoveryOptions{
		HttpClient: nil, // Optionally specify an http client - defaults to http.DefaultClient
	})

	if err != nil {
		panic(err)
	}

	oauthCredentials := auth.NewOAuth2ClientCredentials(os.Getenv("AUTH_CLIENT_ID"), os.Getenv("AUTH_CLIENT_SECRET"), discovered.TokenEndpoint)

	defaultWorkspace, err := v2.FetchDefaultWorkspace(ctx, "http://localhost:8888", "12345", v2.FetchWorkspaceOptions{
		HttpClient: http.DefaultClient,
		Auth: v2.OAuth2AuthRequest(&oauthCredentials, v2.OAuth2AuthRequestOptions{
			HttpClient: http.DefaultClient,
		}),
	})

	if err != nil {
		panic(err)
	}

	log.Printf("Found default Workspace: %+v", defaultWorkspace)

	rootWorkspace, err := v2.FetchRootWorkspace(ctx, "http://localhost:8888", "12345", v2.FetchWorkspaceOptions{
		HttpClient: http.DefaultClient,
		Auth: v2.OAuth2AuthRequest(&oauthCredentials, v2.OAuth2AuthRequestOptions{
			HttpClient: http.DefaultClient,
		}),
	})

	if err != nil {
		panic(err)
	}

	log.Printf("Found root Workspace: %+v", rootWorkspace)
}

func main() { fetchWorkspace() }
