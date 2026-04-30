package console

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	v1beta2 "github.com/project-kessel/kessel-sdk-go/kessel/inventory/v1beta2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrincipalFromRHIdentity(t *testing.T) {
	tests := []struct {
		name           string
		identity       map[string]any
		domain         string
		expectedResID  string
		expectedErrMsg string
	}{
		{
			name: "user identity",
			identity: map[string]any{
				"type":   "User",
				"org_id": "12345",
				"user":   map[string]any{"user_id": "7393748", "username": "jdoe"},
			},
			expectedResID: "redhat/7393748",
		},
		{
			name: "service account identity",
			identity: map[string]any{
				"type":   "ServiceAccount",
				"org_id": "456",
				"service_account": map[string]any{
					"user_id":  "sa-001",
					"username": "service-account-sa-001",
				},
			},
			expectedResID: "redhat/sa-001",
		},
		{
			name: "custom domain",
			identity: map[string]any{
				"type": "User",
				"user": map[string]any{"user_id": "42"},
			},
			domain:        "customdomain",
			expectedResID: "customdomain/42",
		},
		{
			name: "unsupported type",
			identity: map[string]any{
				"type": "System",
			},
			expectedErrMsg: `unsupported identity type: "System"`,
		},
		{
			name:           "nil identity",
			expectedErrMsg: "identity must not be nil",
		},
		{
			name: "user type missing user field",
			identity: map[string]any{
				"type": "User",
			},
			expectedErrMsg: `identity type "User" is missing the "user" field`,
		},
		{
			name: "service account type missing service_account field",
			identity: map[string]any{
				"type": "ServiceAccount",
			},
			expectedErrMsg: `identity type "ServiceAccount" is missing the "service_account" field`,
		},
		{
			name: "user type empty user_id",
			identity: map[string]any{
				"type": "User",
				"user": map[string]any{"username": "jdoe"},
			},
			expectedErrMsg: "unable to resolve user ID from User identity",
		},
		{
			name: "service account type empty user_id",
			identity: map[string]any{
				"type":            "ServiceAccount",
				"service_account": map[string]any{"username": "sa"},
			},
			expectedErrMsg: "unable to resolve user ID from ServiceAccount identity",
		},
		{
			name: "user field not a map",
			identity: map[string]any{
				"type": "User",
				"user": "not-a-map",
			},
			expectedErrMsg: `identity type "User" is missing the "user" field`,
		},
		{
			name: "missing type field",
			identity: map[string]any{
				"org_id": "123",
			},
			expectedErrMsg: "unsupported identity type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var subj *v1beta2.SubjectReference
			var err error

			if tt.domain != "" {
				subj, err = PrincipalFromRHIdentity(tt.identity, tt.domain)
			} else {
				subj, err = PrincipalFromRHIdentity(tt.identity)
			}

			if tt.expectedErrMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedResID, subj.Resource.ResourceId)
		})
	}
}

func TestPrincipalFromRHIdentity_SubjectFields(t *testing.T) {
	identity := map[string]any{
		"type": "User",
		"user": map[string]any{"user_id": "7393748"},
	}

	subj, err := PrincipalFromRHIdentity(identity)
	require.NoError(t, err)

	assert.Equal(t, "principal", subj.Resource.ResourceType)
	assert.Equal(t, "redhat/7393748", subj.Resource.ResourceId)
	assert.Equal(t, "rbac", subj.Resource.Reporter.Type)
	assert.Nil(t, subj.Relation)
}

func TestPrincipalFromRHIdentity_CustomDomainFields(t *testing.T) {
	identity := map[string]any{
		"type": "User",
		"user": map[string]any{"user_id": "42"},
	}

	subj, err := PrincipalFromRHIdentity(identity, "customdomain")
	require.NoError(t, err)

	assert.Equal(t, "customdomain/42", subj.Resource.ResourceId)
}

func TestPrincipalFromRHIdentityHeader(t *testing.T) {
	tests := []struct {
		name           string
		header         string
		domain         string
		expectedResID  string
		expectedErrMsg string
	}{
		{
			name:          "valid user header",
			header:        encodeHeader(t, map[string]any{"identity": map[string]any{"type": "User", "user": map[string]any{"user_id": "123"}}}),
			expectedResID: "redhat/123",
		},
		{
			name:          "valid header with custom domain",
			header:        encodeHeader(t, map[string]any{"identity": map[string]any{"type": "User", "user": map[string]any{"user_id": "456"}}}),
			domain:        "example",
			expectedResID: "example/456",
		},
		{
			name:           "malformed base64",
			header:         "not-valid-base64!@#$",
			expectedErrMsg: "failed to decode identity header",
		},
		{
			name:           "invalid JSON",
			header:         base64.StdEncoding.EncodeToString([]byte("not json")),
			expectedErrMsg: "failed to decode identity header",
		},
		{
			name:           "missing identity envelope key",
			header:         encodeHeader(t, map[string]any{"other": "data"}),
			expectedErrMsg: `identity header is missing the "identity" envelope key`,
		},
		{
			name:           "unsupported type in header",
			header:         encodeHeader(t, map[string]any{"identity": map[string]any{"type": "X509"}}),
			expectedErrMsg: `unsupported identity type: "X509"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var subj *v1beta2.SubjectReference
			var err error

			if tt.domain != "" {
				subj, err = PrincipalFromRHIdentityHeader(tt.header, tt.domain)
			} else {
				subj, err = PrincipalFromRHIdentityHeader(tt.header)
			}

			if tt.expectedErrMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedResID, subj.Resource.ResourceId)
		})
	}
}

func TestPrincipalFromRHIdentityHeader_SubjectFields(t *testing.T) {
	header := encodeHeader(t, map[string]any{
		"identity": map[string]any{
			"type":   "ServiceAccount",
			"org_id": "org-1",
			"service_account": map[string]any{
				"user_id":  "sa-999",
				"username": "service-account-sa-999",
			},
		},
	})

	subj, err := PrincipalFromRHIdentityHeader(header)
	require.NoError(t, err)

	assert.Equal(t, "principal", subj.Resource.ResourceType)
	assert.Equal(t, "redhat/sa-999", subj.Resource.ResourceId)
	assert.Equal(t, "rbac", subj.Resource.Reporter.Type)
	assert.Nil(t, subj.Relation)
}

func encodeHeader(t *testing.T, v any) string {
	t.Helper()
	data, err := json.Marshal(v)
	require.NoError(t, err)
	return base64.StdEncoding.EncodeToString(data)
}
