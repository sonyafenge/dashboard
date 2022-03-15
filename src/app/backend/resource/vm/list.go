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

package vm

import (
	"k8s.io/apimachinery/pkg/types"
	"log"

	"github.com/kubernetes/dashboard/src/app/backend/api"
	"github.com/kubernetes/dashboard/src/app/backend/errors"
	metricapi "github.com/kubernetes/dashboard/src/app/backend/integration/metric/api"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	"github.com/kubernetes/dashboard/src/app/backend/resource/event"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	time "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sClient "k8s.io/client-go/kubernetes"
)

// VMListComponent contains a list of Pods in the cluster.
type VMList struct {
	ListMeta          api.ListMeta       `json:"listMeta"`
	CumulativeMetrics []metricapi.Metric `json:"cumulativeMetrics"`

	// Basic information about resources status on the list.
	Status common.ResourceStatus `json:"status"`

	// Unordered list of Pods.
	Pods []VM `json:"virtualMachines"`

	// List of non-critical errors, that occurred during resource retrieval.
	Errors []error `json:"errors"`
}

type VirtualMachineStatus struct {
	Status          string              `json:"status"`
	PodPhase        v1.PodPhase         `json:"podPhase"`
	ContainerStates []v1.ContainerState `json:"containerStates"`
}

// Pod is a presentation layer view of Kubernetes Pod resource. This means it is Pod plus additional augmented data
// we can get from other sources (like services that target it).
type VM struct {
	ObjectMeta VirtualMachine `json:"objectMeta"`
	TypeMeta   api.TypeMeta   `json:"typeMeta"`

	// More info on pod status
	PodStatus VirtualMachineStatus `json:"podStatus"`

	// Count of containers restarts.
	RestartCount int32 `json:"restartCount"`

	// Pod metrics.
	//Metrics *PodMetrics `json:"metrics"`

	// Pod warning events
	Warnings []common.Event `json:"warnings"`

	// Name of the Node this Pod runs on.
	NodeName string `json:"nodeName"`
}

var EmptyVMList = &VMList{
	Pods:   make([]VM, 0),
	Errors: make([]error, 0),
	ListMeta: api.ListMeta{
		TotalItems: 0,
	},
}

// GetVMListWithMultiTenancy returns a list of all Pods in the cluster with multi tenancy support.
func GetVMListWithMultiTenancy(client k8sClient.Interface, metricClient metricapi.MetricClient, tenant string, nsQuery *common.NamespaceQuery,
	dsQuery *dataselect.DataSelectQuery) (*VMList, error) {
	log.Print("Getting list of all pods in the cluster")

	channels := &common.ResourceChannels{
		VMList:    common.GetVMListChannelWithMultiTenancyAndOptions(client, tenant, nsQuery, metaV1.ListOptions{}, 1),
		EventList: common.GetEventListChannelWithMultiTenancy(client, tenant, nsQuery, 1),
	}

	return GetVMListFromChannels(channels, dsQuery, metricClient)
}

// GetVMListFromChannels returns a list of all Pods in the cluster
// reading required resource list once from the channels.
func GetVMListFromChannels(channels *common.ResourceChannels, dsQuery *dataselect.DataSelectQuery,
	metricClient metricapi.MetricClient) (*VMList, error) {

	pods := <-channels.VMList.List
	err := <-channels.VMList.Error
	nonCriticalErrors, criticalError := errors.HandleError(err)
	if criticalError != nil {
		return nil, criticalError
	}

	eventList := <-channels.EventList.List
	err = <-channels.EventList.Error
	nonCriticalErrors, criticalError = errors.AppendError(err, nonCriticalErrors)
	if criticalError != nil {
		return nil, criticalError
	}

	VMList := ToVirtualMachineList(pods.Items, eventList.Items, nonCriticalErrors, dsQuery, metricClient)
	VMList.Status = getStatus(pods, eventList.Items)
	return &VMList, nil
}

func GetVolumeListFromChannels(channels *common.ResourceChannels, dsQuery *dataselect.DataSelectQuery,
	metricClient metricapi.MetricClient) (*VMList, error) {

	pods := <-channels.VMList.List
	err := <-channels.VMList.Error
	nonCriticalErrors, criticalError := errors.HandleError(err)
	if criticalError != nil {
		return nil, criticalError
	}

	eventList := <-channels.EventList.List
	err = <-channels.EventList.Error
	nonCriticalErrors, criticalError = errors.AppendError(err, nonCriticalErrors)
	if criticalError != nil {
		return nil, criticalError
	}

	VMList := ToVirtualMachineList(pods.Items, eventList.Items, nonCriticalErrors, dsQuery, metricClient)
	VMList.Status = getStatus(pods, eventList.Items)
	return &VMList, nil
}

func ToVirtualMachineList(pods []v1.Pod, events []v1.Event, nonCriticalErrors []error, dsQuery *dataselect.DataSelectQuery,
	metricClient metricapi.MetricClient) VMList {
	VMList := VMList{
		Pods:   make([]VM, 0),
		Errors: nonCriticalErrors,
	}
	vmPods := make([]v1.Pod, 0)
	for _, pod := range pods {
		if pod.Spec.VirtualMachine != nil {
			warnings := event.GetPodsEventWarnings(events, []v1.Pod{pod})
			podDetail := toVM(&pod, warnings)
			VMList.Pods = append(VMList.Pods, podDetail)
			vmPods = append(vmPods, pod)
		}

	}

	podCells, cumulativeMetricsPromises, filteredTotal := dataselect.
		GenericDataSelectWithFilterAndMetrics(toCells(vmPods), dsQuery, metricapi.NoResourceCache, metricClient)
	vmPods = fromCells(podCells)
	VMList.ListMeta = api.ListMeta{TotalItems: filteredTotal}

	//metrics, err := getMetricsPerPod(pods, metricClient, dsQuery)
	//if err != nil {
	//  log.Printf("Skipping metrics because of error: %s\n", err)
	//}

	//for _, pod := range pods {
	//  if pod.Spec.VirtualMachine != nil {
	//    warnings := event.GetPodsEventWarnings(events, []v1.Pod{pod})
	//    podDetail := toVM(&pod, warnings)
	//    VMList.Pods = append(VMList.Pods, podDetail)
	//  }
	//
	//}

	cumulativeMetrics, err := cumulativeMetricsPromises.GetMetrics()
	VMList.CumulativeMetrics = cumulativeMetrics
	if err != nil {
		VMList.CumulativeMetrics = make([]metricapi.Metric, 0)
	}

	return VMList
}

func ToVMList(pods []v1.Pod, events []v1.Event, nonCriticalErrors []error, dsQuery *dataselect.DataSelectQuery,
	metricClient metricapi.MetricClient) VMList {
	VMList := VMList{
		Pods:   make([]VM, 0),
		Errors: nonCriticalErrors,
	}

	podCells, cumulativeMetricsPromises, filteredTotal := dataselect.
		GenericDataSelectWithFilterAndMetrics(toCells(pods), dsQuery, metricapi.NoResourceCache, metricClient)
	pods = fromCells(podCells)
	VMList.ListMeta = api.ListMeta{TotalItems: filteredTotal}

	//metrics, err := getMetricsPerPod(pods, metricClient, dsQuery)
	//if err != nil {
	//  log.Printf("Skipping metrics because of error: %s\n", err)
	//}

	for _, pod := range pods {
		if pod.Spec.VirtualMachine != nil {
			warnings := event.GetPodsEventWarnings(events, []v1.Pod{pod})
			podDetail := toVM(&pod, warnings)
			VMList.Pods = append(VMList.Pods, podDetail)
		}
	}

	cumulativeMetrics, err := cumulativeMetricsPromises.GetMetrics()
	VMList.CumulativeMetrics = cumulativeMetrics
	if err != nil {
		VMList.CumulativeMetrics = make([]metricapi.Metric, 0)
	}

	return VMList
}

func toVM(pod *v1.Pod, warnings []common.Event) VM {
	podDetail := VM{
		ObjectMeta:   NewObjectMeta(pod.ObjectMeta, pod.Spec),
		TypeMeta:     api.NewTypeMeta(api.ResourceKindVirtualMachine),
		Warnings:     warnings,
		PodStatus:    getPodStatus(*pod, warnings),
		RestartCount: getRestartCount(*pod),
		NodeName:     pod.Spec.NodeName,
	}

	return podDetail
}

type VirtualMachine struct {
	// Name of the container specified as a DNS_LABEL.
	// Each container in a pod must have a unique name (DNS_LABEL).
	// Cannot be updated.
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// Image name.
	// More info: https://kubernetes.io/docs/concepts/containers/images
	// This field is optional to allow higher level config management to default or override
	// container images in workload controllers like Deployments and StatefulSets.
	// +optional
	Image string `json:"image,omitempty" protobuf:"bytes,2,opt,name=image"`
	// Compute Resources required by this container.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
	// +optional
	Resources v1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,3,opt,name=resources"`
	// Node compute resources allocated to the container.
	// +optional
	ResourcesAllocated v1.ResourceList `json:"resourcesAllocated,omitempty" protobuf:"bytes,15,rep,name=resourcesAllocated,casttype=ResourceList,castkey=ResourceName"`
	// Resources resize policy for the container.
	// +optional
	ResizePolicy []v1.ResizePolicy `json:"resizePolicy,omitempty" protobuf:"bytes,16,rep,name=resizePolicy"`
	// Pod volumes to mount into the workload's filesystem.
	// Cannot be updated.
	// +optional
	// +patchMergeKey=mountPath
	// +patchStrategy=merge
	VolumeMounts []v1.VolumeMount `json:"volumeMounts,omitempty" patchStrategy:"merge" patchMergeKey:"mountPath" protobuf:"bytes,4,rep,name=volumeMounts"`
	// Image pull policy.
	// One of Always, Never, IfNotPresent.
	// Defaults to Always if :latest tag is specified, or IfNotPresent otherwise.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/containers/images#updating-images
	// +optional
	ImagePullPolicy v1.PullPolicy `json:"imagePullPolicy,omitempty" protobuf:"bytes,5,opt,name=imagePullPolicy,casttype=PullPolicy"`
	// either keyPair or the publicKeystring must be provided, used to logon to the VM
	// +optional
	KeyPairName string `json:"keyPairName,omitempty" protobuf:"bytes,6,opt,name=keyPairName"`
	// +optional
	PublicKey string `json:"publicKey,omitempty" protobuf:"bytes,7,opt,name=publicKey"`
	// Configuration information or scripts to use upon launch. Must be Base64 encoded. Restricted to 65535 bytes.
	// +optional
	UserData []byte `json:"userData,omitempty" protobuf:"bytes,8,opt,name=userData"`
	// default none, array cert ID that used to verify the image
	// +optional
	TrustedImageCertificate []string `json:"trustedImageCertificate,omitempty" protobuf:"bytes,9,opt,name=trustedImageCertificate"`
	// stop | terminate VM. default to stop
	// +optional
	ShutdownBehavior string `json:"shutdownBehavior,omitempty" protobuf:"bytes,10,opt,name=shutdownBehavior"`
	// if not specified, the first volume in the volume slice will be used
	// +optional
	BootVolume string `json:"bootVolume,omitempty" protobuf:"bytes,11,opt,name=bootVolume"`
	// default running
	// +optional
	PowerSpec v1.VmPowerSpec `json:"powerSpec,omitempty" protobuf:"bytes,12,opt,name=powerSpec,casttype=VmPowerSpec"`
	// volumeDevices is the list of block devices to be used by the container.
	// This is a beta feature.
	// +patchMergeKey=devicePath
	// +patchStrategy=merge
	// +optional
	VolumeDevices []v1.VolumeDevice `json:"volumeDevices,omitempty" patchStrategy:"merge" patchMergeKey:"devicePath" protobuf:"bytes,13,rep,name=volumeDevices"`
	// cloud-init user data script
	// +optional
	CloudInitUserDataScript string `json:"cloudInitUserDataScript,omitempty" protobuf:"bytes,14,opt,name=cloudInitUserDataScript"`
	// Namespace defines the space within which name must be unique. An empty namespace is
	// equivalent to the "default" namespace, but "default" is the canonical representation.
	// Not all objects are required to be scoped to a namespace - the value of this field for
	// those objects will be empty.
	Namespace string `json:"namespace,omitempty"`

	// Labels are key value pairs that may be used to scope and select individual resources.
	// Label keys are of the form:
	//     label-key ::= prefixed-name | name
	//     prefixed-name ::= prefix '/' name
	//     prefix ::= DNS_SUBDOMAIN
	//     name ::= DNS_LABEL
	// The prefix is optional.  If the prefix is not specified, the key is assumed to be private
	// to the user.  Other system components that wish to use labels must specify a prefix.
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations are unstructured key value data stored with a resource that may be set by
	// external tooling. They are not queryable and should be preserved when modifying
	// objects.  Annotation keys have the same formatting restrictions as Label keys. See the
	// comments on Labels for details.
	Annotations map[string]string `json:"annotations,omitempty"`

	// CreationTimestamp is a timestamp representing the server time when this object was
	// created. It is not guaranteed to be set in happens-before order across separate operations.
	// Clients may not set this value. It is represented in RFC3339 form and is in UTC.
	CreationTimestamp time.Time `json:"creationTimestamp,omitempty"`

	// UID is a type that holds unique ID values, including UUIDs.  Because we
	// don't ONLY use UUIDs, this is an alias to string.  Being a type captures
	// intent and helps make sure that UIDs and names do not get conflated.
	UID types.UID `json:"uid,omitempty"`
}

func NewObjectMeta(k8SObjectMeta metaV1.ObjectMeta, spec v1.PodSpec) VirtualMachine {
	return VirtualMachine{
		Name:                    k8SObjectMeta.Name,
		Image:                   spec.VirtualMachine.Image,
		Resources:               spec.VirtualMachine.Resources,
		ResourcesAllocated:      spec.VirtualMachine.ResourcesAllocated,
		ResizePolicy:            spec.VirtualMachine.ResizePolicy,
		VolumeMounts:            spec.VirtualMachine.VolumeMounts,
		ImagePullPolicy:         spec.VirtualMachine.ImagePullPolicy,
		KeyPairName:             spec.VirtualMachine.KeyPairName,
		PublicKey:               spec.VirtualMachine.PublicKey,
		UserData:                spec.VirtualMachine.UserData,
		TrustedImageCertificate: spec.VirtualMachine.TrustedImageCertificate,
		ShutdownBehavior:        spec.VirtualMachine.ShutdownBehavior,
		BootVolume:              spec.VirtualMachine.BootVolume,
		PowerSpec:               spec.VirtualMachine.PowerSpec,
		VolumeDevices:           spec.VirtualMachine.VolumeDevices,
		CloudInitUserDataScript: spec.VirtualMachine.CloudInitUserDataScript,
		Namespace:               k8SObjectMeta.Namespace,
		Labels:                  k8SObjectMeta.Labels,
		CreationTimestamp:       k8SObjectMeta.CreationTimestamp,
		Annotations:             k8SObjectMeta.Annotations,
		UID:                     k8SObjectMeta.UID,
	}
}

//Name:              ,
//Namespace:
