package serviceaccount

import (
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	client "k8s.io/client-go/kubernetes"
	"log"
)

// ServiceAccountDetail contains detailed information about a service account.
type ServiceAccountDetail struct {
	ServiceAccount `json:",inline"`
	Errors         []error `json:"errors"`
}

// GetServiceAccountDetail returns detailed information about a service account.
func GetServiceAccountDetail(client client.Interface, namespace string, name string) (*ServiceAccountDetail, error) {
	log.Printf("Getting details of %s service account in %s namespace", name, namespace)

	raw, err := client.CoreV1().ServiceAccounts(namespace).Get(name, metaV1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return getServiceAccountDetail(raw), nil
}

// GetServiceAccountDetailWithMultiTenancy returns detailed information about a service account.
func GetServiceAccountDetailWithMultiTenancy(client client.Interface, tenant string, namespace string, name string) (*ServiceAccountDetail, error) {
	log.Printf("Getting details of %s service account in %s namespace", name, namespace)

	raw, err := client.CoreV1().ServiceAccountsWithMultiTenancy(namespace, tenant).Get(name, metaV1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return getServiceAccountDetail(raw), nil
}

func getServiceAccountDetail(sa *v1.ServiceAccount) *ServiceAccountDetail {
	return &ServiceAccountDetail{
		ServiceAccount: toServiceAccount(sa),
	}
}
