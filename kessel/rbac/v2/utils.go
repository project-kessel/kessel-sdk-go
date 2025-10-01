package v2

import (
	"fmt"

	v1beta2 "github.com/project-kessel/kessel-sdk-go/kessel/inventory/v1beta2"
)

func WorkspaceType() *v1beta2.RepresentationType {
	reporterType := "rbac"
	return &v1beta2.RepresentationType{
		ResourceType: "workspace",
		ReporterType: &reporterType,
	}
}

func RoleType() *v1beta2.RepresentationType {
	reporterType := "rbac"
	return &v1beta2.RepresentationType{
		ResourceType: "role",
		ReporterType: &reporterType,
	}
}

func PrincipalResource(id string, domain string) *v1beta2.ResourceReference {
	return &v1beta2.ResourceReference{
		ResourceType: "principal",
		ResourceId:   fmt.Sprintf("%s/%s", domain, id),
		Reporter: &v1beta2.ReporterReference{
			Type: "rbac",
		},
	}
}

func RoleResource(resourceId string) *v1beta2.ResourceReference {
	return &v1beta2.ResourceReference{
		ResourceType: "role",
		ResourceId:   resourceId,
		Reporter: &v1beta2.ReporterReference{
			Type: "rbac",
		},
	}
}

func WorkspaceResource(resourceId string) *v1beta2.ResourceReference {
	return &v1beta2.ResourceReference{
		ResourceType: "workspace",
		ResourceId:   resourceId,
		Reporter: &v1beta2.ReporterReference{
			Type: "rbac",
		},
	}
}

func PrincipalSubject(id string, domain string) *v1beta2.SubjectReference {
	return &v1beta2.SubjectReference{
		Resource: PrincipalResource(id, domain),
	}
}

func Subject(resourceRef *v1beta2.ResourceReference, relation string) *v1beta2.SubjectReference {
	if relation != "" {
		return &v1beta2.SubjectReference{
			Resource: resourceRef,
			Relation: &relation,
		}
	}
	return &v1beta2.SubjectReference{
		Resource: resourceRef,
	}
}
