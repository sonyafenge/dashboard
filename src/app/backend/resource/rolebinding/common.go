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

package rolebinding

import (
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/dataselect"
	v1 "k8s.io/api/rbac/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
)

// RoleBindingSpec is a specification of RoleBinding to create.
type RoleBindingSpec struct {
	// Name of the role-binding.
	Name string `json:"name"`
	// Namespace for which role binding is to be created.
	Namespace string `json:"namespace"`
	// Tenant of the role-binding
	Tenant string `json:"tenant"`
	// Subject contains a reference to the object or user identities a role binding applies to.  This can either hold a direct API object reference,
	// or a value for non-objects such as user and group names.
	Subject v1.Subject `json:"subject"`
	// RoleRef contains information that points to the role being used
	RoleRef v1.RoleRef `json:"role_ref"`
}

// CreateRoleBindings creates role-binding based on given specification.
func CreateRoleBindings(spec *RoleBindingSpec, client kubernetes.Interface) error {
	log.Printf("Creating Role-binding %s", spec.Name)

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

	rolebinding := &v1.RoleBinding{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      spec.Name,
			Namespace: spec.Namespace,
			Tenant:    spec.Tenant,
		},
		Subjects: subjects,
		RoleRef:  roleRef,
	}
	_, err := client.RbacV1().RoleBindings(spec.Namespace).Create(rolebinding)
	return err
}

// DeleteRoleBindings deletes role-binding based on given specification.
func DeleteRoleBindings(namespaceName string, rolebindingName string, client kubernetes.Interface) error {
	log.Printf("Deleting Rolebinding %s", rolebindingName)
	err := client.RbacV1().RoleBindings(namespaceName).Delete(rolebindingName, &metaV1.DeleteOptions{})
	return err
}

// CreateRoleBindingsWithMultiTenancy creates role-binding based on given specification.
func CreateRoleBindingsWithMultiTenancy(spec *RoleBindingSpec, client kubernetes.Interface) error {
	log.Printf("Creating Role-binding %v", spec)

	var subjects []v1.Subject
	subject := v1.Subject{
		Kind:      spec.Subject.Kind,
		APIGroup:  "",
		Name:      spec.Subject.Name,
		Namespace: spec.Subject.Namespace,
	}
	subjects = append(subjects, subject)

	roleRef := v1.RoleRef{
		APIGroup: spec.RoleRef.APIGroup,
		Kind:     spec.RoleRef.Kind,
		Name:     spec.RoleRef.Name,
	}

	rolebinding := &v1.RoleBinding{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      spec.Name,
			Namespace: spec.Namespace,
			Tenant:    spec.Tenant,
		},
		Subjects: subjects,
		RoleRef:  roleRef,
	}
	_, err := client.RbacV1().RoleBindingsWithMultiTenancy(spec.Namespace, spec.Tenant).Create(rolebinding)
	return err
}

// DeleteRoleBindingsWithMultiTenancy deletes role-binding based on given specification.
func DeleteRoleBindingsWithMultiTenancy(tenantName string, namespaceName string, rolebindingName string, client kubernetes.Interface) error {
	log.Printf("Deleting Rolebinding %s", rolebindingName)
	err := client.RbacV1().RoleBindingsWithMultiTenancy(namespaceName, tenantName).Delete(rolebindingName, &metaV1.DeleteOptions{})
	return err
}

// The code below allows to perform complex data section on []RoleBinding

type RoleBindingCell RoleBinding

func (self RoleBindingCell) GetProperty(name dataselect.PropertyName) dataselect.ComparableValue {
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

func toCells(std []RoleBinding) []dataselect.DataCell {
	cells := make([]dataselect.DataCell, len(std))
	for i := range std {
		cells[i] = RoleBindingCell(std[i])
	}
	return cells
}

func fromCells(cells []dataselect.DataCell) []RoleBinding {
	std := make([]RoleBinding, len(cells))
	for i := range std {
		std[i] = RoleBinding(cells[i].(RoleBindingCell))
	}
	return std
}
