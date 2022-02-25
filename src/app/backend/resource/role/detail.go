package role

import (
	rbac "k8s.io/api/rbac/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sClient "k8s.io/client-go/kubernetes"
)

// RoleDetail contains Role details.
type RoleDetail struct {
	// Extends list item structure.
	Role `json:",inline"`

	Rules []rbac.PolicyRule `json:"rules"`

	// List of non-critical errors, that occurred during resource retrieval.
	Errors []error `json:"errors"`
}

// GetRoleDetail gets Role details.
func GetRoleDetail(client k8sClient.Interface, namespace string, name string) (*RoleDetail, error) {
	rawObject, err := client.RbacV1().Roles(namespace).Get(name, metaV1.GetOptions{})
	if err != nil {
		return nil, err
	}

	cr := toRoleDetail(*rawObject)
	return &cr, nil
}

// GetRoleDetailWithMultiTenancy gets Role details.
func GetRoleDetailWithMultiTenancy(client k8sClient.Interface, tenant string, namespace string, name string) (*RoleDetail, error) {
	rawObject, err := client.RbacV1().RolesWithMultiTenancy(namespace, tenant).Get(name, metaV1.GetOptions{})
	if err != nil {
		return nil, err
	}

	cr := toRoleDetail(*rawObject)
	return &cr, nil
}

func toRoleDetail(cr rbac.Role) RoleDetail {
	return RoleDetail{
		Role:   toRole(cr),
		Rules:  cr.Rules,
		Errors: []error{},
	}
}
