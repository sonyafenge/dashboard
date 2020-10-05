// Copyright 2020 Authors of Arktos.

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

	"github.com/kubernetes/dashboard/src/app/backend/errors"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sClient "k8s.io/client-go/kubernetes"
)

// TenantDetail is a presentation layer view of Arktos Tenant resource.
type TenantDetail struct {
	// Extends from tenant list.
	Tenant `json:",inline"`

	// List of non-critical errors, that occurred during resource retrieval.
	Errors []error `json:"errors"`
}

// GetTenantDetail gets tenant details.
func GetTenantDetail(client k8sClient.Interface, name string) (*TenantDetail, error) {
	log.Printf("Getting details of %s tenant\n", name)

	tenant, err := client.CoreV1().Tenants().Get(name, metaV1.GetOptions{})
	nonCriticalErrors, criticalError := errors.HandleError(err)
	if criticalError != nil {
		return nil, criticalError
	}

	tenantDetail := toTenantDetail(*tenant, nonCriticalErrors)
	log.Printf("tenantDetail is: %v", tenantDetail)
	return &tenantDetail, nil
}

func toTenantDetail(tenant v1.Tenant, nonCriticalErrors []error) TenantDetail {
	return TenantDetail{
		Tenant: toTenant(tenant),
		Errors: nonCriticalErrors,
	}
}
