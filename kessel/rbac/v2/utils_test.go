package v2

import (
	"testing"

	"github.com/stretchr/testify/assert"

	v1beta2 "github.com/project-kessel/kessel-sdk-go/kessel/inventory/v1beta2"
)

func TestWorkspaceType(t *testing.T) {
	rt := WorkspaceType()

	assert.NotNil(t, rt)
	assert.Equal(t, "workspace", rt.ResourceType)
	assert.NotNil(t, rt.ReporterType)
	assert.Equal(t, "rbac", *rt.ReporterType)
}

func TestRoleType(t *testing.T) {
	rt := RoleType()

	assert.NotNil(t, rt)
	assert.Equal(t, "role", rt.ResourceType)
	assert.NotNil(t, rt.ReporterType)
	assert.Equal(t, "rbac", *rt.ReporterType)
}

func TestPrincipalResource(t *testing.T) {
	tests := []struct {
		name            string
		id              string
		domain          string
		expectedResId   string
		expectedResType string
	}{
		{
			name:            "basic principal",
			id:              "user123",
			domain:          "redhat.com",
			expectedResId:   "redhat.com/user123",
			expectedResType: "principal",
		},
		{
			name:            "principal with numeric id",
			id:              "12345",
			domain:          "example.org",
			expectedResId:   "example.org/12345",
			expectedResType: "principal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := PrincipalResource(tt.id, tt.domain)

			assert.NotNil(t, ref)
			assert.Equal(t, tt.expectedResType, ref.ResourceType)
			assert.Equal(t, tt.expectedResId, ref.ResourceId)
			assert.NotNil(t, ref.Reporter)
			assert.Equal(t, "rbac", ref.Reporter.Type)
		})
	}
}

func TestRoleResource(t *testing.T) {
	ref := RoleResource("admin")

	assert.NotNil(t, ref)
	assert.Equal(t, "role", ref.ResourceType)
	assert.Equal(t, "admin", ref.ResourceId)
	assert.NotNil(t, ref.Reporter)
	assert.Equal(t, "rbac", ref.Reporter.Type)
}

func TestWorkspaceResource(t *testing.T) {
	ref := WorkspaceResource("ws-123")

	assert.NotNil(t, ref)
	assert.Equal(t, "workspace", ref.ResourceType)
	assert.Equal(t, "ws-123", ref.ResourceId)
	assert.NotNil(t, ref.Reporter)
	assert.Equal(t, "rbac", ref.Reporter.Type)
}

func TestPrincipalSubject(t *testing.T) {
	subj := PrincipalSubject("user123", "redhat.com")

	assert.NotNil(t, subj)
	assert.NotNil(t, subj.Resource)
	assert.Equal(t, "principal", subj.Resource.ResourceType)
	assert.Equal(t, "redhat.com/user123", subj.Resource.ResourceId)
	assert.Nil(t, subj.Relation)
}

func TestSubject(t *testing.T) {
	tests := []struct {
		name             string
		resourceRef      *v1beta2.ResourceReference
		relation         string
		expectRelation   bool
		expectedRelation string
	}{
		{
			name:             "subject with relation",
			resourceRef:      WorkspaceResource("ws-123"),
			relation:         "member",
			expectRelation:   true,
			expectedRelation: "member",
		},
		{
			name:           "subject without relation",
			resourceRef:    PrincipalResource("user123", "redhat.com"),
			relation:       "",
			expectRelation: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subj := Subject(tt.resourceRef, tt.relation)

			assert.NotNil(t, subj)
			assert.NotNil(t, subj.Resource)

			if tt.expectRelation {
				assert.NotNil(t, subj.Relation)
				assert.Equal(t, tt.expectedRelation, *subj.Relation)
			} else {
				assert.Nil(t, subj.Relation)
			}
		})
	}
}
