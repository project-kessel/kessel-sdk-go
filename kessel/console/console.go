package console

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	v1beta2 "github.com/project-kessel/kessel-sdk-go/kessel/inventory/v1beta2"
	v2 "github.com/project-kessel/kessel-sdk-go/kessel/rbac/v2"
)

const defaultDomain = "redhat"

var identityTypeFields = map[string]string{
	"User":           "user",
	"ServiceAccount": "service_account",
}

func extractUserID(identity map[string]any) (string, error) {
	if identity == nil {
		return "", fmt.Errorf("identity must not be nil")
	}

	identityType, _ := identity["type"].(string)
	field, ok := identityTypeFields[identityType]
	if !ok {
		supported := make([]string, 0, len(identityTypeFields))
		for k := range identityTypeFields {
			supported = append(supported, k)
		}
		return "", fmt.Errorf("unsupported identity type: %q (supported: %v)", identityType, supported)
	}

	details, ok := identity[field].(map[string]any)
	if !ok {
		return "", fmt.Errorf("identity type %q is missing the %q field", identityType, field)
	}

	userID, _ := details["user_id"].(string)
	if userID == "" {
		return "", fmt.Errorf("unable to resolve user ID from %s identity (tried: user_id)", identityType)
	}

	return userID, nil
}

func PrincipalFromRHIdentity(identity map[string]any, domain ...string) (*v1beta2.SubjectReference, error) {
	d := defaultDomain
	if len(domain) > 0 {
		d = domain[0]
	}

	userID, err := extractUserID(identity)
	if err != nil {
		return nil, err
	}

	return v2.PrincipalSubject(userID, d), nil
}

func PrincipalFromRHIdentityHeader(header string, domain ...string) (*v1beta2.SubjectReference, error) {
	decoded, err := base64.StdEncoding.DecodeString(header)
	if err != nil {
		return nil, fmt.Errorf("failed to decode identity header: %w", err)
	}

	var envelope map[string]any
	if err := json.Unmarshal(decoded, &envelope); err != nil {
		return nil, fmt.Errorf("failed to decode identity header: %w", err)
	}

	identity, ok := envelope["identity"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("identity header is missing the \"identity\" envelope key")
	}

	return PrincipalFromRHIdentity(identity, domain...)
}
