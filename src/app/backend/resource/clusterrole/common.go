package clusterrole

import (
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	v1 "k8s.io/api/rbac/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
)

type ClusterRoleSpec struct {
	// Name of the cluster-role.
	Name string `json:"name"`

	// Tenant of the cluster-role
	Tenant string `json:"tenant"`

	// Verbs is a list of Verbs that apply to ALL the ResourceKinds and AttributeRestrictions contained in this rule.  VerbAll represents all kinds.
	Verbs []string `json:"verbs" protobuf:"bytes,1,rep,name=verbs"`

	// APIGroups is the name of the APIGroup that contains the resources.  If multiple API groups are specified, any action requested against one of
	// the enumerated resources in any API group will be allowed.
	// +optional
	APIGroups []string `json:"apiGroups,omitempty" protobuf:"bytes,2,rep,name=apiGroups"`
	// Resources is a list of resources this rule applies to.  ResourceAll represents all resources.
	// +optional
	Resources []string `json:"resources,omitempty" protobuf:"bytes,3,rep,name=resources"`
	// ResourceNames is an optional white list of names that the rule applies to.  An empty set means that everything is allowed.
	// +optional
	ResourceNames []string `json:"resourceNames,omitempty" protobuf:"bytes,4,rep,name=resourceNames"`

	// NonResourceURLs is a set of partial urls that a user should have access to.  *s are allowed, but only as the full, final step in the path
	// Since non-resource URLs are not namespaced, this field is only applicable for ClusterRoles referenced from a ClusterRoleBinding.
	// Rules can either apply to API resources (such as "pods" or "secrets") or non-resource URL paths (such as "/api"),  but not both.
	// +optional
	NonResourceURLs []string `json:"nonResourceURLs,omitempty" protobuf:"bytes,5,rep,name=nonResourceURLs"`
}

// CreateClusterRole creates Cluster-role based on given specification.
func CreateClusterRole(spec *ClusterRoleSpec, client kubernetes.Interface) error {
	log.Printf("Creating Cluster-role %s", spec.Name)

	var policies []v1.PolicyRule
	policy := v1.PolicyRule{
		Verbs:           spec.Verbs,
		APIGroups:       spec.APIGroups,
		Resources:       spec.Resources,
		ResourceNames:   spec.ResourceNames,
		NonResourceURLs: spec.NonResourceURLs,
	}
	policies = append(policies, policy)
	clusterrole := &v1.ClusterRole{
		ObjectMeta: metaV1.ObjectMeta{
			Name: spec.Name,
		},
		Rules: policies,
	}

	_, err := client.RbacV1().ClusterRoles().Create(clusterrole)
	return err
}

// CreateClusterRolesWithMultiTenancy creates Cluster-role based on given specification.
func CreateClusterRolesWithMultiTenancy(spec *ClusterRoleSpec, client kubernetes.Interface) error {
	log.Printf("Creating Cluster-role %s", spec.Name)

	var policies []v1.PolicyRule
	policy := v1.PolicyRule{
		Verbs:           spec.Verbs,
		APIGroups:       spec.APIGroups,
		Resources:       spec.Resources,
		ResourceNames:   spec.ResourceNames,
		NonResourceURLs: spec.NonResourceURLs,
	}
	policies = append(policies, policy)
	clusterrole := &v1.ClusterRole{
		ObjectMeta: metaV1.ObjectMeta{
			Name:   spec.Name,
			Tenant: spec.Tenant,
		},
		Rules: policies,
	}

	_, err := client.RbacV1().ClusterRolesWithMultiTenancy(spec.Tenant).Create(clusterrole)
	return err
}

func DeleteClusterRole(clusterroleName string, client kubernetes.Interface) error {
	log.Printf("Deleting clusterrole %s", clusterroleName)
	err := client.RbacV1().ClusterRoles().Delete(clusterroleName, &metaV1.DeleteOptions{})
	return err
}

// The code below allows to perform complex data section on []ClusterRole

type RoleCell ClusterRole

func (self RoleCell) GetProperty(name dataselect.PropertyName) dataselect.ComparableValue {
	switch name {
	case dataselect.NameProperty:
		return dataselect.StdComparableString(self.ObjectMeta.Name)
	case dataselect.CreationTimestampProperty:
		return dataselect.StdComparableTime(self.ObjectMeta.CreationTimestamp.Time)
	case dataselect.NamespaceProperty:
		return dataselect.StdComparableString(self.ObjectMeta.Namespace)
	default:
		return nil
	}
}

func toCells(std []ClusterRole) []dataselect.DataCell {
	cells := make([]dataselect.DataCell, len(std))
	for i := range std {
		cells[i] = RoleCell(std[i])
	}
	return cells
}

func fromCells(cells []dataselect.DataCell) []ClusterRole {
	std := make([]ClusterRole, len(cells))
	for i := range std {
		std[i] = ClusterRole(cells[i].(RoleCell))
	}
	return std
}
