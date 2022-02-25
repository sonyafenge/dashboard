package node

import (
	"log"

	v1 "k8s.io/api/core/v1"
	client "k8s.io/client-go/kubernetes"

	"github.com/kubernetes/dashboard/src/app/backend/api"
	"github.com/kubernetes/dashboard/src/app/backend/errors"
	metricapi "github.com/kubernetes/dashboard/src/app/backend/integration/metric/api"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
)

// NodeList contains a list of nodes in the cluster.
type NodeList struct {
	ListMeta          api.ListMeta       `json:"listMeta"`
	Nodes             []Node             `json:"nodes"`
	CumulativeMetrics []metricapi.Metric `json:"cumulativeMetrics"`

	// List of non-critical errors, that occurred during resource retrieval.
	Errors []error `json:"errors"`
}

// Node is a presentation layer view of Kubernetes nodes. This means it is node plus additional
// augmented data we can get from other sources.
type Node struct {
	ObjectMeta         api.ObjectMeta         `json:"objectMeta"`
	TypeMeta           api.TypeMeta           `json:"typeMeta"`
	Ready              v1.ConditionStatus     `json:"ready"`
	AllocatedResources NodeAllocatedResources `json:"allocatedResources"`
	ClusterName        string                 `json:"clusterName"`
}

// GetNodeList returns a list of all Nodes in the cluster.
func GetNodeList(client client.Interface, dsQuery *dataselect.DataSelectQuery, metricClient metricapi.MetricClient, cLusterName string) (*NodeList, error) {
	nodes, err := client.CoreV1().Nodes().List(api.ListEverything)

	nonCriticalErrors, criticalError := errors.HandleError(err)
	if criticalError != nil {
		return nil, criticalError
	}

	return toNodeList(client, nodes.Items, nonCriticalErrors, dsQuery, metricClient, cLusterName), nil
}

func toNodeList(client client.Interface, nodes []v1.Node, nonCriticalErrors []error, dsQuery *dataselect.DataSelectQuery,
	metricClient metricapi.MetricClient, clusterName string) *NodeList {
	nodeList := &NodeList{
		Nodes:    make([]Node, 0),
		ListMeta: api.ListMeta{TotalItems: len(nodes)},
		Errors:   nonCriticalErrors,
	}

	nodeCells, metricPromises, filteredTotal := dataselect.GenericDataSelectWithFilterAndMetrics(toCells(nodes),
		dsQuery, metricapi.NoResourceCache, metricClient)
	nodes = fromCells(nodeCells)
	nodeList.ListMeta = api.ListMeta{TotalItems: filteredTotal}

	for _, node := range nodes {
		pods, err := GetNodePodsDetails(client, node)
		if err != nil {
			log.Printf("Couldn't get pods of %s node: %s\n", node.Name, err)
		}
		node.ClusterName = clusterName
		nodeList.Nodes = append(nodeList.Nodes, toNode(node, pods))

	}

	cumulativeMetrics, err := metricPromises.GetMetrics()
	nodeList.CumulativeMetrics = cumulativeMetrics
	if err != nil {
		nodeList.CumulativeMetrics = make([]metricapi.Metric, 0)
	}

	return nodeList
}

func toNode(node v1.Node, pods *v1.PodList) Node {
	allocatedResources, err := GetNodeAllocatedResources(node, pods)
	if err != nil {
		log.Printf("Couldn't get allocated resources of %s node: %s\n", node.Name, err)
	}

	return Node{
		ObjectMeta:         api.NewObjectMeta(node.ObjectMeta),
		TypeMeta:           api.NewTypeMeta(api.ResourceKindNode),
		Ready:              getNodeConditionStatus(node, v1.NodeReady),
		AllocatedResources: allocatedResources,
		ClusterName:        node.ClusterName,
	}
}

func getNodeConditionStatus(node v1.Node, conditionType v1.NodeConditionType) v1.ConditionStatus {
	for _, condition := range node.Status.Conditions {
		if condition.Type == conditionType {
			return condition.Status
		}
	}
	return v1.ConditionUnknown
}
