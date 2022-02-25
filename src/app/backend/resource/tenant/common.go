package tenant

import (
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
)

// The code below allows to perform complex data section on []api.Tenant

type TenantSpec struct {
	// Name of the tenant.
	Name             string `json:"name"`
	StorageClusterId string `json:"storageclusterid"` // Storage Cluster Id of the Tenant
}

// CreateTenant creates tenant based on given specification.
func CreateTenant(spec *TenantSpec, client kubernetes.Interface, clusterName string) error {
	log.Printf("Creating tenant %s", spec.Name)

	// setting default values if no values passed
	if spec.StorageClusterId == "" {
		spec.StorageClusterId = "0"
	}
	clusterLabel := make(map[string]string)
	clusterLabel["clusterName"] = clusterName
	tenant := &v1.Tenant{
		ObjectMeta: metaV1.ObjectMeta{
			Name:   spec.Name,
			Labels: clusterLabel,
		},
		Spec: v1.TenantSpec{

			StorageClusterId: spec.StorageClusterId,
		},
	}

	_, err := client.CoreV1().Tenants().Create(tenant)
	return err
}

// DeleteTenant deletes tenant based on given specification.
func DeleteTenant(tenantName string, client kubernetes.Interface) error {
	log.Printf("Deleting Tenant %s", tenantName)
	err := client.CoreV1().Tenants().Delete(tenantName, &metaV1.DeleteOptions{})
	return err
}

type TenantCell v1.Tenant

func (self TenantCell) GetProperty(name dataselect.PropertyName) dataselect.ComparableValue {
	switch name {
	case dataselect.NameProperty:
		return dataselect.StdComparableString(self.ObjectMeta.Name)
	case dataselect.CreationTimestampProperty:
		return dataselect.StdComparableTime(self.ObjectMeta.CreationTimestamp.Time)
	default:
		return nil
	}
}

func toCells(std []v1.Tenant) []dataselect.DataCell {
	cells := make([]dataselect.DataCell, len(std))
	for i := range std {
		cells[i] = TenantCell(std[i])
	}
	return cells
}

func fromCells(cells []dataselect.DataCell) []v1.Tenant {
	std := make([]v1.Tenant, len(cells))
	for i := range std {
		std[i] = v1.Tenant(cells[i].(TenantCell))
	}
	return std
}
