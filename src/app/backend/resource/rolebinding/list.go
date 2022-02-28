package rolebinding

import (
	"github.com/kubernetes/dashboard/src/app/backend/api"
	"github.com/kubernetes/dashboard/src/app/backend/errors"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/client-go/kubernetes"
	"log"
)

// RoleBindingList contains a list of roleBindings in the cluster.
type RoleBindingList struct {
	ListMeta api.ListMeta  `json:"listMeta"`
	Items    []RoleBinding `json:"items"`

	// List of non-critical errors, that occurred during resource retrieval.
	Errors []error `json:"errors"`
}

// RoleBinding is a presentation layer view of Kubernetes roleBinding. This means it is roleBinding plus additional
// augmented data we can get from other sources.
type RoleBinding struct {
	ObjectMeta api.ObjectMeta `json:"objectMeta"`
	TypeMeta   api.TypeMeta   `json:"typeMeta"`
}

// GetRoleBindingList returns a list of all RoleBindings in the cluster.
func GetRoleBindingList(client kubernetes.Interface, nsQuery *common.NamespaceQuery, dsQuery *dataselect.DataSelectQuery) (*RoleBindingList, error) {
	log.Print("Getting list of all roleBindings in the cluster")
	channels := &common.ResourceChannels{
		RoleBindingList: common.GetRoleBindingListChannel(client, 1),
	}

	return GetRoleBindingListFromChannels(channels, dsQuery)
}

// GetRoleBindingListFromChannels returns a list of all RoleBindings in the cluster
// reading required resource list once from the channels.
func GetRoleBindingListFromChannels(channels *common.ResourceChannels, dsQuery *dataselect.DataSelectQuery) (*RoleBindingList, error) {
	roleBindings := <-channels.RoleBindingList.List
	err := <-channels.RoleBindingList.Error
	nonCriticalErrors, criticalError := errors.HandleError(err)
	if criticalError != nil {
		return nil, criticalError
	}
	roleBindingList := toRoleBindingList(roleBindings.Items, nonCriticalErrors, dsQuery)
	return roleBindingList, nil
}

func toRoleBinding(roleBinding rbac.RoleBinding) RoleBinding {
	return RoleBinding{
		ObjectMeta: api.NewObjectMeta(roleBinding.ObjectMeta),
		TypeMeta:   api.NewTypeMeta(api.ResourceKindRoleBinding),
	}
}

func toRoleBindingList(roleBindings []rbac.RoleBinding, nonCriticalErrors []error, dsQuery *dataselect.DataSelectQuery) *RoleBindingList {
	result := &RoleBindingList{
		ListMeta: api.ListMeta{TotalItems: len(roleBindings)},
		Errors:   nonCriticalErrors,
	}

	items := make([]RoleBinding, 0)
	for _, item := range roleBindings {
		items = append(items, toRoleBinding(item))
	}

	roleBindingCells, filteredTotal := dataselect.GenericDataSelectWithFilter(toCells(items), dsQuery)
	result.ListMeta = api.ListMeta{TotalItems: filteredTotal}
	result.Items = fromCells(roleBindingCells)
	return result
}
