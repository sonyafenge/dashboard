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
	"github.com/kubernetes/dashboard/src/app/backend/api"
	"github.com/kubernetes/dashboard/src/app/backend/errors"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	v1 "k8s.io/api/core/v1"
	client "k8s.io/client-go/kubernetes"
	"log"
)

type Tenant struct {
	ObjectMeta api.ObjectMeta `json:"objectMeta"`
	TypeMeta   api.TypeMeta   `json:"typeMeta"`

	// Phase is the current lifecycle phase of the tenant
	Phase       v1.TenantPhase `json:"phase"`
	ClusterName string         `json:"clusterName"`
}

// TenantList contains a list of tenants in the cluster.
type TenantList struct {
	ListMeta api.ListMeta `json:"listMeta"`
	Tenants  []Tenant     `json:"tenants"`

	// List of non-critical errors, that occurred during resource retrieval.
	Errors []error `json:"errors"`
}

func GetTenantList(client client.Interface, dsQuery *dataselect.DataSelectQuery, clusterName string, tenant string) (*TenantList, error) {
	log.Println("Getting list of tenants")
	tenants, err := client.CoreV1().Tenants().List(api.ListEverything)

	nonCriticalErrors, criticalError := errors.HandleError(err)
	if criticalError != nil {
		return nil, criticalError
	}
	return toTenantList(tenants.Items, nonCriticalErrors, dsQuery, clusterName, tenant), nil
}

func toTenantList(tenants []v1.Tenant, nonCriticalErrors []error, dsQuery *dataselect.DataSelectQuery, clusterName string, tenantName string) *TenantList {
	tenantList := &TenantList{
		ListMeta: api.ListMeta{TotalItems: len(tenants)},
		Tenants:  make([]Tenant, 0),
	}

	tenantCells, filteredTotal := dataselect.GenericDataSelectWithFilter(toCells(tenants), dsQuery)
	tenants = fromCells(tenantCells)
	tenantList.ListMeta = api.ListMeta{TotalItems: filteredTotal}
	tenantList.Errors = nonCriticalErrors

	for _, tenant := range tenants {
		if tenantName == "system" {
			tenant.ClusterName = clusterName
			tenantList.Tenants = append(tenantList.Tenants, toTenant(tenant))
		} else {
			if tenantName == tenant.Name {
				tenant.ClusterName = clusterName
				tenantList.Tenants = append(tenantList.Tenants, toTenant(tenant))
			}
		}
	}
	return tenantList
}

func toTenant(tenant v1.Tenant) Tenant {
	return Tenant{
		ObjectMeta: api.NewObjectMeta(tenant.ObjectMeta),
		TypeMeta:   api.NewTypeMeta(api.ResourceKindTenant),
		Phase:      tenant.Status.Phase,
	}
}
