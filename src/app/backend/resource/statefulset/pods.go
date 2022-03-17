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

package statefulset

import (
	"log"

	"github.com/CentaurusInfra/dashboard/src/app/backend/errors"
	metricapi "github.com/CentaurusInfra/dashboard/src/app/backend/integration/metric/api"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/common"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/dataselect"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/event"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/pod"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetStatefulSetPods return list of pods targeting pet set.
func GetStatefulSetPods(client kubernetes.Interface, metricClient metricapi.MetricClient,
	dsQuery *dataselect.DataSelectQuery, name, namespace string) (*pod.PodList, error) {

	log.Printf("Getting replication controller %s pods in namespace %s", name, namespace)

	pods, err := getRawStatefulSetPods(client, name, namespace)
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

// getRawStatefulSetPods return array of api pods targeting pet set with given name.
func getRawStatefulSetPods(client kubernetes.Interface, name, namespace string) ([]v1.Pod, error) {
	statefulSet, err := client.AppsV1().StatefulSetsWithMultiTenancy(namespace, "").Get(name, metaV1.GetOptions{})
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

	return common.FilterPodsByControllerRef(statefulSet, podList.Items), nil
}

// getRawStatefulSetPodsWithMultiTenancy return array of api pods targeting pet set with given name.
func getRawStatefulSetPodsWithMultiTenancy(client kubernetes.Interface, tenant, name, namespace string) ([]v1.Pod, error) {
	statefulSet, err := client.AppsV1().StatefulSetsWithMultiTenancy(namespace, tenant).Get(name, metaV1.GetOptions{})
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

	return common.FilterPodsByControllerRef(statefulSet, podList.Items), nil
}

// Returns simple info about pods(running, desired, failing, etc.) related to given pet set.
func getStatefulSetPodInfo(client kubernetes.Interface, statefulSet *apps.StatefulSet) (*common.PodInfo, error) {
	pods, err := getRawStatefulSetPods(client, statefulSet.Name, statefulSet.Namespace)
	if err != nil {
		return nil, err
	}

	podInfo := common.GetPodInfo(statefulSet.Status.Replicas, statefulSet.Spec.Replicas, pods)
	return &podInfo, nil
}

// Returns simple info about pods(running, desired, failing, etc.) related to given pet set.
func getStatefulSetPodInfoWithMultiTenancy(client kubernetes.Interface, tenant string, statefulSet *apps.StatefulSet) (*common.PodInfo, error) {
	pods, err := getRawStatefulSetPodsWithMultiTenancy(client, tenant, statefulSet.Name, statefulSet.Namespace)
	if err != nil {
		return nil, err
	}

	podInfo := common.GetPodInfo(statefulSet.Status.Replicas, statefulSet.Spec.Replicas, pods)
	return &podInfo, nil
}
