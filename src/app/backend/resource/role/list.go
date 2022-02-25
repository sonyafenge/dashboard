package role

import (
	"github.com/kubernetes/dashboard/src/app/backend/api"
	"github.com/kubernetes/dashboard/src/app/backend/errors"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	rbac "k8s.io/api/rbac/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
)

// RoleList contains a list of role in the cluster.
type RoleList struct {
	ListMeta api.ListMeta `json:"listMeta"`
	Items    []Role       `json:"items"`

	// List of non-critical errors, that occurred during resource retrieval.
	Errors []error `json:"errors"`
}

// Role is a presentation layer view of Kubernetes role. This means it is role plus additional
// augmented data we can get from other sources.
type Role struct {
	ObjectMeta api.ObjectMeta `json:"objectMeta"`
	TypeMeta   api.TypeMeta   `json:"typeMeta"`
}

// GetRoleList returns a list of all Roles in the cluster.
func GetRoleList(client kubernetes.Interface, nsQuery *common.NamespaceQuery, dsQuery *dataselect.DataSelectQuery) (*RoleList, error) {
	log.Print("Getting list of all roles in the cluster")
	channels := &common.ResourceChannels{
		RoleList: common.GetRoleListChannel(client, 1),
	}

	return GetRoleListFromChannels(channels, dsQuery)
}

// GetRolesWithMultiTenancy returns a list of all Roles under specific tenant.
func GetRolesWithMultiTenancy(client kubernetes.Interface, tenant string, namespace string) (*RoleList, error) {
	rawObject, err := client.RbacV1().RolesWithMultiTenancy(namespace, tenant).List(metaV1.ListOptions{})

	nonCriticalErrors, criticalError := errors.HandleError(err)
	if criticalError != nil {
		return nil, criticalError
	}
	dsQuery := dataselect.NoDataSelect
	roleList := toRoleList(rawObject.Items, nonCriticalErrors, dsQuery)
	return roleList, nil
}

// GetRoleListFromChannels returns a list of all Roles in the cluster
// reading required resource list once from the channels.
func GetRoleListFromChannels(channels *common.ResourceChannels, dsQuery *dataselect.DataSelectQuery) (*RoleList, error) {
	roles := <-channels.RoleList.List
	err := <-channels.RoleList.Error
	nonCriticalErrors, criticalError := errors.HandleError(err)
	if criticalError != nil {
		return nil, criticalError
	}
	roleList := toRoleList(roles.Items, nonCriticalErrors, dsQuery)
	return roleList, nil
}

func toRole(role rbac.Role) Role {
	return Role{
		ObjectMeta: api.NewObjectMeta(role.ObjectMeta),
		TypeMeta:   api.NewTypeMeta(api.ResourceKindRole),
	}
}

func toRoleList(roles []rbac.Role, nonCriticalErrors []error, dsQuery *dataselect.DataSelectQuery) *RoleList {
	result := &RoleList{
		ListMeta: api.ListMeta{TotalItems: len(roles)},
		Errors:   nonCriticalErrors,
	}

	items := make([]Role, 0)
	for _, item := range roles {
		items = append(items, toRole(item))
	}

	roleCells, filteredTotal := dataselect.GenericDataSelectWithFilter(toCells(items), dsQuery)
	result.ListMeta = api.ListMeta{TotalItems: filteredTotal}
	result.Items = fromCells(roleCells)
	return result
}
