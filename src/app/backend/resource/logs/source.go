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

package logs

import (
	"github.com/CentaurusInfra/dashboard/src/app/backend/api"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/common"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/controller"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetLogSources returns all log sources for a given resource. A log source identifies a log file through the combination of pod & container
func GetLogSources(k8sClient kubernetes.Interface, ns string, resourceName string, resourceType string) (controller.LogSources, error) {
	if resourceType == "pod" {
		return getLogSourcesFromPod(k8sClient, ns, resourceName)
	}
	return getLogSourcesFromController(k8sClient, ns, resourceName, resourceType)
}

// GetLogSourcesWithMultiTenancy returns all log sources for a given resource. A log source identifies a log file through the combination of pod & container
func GetLogSourcesWithMultiTenancy(k8sClient kubernetes.Interface, tenant string, ns string, resourceName string, resourceType string) (controller.LogSources, error) {
	if resourceType == "pod" {
		return getLogSourcesFromPodWithMultiTenancy(k8sClient, tenant, ns, resourceName)
	}
	return getLogSourcesFromControllerWithMultiTenancy(k8sClient, tenant, ns, resourceName, resourceType)
}

// GetLogSourcesFromPod returns all containers for a given pod
func getLogSourcesFromPod(k8sClient kubernetes.Interface, ns, resourceName string) (controller.LogSources, error) {
	pod, err := k8sClient.CoreV1().PodsWithMultiTenancy(ns, "").Get(resourceName, meta.GetOptions{})
	if err != nil {
		return controller.LogSources{}, err
	}
	return controller.LogSources{
		ContainerNames:     common.GetContainerNames(&pod.Spec),
		InitContainerNames: common.GetInitContainerNames(&pod.Spec),
		PodNames:           []string{resourceName},
	}, nil
}

// getLogSourcesFromPodWithMultiTenancy returns all containers for a given pod
func getLogSourcesFromPodWithMultiTenancy(k8sClient kubernetes.Interface, tenant, ns, resourceName string) (controller.LogSources, error) {
	pod, err := k8sClient.CoreV1().PodsWithMultiTenancy(ns, tenant).Get(resourceName, meta.GetOptions{})
	if err != nil {
		return controller.LogSources{}, err
	}
	return controller.LogSources{
		ContainerNames:     common.GetContainerNames(&pod.Spec),
		InitContainerNames: common.GetInitContainerNames(&pod.Spec),
		PodNames:           []string{resourceName},
	}, nil
}

// GetLogSourcesFromController returns all pods and containers for a controller object, such as ReplicaSet
func getLogSourcesFromController(k8sClient kubernetes.Interface, ns, resourceName, resourceType string) (controller.LogSources, error) {
	ref := meta.OwnerReference{Kind: resourceType, Name: resourceName}
	rc, err := controller.NewResourceController(ref, ns, k8sClient)
	if err != nil {
		return controller.LogSources{}, err
	}
	allPods, err := k8sClient.CoreV1().PodsWithMultiTenancy(ns, "").List(api.ListEverything)
	if err != nil {
		return controller.LogSources{}, err
	}
	return rc.GetLogSources(allPods.Items), nil
}

// getLogSourcesFromControllerWithMultiTenancy returns all pods and containers for a controller object, such as ReplicaSet
func getLogSourcesFromControllerWithMultiTenancy(k8sClient kubernetes.Interface, tenant, ns, resourceName, resourceType string) (controller.LogSources, error) {
	ref := meta.OwnerReference{Kind: resourceType, Name: resourceName}
	rc, err := controller.NewResourceController(ref, ns, k8sClient)
	if err != nil {
		return controller.LogSources{}, err
	}
	allPods, err := k8sClient.CoreV1().PodsWithMultiTenancy(ns, tenant).List(api.ListEverything)
	if err != nil {
		return controller.LogSources{}, err
	}
	return rc.GetLogSources(allPods.Items), nil
}
