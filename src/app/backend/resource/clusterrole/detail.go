package clusterrole

import (
	rbac "k8s.io/api/rbac/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sClient "k8s.io/client-go/kubernetes"
)

// ClusterRoleDetail contains Cron Job details.
type ClusterRoleDetail struct {
	// Extends list item structure.
	ClusterRole `json:",inline"`

	Rules []rbac.PolicyRule `json:"rules"`

	// List of non-critical errors, that occurred during resource retrieval.
	Errors []error `json:"errors"`
}

// GetClusterRoleDetail gets Cluster Role details.
func GetClusterRoleDetail(client k8sClient.Interface, name string) (*ClusterRoleDetail, error) {
	rawObject, err := client.RbacV1().ClusterRolesWithMultiTenancy("").Get(name, metaV1.GetOptions{})
	if err != nil {
		return nil, err
	}

	cr := toClusterRoleDetail(*rawObject)
	return &cr, nil
}

// GetClusterRoleDetailWithMultiTenancy gets Cluster Role details.
func GetClusterRoleDetailWithMultiTenancy(client k8sClient.Interface, tenant string, name string) (*ClusterRoleDetail, error) {
	rawObject, err := client.RbacV1().ClusterRolesWithMultiTenancy(tenant).Get(name, metaV1.GetOptions{})
	if err != nil {
		return nil, err
	}

	cr := toClusterRoleDetail(*rawObject)
	return &cr, nil
}

func toClusterRoleDetail(cr rbac.ClusterRole) ClusterRoleDetail {
	return ClusterRoleDetail{
		ClusterRole: toClusterRole(cr),
		Rules:       cr.Rules,
		Errors:      []error{},
	}
}
