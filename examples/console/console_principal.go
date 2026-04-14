package main

import (
	"encoding/base64"
	"encoding/json"
	"log"

	"github.com/project-kessel/kessel-sdk-go/kessel/console"
)

func consolePrincipal() {
	// --- From a parsed User identity ---
	userIdentity := map[string]any{
		"type":   "User",
		"org_id": "12345",
		"user":   map[string]any{"user_id": "7393748", "username": "jdoe"},
	}

	subject, err := console.PrincipalFromRHIdentity(userIdentity)
	if err != nil {
		panic(err)
	}
	log.Printf("User principal:            %s", subject.Resource.ResourceId)

	// --- From a parsed ServiceAccount identity ---
	saIdentity := map[string]any{
		"type":   "ServiceAccount",
		"org_id": "456",
		"service_account": map[string]any{
			"user_id":  "12345",
			"username": "service-account-b69eaf9e",
		},
	}

	subject, err = console.PrincipalFromRHIdentity(saIdentity)
	if err != nil {
		panic(err)
	}
	log.Printf("ServiceAccount principal:  %s", subject.Resource.ResourceId)

	// --- From a raw base64-encoded x-rh-identity header ---
	headerPayload := map[string]any{
		"identity": map[string]any{
			"type":   "User",
			"org_id": "12345",
			"user":   map[string]any{"user_id": "7393748", "username": "jdoe"},
		},
	}

	data, err := json.Marshal(headerPayload)
	if err != nil {
		panic(err)
	}
	header := base64.StdEncoding.EncodeToString(data)

	subject, err = console.PrincipalFromRHIdentityHeader(header)
	if err != nil {
		panic(err)
	}
	log.Printf("From header principal:     %s", subject.Resource.ResourceId)
}

func main() { consolePrincipal() }
