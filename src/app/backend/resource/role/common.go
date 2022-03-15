// Copyright 2017 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package role

import (
	"log"

	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	v1 "k8s.io/api/rbac/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// RoleSpec is a specification of role to create.
type RoleSpec struct {
	// Name of the role.
	Name string `json:"name"`

	// Namespace of the role.
	Namespace string `json:"namespace"`

	// Tenant of the role
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

// CreateRole creates Role based on given specification.
func CreateRole(spec *RoleSpec, client kubernetes.Interface) error {
	log.Printf("Creating role %s", spec.Name)

	// setting default namespace if no namespace is specified
	if spec.Namespace == "" {
		spec.Namespace = "default"
	}

	var policies []v1.PolicyRule
	policy := v1.PolicyRule{
		Verbs:           spec.Verbs,
		APIGroups:       spec.APIGroups,
		Resources:       spec.Resources,
		ResourceNames:   spec.ResourceNames,
		NonResourceURLs: spec.NonResourceURLs,
	}
	policies = append(policies, policy)
	role := &v1.Role{
		ObjectMeta: metaV1.ObjectMeta{
			Name: spec.Name,
		},
		Rules: policies,
	}

	_, err := client.RbacV1().Roles(spec.Namespace).Create(role)
	return err
}

// DeleteRole deletes role based on given specification.
func DeleteRole(namespace string, roleName string, client kubernetes.Interface) error {
	log.Printf("Deleting role %s", roleName)
	err := client.RbacV1().Roles(namespace).Delete(roleName, &metaV1.DeleteOptions{})
	return err
}

// CreateRolesWithMultiTenancy creates Role based on given specification.
func CreateRolesWithMultiTenancy(spec *RoleSpec, client kubernetes.Interface) error {
	log.Printf("Creating role %s", spec.Name)

	// setting default namespace if no namespace is specified
	if spec.Namespace == "" {
		spec.Namespace = "default"
	}

	var policies []v1.PolicyRule
	policy := v1.PolicyRule{
		Verbs:           spec.Verbs,
		APIGroups:       spec.APIGroups,
		Resources:       spec.Resources,
		ResourceNames:   spec.ResourceNames,
		NonResourceURLs: spec.NonResourceURLs,
	}
	policies = append(policies, policy)
	role := &v1.Role{
		ObjectMeta: metaV1.ObjectMeta{
			Name:   spec.Name,
			Tenant: spec.Tenant,
		},
		Rules: policies,
	}

	_, err := client.RbacV1().RolesWithMultiTenancy(spec.Namespace, spec.Tenant).Create(role)
	return err
}

// DeleteRolesWithMultiTenancy deletes role based on given specification.
func DeleteRolesWithMultiTenancy(tenant string, namespace string, roleName string, client kubernetes.Interface) error {
	log.Printf("Deleting role %s", roleName)
	err := client.RbacV1().RolesWithMultiTenancy(namespace, tenant).Delete(roleName, &metaV1.DeleteOptions{})
	return err
}

// The code below allows to perform complex data section on []Role

type RoleCell Role

func (self RoleCell) GetProperty(name dataselect.PropertyName) dataselect.ComparableValue {
	switch name {
	case dataselect.NameProperty:
		return dataselect.StdComparableString(self.ObjectMeta.Name)
	case dataselect.CreationTimestampProperty:
		return dataselect.StdComparableTime(self.ObjectMeta.CreationTimestamp.Time)
	case dataselect.NamespaceProperty:
		return dataselect.StdComparableString(self.ObjectMeta.Namespace)
	default:
		// if name is not supported then just return a constant dummy value, sort will have no effect.
		return nil
	}
}

func toCells(std []Role) []dataselect.DataCell {
	cells := make([]dataselect.DataCell, len(std))
	for i := range std {
		cells[i] = RoleCell(std[i])
	}
	return cells
}

func fromCells(cells []dataselect.DataCell) []Role {
	std := make([]Role, len(cells))
	for i := range std {
		std[i] = Role(cells[i].(RoleCell))
	}
	return std
}
