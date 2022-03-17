// Copyright 2017 The Kubernetes Authors.
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

package resourcequota

import (
	"github.com/CentaurusInfra/dashboard/src/app/backend/api"
	"github.com/CentaurusInfra/dashboard/src/app/backend/errors"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/common"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/dataselect"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sClient "k8s.io/client-go/kubernetes"
	"log"
)

type ResourceQuotaSpec struct {
	// Name of the resource quota.
	Name                     string `json:"name"`
	Tenant                   string `json:"tenant"`
	NameSpace                string `json:"namespace"`
	ResourceCPU              string `json:"cpu"`
	ResourceMemory           string `json:"memory"`
	ResourcePods             string `json:"pods"`
	ResourceConfigMaps       string `json:"config_maps"`
	ResourcePVC              string `json:"pvc"`
	ResourceSecrets          string `json:"secrets"`
	ResourceServices         string `json:"services"`
	ResourceStorage          string `json:"storage"`
	ResourceEphemeralStorage string `json:"ephemeral_storage"`
}

// ResourceStatus provides the status of the resource defined by a resource quota.
type ResourceStatus struct {
	Used string `json:"used,omitempty"`
	Hard string `json:"hard,omitempty"`
}

func toResourceQuota(resourcequota v1.ResourceQuota) ResourceQuota {
	return ResourceQuota{
		ObjectMeta: api.NewObjectMeta(resourcequota.ObjectMeta),
		TypeMeta:   api.NewTypeMeta(api.ResourceKindResourceQuota),
	}
}

type ResourceQuota struct {
	ObjectMeta api.ObjectMeta `json:"objectMeta"`
	TypeMeta   api.TypeMeta   `json:"typeMeta"`
}

// ResourceQuotaDetail provides the presentation layer view of Kubernetes Resource Quotas resource.
type ResourceQuotaDetail struct {
	ResourceQuota `json:",inline"`

	ObjectMeta api.ObjectMeta `json:"objectMeta"`
	TypeMeta   api.TypeMeta   `json:"typeMeta"`

	// Scopes defines quota scopes
	Scopes []v1.ResourceQuotaScope `json:"scopes,omitempty"`

	// StatusList is a set of (resource name, Used, Hard) tuple.
	StatusList map[v1.ResourceName]ResourceStatus `json:"statusList,omitempty"`

	// List of non-critical errors, that occurred during resource retrieval.
	Errors []error `json:"errors"`
}

// ResourceQuotaDetailList
type ResourceQuotaDetailList struct {
	ListMeta api.ListMeta          `json:"listMeta"`
	Items    []ResourceQuotaDetail `json:"items"`
}

func AddResourceQuotas(client k8sClient.Interface, namespace string, tenant string, spec *ResourceQuotaSpec) (*v1.ResourceQuota, error) {
	if tenant == "" {
		tenant = "default"
	}
	ns, err := client.CoreV1().NamespacesWithMultiTenancy(tenant).Get(namespace, metaV1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var resList = make(v1.ResourceList)
	if spec.ResourceMemory != "" {
		resList[v1.ResourceMemory] = resource.MustParse(spec.ResourceMemory)
	}
	if spec.ResourceCPU != "" {
		resList[v1.ResourceCPU] = resource.MustParse(spec.ResourceCPU)
	}
	if spec.ResourceConfigMaps != "" {
		resList[v1.ResourceConfigMaps] = resource.MustParse(spec.ResourceConfigMaps)
	}
	if spec.ResourceMemory != "" {
		resList[v1.ResourceMemory] = resource.MustParse(spec.ResourceMemory)
	}
	if spec.ResourcePVC != "" {
		resList[v1.ResourcePersistentVolumeClaims] = resource.MustParse(spec.ResourcePVC)
	}
	if spec.ResourcePods != "" {
		resList[v1.ResourcePods] = resource.MustParse(spec.ResourcePods)
	}
	if spec.ResourceServices != "" {
		resList[v1.ResourceServices] = resource.MustParse(spec.ResourceServices)
	}
	if spec.ResourceSecrets != "" {
		resList[v1.ResourceSecrets] = resource.MustParse(spec.ResourceSecrets)
	}
	if spec.ResourceStorage != "" {
		resList[v1.ResourceStorage] = resource.MustParse(spec.ResourceStorage)
	}
	if spec.ResourceEphemeralStorage != "" {
		resList[v1.ResourceEphemeralStorage] = resource.MustParse(spec.ResourceEphemeralStorage)
	}
	if spec.Tenant == "" {
		spec.Tenant = tenant
	}
	if spec.NameSpace == "" {
		spec.NameSpace = namespace
	}
	if spec.Name == "" {
		err := errors.NewInternal("empty resource-quota name error")
		return nil, err
	}
	resQuota, err := client.CoreV1().ResourceQuotasWithMultiTenancy(namespace, ns.Tenant).Create(&v1.ResourceQuota{
		TypeMeta: metaV1.TypeMeta{},
		ObjectMeta: metaV1.ObjectMeta{
			Name:      spec.Name,
			Tenant:    spec.Tenant,
			Namespace: spec.NameSpace,
		},

		Spec: v1.ResourceQuotaSpec{
			Hard:          resList,
			Scopes:        nil,
			ScopeSelector: nil,
		},
		Status: v1.ResourceQuotaStatus{},
	})
	if err != nil {
		return nil, err
	}

	return resQuota, nil
}

// DeleteResourceQuota

func DeleteResourceQuota(client k8sClient.Interface, namespace string, tenant string, name string) error {
	if tenant == "" {
		tenant = "default"
	}
	ns, err := client.CoreV1().NamespacesWithMultiTenancy(tenant).Get(namespace, metaV1.GetOptions{})
	if err != nil {
		return nil
	}

	err = client.CoreV1().ResourceQuotasWithMultiTenancy(namespace, ns.Tenant).Delete(name, &metaV1.DeleteOptions{})
	if err != nil {
		return nil
	}
	return nil
}

func GetResourceQuotaList(client k8sClient.Interface, namespace *common.NamespaceQuery, tenant string, dsQuery *dataselect.DataSelectQuery) (*ResourceQuotaDetailList, error) {
	log.Println("Getting list of Resource quotas")
	rqlist, err := client.CoreV1().ResourceQuotasWithMultiTenancy(namespace.ToRequestParam(), tenant).List(metaV1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := &ResourceQuotaDetailList{
		Items:    make([]ResourceQuotaDetail, 0),
		ListMeta: api.ListMeta{TotalItems: len(rqlist.Items)},
	}
	for _, item := range rqlist.Items {
		detail := ToResourceQuotaDetail(&item)
		result.Items = append(result.Items, *detail)
	}
	return result, nil
}

func GetResourceQuotaListsWithMultiTenancy(client k8sClient.Interface, namespace string, tenant string) (*ResourceQuotaDetailList, error) {
	if tenant == "" {
		tenant = "default"
	}
	ns, err := client.CoreV1().NamespacesWithMultiTenancy(tenant).Get(namespace, metaV1.GetOptions{})
	if err != nil {
		return nil, err
	}
	list, err := client.CoreV1().ResourceQuotasWithMultiTenancy(namespace, ns.Tenant).List(metaV1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := &ResourceQuotaDetailList{
		Items:    make([]ResourceQuotaDetail, 0),
		ListMeta: api.ListMeta{TotalItems: len(list.Items)},
	}
	for _, item := range list.Items {
		detail := ToResourceQuotaDetail(&item)
		result.Items = append(result.Items, *detail)
	}
	return result, nil
}

func GetResourceQuotaDetails(client k8sClient.Interface, namespace string, tenant string, name string) (*ResourceQuotaDetail, error) {
	if tenant == "" {
		tenant = "default"
	}
	ns, err := client.CoreV1().NamespacesWithMultiTenancy(tenant).Get(namespace, metaV1.GetOptions{})
	if err != nil {
		return nil, err
	}
	list, err := client.CoreV1().ResourceQuotasWithMultiTenancy(namespace, ns.Tenant).List(metaV1.ListOptions{})
	if err != nil {
		return nil, err
	}
	itemNew := new(ResourceQuotaDetail)

	for _, item := range list.Items {
		if name == item.GetName() {
			detail := ToResourceQuotaDetail(&item)
			itemNew = detail
		}
	}
	return itemNew, nil
}

type ResourceQuotaCell ResourceQuota

func (self ResourceQuotaCell) GetProperty(name dataselect.PropertyName) dataselect.ComparableValue {
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

func toCells(std []ResourceQuota) []dataselect.DataCell {
	cells := make([]dataselect.DataCell, len(std))
	for i := range std {
		cells[i] = ResourceQuotaCell(std[i])
	}
	return cells
}

func fromCells(cells []dataselect.DataCell) []ResourceQuota {
	std := make([]ResourceQuota, len(cells))
	for i := range std {
		std[i] = ResourceQuota(cells[i].(ResourceQuotaCell))
	}
	return std
}

func GetResourceQuotaDetail(client k8sClient.Interface, namespace string, name string) (*ResourceQuotaDetail, error) {
	rawObject, err := client.CoreV1().ResourceQuotas(namespace).Get(name, metaV1.GetOptions{})
	if err != nil {
		return nil, err
	}

	cr := toResourceQuotaDetail(*rawObject)
	return &cr, nil
}

func ToResourceQuotaDetail(rawResourceQuota *v1.ResourceQuota) *ResourceQuotaDetail {
	statusList := make(map[v1.ResourceName]ResourceStatus)

	for key, value := range rawResourceQuota.Status.Hard {
		used := rawResourceQuota.Status.Used[key]
		statusList[key] = ResourceStatus{
			Used: used.String(),
			Hard: value.String(),
		}
	}

	return &ResourceQuotaDetail{
		ObjectMeta: api.NewObjectMeta(rawResourceQuota.ObjectMeta),
		TypeMeta:   api.NewTypeMeta(api.ResourceKindResourceQuota),
		Scopes:     rawResourceQuota.Spec.Scopes,
		StatusList: statusList,
	}
}
func toResourceQuotaDetail(cr v1.ResourceQuota) ResourceQuotaDetail {
	return ResourceQuotaDetail{
		ResourceQuota: toResourceQuota(cr),
		Errors:        []error{},
	}
}
