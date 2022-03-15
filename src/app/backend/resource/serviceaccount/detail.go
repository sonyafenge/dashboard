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

package serviceaccount

import (
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	client "k8s.io/client-go/kubernetes"
	"log"
)

// ServiceAccountDetail contains detailed information about a service account.
type ServiceAccountDetail struct {
	ServiceAccount `json:",inline"`
	Errors         []error `json:"errors"`
}

// GetServiceAccountDetail returns detailed information about a service account.
func GetServiceAccountDetail(client client.Interface, namespace string, name string) (*ServiceAccountDetail, error) {
	log.Printf("Getting details of %s service account in %s namespace", name, namespace)

	raw, err := client.CoreV1().ServiceAccounts(namespace).Get(name, metaV1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return getServiceAccountDetail(raw), nil
}

// GetServiceAccountDetailWithMultiTenancy returns detailed information about a service account.
func GetServiceAccountDetailWithMultiTenancy(client client.Interface, tenant string, namespace string, name string) (*ServiceAccountDetail, error) {
	log.Printf("Getting details of %s service account in %s namespace", name, namespace)

	raw, err := client.CoreV1().ServiceAccountsWithMultiTenancy(namespace, tenant).Get(name, metaV1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return getServiceAccountDetail(raw), nil
}

func getServiceAccountDetail(sa *v1.ServiceAccount) *ServiceAccountDetail {
	return &ServiceAccountDetail{
		ServiceAccount: toServiceAccount(sa),
	}
}
