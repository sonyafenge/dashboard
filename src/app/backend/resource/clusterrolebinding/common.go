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

package clusterrolebinding

import (
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/dataselect"
	v1 "k8s.io/api/rbac/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
)

// ClusterRoleBindingSpec is a specification of CLusterRoleBinding to create.
type ClusterRoleBindingSpec struct {
	// Name of the cluster-role-binding.
	Name string `json:"name"`

	// Tenant of the cluster-role-binding
	Tenant string `json:"tenant"`

	// Subject contains a reference to the object or user identities a role binding applies to.  This can either hold a direct API object reference,
	// or a value for non-objects such as user and group names.
	Subject v1.Subject `json:"subject"`

	// RoleRef contains information that points to the role being used
	RoleRef v1.RoleRef `json:"role_ref"`
}

// CreateClusterRoleBindings creates Cluster-role-binding based on given specification.
func CreateClusterRoleBindings(spec *ClusterRoleBindingSpec, client kubernetes.Interface) error {
	log.Printf("Creating Cluster-role-binding %s", spec.Name)

	var subjects []v1.Subject
	subject := v1.Subject{
		Kind:      spec.Subject.Kind,
		APIGroup:  spec.Subject.APIGroup,
		Name:      spec.Subject.Name,
		Namespace: spec.Subject.Namespace,
	}
	subjects = append(subjects, subject)

	roleRef := v1.RoleRef{
		APIGroup: spec.RoleRef.APIGroup,
		Kind:     spec.RoleRef.Kind,
		Name:     spec.RoleRef.Name,
	}

	clusterrolebinding := &v1.ClusterRoleBinding{
		ObjectMeta: metaV1.ObjectMeta{
			Name: spec.Name,
		},
		Subjects: subjects,
		RoleRef:  roleRef,
	}

	_, err := client.RbacV1().ClusterRoleBindingsWithMultiTenancy(spec.Tenant).Create(clusterrolebinding)
	return err
}

// DeleteClusterRoleBindings deletes clusterrolebinding based on given specification.
func DeleteClusterRoleBindings(clusterrolebindingName string, client kubernetes.Interface) error {
	log.Printf("Deleting clusterrolebinding %s", clusterrolebindingName)
	err := client.RbacV1().ClusterRoleBindings().Delete(clusterrolebindingName, &metaV1.DeleteOptions{})
	return err
}

// CreateClusterRoleBindingsWithMultiTenancy creates Cluster-role-binding based on given specification.
func CreateClusterRoleBindingsWithMultiTenancy(spec *ClusterRoleBindingSpec, client kubernetes.Interface) error {
	log.Printf("Creating Cluster-role-binding %s", spec.Name)

	var subjects []v1.Subject
	subject := v1.Subject{
		Kind:      spec.Subject.Kind,
		APIGroup:  spec.Subject.APIGroup,
		Name:      spec.Subject.Name,
		Namespace: spec.Subject.Namespace,
	}
	subjects = append(subjects, subject)

	roleRef := v1.RoleRef{
		APIGroup: spec.RoleRef.APIGroup,
		Kind:     spec.RoleRef.Kind,
		Name:     spec.RoleRef.Name,
	}

	clusterrolebinding := &v1.ClusterRoleBinding{
		ObjectMeta: metaV1.ObjectMeta{
			Name:   spec.Name,
			Tenant: spec.Tenant,
		},
		Subjects: subjects,
		RoleRef:  roleRef,
	}

	_, err := client.RbacV1().ClusterRoleBindingsWithMultiTenancy(spec.Tenant).Create(clusterrolebinding)
	return err
}

// DeleteClusterRoleBindingsWithMultiTenancy deletes clusterrolebinding based on given specification.
func DeleteClusterRoleBindingsWithMultiTenancy(tenantName string, clusterrolebindingName string, client kubernetes.Interface) error {
	log.Printf("Deleting clusterrolebinding %s", clusterrolebindingName)
	err := client.RbacV1().ClusterRoleBindingsWithMultiTenancy(tenantName).Delete(clusterrolebindingName, &metaV1.DeleteOptions{})
	return err
}

// The code below allows to perform complex data section on []ClusterRoleBinding

type ClusterRoleBindingCell ClusterRoleBinding

func (self ClusterRoleBindingCell) GetProperty(name dataselect.PropertyName) dataselect.ComparableValue {
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

func toCells(std []ClusterRoleBinding) []dataselect.DataCell {
	cells := make([]dataselect.DataCell, len(std))
	for i := range std {
		cells[i] = ClusterRoleBindingCell(std[i])
	}
	return cells
}

func fromCells(cells []dataselect.DataCell) []ClusterRoleBinding {
	std := make([]ClusterRoleBinding, len(cells))
	for i := range std {
		std[i] = ClusterRoleBinding(cells[i].(ClusterRoleBindingCell))
	}
	return std
}
