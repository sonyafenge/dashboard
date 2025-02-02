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

package secret

import (
	"log"
	"time"

	"github.com/CentaurusInfra/dashboard/src/app/backend/api"
	"github.com/CentaurusInfra/dashboard/src/app/backend/errors"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/common"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/dataselect"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// SecretSpec is a common interface for the specification of different secrets.
type SecretSpec interface {
	GetName() string
	GetType() v1.SecretType
	GetNamespace() string
	GetData() map[string][]byte
}

// ImagePullSecretSpec is a specification of an image pull secret implements SecretSpec
type ImagePullSecretSpec struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`

	// The value of the .dockercfg property. It must be Base64 encoded.
	Data []byte `json:"data"`
}

// Secret is a single secret returned to the frontend.
type SecretDetailSpec struct {
	ObjectMeta struct {
		Name        string `json:"name"`
		Namespace   string `json:"namespace"`
		Annotations struct {
			KubernetesIoServiceAccountName string `json:"kubernetes.io/service-account.name"`
			KubernetesIoServiceAccountUID  string `json:"kubernetes.io/service-account.uid"`
		} `json:"annotations"`
		CreationTimestamp time.Time `json:"creationTimestamp"`
		UID               string    `json:"uid"`
	} `json:"objectMeta"`
	TypeMeta struct {
		Kind string `json:"kind"`
	} `json:"typeMeta"`
	Type string `json:"type"`
	Data struct {
		CaCrt     string `json:"ca.crt"`
		Namespace string `json:"namespace"`
		Token     string `json:"token"`
	} `json:"data"`
}

// GetName returns the name of the ImagePullSecret
func (spec *ImagePullSecretSpec) GetName() string {
	return spec.Name
}

// GetType returns the type of the ImagePullSecret, which is always api.SecretTypeDockercfg
func (spec *ImagePullSecretSpec) GetType() v1.SecretType {
	return v1.SecretTypeDockercfg
}

// GetNamespace returns the namespace of the ImagePullSecret
func (spec *ImagePullSecretSpec) GetNamespace() string {
	return spec.Namespace
}

// GetData returns the data the secret carries, it is a single key-value pair
func (spec *ImagePullSecretSpec) GetData() map[string][]byte {
	return map[string][]byte{v1.DockerConfigKey: spec.Data}
}

// Secret is a single secret returned to the frontend.
type Secret struct {
	ObjectMeta api.ObjectMeta `json:"objectMeta"`
	TypeMeta   api.TypeMeta   `json:"typeMeta"`
	Type       v1.SecretType  `json:"type"`
}

// SecretsList is a response structure for a queried secrets list.
type SecretList struct {
	api.ListMeta `json:"listMeta"`

	// Unordered list of Secrets.
	Secrets []Secret `json:"secrets"`

	// List of non-critical errors, that occurred during resource retrieval.
	Errors []error `json:"errors"`
}

// GetSecretList returns all secrets in the given namespace.
func GetSecretList(client kubernetes.Interface, namespace *common.NamespaceQuery,
	dsQuery *dataselect.DataSelectQuery) (*SecretList, error) {
	log.Printf("Getting list of secrets in %s namespace\n", namespace)
	secretList, err := client.CoreV1().SecretsWithMultiTenancy(namespace.ToRequestParam(), "").List(api.ListEverything)

	nonCriticalErrors, criticalError := errors.HandleError(err)
	if criticalError != nil {
		return nil, criticalError
	}

	return toSecretList(secretList.Items, nonCriticalErrors, dsQuery), nil
}

// GetSecretListWithMultiTenancy returns all secrets in the given namespace.
func GetSecretListWithMultiTenancy(client kubernetes.Interface, tenant string, namespace *common.NamespaceQuery,
	dsQuery *dataselect.DataSelectQuery) (*SecretList, error) {
	log.Printf("Getting list of secrets in %s namespace for %s\n", namespace, tenant)
	secretList, err := client.CoreV1().SecretsWithMultiTenancy(namespace.ToRequestParam(), tenant).List(api.ListEverything)

	nonCriticalErrors, criticalError := errors.HandleError(err)
	if criticalError != nil {
		return nil, criticalError
	}

	return toSecretList(secretList.Items, nonCriticalErrors, dsQuery), nil
}

// CreateSecret creates a single secret using the cluster API client
func CreateSecret(client kubernetes.Interface, spec SecretSpec) (*Secret, error) {
	namespace := spec.GetNamespace()
	secret := &v1.Secret{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      spec.GetName(),
			Namespace: namespace,
		},
		Type: spec.GetType(),
		Data: spec.GetData(),
	}
	_, err := client.CoreV1().SecretsWithMultiTenancy(namespace, "").Create(secret)
	result := toSecret(secret)
	return &result, err
}

// CreateSecretWithMultiTenancy creates a single secret using the cluster API client
func CreateSecretWithMultiTenancy(client kubernetes.Interface, tenant string, spec SecretSpec) (*Secret, error) {
	namespace := spec.GetNamespace()
	secret := &v1.Secret{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      spec.GetName(),
			Namespace: namespace,
		},
		Type: spec.GetType(),
		Data: spec.GetData(),
	}
	_, err := client.CoreV1().SecretsWithMultiTenancy(namespace, tenant).Create(secret)
	result := toSecret(secret)
	return &result, err
}

func toSecret(secret *v1.Secret) Secret {
	return Secret{
		ObjectMeta: api.NewObjectMeta(secret.ObjectMeta),
		TypeMeta:   api.NewTypeMeta(api.ResourceKindSecret),
		Type:       secret.Type,
	}
}

func toSecretList(secrets []v1.Secret, nonCriticalErrors []error, dsQuery *dataselect.DataSelectQuery) *SecretList {
	newSecretList := &SecretList{
		ListMeta: api.ListMeta{TotalItems: len(secrets)},
		Secrets:  make([]Secret, 0),
		Errors:   nonCriticalErrors,
	}

	secretCells, filteredTotal := dataselect.GenericDataSelectWithFilter(toCells(secrets), dsQuery)
	secrets = fromCells(secretCells)
	newSecretList.ListMeta = api.ListMeta{TotalItems: filteredTotal}

	for _, secret := range secrets {
		newSecretList.Secrets = append(newSecretList.Secrets, toSecret(&secret))
	}

	return newSecretList
}
