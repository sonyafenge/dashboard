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

package daemonset

import (
	metricapi "github.com/CentaurusInfra/dashboard/src/app/backend/integration/metric/api"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/common"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sClient "k8s.io/client-go/kubernetes"
	"log"
)

// DaemonSeDetail represents detailed information about a Daemon Set.
type DaemonSetDetail struct {
	// Extends list item structure.
	DaemonSet `json:",inline"`

	LabelSelector *v1.LabelSelector `json:"labelSelector,omitempty"`

	// List of non-critical errors, that occurred during resource retrieval.
	Errors []error `json:"errors"`
}

// Returns detailed information about the given daemon set in the given namespace.
func GetDaemonSetDetail(client k8sClient.Interface, metricClient metricapi.MetricClient,
	namespace, name string) (*DaemonSetDetail, error) {

	log.Printf("Getting details of %s daemon set in %s namespace", name, namespace)
	daemonSet, err := client.AppsV1().DaemonSetsWithMultiTenancy(namespace, "").Get(name, metaV1.GetOptions{})
	if err != nil {
		return nil, err
	}

	channels := &common.ResourceChannels{
		EventList: common.GetEventListChannel(client, common.NewSameNamespaceQuery(namespace), 1),
		PodList:   common.GetPodListChannel(client, common.NewSameNamespaceQuery(namespace), 1),
	}

	eventList := <-channels.EventList.List
	if err := <-channels.EventList.Error; err != nil {
		return nil, err
	}

	podList := <-channels.PodList.List
	if err := <-channels.PodList.Error; err != nil {
		return nil, err
	}

	return &DaemonSetDetail{
		DaemonSet:     toDaemonSet(*daemonSet, podList.Items, eventList.Items),
		LabelSelector: daemonSet.Spec.Selector,
		Errors:        []error{},
	}, nil
}

// Returns detailed information about the given daemon set in the given namespace.
func GetDaemonSetDetailWithMultiTenancy(client k8sClient.Interface, metricClient metricapi.MetricClient, tenant,
	namespace, name string) (*DaemonSetDetail, error) {

	log.Printf("Getting details of %s daemon set in %s namespace", name, namespace)
	daemonSet, err := client.AppsV1().DaemonSetsWithMultiTenancy(namespace, tenant).Get(name, metaV1.GetOptions{})
	if err != nil {
		return nil, err
	}

	channels := &common.ResourceChannels{
		EventList: common.GetEventListChannelWithMultiTenancy(client, tenant, common.NewSameNamespaceQuery(namespace), 1),
		PodList:   common.GetPodListChannelWithMultiTenancy(client, tenant, common.NewSameNamespaceQuery(namespace), 1),
	}

	eventList := <-channels.EventList.List
	if err := <-channels.EventList.Error; err != nil {
		return nil, err
	}

	podList := <-channels.PodList.List
	if err := <-channels.PodList.Error; err != nil {
		return nil, err
	}

	return &DaemonSetDetail{
		DaemonSet:     toDaemonSet(*daemonSet, podList.Items, eventList.Items),
		LabelSelector: daemonSet.Spec.Selector,
		Errors:        []error{},
	}, nil
}
