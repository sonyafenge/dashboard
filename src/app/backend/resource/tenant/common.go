// Copyright 2020 Authors of Arktos.
// Copyright 2020 Authors of Arktos - file modified.
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

package tenant

import (
	"log"

	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/dataselect"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// TenantSpec is a specification of tenant to create.
type TenantSpec struct {
	// Name of the tenant.
	Name             string `json:"name"`
	StorageClusterId string `json:"storageclusterid"` // Storage Cluster Id of the Tenant
	Username         string `json:"username"`         // Username of the Tenant
	Password         string `json:"password"`         // Password of the Tenant
}

// CreateTenant creates tenant based on given specification.
func CreateTenant(spec *TenantSpec, client kubernetes.Interface, clusterName string) error {
	log.Printf("Creating tenant %s", spec.Name)

	// setting default values if no values passed
	if spec.StorageClusterId == "" {
		spec.StorageClusterId = "0"
	}
	clusterLabel := make(map[string]string)
	clusterLabel["clusterName"] = clusterName
	tenant := &v1.Tenant{
		ObjectMeta: metaV1.ObjectMeta{
			Name:   spec.Name,
			Labels: clusterLabel,
		},
		Spec: v1.TenantSpec{

			StorageClusterId: spec.StorageClusterId,
		},
	}

	_, err := client.CoreV1().Tenants().Create(tenant)
	return err
}

// DeleteTenant deletes tenant based on given specification.
func DeleteTenant(tenantName string, client kubernetes.Interface) error {
	log.Printf("Deleting Tenant %s", tenantName)
	err := client.CoreV1().Tenants().Delete(tenantName, &metaV1.DeleteOptions{})
	return err
}

// The code below allows to perform complex data section on []api.Tenant
type TenantCell v1.Tenant

func (self TenantCell) GetProperty(name dataselect.PropertyName) dataselect.ComparableValue {
	switch name {
	case dataselect.NameProperty:
		return dataselect.StdComparableString(self.ObjectMeta.Name)
	case dataselect.CreationTimestampProperty:
		return dataselect.StdComparableTime(self.ObjectMeta.CreationTimestamp.Time)
	default:
		return nil
	}
}

func toCells(std []v1.Tenant) []dataselect.DataCell {
	cells := make([]dataselect.DataCell, len(std))
	for i := range std {
		cells[i] = TenantCell(std[i])
	}
	return cells
}

func fromCells(cells []dataselect.DataCell) []v1.Tenant {
	std := make([]v1.Tenant, len(cells))
	for i := range std {
		std[i] = v1.Tenant(cells[i].(TenantCell))
	}
	return std
}
