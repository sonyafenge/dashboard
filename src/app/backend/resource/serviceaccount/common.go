package serviceaccount

import (
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
)

// ServiceAccountSpec is a specification of service account to create.
type ServiceAccountSpec struct {
	// Name of the service account.
	Name string `json:"name"`

	// Namespace under which service account is to be created.
	Namespace string `json:"namespace"`

	// Tenant of the service-account.
	Tenant string `json:"tenant"`
}

// CreateServiceAccount creates Service Account based on given specification.
func CreateServiceAccount(spec *ServiceAccountSpec, client kubernetes.Interface) error {
	log.Printf("Creating Service-account %s", spec.Name)

	serviceaccount := &v1.ServiceAccount{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      spec.Name,
			Namespace: spec.Namespace,
		},
	}

	_, err := client.CoreV1().ServiceAccounts(spec.Namespace).Create(serviceaccount)
	return err
}

// DeleteServiceAccount deletes service-account based on given specification.
func DeleteServiceAccount(namespaceName string, serviceaccountName string, client kubernetes.Interface) error {
	log.Printf("Deleting Service-Account %s", serviceaccountName)
	err := client.CoreV1().ServiceAccounts(namespaceName).Delete(serviceaccountName, &metaV1.DeleteOptions{})
	return err
}

// CreateServiceAccountsWithMultiTenancy creates Service Account based on given specification.
func CreateServiceAccountsWithMultiTenancy(spec *ServiceAccountSpec, client kubernetes.Interface) error {
	log.Printf("Creating Service-account %s", spec.Name)

	serviceaccount := &v1.ServiceAccount{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      spec.Name,
			Namespace: spec.Namespace,
			Tenant:    spec.Tenant,
		},
	}

	_, err := client.CoreV1().ServiceAccountsWithMultiTenancy(spec.Namespace, spec.Tenant).Create(serviceaccount)
	return err
}

// DeleteServiceAccountsWithMultiTenancy deletes service-account based on given specification.
func DeleteServiceAccountsWithMultiTenancy(tenantName string, namespaceName string, serviceaccountName string, client kubernetes.Interface) error {
	log.Printf("Deleting Service-Account %s", serviceaccountName)
	err := client.CoreV1().ServiceAccountsWithMultiTenancy(namespaceName, tenantName).Delete(serviceaccountName, &metaV1.DeleteOptions{})
	return err
}

type ServiceAccountCell v1.ServiceAccount

func (self ServiceAccountCell) GetProperty(name dataselect.PropertyName) dataselect.ComparableValue {
	switch name {
	case dataselect.NameProperty:
		return dataselect.StdComparableString(self.ObjectMeta.Name)
	case dataselect.CreationTimestampProperty:
		return dataselect.StdComparableTime(self.ObjectMeta.CreationTimestamp.Time)
	case dataselect.NamespaceProperty:
		return dataselect.StdComparableString(self.ObjectMeta.Namespace)
	default:
		// If name is not supported then just return a constant dummy value, sort will have no effect.
		return nil
	}
}

func toCells(std []v1.ServiceAccount) []dataselect.DataCell {
	cells := make([]dataselect.DataCell, len(std))
	for i := range std {
		cells[i] = ServiceAccountCell(std[i])
	}
	return cells
}

func fromCells(cells []dataselect.DataCell) []v1.ServiceAccount {
	std := make([]v1.ServiceAccount, len(cells))
	for i := range std {
		std[i] = v1.ServiceAccount(cells[i].(ServiceAccountCell))
	}
	return std
}
