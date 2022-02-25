package tenant

import (
	"log"

	v1 "k8s.io/api/core/v1"
	client "k8s.io/client-go/kubernetes"

	"github.com/kubernetes/dashboard/src/app/backend/api"
	"github.com/kubernetes/dashboard/src/app/backend/errors"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
)

type Tenant struct {
	ObjectMeta api.ObjectMeta `json:"objectMeta"`
	TypeMeta   api.TypeMeta   `json:"typeMeta"`

	// Phase is the current lifecycle phase of the tenant
	Phase       v1.TenantPhase `json:"phase"`
	ClusterName string         `json:"clusterName"`
}

// TenantList contains a list of tenants in the cluster.
type TenantList struct {
	ListMeta api.ListMeta `json:"listMeta"`
	Tenants  []Tenant     `json:"tenants"`

	// List of non-critical errors, that occurred during resource retrieval.
	Errors []error `json:"errors"`
}

func GetTenantList(client client.Interface, dsQuery *dataselect.DataSelectQuery, clusterName string) (*TenantList, error) {
	log.Println("Getting list of tenants")
	tenants, err := client.CoreV1().Tenants().List(api.ListEverything)

	nonCriticalErrors, criticalError := errors.HandleError(err)
	if criticalError != nil {
		return nil, criticalError
	}

	return toTenantList(tenants.Items, nonCriticalErrors, dsQuery, clusterName), nil
}

func toTenantList(tenants []v1.Tenant, nonCriticalErrors []error, dsQuery *dataselect.DataSelectQuery, clusterName string) *TenantList {
	tenantList := &TenantList{
		ListMeta: api.ListMeta{TotalItems: len(tenants)},
		Tenants:  make([]Tenant, 0),
	}

	tenantCells, filteredTotal := dataselect.GenericDataSelectWithFilter(toCells(tenants), dsQuery)
	tenants = fromCells(tenantCells)
	tenantList.ListMeta = api.ListMeta{TotalItems: filteredTotal}
	tenantList.Errors = nonCriticalErrors

	for _, tenant := range tenants {

		tenant.ClusterName = clusterName
		tenantList.Tenants = append(tenantList.Tenants, toTenant(tenant))
	}

	return tenantList
}

func toTenant(tenant v1.Tenant) Tenant {
	return Tenant{
		ObjectMeta: api.NewObjectMeta(tenant.ObjectMeta),
		TypeMeta:   api.NewTypeMeta(api.ResourceKindTenant),
		Phase:      tenant.Status.Phase,
	}
}
