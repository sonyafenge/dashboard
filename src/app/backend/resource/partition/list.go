package partition

import (
	"github.com/CentaurusInfra/dashboard/src/app/backend/api"
	metricapi "github.com/CentaurusInfra/dashboard/src/app/backend/integration/metric/api"
	resource "github.com/CentaurusInfra/dashboard/src/app/backend/resource/node"
	v1 "k8s.io/api/core/v1"
	client "k8s.io/client-go/kubernetes"
)

type ResourcePartitionList struct {
	ListMeta          api.ListMeta               `json:"listMeta"`
	Partitions        []*ResourcePartitionDetail `json:"resourcePartitions"`
	CumulativeMetrics []metricapi.Metric         `json:"cumulativeMetrics"`

	// List of non-critical errors, that occurred during resource retrieval.
	Errors []error `json:"errors"`
}

type TenantPartitionList struct {
	ListMeta          api.ListMeta             `json:"listMeta"`
	Partitions        []*TenantPartitionDetail `json:"tenantPartitions"`
	CumulativeMetrics []metricapi.Metric       `json:"cumulativeMetrics"`

	// List of non-critical errors, that occurred during resource retrieval.
	Errors []error `json:"errors"`
}
type ResourcePartitionDetail struct {
	ObjectMeta ResourcePartition `json:"objectMeta"`
	TypeMeta   api.TypeMeta      `json:"typeMeta"`
}

type TenantPartitionDetail struct {
	ObjectMeta TenantPartition `json:"objectMeta"`
	TypeMeta   api.TypeMeta    `json:"typeMeta"`
}

type ResourcePartition struct {
	Name             string  `json:"name"`
	NodeCount        int64   `json:"nodeCount"`
	CPULimit         int64   `json:"cpuLimit"`
	CPUUsed          float64 `json:"cpuUsed"`
	MemoryLimit      int64   `json:"memoryLimit"`
	MemoryUsed       float64 `json:"memoryUsed"`
	HealthyNodeCount int64   `json:"healthyNodeCount"`
}

type TenantPartition struct {
	Name             string  `json:"name"`
	NodeName         string  `json:"nodeName"`
	PodCount         int64   `json:"podCount"`
	TotalPods        int64   `json:"totalPods"`
	TenantCount      int64   `json:"tenantCount"`
	CPUUsed          float64 `json:"cpuUsed"`
	CPULimit         int64   `json:"cpuLimit"`
	MemoryUsed       float64 `json:"memoryUsed"`
	MemoryLimit      int64   `json:"memoryLimit"`
	HealthyNodeCount int64   `json:"healthyNodeCount"`
}

type NodeAllocatedResources struct {
	// CPURequests is number of allocated milicores.
	CPURequests int64 `json:"cpuRequests"`

	// CPURequestsFraction is a fraction of CPU, that is allocated.
	CPURequestsFraction float64 `json:"cpuRequestsFraction"`

	// CPULimits is defined CPU limit.
	CPULimits int64 `json:"cpuLimits"`

	// CPULimitsFraction is a fraction of defined CPU limit, can be over 100%, i.e.
	// overcommitted.
	CPULimitsFraction float64 `json:"cpuLimitsFraction"`

	// CPUCapacity is specified node CPU capacity in milicores.
	CPUCapacity int64 `json:"cpuCapacity"`

	// MemoryRequests is a fraction of memory, that is allocated.
	MemoryRequests int64 `json:"memoryRequests"`

	// MemoryRequestsFraction is a fraction of memory, that is allocated.
	MemoryRequestsFraction float64 `json:"memoryRequestsFraction"`

	// MemoryLimits is defined memory limit.
	MemoryLimits int64 `json:"memoryLimits"`

	// MemoryLimitsFraction is a fraction of defined memory limit, can be over 100%, i.e.
	// overcommitted.
	MemoryLimitsFraction float64 `json:"memoryLimitsFraction"`

	// MemoryCapacity is specified node memory capacity in bytes.
	MemoryCapacity int64 `json:"memoryCapacity"`

	// AllocatedPods in number of currently allocated pods on the node.
	AllocatedPods int `json:"allocatedPods"`

	// PodCapacity is maximum number of pods, that can be allocated on the node.
	PodCapacity int64 `json:"podCapacity"`

	// PodFraction is a fraction of pods, that can be allocated on given node.
	PodFraction float64 `json:"podFraction"`
}

func GetResourcePartitionDetail(client client.Interface, clusterName string) (*ResourcePartitionDetail, error) {
	nodes, err := client.CoreV1().Nodes().List(api.ListEverything)
	if err != nil {
		return nil, err
	}
	var cpuLimit int64 = 0
	var cpuUsed float64 = 0
	var memoryLimit int64 = 0
	var memoryUsed float64 = 0
	var healthyNodeCount int64 = 0
	for _, node := range nodes.Items {
		pods, _ := resource.GetNodePodsDetails(client, node)
		allocatedResources, _ := resource.GetNodeAllocatedResources(node, pods)

		cpuLimit += allocatedResources.CPUCapacity
		cpuUsed += allocatedResources.CPURequestsFraction

		memoryLimit += allocatedResources.MemoryCapacity
		memoryUsed += allocatedResources.MemoryRequestsFraction

		for _, condition := range node.Status.Conditions {
			if condition.Type == v1.NodeConditionType("Ready") && condition.Status == v1.ConditionTrue {
				healthyNodeCount++
				break
			}
		}
	}

	partitionDetail := new(ResourcePartitionDetail)
	partitionDetail.ObjectMeta.NodeCount = int64(len(nodes.Items))
	partitionDetail.ObjectMeta.CPUUsed = cpuUsed
	partitionDetail.ObjectMeta.CPULimit = cpuLimit
	partitionDetail.ObjectMeta.MemoryUsed = memoryUsed
	partitionDetail.ObjectMeta.MemoryLimit = memoryLimit
	partitionDetail.ObjectMeta.HealthyNodeCount = healthyNodeCount
	partitionDetail.ObjectMeta.Name = clusterName
	partitionDetail.TypeMeta.Kind = "ResourcePartition"
	return partitionDetail, nil
}

func GetTenantPartitionDetail(client client.Interface, clusterName string) (*TenantPartitionDetail, error) {
	nodes, err := client.CoreV1().Nodes().List(api.ListEverything)
	if err != nil {
		return nil, err
	}
	var cpuLimit int64 = 0
	var cpuUsed float64 = 0
	var memoryLimit int64 = 0
	var memoryUsed float64 = 0
	var healthyNodeCount int64 = 0
	var podCount int64 = 0
	var nodePods int64 = 0
	nodeName := ``

	for _, node := range nodes.Items {
		nodeName = node.Name
		pods, _ := resource.GetNodePodsDetails(client, node)
		allocatedResources, _ := resource.GetNodeAllocatedResources(node, pods)

		cpuLimit += allocatedResources.CPUCapacity
		cpuUsed += allocatedResources.CPURequestsFraction

		memoryLimit += allocatedResources.MemoryCapacity
		memoryUsed += allocatedResources.MemoryRequestsFraction

		nodePods = node.Status.Capacity.Pods().Value()
		podCount += int64(allocatedResources.AllocatedPods)

		for _, condition := range node.Status.Conditions {
			if condition.Type == v1.NodeConditionType("Ready") && condition.Status == v1.ConditionTrue {
				healthyNodeCount++
				break
			}
		}
	}
	tenants, err := client.CoreV1().Tenants().List(api.ListEverything)
	if err != nil {
		return nil, err
	}
	partitionDetail := new(TenantPartitionDetail)
	partitionDetail.ObjectMeta.TenantCount = int64(len(tenants.Items))
	partitionDetail.ObjectMeta.CPUUsed = cpuUsed
	partitionDetail.ObjectMeta.CPULimit = cpuLimit
	partitionDetail.ObjectMeta.MemoryUsed = memoryUsed
	partitionDetail.ObjectMeta.MemoryLimit = memoryLimit
	partitionDetail.ObjectMeta.HealthyNodeCount = healthyNodeCount
	partitionDetail.ObjectMeta.PodCount = podCount
	partitionDetail.ObjectMeta.TotalPods = nodePods
	partitionDetail.ObjectMeta.Name = clusterName
	partitionDetail.ObjectMeta.NodeName = nodeName
	partitionDetail.TypeMeta.Kind = "TenantPartition"
	return partitionDetail, nil
}

func GetWorkerCount(client client.Interface) int64 {
	nodes, _ := client.CoreV1().Nodes().List(api.ListEverything)
	var workerCount int64 = 0
	if len(nodes.Items) == 1 {
		workerCount = 1
	} else {
		workerCount = int64(len(nodes.Items) - 1)
	}
	return workerCount
}
