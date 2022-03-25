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
	"github.com/CentaurusInfra/dashboard/src/app/backend/errors"
	metricapi "github.com/CentaurusInfra/dashboard/src/app/backend/integration/metric/api"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/common"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/dataselect"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/event"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/pod"
	api "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sClient "k8s.io/client-go/kubernetes"
	"log"
)

// GetDaemonSetPods return list of pods targeting daemon set.
func GetDaemonSetPods(client k8sClient.Interface, metricClient metricapi.MetricClient,
	dsQuery *dataselect.DataSelectQuery, daemonSetName, namespace string) (*pod.PodList, error) {
	log.Printf("Getting replication controller %s pods in namespace %s", daemonSetName, namespace)

	pods, err := getRawDaemonSetPods(client, daemonSetName, namespace)
	if err != nil {
		return pod.EmptyPodList, err
	}

	events, err := event.GetPodsEvents(client, namespace, pods)
	nonCriticalErrors, criticalError := errors.HandleError(err)
	if criticalError != nil {
		return nil, criticalError
	}

	podList := pod.ToPodList(pods, events, nonCriticalErrors, dsQuery, metricClient)
	return &podList, nil
}

// GetDaemonSetPodsWithMultiTenancy return list of pods targeting daemon set.
func GetDaemonSetPodsWithMultiTenancy(client k8sClient.Interface, metricClient metricapi.MetricClient,
	dsQuery *dataselect.DataSelectQuery, tenant, daemonSetName, namespace string) (*pod.PodList, error) {
	log.Printf("Getting replication controller %s pods in namespace %s for %s", daemonSetName, namespace, tenant)

	pods, err := getRawDaemonSetPodsWithMultiTenancy(client, tenant, daemonSetName, namespace)
	if err != nil {
		return pod.EmptyPodList, err
	}

	events, err := event.GetPodsEventsWithMultiTenancy(client, tenant, namespace, pods)
	nonCriticalErrors, criticalError := errors.HandleError(err)
	if criticalError != nil {
		return nil, criticalError
	}

	podList := pod.ToPodList(pods, events, nonCriticalErrors, dsQuery, metricClient)
	return &podList, nil
}

// Returns array of api pods targeting daemon set with given name.
func getRawDaemonSetPods(client k8sClient.Interface, daemonSetName, namespace string) ([]api.Pod, error) {
	daemonSet, err := client.AppsV1().DaemonSetsWithMultiTenancy(namespace, "").Get(daemonSetName, metaV1.GetOptions{})
	if err != nil {
		return nil, err
	}

	channels := &common.ResourceChannels{
		PodList: common.GetPodListChannel(client, common.NewSameNamespaceQuery(namespace), 1),
	}

	podList := <-channels.PodList.List
	if err := <-channels.PodList.Error; err != nil {
		return nil, err
	}

	matchingPods := common.FilterPodsByControllerRef(daemonSet, podList.Items)
	return matchingPods, nil
}

// Returns array of api pods targeting daemon set with given name.
func getRawDaemonSetPodsWithMultiTenancy(client k8sClient.Interface, tenant, daemonSetName, namespace string) ([]api.Pod, error) {
	daemonSet, err := client.AppsV1().DaemonSetsWithMultiTenancy(namespace, tenant).Get(daemonSetName, metaV1.GetOptions{})
	if err != nil {
		return nil, err
	}

	channels := &common.ResourceChannels{
		PodList: common.GetPodListChannelWithMultiTenancy(client, tenant, common.NewSameNamespaceQuery(namespace), 1),
	}

	podList := <-channels.PodList.List
	if err := <-channels.PodList.Error; err != nil {
		return nil, err
	}

	matchingPods := common.FilterPodsByControllerRef(daemonSet, podList.Items)
	return matchingPods, nil
}
