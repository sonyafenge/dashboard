package namespace

import (
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	api "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
)

// NamespaceSpec is a specification of namespace to create.
type NamespaceSpec struct {
	// Name of the namespace.
	Name   string `json:"name"`
	Tenant string `json:"tenant"`
}

// CreateNamespace creates namespace based on given specification.
func CreateNamespace(spec *NamespaceSpec, tenant string, client kubernetes.Interface) error {
	log.Printf("Creating namespace %s", spec.Name)

	if tenant == "" {
		tenant = "default"
	}

	if spec.Tenant == "" {
		spec.Tenant = tenant
	}
	namespace := &api.Namespace{
		ObjectMeta: metaV1.ObjectMeta{
			Name:   spec.Name,
			Tenant: spec.Tenant,
		},
	}

	_, err := client.CoreV1().Namespaces().Create(namespace)
	return err
}

// The code below allows to perform complex data section on []api.Namespace

type NamespaceCell api.Namespace

func (self NamespaceCell) GetProperty(name dataselect.PropertyName) dataselect.ComparableValue {
	switch name {
	case dataselect.NameProperty:
		return dataselect.StdComparableString(self.ObjectMeta.Name)
	case dataselect.CreationTimestampProperty:
		return dataselect.StdComparableTime(self.ObjectMeta.CreationTimestamp.Time)
	case dataselect.NamespaceProperty:
		return dataselect.StdComparableString(self.ObjectMeta.Namespace)
	default:
		// if name is not supported then just return a constant dummy value, sort will have no effect.
		return nil
	}
}

func toCells(std []api.Namespace) []dataselect.DataCell {
	cells := make([]dataselect.DataCell, len(std))
	for i := range std {
		cells[i] = NamespaceCell(std[i])
	}
	return cells
}

func fromCells(cells []dataselect.DataCell) []api.Namespace {
	std := make([]api.Namespace, len(cells))
	for i := range std {
		std[i] = api.Namespace(cells[i].(NamespaceCell))
	}
	return std
}
