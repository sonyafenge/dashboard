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

package handler

import (
	"encoding/base64"
	er "errors"
	"fmt"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/partition"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/vm"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/CentaurusInfra/dashboard/src/app/backend/iam/db"
	"github.com/CentaurusInfra/dashboard/src/app/backend/iam/model"
	_ "github.com/lib/pq" // postgres golang driver

	"github.com/CentaurusInfra/dashboard/src/app/backend/api"
	"github.com/CentaurusInfra/dashboard/src/app/backend/auth"
	authApi "github.com/CentaurusInfra/dashboard/src/app/backend/auth/api"
	clientapi "github.com/CentaurusInfra/dashboard/src/app/backend/client/api"
	"github.com/CentaurusInfra/dashboard/src/app/backend/errors"
	"github.com/CentaurusInfra/dashboard/src/app/backend/integration"
	metricapi "github.com/CentaurusInfra/dashboard/src/app/backend/integration/metric/api"
	"github.com/CentaurusInfra/dashboard/src/app/backend/plugin"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/clusterrole"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/clusterrolebinding"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/common"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/configmap"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/container"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/controller"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/cronjob"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/customresourcedefinition"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/daemonset"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/dataselect"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/deployment"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/event"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/horizontalpodautoscaler"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/ingress"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/job"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/logs"
	ns "github.com/CentaurusInfra/dashboard/src/app/backend/resource/namespace"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/node"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/persistentvolume"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/persistentvolumeclaim"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/pod"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/replicaset"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/replicationcontroller"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/resourcequota"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/role"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/rolebinding"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/secret"
	resourceService "github.com/CentaurusInfra/dashboard/src/app/backend/resource/service"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/serviceaccount"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/statefulset"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/storageclass"
	"github.com/CentaurusInfra/dashboard/src/app/backend/resource/tenant"
	"github.com/CentaurusInfra/dashboard/src/app/backend/scaling"
	"github.com/CentaurusInfra/dashboard/src/app/backend/settings"
	settingsApi "github.com/CentaurusInfra/dashboard/src/app/backend/settings/api"
	"github.com/CentaurusInfra/dashboard/src/app/backend/systembanner"
	"github.com/CentaurusInfra/dashboard/src/app/backend/validation"
	restful "github.com/emicklei/go-restful"
	"golang.org/x/net/xsrftoken"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/remotecommand"
)

const (
	// RequestLogString is a template for request log message.
	RequestLogString = "[%s] Incoming %s %s %s request from %s: %s"

	// ResponseLogString is a template for response log message.
	ResponseLogString = "[%s] Outcoming response to %s with %d status code"
)

// APIHandler is a representation of API handler. Structure contains clientapi, Heapster clientapi and clientapi configuration.
type APIHandler struct {
	iManager  integration.IntegrationManager
	tpManager clientapi.ClientManager
	sManager  settingsApi.SettingsManager
}
type APIHandlerV2 struct {
	iManager             integration.IntegrationManager
	defaultClientmanager clientapi.ClientManager
	tpManager            []clientapi.ClientManager
	rpManager            []clientapi.ClientManager
	sManager             settingsApi.SettingsManager
	podInformerManager   []cache.SharedIndexInformer
}

// TerminalResponse is sent by handleExecShell. The Id is a random session id that binds the original REST request and the SockJS connection.
// Any clientapi in possession of this Id can hijack the terminal session.
type TerminalResponse struct {
	Id string `json:"id"`
}

func ResourceAllocator(tenant string, clients []clientapi.ClientManager) clientapi.ClientManager {
	if tenant == "system" || tenant == "" {
		log.Printf("selected config of %s cluster", clients[0].GetClusterName())
		return clients[0]
	}
	if clienlen := len(clients); clienlen > 1 {
		pref := []rune(strings.ToUpper(tenant))
		log.Printf("prefix:%v", pref[0])
		if pref[0] <= rune(77) {
			log.Printf("selected config of %s cluster", clients[0].GetClusterName())
			return clients[0]
		} else {
			log.Printf("selected config of %s cluster", clients[1].GetClusterName())
			return clients[1]
		}
	}
	log.Printf("selected config of %s cluster", clients[0].GetClusterName())
	return clients[0]
}

//struct for already exists
type ErrorMsg struct {
	Msg string `json:"msg"`
}

// CreateHTTPAPIHandler creates a new HTTP handler that handles all requests to the API of the backend.
func CreateHTTPAPIHandler(iManager integration.IntegrationManager, tpManager clientapi.ClientManager, tpManagers []clientapi.ClientManager, rpManagers []clientapi.ClientManager,
	authManager []authApi.AuthManager, sManager settingsApi.SettingsManager,
	sbManager systembanner.SystemBannerManager, podInformers []cache.SharedIndexInformer) (

	http.Handler, error) {
	//apiHandler1 := APIHandler{iManager: iManager, tpManager: tpManager, sManager: sManager}
	apiHandler := APIHandlerV2{iManager: iManager, defaultClientmanager: tpManager, tpManager: tpManagers, rpManager: rpManagers, sManager: sManager, podInformerManager: podInformers}
	wsContainer := restful.NewContainer()
	wsContainer.EnableContentEncoding(true)

	apiV1Ws := new(restful.WebService)
	for _, cManager := range tpManagers {
		InstallFilters(apiV1Ws, cManager)
	}

	apiV1Ws.Path("/api/v1").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	wsContainer.Add(apiV1Ws)

	integrationHandler := integration.NewIntegrationHandler(iManager)
	integrationHandler.Install(apiV1Ws)

	pluginHandler := plugin.NewPluginHandler(tpManager)
	pluginHandler.Install(apiV1Ws)

	authHandler := auth.NewAuthHandler(authManager)
	authHandler.Install(apiV1Ws)

	settingsHandler := settings.NewSettingsHandler(sManager, tpManagers)
	settingsHandler.Install(apiV1Ws)

	systemBannerHandler := systembanner.NewSystemBannerHandler(sbManager)
	systemBannerHandler.Install(apiV1Ws)

	apiV1Ws.Route(
		apiV1Ws.GET("/resourcepartition").
			To(apiHandler.handleGetResourcePartitionDetail).
			Writes(partition.ResourcePartitionList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenantpartition").
			To(apiHandler.handleGetTenantPartitionDetail).
			Writes(partition.TenantPartitionList{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/tenant").
			To(apiHandler.handleGetTenantList).
			Writes(tenant.TenantList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenant/{name}").
			To(apiHandler.handleGetTenantDetail).
			Writes(tenant.TenantDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tptenant").
			To(apiHandler.handleGetTenantList).
			Writes(tenant.TenantList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tptenant/{name}").
			To(apiHandler.handleGetTenantDetail).
			Writes(tenant.TenantDetail{}))
	apiV1Ws.Route(
		apiV1Ws.POST("/tenant").
			To(apiHandler.handleCreateTenant).
			Reads(tenant.TenantSpec{}).
			Writes(tenant.TenantSpec{}))
	apiV1Ws.Route(
		apiV1Ws.DELETE("/tenants/{tenant}").
			To(apiHandler.handleDeleteTenant))

	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/csrftoken/{action}").
			To(apiHandler.handleGetCsrfToken).
			Writes(api.CsrfToken{}))

	apiV1Ws.Route(
		apiV1Ws.POST("/appdeployment").
			To(apiHandler.handleDeploy).
			Reads(deployment.AppDeploymentSpec{}).
			Writes(deployment.AppDeploymentSpec{}))
	apiV1Ws.Route(
		apiV1Ws.POST("/appdeployment/validate/name").
			To(apiHandler.handleNameValidity).
			Reads(validation.AppNameValiditySpec{}).
			Writes(validation.AppNameValidity{}))
	apiV1Ws.Route(
		apiV1Ws.POST("/appdeployment/validate/imagereference").
			To(apiHandler.handleImageReferenceValidity).
			Reads(validation.ImageReferenceValiditySpec{}).
			Writes(validation.ImageReferenceValidity{}))
	apiV1Ws.Route(
		apiV1Ws.POST("/appdeployment/validate/protocol").
			To(apiHandler.handleProtocolValidity).
			Reads(validation.ProtocolValiditySpec{}).
			Writes(validation.ProtocolValidity{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/appdeployment/protocols").
			To(apiHandler.handleGetAvailableProcotols).
			Writes(deployment.Protocols{}))

	apiV1Ws.Route(
		apiV1Ws.POST("/appdeploymentfromfile").
			To(apiHandler.handleDeployFromFile).
			Reads(deployment.AppDeploymentFromFileSpec{}).
			Writes(deployment.AppDeploymentFromFileResponse{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/replicationcontroller").
			To(apiHandler.handleGetReplicationControllerList).
			Writes(replicationcontroller.ReplicationControllerList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/replicationcontroller/{namespace}").
			To(apiHandler.handleGetReplicationControllerList).
			Writes(replicationcontroller.ReplicationControllerList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/replicationcontroller/{namespace}/{replicationController}").
			To(apiHandler.handleGetReplicationControllerDetail).
			Writes(replicationcontroller.ReplicationControllerDetail{}))
	apiV1Ws.Route(
		apiV1Ws.POST("/replicationcontroller/{namespace}/{replicationController}/update/pod").
			To(apiHandler.handleUpdateReplicasCount).
			Reads(replicationcontroller.ReplicationControllerSpec{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/replicationcontroller/{namespace}/{replicationController}/pod").
			To(apiHandler.handleGetReplicationControllerPods).
			Writes(pod.PodList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/replicationcontroller/{namespace}/{replicationController}/event").
			To(apiHandler.handleGetReplicationControllerEvents).
			Writes(common.EventList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/replicationcontroller/{namespace}/{replicationController}/service").
			To(apiHandler.handleGetReplicationControllerServices).
			Writes(resourceService.ServiceList{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/replicationcontroller").
			To(apiHandler.handleGetReplicationControllerListWithMultiTenancy).
			Writes(replicationcontroller.ReplicationControllerList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/replicationcontroller/{namespace}").
			To(apiHandler.handleGetReplicationControllerListWithMultiTenancy).
			Writes(replicationcontroller.ReplicationControllerList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/replicationcontroller/{namespace}/{replicationController}").
			To(apiHandler.handleGetReplicationControllerDetailWithMultiTenancy).
			Writes(replicationcontroller.ReplicationControllerDetail{}))
	apiV1Ws.Route(
		apiV1Ws.POST("/tenants/{tenant}/replicationcontroller/{namespace}/{replicationController}/update/pod").
			To(apiHandler.handleUpdateReplicasCountWithMultiTenancy).
			Reads(replicationcontroller.ReplicationControllerSpec{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/replicationcontroller/{namespace}/{replicationController}/pod").
			To(apiHandler.handleGetReplicationControllerPodsWithMultiTenancy).
			Writes(pod.PodList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/replicationcontroller/{namespace}/{replicationController}/event").
			To(apiHandler.handleGetReplicationControllerEventsWithMultiTenancy).
			Writes(common.EventList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/replicationcontroller/{namespace}/{replicationController}/service").
			To(apiHandler.handleGetReplicationControllerServicesWithMultiTenancy).
			Writes(resourceService.ServiceList{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/replicaset").
			To(apiHandler.handleGetReplicaSets).
			Writes(replicaset.ReplicaSetList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/replicaset/{namespace}").
			To(apiHandler.handleGetReplicaSets).
			Writes(replicaset.ReplicaSetList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/replicaset/{namespace}/{replicaSet}").
			To(apiHandler.handleGetReplicaSetDetail).
			Writes(replicaset.ReplicaSetDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/replicaset/{namespace}/{replicaSet}/pod").
			To(apiHandler.handleGetReplicaSetPods).
			Writes(pod.PodList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/replicaset/{namespace}/{replicaSet}/service").
			To(apiHandler.handleGetReplicaSetServices).
			Writes(pod.PodList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/replicaset/{namespace}/{replicaSet}/event").
			To(apiHandler.handleGetReplicaSetEvents).
			Writes(common.EventList{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/replicaset").
			To(apiHandler.handleGetReplicaSetsWithMultiTenancy).
			Writes(replicaset.ReplicaSetList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/replicaset/{namespace}").
			To(apiHandler.handleGetReplicaSetsWithMultiTenancy).
			Writes(replicaset.ReplicaSetList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/replicaset/{namespace}/{replicaSet}").
			To(apiHandler.handleGetReplicaSetDetailWithMultiTenancy).
			Writes(replicaset.ReplicaSetDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/replicaset/{namespace}/{replicaSet}/pod").
			To(apiHandler.handleGetReplicaSetPodsWithMutiTenancy).
			Writes(pod.PodList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/replicaset/{namespace}/{replicaSet}/service").
			To(apiHandler.handleGetReplicaSetServicesWithMultiTenancy).
			Writes(pod.PodList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/replicaset/{namespace}/{replicaSet}/event").
			To(apiHandler.handleGetReplicaSetEventsWithMultiTenancy).
			Writes(common.EventList{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/pod").
			To(apiHandler.handleGetPods).
			Writes(pod.PodList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/pod/{namespace}").
			To(apiHandler.handleGetPods).
			Writes(pod.PodList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/pod/{namespace}/{pod}").
			To(apiHandler.handleGetPodDetail).
			Writes(pod.PodDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/pod/{namespace}/{pod}/container").
			To(apiHandler.handleGetPodContainers).
			Writes(pod.PodDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/pod/{namespace}/{pod}/event").
			To(apiHandler.handleGetPodEvents).
			Writes(common.EventList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/pod/{namespace}/{pod}/shell/{container}").
			To(apiHandler.handleExecShell).
			Writes(TerminalResponse{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/pod/{namespace}/{pod}/persistentvolumeclaim").
			To(apiHandler.handleGetPodPersistentVolumeClaims).
			Writes(persistentvolumeclaim.PersistentVolumeClaimList{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/pod").
			To(apiHandler.handleGetPodsWithMultiTenancy).
			Writes(pod.PodList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/virtualmachine").
			To(apiHandler.handleGetVMsWithMultiTenancy).
			Writes(vm.VMList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/pod/{namespace}").
			To(apiHandler.handleGetPodsWithMultiTenancy).
			Writes(pod.PodList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/virtualmachine/{namespace}").
			To(apiHandler.handleGetVMsWithMultiTenancy).
			Writes(vm.VMList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/pod/{namespace}/{pod}").
			To(apiHandler.handleGetPodDetailWithMultiTenancy).
			Writes(pod.PodDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/virtualmachine/{namespace}/{virtualmachine}").
			To(apiHandler.handleGetVMDetailWithMultiTenancy).
			Writes(vm.VirtualMachineDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/pod/{namespace}/{pod}/container").
			To(apiHandler.handleGetPodContainersWithMultiTenancy).
			Writes(pod.PodDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/pod/{namespace}/{pod}/event").
			To(apiHandler.handleGetPodEventsWithMultiTenancy).
			Writes(common.EventList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/pod/{namespace}/{pod}/shell/{container}").
			To(apiHandler.handleExecShellWithMultiTenancy).
			Writes(TerminalResponse{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/pod/{namespace}/{pod}/persistentvolumeclaim").
			To(apiHandler.handleGetPodPersistentVolumeClaims).
			Writes(persistentvolumeclaim.PersistentVolumeClaimList{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/deployment").
			To(apiHandler.handleGetDeployments).
			Writes(deployment.DeploymentList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/deployment/{namespace}").
			To(apiHandler.handleGetDeployments).
			Writes(deployment.DeploymentList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/deployment/{namespace}/{deployment}").
			To(apiHandler.handleGetDeploymentDetail).
			Writes(deployment.DeploymentDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/deployment/{namespace}/{deployment}/event").
			To(apiHandler.handleGetDeploymentEvents).
			Writes(common.EventList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/deployment/{namespace}/{deployment}/oldreplicaset").
			To(apiHandler.handleGetDeploymentOldReplicaSets).
			Writes(replicaset.ReplicaSetList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/deployment/{namespace}/{deployment}/newreplicaset").
			To(apiHandler.handleGetDeploymentNewReplicaSet).
			Writes(replicaset.ReplicaSet{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/deployment").
			To(apiHandler.handleGetDeploymentsWithMultiTenancy).
			Writes(deployment.DeploymentList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/deployment/{namespace}").
			To(apiHandler.handleGetDeploymentsWithMultiTenancy).
			Writes(deployment.DeploymentList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/deployment/{namespace}/{deployment}").
			To(apiHandler.handleGetDeploymentDetailWithMultiTenancy).
			Writes(deployment.DeploymentDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/deployment/{namespace}/{deployment}/event").
			To(apiHandler.handleGetDeploymentEventsWithMultiTenancy).
			Writes(common.EventList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/deployment/{namespace}/{deployment}/oldreplicaset").
			To(apiHandler.handleGetDeploymentOldReplicaSetsWithMultiTenancy).
			Writes(replicaset.ReplicaSetList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/deployment/{namespace}/{deployment}/newreplicaset").
			To(apiHandler.handleGetDeploymentNewReplicaSetWithMultiTenancy).
			Writes(replicaset.ReplicaSet{}))

	apiV1Ws.Route(
		apiV1Ws.PUT("/scale/{kind}/{namespace}/{name}/").
			To(apiHandler.handleScaleResource).
			Writes(scaling.ReplicaCounts{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/scale/{kind}/{namespace}/{name}").
			To(apiHandler.handleGetReplicaCount).
			Writes(scaling.ReplicaCounts{}))
	apiV1Ws.Route(
		apiV1Ws.PUT("/tenants/{tenant}/scale/{kind}/{namespace}/{name}/").
			To(apiHandler.handleScaleResourceWithMultiTenancy).
			Writes(scaling.ReplicaCounts{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/scale/{kind}/{namespace}/{name}").
			To(apiHandler.handleGetReplicaCountWithMultiTenancy).
			Writes(scaling.ReplicaCounts{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/daemonset").
			To(apiHandler.handleGetDaemonSetList).
			Writes(daemonset.DaemonSetList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/daemonset/{namespace}").
			To(apiHandler.handleGetDaemonSetList).
			Writes(daemonset.DaemonSetList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/daemonset/{namespace}/{daemonSet}").
			To(apiHandler.handleGetDaemonSetDetail).
			Writes(daemonset.DaemonSetDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/daemonset/{namespace}/{daemonSet}/pod").
			To(apiHandler.handleGetDaemonSetPods).
			Writes(pod.PodList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/daemonset/{namespace}/{daemonSet}/service").
			To(apiHandler.handleGetDaemonSetServices).
			Writes(resourceService.ServiceList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/daemonset/{namespace}/{daemonSet}/event").
			To(apiHandler.handleGetDaemonSetEvents).
			Writes(common.EventList{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/daemonset").
			To(apiHandler.handleGetDaemonSetListWithMultiTenancy).
			Writes(daemonset.DaemonSetList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/daemonset/{namespace}").
			To(apiHandler.handleGetDaemonSetListWithMultiTenancy).
			Writes(daemonset.DaemonSetList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/daemonset/{namespace}/{daemonSet}").
			To(apiHandler.handleGetDaemonSetDetailWithMultiTenancy).
			Writes(daemonset.DaemonSetDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/daemonset/{namespace}/{daemonSet}/pod").
			To(apiHandler.handleGetDaemonSetPodsWithMultiTenancy).
			Writes(pod.PodList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/daemonset/{namespace}/{daemonSet}/service").
			To(apiHandler.handleGetDaemonSetServicesWithMultiTenancy).
			Writes(resourceService.ServiceList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/daemonset/{namespace}/{daemonSet}/event").
			To(apiHandler.handleGetDaemonSetEventsWithMultiTenancy).
			Writes(common.EventList{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/horizontalpodautoscaler").
			To(apiHandler.handleGetHorizontalPodAutoscalerList).
			Writes(horizontalpodautoscaler.HorizontalPodAutoscalerList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/horizontalpodautoscaler/{namespace}").
			To(apiHandler.handleGetHorizontalPodAutoscalerList).
			Writes(horizontalpodautoscaler.HorizontalPodAutoscalerList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/horizontalpodautoscaler/{namespace}/{horizontalpodautoscaler}").
			To(apiHandler.handleGetHorizontalPodAutoscalerDetail).
			Writes(horizontalpodautoscaler.HorizontalPodAutoscalerDetail{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/job").
			To(apiHandler.handleGetJobList).
			Writes(job.JobList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/job/{namespace}").
			To(apiHandler.handleGetJobList).
			Writes(job.JobList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/job/{namespace}/{name}").
			To(apiHandler.handleGetJobDetail).
			Writes(job.JobDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/job/{namespace}/{name}/pod").
			To(apiHandler.handleGetJobPods).
			Writes(pod.PodList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/job/{namespace}/{name}/event").
			To(apiHandler.handleGetJobEvents).
			Writes(common.EventList{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/job").
			To(apiHandler.handleGetJobListWithMultiTenancy).
			Writes(job.JobList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/job/{namespace}").
			To(apiHandler.handleGetJobListWithMultiTenancy).
			Writes(job.JobList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/job/{namespace}/{name}").
			To(apiHandler.handleGetJobDetailWithMultitenancy).
			Writes(job.JobDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/job/{namespace}/{name}/pod").
			To(apiHandler.handleGetJobPodsWithMultiTenancy).
			Writes(pod.PodList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/job/{namespace}/{name}/event").
			To(apiHandler.handleGetJobEventsWithMultiTenancy).
			Writes(common.EventList{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/cronjob").
			To(apiHandler.handleGetCronJobList).
			Writes(cronjob.CronJobList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/cronjob/{namespace}").
			To(apiHandler.handleGetCronJobList).
			Writes(cronjob.CronJobList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/cronjob/{namespace}/{name}").
			To(apiHandler.handleGetCronJobDetail).
			Writes(cronjob.CronJobDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/cronjob/{namespace}/{name}/job").
			To(apiHandler.handleGetCronJobJobs).
			Writes(job.JobList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/cronjob/{namespace}/{name}/event").
			To(apiHandler.handleGetCronJobEvents).
			Writes(common.EventList{}))
	apiV1Ws.Route(
		apiV1Ws.PUT("/cronjob/{namespace}/{name}/trigger").
			To(apiHandler.handleTriggerCronJob))

	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/cronjob").
			To(apiHandler.handleGetCronJobListWithMultiTenancy).
			Writes(cronjob.CronJobList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/cronjob/{namespace}").
			To(apiHandler.handleGetCronJobListWithMultiTenancy).
			Writes(cronjob.CronJobList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/cronjob/{namespace}/{name}").
			To(apiHandler.handleGetCronJobDetailWithMultiTenancy).
			Writes(cronjob.CronJobDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/cronjob/{namespace}/{name}/job").
			To(apiHandler.handleGetCronJobJobsWithMultiTenancy).
			Writes(job.JobList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/cronjob/{namespace}/{name}/event").
			To(apiHandler.handleGetCronJobEventsWithMultiTenancy).
			Writes(common.EventList{}))
	apiV1Ws.Route(
		apiV1Ws.PUT("/tenants/{tenant}/cronjob/{namespace}/{name}/trigger").
			To(apiHandler.handleTriggerCronJobWithMultiTenancy))

	apiV1Ws.Route(
		apiV1Ws.POST("/namespace").
			To(apiHandler.handleCreateNamespace).
			Reads(ns.NamespaceSpec{}).
			Writes(ns.NamespaceSpec{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/namespace").
			To(apiHandler.handleGetNamespaces).
			Writes(ns.NamespaceList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/namespace/{name}").
			To(apiHandler.handleGetNamespaceDetail).
			Writes(ns.NamespaceDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/namespace/{name}/event").
			To(apiHandler.handleGetNamespaceEvents).
			Writes(common.EventList{}))
	apiV1Ws.Route(
		apiV1Ws.POST("/resourcequota").
			To(apiHandler.handleAddResourceQuota).
			Reads(resourcequota.ResourceQuotaSpec{}).
			Writes(v1.ResourceQuota{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/resourcequota").
			To(apiHandler.handleGetResourceQuotaList).
			Writes(v1.ResourceQuotaList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/resourcequota/{namespace}").
			To(apiHandler.handleGetResourceQuotaListWithMultiTenancy).
			Writes(v1.ResourceQuotaList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/resourcequota/{namespace}/{name}"). // TODO
											To(apiHandler.handleGetResourceQuotaDetails).
											Writes(v1.ResourceQuotaList{}))
	apiV1Ws.Route(
		apiV1Ws.DELETE("/tenants/{tenant}/namespace/{namespace}/resourcequota/{name}").
			To(apiHandler.handleDeleteResourceQuota))
	apiV1Ws.Route(
		apiV1Ws.POST("/tenants/{tenant}/namespace"). // TODO
								To(apiHandler.handleCreateNamespace).
								Reads(ns.NamespaceSpec{}).
								Writes(ns.NamespaceSpec{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/namespace").
			To(apiHandler.handleGetNamespacesWithMultiTenancy).
			Writes(ns.NamespaceList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/namespace/{name}").
			To(apiHandler.handleGetNamespaceDetailWithMultiTenancy).
			Writes(ns.NamespaceDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/namespace/{name}/event").
			To(apiHandler.handleGetNamespaceEventsWithMultiTenancy).
			Writes(common.EventList{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/secret").
			To(apiHandler.handleGetSecretList).
			Writes(secret.SecretList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/secret/{namespace}").
			To(apiHandler.handleGetSecretList).
			Writes(secret.SecretList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/secret/{namespace}/{name}").
			To(apiHandler.handleGetSecretDetail).
			// TODO		Writes(secret.SecretDetail{}))
			Writes(secret.SecretDetailSpec{}))

	apiV1Ws.Route(
		apiV1Ws.POST("/secret").
			To(apiHandler.handleCreateImagePullSecret).
			Reads(secret.ImagePullSecretSpec{}).
			Writes(secret.Secret{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/secret").
			To(apiHandler.handleGetSecretListWithMultiTenancy).
			Writes(secret.SecretList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/secret/{namespace}").
			To(apiHandler.handleGetSecretListWithMultiTenancy).
			Writes(secret.SecretList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/secret/{namespace}/{name}").
			To(apiHandler.handleGetSecretDetailWithMultiTenancy).
			Writes(secret.SecretDetail{}))
	apiV1Ws.Route(
		apiV1Ws.POST("/tenants/{tenant}/secret").
			To(apiHandler.handleCreateImagePullSecretWithMultiTenancy).
			Reads(secret.ImagePullSecretSpec{}).
			Writes(secret.Secret{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/configmap").
			To(apiHandler.handleGetConfigMapList).
			Writes(configmap.ConfigMapList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/configmap/{namespace}").
			To(apiHandler.handleGetConfigMapList).
			Writes(configmap.ConfigMapList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/configmap/{namespace}/{configmap}").
			To(apiHandler.handleGetConfigMapDetail).
			Writes(configmap.ConfigMapDetail{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/configmap").
			To(apiHandler.handleGetConfigMapList).
			Writes(configmap.ConfigMapList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/configmap/{namespace}").
			To(apiHandler.handleGetConfigMapList).
			Writes(configmap.ConfigMapList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/configmap/{namespace}/{configmap}").
			To(apiHandler.handleGetConfigMapDetail).
			Writes(configmap.ConfigMapDetail{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/service").
			To(apiHandler.handleGetServiceList).
			Writes(resourceService.ServiceList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/service/{namespace}").
			To(apiHandler.handleGetServiceList).
			Writes(resourceService.ServiceList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/service/{namespace}/{service}").
			To(apiHandler.handleGetServiceDetail).
			Writes(resourceService.ServiceDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/service/{namespace}/{service}/event").
			To(apiHandler.handleGetServiceEvent).
			Writes(common.EventList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/service/{namespace}/{service}/pod").
			To(apiHandler.handleGetServicePods).
			Writes(pod.PodList{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/service").
			To(apiHandler.handleGetServiceListWithMultiTenancy).
			Writes(resourceService.ServiceList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/service/{namespace}").
			To(apiHandler.handleGetServiceListWithMultiTenancy).
			Writes(resourceService.ServiceList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/service/{namespace}/{service}").
			To(apiHandler.handleGetServiceDetailWithMultiTenancy).
			Writes(resourceService.ServiceDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/service/{namespace}/{service}/event").
			To(apiHandler.handleGetServiceEventWithMultiTenancy).
			Writes(common.EventList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/service/{namespace}/{service}/pod").
			To(apiHandler.handleGetServicePodsWithMultiTenancy).
			Writes(pod.PodList{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/ingress").
			To(apiHandler.handleGetIngressList).
			Writes(ingress.IngressList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/ingress/{namespace}").
			To(apiHandler.handleGetIngressList).
			Writes(ingress.IngressList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/ingress/{namespace}/{name}").
			To(apiHandler.handleGetIngressDetail).
			Writes(ingress.IngressDetail{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/ingress").
			To(apiHandler.handleGetIngressListWithMultiTenancy).
			Writes(ingress.IngressList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/ingress/{namespace}").
			To(apiHandler.handleGetIngressListWithMultiTenancy).
			Writes(ingress.IngressList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/ingress/{namespace}/{name}").
			To(apiHandler.handleGetIngressDetailWithMultiTenancy).
			Writes(ingress.IngressDetail{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/statefulset").
			To(apiHandler.handleGetStatefulSetList).
			Writes(statefulset.StatefulSetList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/statefulset/{namespace}").
			To(apiHandler.handleGetStatefulSetList).
			Writes(statefulset.StatefulSetList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/statefulset/{namespace}/{statefulset}").
			To(apiHandler.handleGetStatefulSetDetail).
			Writes(statefulset.StatefulSetDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/statefulset/{namespace}/{statefulset}/pod").
			To(apiHandler.handleGetStatefulSetPods).
			Writes(pod.PodList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/statefulset/{namespace}/{statefulset}/event").
			To(apiHandler.handleGetStatefulSetEvents).
			Writes(common.EventList{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/statefulset").
			To(apiHandler.handleGetStatefulSetListWithMultitenancy).
			Writes(statefulset.StatefulSetList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/statefulset/{namespace}").
			To(apiHandler.handleGetStatefulSetListWithMultitenancy).
			Writes(statefulset.StatefulSetList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/statefulset/{namespace}/{statefulset}").
			To(apiHandler.handleGetStatefulSetDetailWithMultiTenancy).
			Writes(statefulset.StatefulSetDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/statefulset/{namespace}/{statefulset}/pod").
			To(apiHandler.handleGetStatefulSetPodsWithMultiTenancy).
			Writes(pod.PodList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/statefulset/{namespace}/{statefulset}/event").
			To(apiHandler.handleGetStatefulSetEventsWithMultiTenancy).
			Writes(common.EventList{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/node").
			To(apiHandler.handleGetNodeLists).
			Writes(node.NodeList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/node/{name}").
			To(apiHandler.handleGetNodeDetail).
			Writes(node.NodeDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/node/{name}/event").
			To(apiHandler.handleGetNodeEvents).
			Writes(common.EventList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/node/{name}/pod").
			To(apiHandler.handleGetNodePods).
			Writes(pod.PodList{}))

	apiV1Ws.Route(
		apiV1Ws.DELETE("/_raw/{kind}/namespace/{namespace}/name/{name}").
			To(apiHandler.handleDeleteResource))
	apiV1Ws.Route(
		apiV1Ws.GET("/_raw/{kind}/namespace/{namespace}/name/{name}").
			To(apiHandler.handleGetResource))
	apiV1Ws.Route(
		apiV1Ws.PUT("/_raw/{kind}/namespace/{namespace}/name/{name}").
			To(apiHandler.handlePutResource))

	apiV1Ws.Route(
		apiV1Ws.DELETE("/tenants/{tenant}/_raw/{kind}/namespace/{namespace}/name/{name}").
			To(apiHandler.handleDeleteResourceWithMultiTenancy))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/_raw/{kind}/namespace/{namespace}/name/{name}").
			To(apiHandler.handleGetResourceWithMultiTenancy))
	apiV1Ws.Route(
		apiV1Ws.PUT("/tenants/{tenant}/_raw/{kind}/namespace/{namespace}/name/{name}").
			To(apiHandler.handlePutResourceWithMultiTenancy))

	apiV1Ws.Route(
		apiV1Ws.DELETE("/_raw/{kind}/name/{name}").
			To(apiHandler.handleDeleteResource))
	apiV1Ws.Route(
		apiV1Ws.GET("/_raw/{kind}/name/{name}").
			To(apiHandler.handleGetResource))
	apiV1Ws.Route(
		apiV1Ws.PUT("/_raw/{kind}/name/{name}").
			To(apiHandler.handlePutResource))

	apiV1Ws.Route(
		apiV1Ws.DELETE("/tenants/{tenant}/_raw/{kind}/name/{name}").
			To(apiHandler.handleDeleteResourceWithMultiTenancy))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/_raw/{kind}/name/{name}").
			To(apiHandler.handleGetResourceWithMultiTenancy))
	apiV1Ws.Route(
		apiV1Ws.PUT("/tenants/{tenant}/_raw/{kind}/name/{name}").
			To(apiHandler.handlePutResourceWithMultiTenancy))

	apiV1Ws.Route(
		apiV1Ws.GET("/role/{namespace}").
			To(apiHandler.handleGetRoleList).
			Writes(role.RoleList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/role/{namespace}/{name}").
			To(apiHandler.handleGetRoleDetail).
			Writes(role.RoleDetail{}))
	apiV1Ws.Route(
		apiV1Ws.POST("/role").
			To(apiHandler.handleCreateRole).
			Reads(role.Role{}).
			Writes(role.Role{}))
	apiV1Ws.Route(
		apiV1Ws.DELETE("/namespaces/{namespace}/role/{role}").
			To(apiHandler.handleDeleteRole))

	apiV1Ws.Route(
		apiV1Ws.GET("/clusterrole").
			To(apiHandler.handleGetClusterRoleList).
			Writes(clusterrole.ClusterRoleList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/clusterrole/{name}").
			To(apiHandler.handleGetClusterRoleDetail).
			Writes(clusterrole.ClusterRoleDetail{}))
	apiV1Ws.Route(
		apiV1Ws.POST("/clusterrole").
			To(apiHandler.handleCreateCreateClusterRole).
			Reads(clusterrole.ClusterRoleSpec{}).
			Writes(clusterrole.ClusterRoleSpec{}))

	apiV1Ws.Route(
		apiV1Ws.POST("/clusterroles").
			To(apiHandler.handleCreateCreateClusterRolesWithMultiTenancy).
			Reads(clusterrole.ClusterRoleSpec{}).
			Writes(clusterrole.ClusterRoleSpec{}))

	apiV1Ws.Route(
		apiV1Ws.POST("/rolebinding").
			To(apiHandler.handleCreateRoleBindings).
			Reads(rolebinding.RoleBinding{}).
			Writes(rolebinding.RoleBinding{}))
	apiV1Ws.Route(
		apiV1Ws.DELETE("/namespaces/{namespace}/rolebindings/{rolebinding}").
			To(apiHandler.handleDeleteRoleBindings))

	apiV1Ws.Route(
		apiV1Ws.POST("/rolebindings").
			To(apiHandler.handleCreateRoleBindingsWithMultiTenancy).
			Reads(rolebinding.RoleBinding{}).
			Writes(rolebinding.RoleBinding{}))
	apiV1Ws.Route(
		apiV1Ws.DELETE("tenants/{tenant}/namespaces/{namespace}/rolebindings/{rolebinding}").
			To(apiHandler.handleDeleteRoleBindingsWithMultiTenancy))

	apiV1Ws.Route(
		apiV1Ws.POST("/clusterrolebinding").
			To(apiHandler.handleCreateClusterRoleBindings).
			Reads(clusterrolebinding.ClusterRoleBinding{}).
			Writes(clusterrolebinding.ClusterRoleBinding{}))
	apiV1Ws.Route(
		apiV1Ws.DELETE("/clusterrolebindings/{clusterrolebinding}").
			To(apiHandler.handleDeleteClusterRoleBindings))

	apiV1Ws.Route(
		apiV1Ws.DELETE("/clusterrole/{clusterrole}").
			To(apiHandler.handleDeleteClusterRole))

	apiV1Ws.Route(
		apiV1Ws.POST("/clusterrolebindings").
			To(apiHandler.handleCreateClusterRoleBindingsWithMultiTenancy).
			Reads(clusterrolebinding.ClusterRoleBinding{}).
			Writes(clusterrolebinding.ClusterRoleBinding{}))
	apiV1Ws.Route(
		apiV1Ws.DELETE("tenants/{tenant}/clusterrolebindings/{clusterrolebinding}").
			To(apiHandler.handleDeleteClusterRoleBindingsWithMultiTenancy))

	apiV1Ws.Route(
		apiV1Ws.GET("/role").
			To(apiHandler.handleGetRoles).
			Writes(role.RoleList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/role/{namespace}").
			To(apiHandler.handleGetRolesWithMultiTenancy).
			Writes(role.RoleList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/role").
			To(apiHandler.handleGetRolesWithMultiTenancy).
			Writes(role.RoleList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenant/{tenant}/namespace/{namespace}/roles").
			To(apiHandler.handleGetRolesWithMultiTenancy).
			Writes(role.RoleList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/role/{namespace}/{name}").
			To(apiHandler.handleGetRoleDetailWithMultiTenancy).
			Writes(role.RoleDetail{}))
	apiV1Ws.Route(
		apiV1Ws.POST("/roles").
			To(apiHandler.handleCreateRolesWithMultiTenancy).
			Reads(role.Role{}).
			Writes(role.Role{}))
	apiV1Ws.Route(
		apiV1Ws.DELETE("/tenants/{tenant}/namespaces/{namespace}/role/{role}").
			To(apiHandler.handleDeleteRolesWithMultiTenancy))

	apiV1Ws.Route(
		apiV1Ws.GET("/serviceaccounts").
			To(apiHandler.handleGetServiceAccountList).
			Writes(serviceaccount.ServiceAccountList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/serviceaccount/{namespace}/{serviceaccount}").
			To(apiHandler.handleGetServiceAccountDetail).
			Writes(serviceaccount.ServiceAccountDetail{}))
	apiV1Ws.Route(
		apiV1Ws.POST("/serviceaccounts").
			To(apiHandler.handleCreateServiceAccount).
			Reads(serviceaccount.ServiceAccount{}).
			Writes(serviceaccount.ServiceAccount{}))
	apiV1Ws.Route(
		apiV1Ws.DELETE("/namespace/{namespace}/serviceaccount/{serviceaccount}").
			To(apiHandler.handleDeleteServiceAccount))

	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/namespaces/{namespace}/serviceaccount").
			To(apiHandler.handleGetServiceAccountListWithMultiTenancy).
			Writes(serviceaccount.ServiceAccountList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/namespaces/{namespace}/serviceaccounts/{name}").
			To(apiHandler.handleGetServiceAccountDetailWithMultiTenancy).
			Writes(serviceaccount.ServiceAccountDetail{}))
	apiV1Ws.Route(
		apiV1Ws.POST("/serviceaccounts").
			To(apiHandler.handleCreateServiceAccountsWithMultiTenancy).
			Reads(serviceaccount.ServiceAccount{}).
			Writes(serviceaccount.ServiceAccount{}))
	apiV1Ws.Route(
		apiV1Ws.DELETE("/tenants/{tenant}/namespaces/{namespace}/serviceaccounts/{serviceaccount}").
			To(apiHandler.handleDeleteServiceAccountsWithMultiTenancy))

	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/clusterrole").
			To(apiHandler.handleGetClusterRoleListWithMultiTenancy).
			Writes(clusterrole.ClusterRoleList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/clusterrole/{name}").
			To(apiHandler.handleGetClusterRoleDetailWithMultiTenancy).
			Writes(clusterrole.ClusterRoleDetail{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/persistentvolume").
			To(apiHandler.handleGetPersistentVolumeList).
			Writes(persistentvolume.PersistentVolumeList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/persistentvolume/{persistentvolume}").
			To(apiHandler.handleGetPersistentVolumeDetail).
			Writes(persistentvolume.PersistentVolumeDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/persistentvolume/namespace/{namespace}/name/{persistentvolume}").
			To(apiHandler.handleGetPersistentVolumeDetail).
			Writes(persistentvolume.PersistentVolumeDetail{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/persistentvolume").
			To(apiHandler.handleGetPersistentVolumeListWithMultiTenancy).
			Writes(persistentvolume.PersistentVolumeList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/persistentvolume/{persistentvolume}").
			To(apiHandler.handleGetPersistentVolumeDetailWithMultiTenancy).
			Writes(persistentvolume.PersistentVolumeDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/persistentvolume/namespace/{namespace}/name/{persistentvolume}").
			To(apiHandler.handleGetPersistentVolumeDetailWithMultiTenancy).
			Writes(persistentvolume.PersistentVolumeDetail{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/persistentvolumeclaim/").
			To(apiHandler.handleGetPersistentVolumeClaimList).
			Writes(persistentvolumeclaim.PersistentVolumeClaimList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/persistentvolumeclaim/{namespace}").
			To(apiHandler.handleGetPersistentVolumeClaimList).
			Writes(persistentvolumeclaim.PersistentVolumeClaimList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/persistentvolumeclaim/{namespace}/{name}").
			To(apiHandler.handleGetPersistentVolumeClaimDetail).
			Writes(persistentvolumeclaim.PersistentVolumeClaimDetail{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/persistentvolumeclaim/").
			To(apiHandler.handleGetPersistentVolumeClaimList).
			Writes(persistentvolumeclaim.PersistentVolumeClaimList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/persistentvolumeclaim/{namespace}").
			To(apiHandler.handleGetPersistentVolumeClaimList).
			Writes(persistentvolumeclaim.PersistentVolumeClaimList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/persistentvolumeclaim/{namespace}/{name}").
			To(apiHandler.handleGetPersistentVolumeClaimDetail).
			Writes(persistentvolumeclaim.PersistentVolumeClaimDetail{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/crd").
			To(apiHandler.handleGetCustomResourceDefinitionList).
			Writes(customresourcedefinition.CustomResourceDefinitionList{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/crd/{crd}").
			To(apiHandler.handleGetCustomResourceDefinitionDetail).
			Writes(customresourcedefinition.CustomResourceDefinitionDetail{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/crd/{namespace}/{crd}/object").
			To(apiHandler.handleGetCustomResourceObjectList).
			Writes(customresourcedefinition.CustomResourceObjectList{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/crd/{namespace}/{crd}/{object}").
			To(apiHandler.handleGetCustomResourceObjectDetail).
			Writes(customresourcedefinition.CustomResourceObject{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/crd/{namespace}/{crd}/{object}/event").
			To(apiHandler.handleGetCustomResourceObjectEvents).
			Writes(common.EventList{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/crd").
			To(apiHandler.handleGetCustomResourceDefinitionListWithMultiTenancy).
			Writes(customresourcedefinition.CustomResourceDefinitionList{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/crd/{crd}").
			To(apiHandler.handleGetCustomResourceDefinitionDetailWithMultiTenancy).
			Writes(customresourcedefinition.CustomResourceDefinitionDetail{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/crd/{crd}/object").
			To(apiHandler.handleGetCustomResourceObjectListWithMultiTenancy).
			Writes(customresourcedefinition.CustomResourceObjectList{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/crd/{crd}/{object}").
			To(apiHandler.handleGetCustomResourceObjectDetailWithMultiTenancy).
			Writes(customresourcedefinition.CustomResourceObject{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/crd/{namespace}/{crd}/{object}/event").
			To(apiHandler.handleGetCustomResourceObjectEventsWithMultiTenancy).
			Writes(common.EventList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/crd/{crd}/{object}/event").
			To(apiHandler.handleGetCustomResourceObjectEventsWithMultiTenancy).
			Writes(common.EventList{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/storageclass").
			To(apiHandler.handleGetStorageClassList).
			Writes(storageclass.StorageClassList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/storageclass/{storageclass}").
			To(apiHandler.handleGetStorageClass).
			Writes(storageclass.StorageClass{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/storageclass/{storageclass}/persistentvolume").
			To(apiHandler.handleGetStorageClassPersistentVolumes).
			Writes(persistentvolume.PersistentVolumeList{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/storageclass").
			To(apiHandler.handleGetStorageClassListWithMultiTenancy).
			Writes(storageclass.StorageClassList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/storageclass/{storageclass}").
			To(apiHandler.handleGetStorageClassWithMultiTenancy).
			Writes(storageclass.StorageClass{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/storageclass/{storageclass}/persistentvolume").
			To(apiHandler.handleGetStorageClassPersistentVolumesWithMultiTenancy).
			Writes(persistentvolume.PersistentVolumeList{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/log/source/{namespace}/{resourceName}/{resourceType}").
			To(apiHandler.handleLogSource).
			Writes(controller.LogSources{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/log/{namespace}/{pod}").
			To(apiHandler.handleLogs).
			Writes(logs.LogDetails{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/log/{namespace}/{pod}/{container}").
			To(apiHandler.handleLogs).
			Writes(logs.LogDetails{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/log/file/{namespace}/{pod}/{container}").
			To(apiHandler.handleLogFile).
			Writes(logs.LogDetails{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/log/source/{namespace}/{resourceName}/{resourceType}").
			To(apiHandler.handleLogSourceWithMultiTenancy).
			Writes(controller.LogSources{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/log/{namespace}/{pod}").
			To(apiHandler.handleLogsWithMultiTenancy).
			Writes(logs.LogDetails{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/log/{namespace}/{pod}/{container}").
			To(apiHandler.handleLogsWithMultiTenancy).
			Writes(logs.LogDetails{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/log/file/{namespace}/{pod}/{container}").
			To(apiHandler.handleLogFileWithMultiTenancy).
			Writes(logs.LogDetails{}))

	// IAM User related routes
	apiV1Ws.Route(
		apiV1Ws.POST("/users").
			To(apiHandler.handleCreateUser).
			Reads(model.User{}).
			Writes(model.User{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/tenants/{tenant}/users").
			To(apiHandler.handleGetAllUser).
			Writes(model.User{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/users/{username}").
			To(apiHandler.handleGetUser).
			Writes(model.User{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/user/{username}").
			To(apiHandler.handleGetUserDetail).
			Writes(model.User{}))
	apiV1Ws.Route(
		apiV1Ws.DELETE("/tenants/{tenant}/users/{username}/{userid}").
			To(apiHandler.handleDeleteUser).
			Writes(model.User{}))

	return wsContainer, nil
}

//error struct for already exists
type Error struct {
	// Name of the tenant.
	Msg        string `json:"msg"`
	StatusCode int    `json:"statusCode"`
}

//for tenant handlerCreateTenant method
func (apiHandler *APIHandlerV2) handleCreateTenant(request *restful.Request, response *restful.Response) {
	tenantSpec := new(tenant.TenantSpec)
	if err := request.ReadEntity(tenantSpec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	if len(apiHandler.tpManager) == 0 {
		apiHandler.tpManager = append(apiHandler.tpManager, apiHandler.defaultClientmanager)
	}
	userdetail, _ := db.GetUser(tenantSpec.Username)
	if userdetail.ObjectMeta.Username != "" {
		errors.HandleInternalError(response, errors.NewInternal("User already exists"))
		return
	}
	client := ResourceAllocator(tenantSpec.Name, apiHandler.tpManager)
	k8sClient := client.InsecureClient()
	//k8sClient, err := client.Client(request)
	//if err != nil {
	//	errors.HandleInternalError(response, err)
	//	return
	//}

	if err := tenant.CreateTenant(tenantSpec, k8sClient, client.GetClusterName()); err != nil {
		errorMsg := Error{Msg: err.Error(), StatusCode: http.StatusConflict}
		response.WriteHeaderAndEntity(http.StatusConflict, errorMsg)
		return
	}
	userSpec := model.User{
		Username:          tenantSpec.Username,
		Password:          tenantSpec.Password,
		Token:             "",
		Type:              "tenant-admin",
		Tenant:            tenantSpec.Name,
		Role:              tenantSpec.Name + "-admin",
		NameSpace:         "default",
		CreationTimestamp: time.Time{},
	}
	user, err := iam.TenantAdmin(userSpec, client)
	if err != nil {
		log.Printf("Error creating tenant admin user: %s", err.Error())
	}
	_ = db.InsertUser(user)
	tenantSpec.Username = user.Username
	tenantSpec.Password = user.Password
	response.WriteHeaderAndEntity(http.StatusCreated, tenantSpec)
}

//for delete tenant
func (apiHandler *APIHandlerV2) handleDeleteTenant(request *restful.Request, response *restful.Response) {
	//tenant := request.PathParameter("tenant")
	tenantName := request.PathParameter("tenant")
	client := ResourceAllocator(tenantName, apiHandler.tpManager)
	k8sClient := client.InsecureClient()

	if err := tenant.DeleteTenant(tenantName, k8sClient); err != nil {
		errors.HandleInternalError(response, err)
		return
	} else {
		db.DeleteTenantUser(tenantName)
	}
	response.WriteHeader(http.StatusOK)
}

func (apiHandler *APIHandlerV2) handleGetTenantList(request *restful.Request, response *restful.Response) {
	var tenantsList tenant.TenantList
	tenantName := request.PathParameter("tenant")
	cManager := ResourceAllocator(tenantName, apiHandler.tpManager)
	_, err := cManager.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	if len(apiHandler.tpManager) == 0 {
		apiHandler.tpManager = append(apiHandler.tpManager, apiHandler.defaultClientmanager)
	}

	dataSelect := parseDataSelectPathParameter(request)
	for _, tpManager := range apiHandler.tpManager {
		k8sClient := tpManager.InsecureClient()
		dataSelect := dataselect.NewDataSelectQuery(dataselect.NoPagination, dataselect.NoSort, dataselect.NoFilter, dataselect.NoMetrics)
		result, err := tenant.GetTenantList(k8sClient, dataSelect, tpManager.GetClusterName(), tenantName)
		if err != nil {
			errors.HandleInternalError(response, err)
			return
		}
		for _, tenants := range result.Tenants {
			tenants.ClusterName = tpManager.GetClusterName()
			tenantsList.Tenants = append(tenantsList.Tenants, tenants)
			tenantsList.ListMeta.TotalItems++
		}

	}
	tenantCells, filteredTotal := dataselect.GenericDataSelectWithFilter(toCells(tenantsList.Tenants), dataSelect)
	tenantsList.Tenants = fromCells(tenantCells)

	tenantsList.ListMeta = api.ListMeta{TotalItems: filteredTotal}

	response.WriteHeaderAndEntity(http.StatusOK, tenantsList)
}

type TenantCell tenant.Tenant

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
func toCells(std []tenant.Tenant) []dataselect.DataCell {
	cells := make([]dataselect.DataCell, len(std))
	for i := range std {
		cells[i] = TenantCell(std[i])
	}
	return cells
}
func fromCells(cells []dataselect.DataCell) []tenant.Tenant {
	std := make([]tenant.Tenant, len(cells))
	for i := range std {
		std[i] = tenant.Tenant(cells[i].(TenantCell))
	}
	return std
}
func (apiHandler *APIHandlerV2) handleGetTenantDetail(request *restful.Request, response *restful.Response) {
	tenantName := request.PathParameter("name")
	if len(apiHandler.tpManager) == 0 {
		apiHandler.tpManager = append(apiHandler.tpManager, apiHandler.defaultClientmanager)
	}
	client := ResourceAllocator(tenantName, apiHandler.tpManager)
	c, err := request.Request.Cookie("tenant")
	var CookieTenant string
	if err != nil {
		log.Printf("Cookie error: %v", err)
		CookieTenant = tenantName
	} else {
		CookieTenant = c.Value
	}
	log.Printf("cookie_tenant is: %s", CookieTenant)
	var k8sClient kubernetes.Interface
	if tenantName != CookieTenant {
		k8sClient = client.InsecureClient()
	} else {
		k8sClient, err = client.Client(request)
		if err != nil {
			errors.HandleInternalError(response, err)
			return
		}
	}

	result, err := tenant.GetTenantDetail(k8sClient, tenantName)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetRoleList(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := parseNamespacePathParameter(request)
	dataSelect := parseDataSelectPathParameter(request)
	result, err := role.GetRoleList(k8sClient, namespace, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetClusterRoleList(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	dataSelect := parseDataSelectPathParameter(request)
	result, err := clusterrole.GetClusterRoleList(k8sClient, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetClusterRoleListWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	dataSelect := parseDataSelectPathParameter(request)
	result, err := clusterrole.GetClusterRoleListWithMultiTenancy(k8sClient, tenant, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetClusterRoleDetail(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	name := request.PathParameter("name")
	result, err := clusterrole.GetClusterRoleDetail(k8sClient, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetClusterRoleDetailWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	name := request.PathParameter("name")
	result, err := clusterrole.GetClusterRoleDetailWithMultiTenancy(k8sClient, tenant, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetCsrfToken(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	if tenant == "" {
		tenant = "system"
	}

	client := ResourceAllocator(tenant, apiHandler.tpManager)

	action := request.PathParameter("action")
	token := xsrftoken.Generate(client.CSRFKey(), "none", action)
	response.WriteHeaderAndEntity(http.StatusOK, api.CsrfToken{Token: token})
}

func (apiHandler *APIHandlerV2) handleGetStatefulSetList(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := parseNamespacePathParameter(request)
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := statefulset.GetStatefulSetList(k8sClient, namespace, dataSelect,
		apiHandler.iManager.Metric().Client())
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetStatefulSetListWithMultitenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := parseNamespacePathParameter(request)
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := statefulset.GetStatefulSetListWithMultiTenancy(k8sClient, tenant, namespace, dataSelect,
		apiHandler.iManager.Metric().Client())
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetStatefulSetDetail(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("statefulset")
	result, err := statefulset.GetStatefulSetDetail(k8sClient, apiHandler.iManager.Metric().Client(), namespace, name)

	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetStatefulSetDetailWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("statefulset")
	result, err := statefulset.GetStatefulSetDetailWithMultiTenancy(k8sClient, apiHandler.iManager.Metric().Client(), tenant, namespace, name)

	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetStatefulSetPods(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("statefulset")
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := statefulset.GetStatefulSetPods(k8sClient, apiHandler.iManager.Metric().Client(), dataSelect, name, namespace)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetStatefulSetPodsWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("statefulset")
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := statefulset.GetStatefulSetPods(k8sClient, apiHandler.iManager.Metric().Client(), dataSelect, name, namespace)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetStatefulSetEvents(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("statefulset")
	dataSelect := parseDataSelectPathParameter(request)
	result, err := event.GetResourceEvents(k8sClient, dataSelect, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetStatefulSetEventsWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("statefulset")
	dataSelect := parseDataSelectPathParameter(request)
	result, err := event.GetResourceEventsWithMultiTenancy(k8sClient, dataSelect, tenant, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetServiceList(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := parseNamespacePathParameter(request)
	dataSelect := parseDataSelectPathParameter(request)
	result, err := resourceService.GetServiceList(k8sClient, namespace, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetServiceListWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := parseNamespacePathParameter(request)
	dataSelect := parseDataSelectPathParameter(request)
	result, err := resourceService.GetServiceListWithMultiTenancy(k8sClient, tenant, namespace, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetServiceDetail(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("service")
	result, err := resourceService.GetServiceDetail(k8sClient, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetServiceDetailWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("service")
	result, err := resourceService.GetServiceDetail(k8sClient, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetServiceEvent(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("service")
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := resourceService.GetServiceEvents(k8sClient, dataSelect, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetServiceEventWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("service")
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := resourceService.GetServiceEventsWithMultiTenancy(k8sClient, dataSelect, tenant, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetIngressDetail(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")
	result, err := ingress.GetIngressDetail(k8sClient, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetIngressDetailWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")
	result, err := ingress.GetIngressDetailWithMultiTenancy(k8sClient, tenant, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetIngressList(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	dataSelect := parseDataSelectPathParameter(request)
	namespace := parseNamespacePathParameter(request)
	result, err := ingress.GetIngressList(k8sClient, namespace, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetIngressListWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	dataSelect := parseDataSelectPathParameter(request)
	namespace := parseNamespacePathParameter(request)
	result, err := ingress.GetIngressListWithMultiTenancy(k8sClient, tenant, namespace, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetServicePods(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("service")
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := resourceService.GetServicePods(k8sClient, apiHandler.iManager.Metric().Client(), namespace, name, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetServicePodsWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("service")
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := resourceService.GetServicePodsWithMultiTenancy(k8sClient, apiHandler.iManager.Metric().Client(), tenant, namespace, name, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetNodeLists(request *restful.Request, response *restful.Response) {
	var nodeLists node.NodeList
	//For tpclients
	var clients []clientapi.ClientManager

	if len(apiHandler.tpManager) == 0 && len(apiHandler.rpManager) == 0 {
		clients = append(clients, apiHandler.defaultClientmanager)
	}
	for _, tpManager := range apiHandler.tpManager {
		k8sClient := tpManager.InsecureClient()
		//if err != nil {
		//	errors.HandleInternalError(response, err)
		//	return
		//}
		dataSelect := parseDataSelectPathParameter(request)
		dataSelect.MetricQuery = dataselect.StandardMetrics
		result, err := node.GetNodeList(k8sClient, dataSelect, apiHandler.iManager.Metric().Client(), tpManager.GetClusterName())
		if err != nil {
			errors.HandleInternalError(response, err)
			return
		}
		for _, nodes := range result.Nodes {
			nodeLists.Nodes = append(nodeLists.Nodes, nodes)
			nodeLists.ListMeta.TotalItems++
		}
	}
	//For rpclients
	for _, rpManager := range apiHandler.rpManager {
		//k8sClient, err := rpManager.Client(request)
		//if err != nil {
		//	errors.HandleInternalError(response, err)
		//	return
		//}
		if apiHandler.tpManager[0].GetClusterName() == rpManager.GetClusterName() {
			continue
		}
		k8sClient := rpManager.InsecureClient()
		dataSelect := parseDataSelectPathParameter(request)
		dataSelect.MetricQuery = dataselect.StandardMetrics
		result, err := node.GetNodeList(k8sClient, dataSelect, apiHandler.iManager.Metric().Client(), rpManager.GetClusterName())
		if err != nil {
			errors.HandleInternalError(response, err)
			return
		}
		for _, nodes := range result.Nodes {
			nodeLists.Nodes = append(nodeLists.Nodes, nodes)
			nodeLists.ListMeta.TotalItems++
		}
	}
	response.WriteHeaderAndEntity(http.StatusOK, nodeLists)

}
func (apiHandler *APIHandlerV2) handleGetResourcePartitionDetail(request *restful.Request, response *restful.Response) {
	//For rpclients
	if len(apiHandler.rpManager) == 0 {
		apiHandler.rpManager = append(apiHandler.rpManager, apiHandler.defaultClientmanager)
	}
	result := new(partition.ResourcePartitionList)
	for _, rpManager := range apiHandler.rpManager {
		k8sClient := rpManager.InsecureClient()
		//if err != nil {
		//	errors.HandleInternalError(response, err)
		//	return
		//}
		dataSelect := parseDataSelectPathParameter(request)
		dataSelect.MetricQuery = dataselect.StandardMetrics
		partitionDetail, err := partition.GetResourcePartitionDetail(k8sClient, rpManager.GetClusterName())
		if err != nil {
			errors.HandleInternalError(response, err)
			return
		}
		result.Partitions = append(result.Partitions, partitionDetail)
	}
	result.ListMeta.TotalItems = len(result.Partitions)

	response.WriteHeaderAndEntity(http.StatusOK, result)

}

func (apiHandler *APIHandlerV2) handleGetTenantPartitionDetail(request *restful.Request, response *restful.Response) {
	//For tpclients
	result := new(partition.TenantPartitionList)
	if len(apiHandler.tpManager) == 0 {
		apiHandler.tpManager = append(apiHandler.tpManager, apiHandler.defaultClientmanager)
	}

	if len(apiHandler.rpManager) == 0 {
		apiHandler.rpManager = append(apiHandler.rpManager, apiHandler.defaultClientmanager)
	}
	var workerCount int64 = 0
	for _, rpManager := range apiHandler.rpManager {
		k8sClient := rpManager.InsecureClient()
		workerCount += partition.GetWorkerCount(k8sClient)
	}

	var PodList []interface{}
	for i, tpManager := range apiHandler.tpManager {
		k8sClient := tpManager.InsecureClient()
		dataSelect := parseDataSelectPathParameter(request)
		dataSelect.MetricQuery = dataselect.StandardMetrics
		if len(apiHandler.podInformerManager) != 0 {
			PodInformer := apiHandler.podInformerManager[i]
			PodList = PodInformer.GetStore().List()
			fmt.Printf("Checking nodes length: %v %v", len(PodList), PodList)
		}
		partitionDetail, err := partition.GetTenantPartitionDetail(k8sClient, tpManager.GetClusterName())
		if err != nil {
			errors.HandleInternalError(response, err)
			return
		}
		if len(PodList) != 0 {
			partitionDetail.ObjectMeta.PodCount = int64(len(PodList))
		}
		partitionDetail.ObjectMeta.TotalPods = partitionDetail.ObjectMeta.TotalPods * workerCount
		result.Partitions = append(result.Partitions, partitionDetail)
	}
	result.ListMeta.TotalItems = len(result.Partitions)

	response.WriteHeaderAndEntity(http.StatusOK, result)

}

func (apiHandler *APIHandlerV2) handleGetNodeDetail(request *restful.Request, response *restful.Response) {
	//k8sClient, err := apiHandler.tpManager.Client(request)
	//if err != nil {
	//	errors.HandleInternalError(response, err)
	//	return
	//}
	if len(apiHandler.rpManager) == 0 {
		apiHandler.rpManager = append(apiHandler.rpManager, apiHandler.defaultClientmanager)
	}
	if len(apiHandler.tpManager) == 0 {
		apiHandler.tpManager = append(apiHandler.tpManager, apiHandler.defaultClientmanager)
	}
	name := request.PathParameter("name")
	var k8sClient kubernetes.Interface
	var err error
	var clusterName string
	for _, rpManager := range apiHandler.rpManager {
		k8sClient = rpManager.InsecureClient()
		dataSelect := parseDataSelectPathParameter(request)
		dataSelect.MetricQuery = dataselect.StandardMetrics
		clusterName = rpManager.GetClusterName()
		_, err = node.GetNodeDetail(k8sClient, apiHandler.iManager.Metric().Client(), name, dataSelect, clusterName)
		if err != nil {
			log.Printf("Invalid Client or Internal Error %s", err.Error())
			//errors.HandleInternalError(response, err)
		} else {
			break
		}
	}
	if err != nil {
		for _, tpManager := range apiHandler.tpManager {
			k8sClient = tpManager.InsecureClient()
			dataSelect := parseDataSelectPathParameter(request)
			clusterName = tpManager.GetClusterName()
			dataSelect.MetricQuery = dataselect.StandardMetrics
			_, err = node.GetNodeDetail(k8sClient, apiHandler.iManager.Metric().Client(), name, dataSelect, clusterName)
			if err != nil {
				log.Printf("Invalid Client or Internal Error %s", err.Error())
				//errors.HandleInternalError(response, err)
			} else {
				break
			}
		}
	}
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := node.GetNodeDetail(k8sClient, apiHandler.iManager.Metric().Client(), name, dataSelect, clusterName)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetNodeEvents(request *restful.Request, response *restful.Response) {
	//k8sClient, err := apiHandler.tpManager.Client(request)
	//if err != nil {
	//	errors.HandleInternalError(response, err)
	//	return
	//}
	name := request.PathParameter("name")
	var k8sClient kubernetes.Interface
	var err error
	var clusterName string
	if len(apiHandler.tpManager) == 0 {
		apiHandler.tpManager = append(apiHandler.tpManager, apiHandler.defaultClientmanager)
	}
	if len(apiHandler.rpManager) == 0 {
		apiHandler.rpManager = append(apiHandler.rpManager, apiHandler.defaultClientmanager)
	}
	for _, rpManager := range apiHandler.rpManager {
		k8sClient = rpManager.InsecureClient()
		dataSelect := parseDataSelectPathParameter(request)
		dataSelect.MetricQuery = dataselect.StandardMetrics
		clusterName = rpManager.GetClusterName()
		_, err = node.GetNodeDetail(k8sClient, apiHandler.iManager.Metric().Client(), name, dataSelect, clusterName)
		if err != nil {
			log.Printf("Invalid Client or Internal Error %s", err.Error())
			//errors.HandleInternalError(response, err)
		} else {
			break
		}
	}
	if err != nil {
		for _, tpManager := range apiHandler.tpManager {
			k8sClient = tpManager.InsecureClient()
			dataSelect := parseDataSelectPathParameter(request)
			dataSelect.MetricQuery = dataselect.StandardMetrics
			clusterName = tpManager.GetClusterName()
			_, err = node.GetNodeDetail(k8sClient, apiHandler.iManager.Metric().Client(), name, dataSelect, clusterName)
			if err != nil {
				log.Printf("Invalid Client or Internal Error %s", err.Error())
				//errors.HandleInternalError(response, err)
			} else {
				break
			}
		}
	}
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := event.GetNodeEvents(k8sClient, dataSelect, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetNodePods(request *restful.Request, response *restful.Response) {
	//k8sClient, err := apiHandler.tpManager.Client(request)
	//if err != nil {
	//	errors.HandleInternalError(response, err)
	//	return
	//}
	if len(apiHandler.tpManager) == 0 {
		apiHandler.tpManager = append(apiHandler.tpManager, apiHandler.defaultClientmanager)
	}
	if len(apiHandler.rpManager) == 0 {
		apiHandler.rpManager = append(apiHandler.rpManager, apiHandler.defaultClientmanager)
	}
	name := request.PathParameter("name")
	var k8sClient kubernetes.Interface
	var err error
	var clusterName string
	for _, rpManager := range apiHandler.rpManager {
		k8sClient = rpManager.InsecureClient()
		dataSelect := parseDataSelectPathParameter(request)
		dataSelect.MetricQuery = dataselect.StandardMetrics
		clusterName = rpManager.GetClusterName()
		_, err = node.GetNodeDetail(k8sClient, apiHandler.iManager.Metric().Client(), name, dataSelect, clusterName)
		if err != nil {
			log.Printf("Invalid Client or Internal Error %s", err.Error())
			//errors.HandleInternalError(response, err)
		} else {
			break
		}
	}
	if err != nil {
		for _, tpManager := range apiHandler.tpManager {
			k8sClient = tpManager.InsecureClient()
			dataSelect := parseDataSelectPathParameter(request)
			dataSelect.MetricQuery = dataselect.StandardMetrics
			clusterName = tpManager.GetClusterName()
			_, err = node.GetNodeDetail(k8sClient, apiHandler.iManager.Metric().Client(), name, dataSelect, clusterName)
			if err != nil {
				log.Printf("Invalid Client or Internal Error %s", err.Error())
				//errors.HandleInternalError(response, err)
			} else {
				break
			}
		}
	}
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := node.GetNodePods(k8sClient, apiHandler.iManager.Metric().Client(), dataSelect, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleDeploy(request *restful.Request, response *restful.Response) {

	appDeploymentSpec := new(deployment.AppDeploymentSpec)
	if err := request.ReadEntity(appDeploymentSpec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	client := ResourceAllocator(appDeploymentSpec.Tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	if err := deployment.DeployApp(appDeploymentSpec, k8sClient); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusCreated, appDeploymentSpec)
}

func (apiHandler *APIHandlerV2) handleScaleResource(request *restful.Request, response *restful.Response) {

	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)

	cfg, err := client.Config(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	kind := request.PathParameter("kind")
	name := request.PathParameter("name")
	count := request.QueryParameter("scaleBy")
	replicaCountSpec, err := scaling.ScaleResource(cfg, kind, namespace, name, count)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, replicaCountSpec)
}

func (apiHandler *APIHandlerV2) handleScaleResourceWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	cfg, err := client.Config(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	kind := request.PathParameter("kind")
	name := request.PathParameter("name")
	count := request.QueryParameter("scaleBy")
	replicaCountSpec, err := scaling.ScaleResourceWithMultiTenancy(cfg, tenant, kind, namespace, name, count)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, replicaCountSpec)
}

func (apiHandler *APIHandlerV2) handleGetReplicaCount(request *restful.Request, response *restful.Response) {
	log.Println("handleGetReplicaCount")
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	cfg, err := client.Config(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	kind := request.PathParameter("kind")
	name := request.PathParameter("name")
	scaleSpec, err := scaling.GetScaleSpec(cfg, kind, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, scaleSpec)
}

func (apiHandler *APIHandlerV2) handleGetReplicaCountWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	cfg, err := client.Config(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	kind := request.PathParameter("kind")
	name := request.PathParameter("name")
	scaleSpec, err := scaling.GetScaleSpecWithMultiTenancy(cfg, tenant, kind, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, scaleSpec)
}

func (apiHandler *APIHandlerV2) handleDeployFromFile(request *restful.Request, response *restful.Response) {
	//tenant := request.PathParameter("tenant")

	deploymentSpec := new(deployment.AppDeploymentFromFileSpec)
	if err := request.ReadEntity(deploymentSpec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	client := ResourceAllocator(deploymentSpec.Tenant, apiHandler.tpManager)
	cfg, err := client.Config(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	isDeployed, err := deployment.DeployAppFromFile(cfg, deploymentSpec)
	if !isDeployed {
		errors.HandleInternalError(response, err)
		return
	}

	errorMessage := ""
	if err != nil {
		errorMessage = err.Error()
	}

	response.WriteHeaderAndEntity(http.StatusCreated, deployment.AppDeploymentFromFileResponse{
		Name:    deploymentSpec.Name,
		Content: deploymentSpec.Content,
		Error:   errorMessage,
	})
}

func (apiHandler *APIHandlerV2) handleNameValidity(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	spec := new(validation.AppNameValiditySpec)
	if err := request.ReadEntity(spec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	validity, err := validation.ValidateAppName(spec, k8sClient)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, validity)
}

func (APIHandler *APIHandlerV2) handleImageReferenceValidity(request *restful.Request, response *restful.Response) {
	spec := new(validation.ImageReferenceValiditySpec)
	if err := request.ReadEntity(spec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	validity, err := validation.ValidateImageReference(spec)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, validity)
}

func (apiHandler *APIHandlerV2) handleProtocolValidity(request *restful.Request, response *restful.Response) {
	spec := new(validation.ProtocolValiditySpec)
	if err := request.ReadEntity(spec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, validation.ValidateProtocol(spec))
}

func (apiHandler *APIHandlerV2) handleGetAvailableProcotols(request *restful.Request, response *restful.Response) {
	response.WriteHeaderAndEntity(http.StatusOK, deployment.GetAvailableProtocols())
}

func (apiHandler *APIHandlerV2) handleGetReplicationControllerList(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := parseNamespacePathParameter(request)
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := replicationcontroller.GetReplicationControllerList(k8sClient, namespace, dataSelect, apiHandler.iManager.Metric().Client())
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetReplicationControllerListWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := parseNamespacePathParameter(request)
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := replicationcontroller.GetReplicationControllerListWithMultiTenancy(k8sClient, tenant, namespace, dataSelect, apiHandler.iManager.Metric().Client())
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetReplicaSets(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := parseNamespacePathParameter(request)
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := replicaset.GetReplicaSetList(k8sClient, namespace, dataSelect, apiHandler.iManager.Metric().Client())
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetReplicaSetsWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	c, err := request.Request.Cookie("tenant")
	var CookieTenant string
	if err != nil {
		log.Printf("Cookie error: %v", err)
		CookieTenant = tenant
	} else {
		CookieTenant = c.Value
	}
	log.Printf("cookie_tenant is: %s", CookieTenant)
	var k8sClient kubernetes.Interface
	if tenant != CookieTenant {
		k8sClient = client.InsecureClient()
	} else {
		k8sClient, err = client.Client(request)
		if err != nil {
			errors.HandleInternalError(response, err)
			return
		}
	}
	namespace := parseNamespacePathParameter(request)
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := replicaset.GetReplicaSetListWithMultiTenancy(k8sClient, tenant, namespace, dataSelect, apiHandler.iManager.Metric().Client())
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetReplicaSetDetail(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	replicaSet := request.PathParameter("replicaSet")
	result, err := replicaset.GetReplicaSetDetail(k8sClient, apiHandler.iManager.Metric().Client(), namespace, replicaSet)

	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetReplicaSetDetailWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	replicaSet := request.PathParameter("replicaSet")
	result, err := replicaset.GetReplicaSetDetailWithMultiTenancy(k8sClient, apiHandler.iManager.Metric().Client(), tenant, namespace, replicaSet)

	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetReplicaSetPods(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	replicaSet := request.PathParameter("replicaSet")
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := replicaset.GetReplicaSetPods(k8sClient, apiHandler.iManager.Metric().Client(), dataSelect, replicaSet, namespace)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetReplicaSetPodsWithMutiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	replicaSet := request.PathParameter("replicaSet")
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := replicaset.GetReplicaSetPodsWithMultiTenancy(k8sClient, apiHandler.iManager.Metric().Client(), tenant, dataSelect, replicaSet, namespace)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetReplicaSetServices(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	replicaSet := request.PathParameter("replicaSet")
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := replicaset.GetReplicaSetServices(k8sClient, dataSelect, namespace, replicaSet)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetReplicaSetServicesWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	replicaSet := request.PathParameter("replicaSet")
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := replicaset.GetReplicaSetServicesWithMultiTenancy(k8sClient, tenant, dataSelect, namespace, replicaSet)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetReplicaSetEvents(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("replicaSet")
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := event.GetResourceEvents(k8sClient, dataSelect, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)

}

func (apiHandler *APIHandlerV2) handleGetReplicaSetEventsWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("replicaSet")
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := event.GetResourceEventsWithMultiTenancy(k8sClient, dataSelect, tenant, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)

}

func (apiHandler *APIHandlerV2) handleGetPodEvents(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	log.Println("Getting events related to a pod in namespace")
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("pod")
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := pod.GetEventsForPod(k8sClient, dataSelect, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetPodEventsWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	log.Println("Getting events related to a pod in namespace")
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("pod")
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := pod.GetEventsForPodWithMultiTenancy(k8sClient, dataSelect, tenant, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

// Handles execute shell API call
func (apiHandler *APIHandlerV2) handleExecShell(request *restful.Request, response *restful.Response) {
	sessionId, err := genTerminalSessionId()
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	cfg, err := client.Config(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	terminalSessions.Set(sessionId, TerminalSession{
		id:       sessionId,
		bound:    make(chan error),
		sizeChan: make(chan remotecommand.TerminalSize),
	})
	go WaitForTerminal(k8sClient, cfg, request, sessionId)
	response.WriteHeaderAndEntity(http.StatusOK, TerminalResponse{Id: sessionId})
}

// Handles execute shell API call
func (apiHandler *APIHandlerV2) handleExecShellWithMultiTenancy(request *restful.Request, response *restful.Response) {
	sessionId, err := genTerminalSessionId()
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	cfg, err := client.Config(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	terminalSessions.Set(sessionId, TerminalSession{
		id:       sessionId,
		bound:    make(chan error),
		sizeChan: make(chan remotecommand.TerminalSize),
	})

	go WaitForTerminalWithMultiTenancy(k8sClient, cfg, request, sessionId, tenant)
	response.WriteHeaderAndEntity(http.StatusOK, TerminalResponse{Id: sessionId})
}

func (apiHandler *APIHandlerV2) handleGetDeployments(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := parseNamespacePathParameter(request)
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := deployment.GetDeploymentList(k8sClient, namespace, dataSelect, apiHandler.iManager.Metric().Client())
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetDeploymentsWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)

	c, err := request.Request.Cookie("tenant")
	var CookieTenant string
	if err != nil {
		log.Printf("Cookie error: %v", err)
		CookieTenant = tenant
	} else {
		CookieTenant = c.Value
	}
	log.Printf("cookie_tenant is: %s", CookieTenant)
	var k8sClient kubernetes.Interface
	if tenant != CookieTenant {
		k8sClient = client.InsecureClient()
	} else {
		k8sClient, err = client.Client(request)
		if err != nil {
			errors.HandleInternalError(response, err)
			return
		}
	}
	namespace := parseNamespacePathParameter(request)
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := deployment.GetDeploymentListWithMultiTenancy(k8sClient, tenant, namespace, dataSelect, apiHandler.iManager.Metric().Client())
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetDeploymentDetail(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("deployment")
	result, err := deployment.GetDeploymentDetail(k8sClient, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetDeploymentDetailWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("deployment")
	result, err := deployment.GetDeploymentDetailWithMultiTenancy(k8sClient, tenant, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetDeploymentEvents(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("deployment")
	dataSelect := parseDataSelectPathParameter(request)
	result, err := event.GetResourceEvents(k8sClient, dataSelect, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetDeploymentEventsWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("deployment")
	dataSelect := parseDataSelectPathParameter(request)
	result, err := event.GetResourceEventsWithMultiTenancy(k8sClient, dataSelect, tenant, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetDeploymentOldReplicaSets(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("deployment")
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := deployment.GetDeploymentOldReplicaSets(k8sClient, dataSelect, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetDeploymentOldReplicaSetsWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("deployment")
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := deployment.GetDeploymentOldReplicaSetsWithMultiTenancy(k8sClient, dataSelect, tenant, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetDeploymentNewReplicaSet(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("deployment")
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := deployment.GetDeploymentNewReplicaSet(k8sClient, dataSelect, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetDeploymentNewReplicaSetWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("deployment")
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := deployment.GetDeploymentNewReplicaSetWithMultiTenancy(k8sClient, dataSelect, tenant, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetPods(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := parseNamespacePathParameter(request)
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics // download standard metrics - cpu, and memory - by default
	result, err := pod.GetPodList(k8sClient, apiHandler.iManager.Metric().Client(), namespace, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetPodsWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	c, err := request.Request.Cookie("tenant")
	var CookieTenant string
	if err != nil {
		log.Printf("Cookie error: %v", err)
		CookieTenant = tenant
	} else {
		CookieTenant = c.Value
	}
	log.Printf("cookie_tenant is: %s", CookieTenant)
	var k8sClient kubernetes.Interface
	if tenant != CookieTenant {
		k8sClient = client.InsecureClient()
	} else {
		k8sClient, err = client.Client(request)
		if err != nil {
			errors.HandleInternalError(response, err)
			return
		}
	}

	namespace := parseNamespacePathParameter(request)
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics // download standard metrics - cpu, and memory - by default
	result, err := pod.GetPodListWithMultiTenancy(k8sClient, apiHandler.iManager.Metric().Client(), tenant, namespace, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetVMsWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := parseNamespacePathParameter(request)
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics // download standard metrics - cpu, and memory - by default
	result, err := vm.GetVMListWithMultiTenancy(k8sClient, apiHandler.iManager.Metric().Client(), tenant, namespace, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetPodDetail(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("pod")
	result, err := pod.GetPodDetail(k8sClient, apiHandler.iManager.Metric().Client(), namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetPodDetailWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("pod")
	result, err := pod.GetPodDetailWithMultiTenancy(k8sClient, apiHandler.iManager.Metric().Client(), tenant, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetVMDetailWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("virtualmachine")
	result, err := vm.GetVirtualMachineDetailWithMultiTenancy(k8sClient, apiHandler.iManager.Metric().Client(), tenant, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetReplicationControllerDetail(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("replicationController")
	result, err := replicationcontroller.GetReplicationControllerDetail(k8sClient, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetReplicationControllerDetailWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("replicationController")
	result, err := replicationcontroller.GetReplicationControllerDetailWithMultiTenancy(k8sClient, tenant, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleUpdateReplicasCount(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("replicationController")
	spec := new(replicationcontroller.ReplicationControllerSpec)
	if err := request.ReadEntity(spec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	if err := replicationcontroller.UpdateReplicasCount(k8sClient, namespace, name, spec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	response.WriteHeader(http.StatusAccepted)
}

func (apiHandler *APIHandlerV2) handleUpdateReplicasCountWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("replicationController")
	spec := new(replicationcontroller.ReplicationControllerSpec)
	if err := request.ReadEntity(spec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	if err := replicationcontroller.UpdateReplicasCountWithMultiTenancy(k8sClient, tenant, namespace, name, spec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	response.WriteHeader(http.StatusAccepted)
}

func (apiHandler *APIHandlerV2) handleGetResource(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)

	config, err := client.Config(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	verber, err := client.VerberClient(request, config)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	kind := request.PathParameter("kind")
	namespace, ok := request.PathParameters()["namespace"]
	name := request.PathParameter("name")
	result, err := verber.Get(kind, ok, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetResourceWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	config, err := client.Config(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	verber, err := client.VerberClient(request, config)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	kind := request.PathParameter("kind")
	namespace, ok := request.PathParameters()["namespace"]
	name := request.PathParameter("name")
	result, err := verber.GetWithMultiTenancy(kind, tenant, ok, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handlePutResource(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	config, err := client.Config(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	verber, err := client.VerberClient(request, config)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	kind := request.PathParameter("kind")
	namespace, ok := request.PathParameters()["namespace"]
	name := request.PathParameter("name")
	putSpec := &runtime.Unknown{}
	if err := request.ReadEntity(putSpec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	if err := verber.Put(kind, ok, namespace, name, putSpec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	response.WriteHeader(http.StatusCreated)
}

func (apiHandler *APIHandlerV2) handlePutResourceWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	config, err := client.Config(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	verber, err := client.VerberClient(request, config)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	kind := request.PathParameter("kind")
	namespace, ok := request.PathParameters()["namespace"]
	name := request.PathParameter("name")
	putSpec := &runtime.Unknown{}
	if err := request.ReadEntity(putSpec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	if err := verber.PutWithMultiTenancy(kind, tenant, ok, namespace, name, putSpec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	response.WriteHeader(http.StatusCreated)
}

func (apiHandler *APIHandlerV2) handleDeleteResource(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	config, err := client.Config(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	verber, err := client.VerberClient(request, config)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	kind := request.PathParameter("kind")
	namespace, ok := request.PathParameters()["namespace"]
	name := request.PathParameter("name")

	if err := verber.Delete(kind, ok, namespace, name); err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	// Try to unpin resource if it was pinned.
	pinnedResource := &settingsApi.PinnedResource{
		Name:      name,
		Kind:      kind,
		Namespace: namespace,
	}

	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	if err = apiHandler.sManager.DeletePinnedResource(k8sClient, pinnedResource); err != nil {
		if !errors.IsNotFoundError(err) {
			log.Printf("error while unpinning resource: %s", err.Error())
		}
	}

	response.WriteHeader(http.StatusOK)
}

func (apiHandler *APIHandlerV2) handleDeleteResourceWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	config, err := client.Config(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	verber, err := client.VerberClient(request, config)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	kind := request.PathParameter("kind")
	namespace, ok := request.PathParameters()["namespace"]
	name := request.PathParameter("name")

	if err := verber.DeleteWithMultiTenancy(kind, tenant, ok, namespace, name); err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	// Try to unpin resource if it was pinned.
	pinnedResource := &settingsApi.PinnedResource{
		Name:      name,
		Kind:      kind,
		Namespace: namespace,
	}

	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	if err = apiHandler.sManager.DeletePinnedResourceWithMultiTenancy(k8sClient, pinnedResource, tenant); err != nil {
		if !errors.IsNotFoundError(err) {
			log.Printf("error while unpinning resource: %s", err.Error())
		}
	}

	response.WriteHeader(http.StatusOK)
}

func (apiHandler *APIHandlerV2) handleGetReplicationControllerPods(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	rc := request.PathParameter("replicationController")
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := replicationcontroller.GetReplicationControllerPods(k8sClient, apiHandler.iManager.Metric().Client(), dataSelect, rc, namespace)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetReplicationControllerPodsWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	rc := request.PathParameter("replicationController")
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := replicationcontroller.GetReplicationControllerPodsWithMultiTenancy(k8sClient, apiHandler.iManager.Metric().Client(), dataSelect, tenant, rc, namespace)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleCreateCreateClusterRole(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	clusterRoleSpec := new(clusterrole.ClusterRoleSpec)
	if err := request.ReadEntity(clusterRoleSpec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	if err := clusterrole.CreateClusterRole(clusterRoleSpec, k8sClient); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			msg := "clusterroles '" + clusterRoleSpec.Name + "' already exists"
			err = er.New(msg)
		}
		errorMsg := Error{Msg: err.Error(), StatusCode: http.StatusConflict}
		response.WriteHeaderAndEntity(http.StatusConflict, errorMsg)
		return
	}
	response.WriteHeaderAndEntity(http.StatusCreated, clusterRoleSpec)
}

func (apiHandler *APIHandlerV2) handleCreateCreateClusterRolesWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	clusterRoleSpec := new(clusterrole.ClusterRoleSpec)
	if err := request.ReadEntity(clusterRoleSpec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	if err := clusterrole.CreateClusterRolesWithMultiTenancy(clusterRoleSpec, k8sClient); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusCreated, clusterRoleSpec)
}

func (apiHandler *APIHandlerV2) handleCreateRoleBindings(request *restful.Request, response *restful.Response) {

	roleBindingSpec := new(rolebinding.RoleBindingSpec)
	if err := request.ReadEntity(roleBindingSpec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	client := ResourceAllocator(roleBindingSpec.Tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	if err := rolebinding.CreateRoleBindings(roleBindingSpec, k8sClient); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusCreated, roleBindingSpec)
}

func (apiHandler *APIHandlerV2) handleDeleteRoleBindings(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespaceName := request.PathParameter("namespace")
	rolebindingName := request.PathParameter("rolebinding")
	if err := rolebinding.DeleteRoleBindings(namespaceName, rolebindingName, k8sClient); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeader(http.StatusOK)
}

func (apiHandler *APIHandlerV2) handleCreateRoleBindingsWithMultiTenancy(request *restful.Request, response *restful.Response) {

	roleBindingSpec := new(rolebinding.RoleBindingSpec)
	if err := request.ReadEntity(roleBindingSpec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	client := ResourceAllocator(roleBindingSpec.Tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	if err := rolebinding.CreateRoleBindingsWithMultiTenancy(roleBindingSpec, k8sClient); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusCreated, roleBindingSpec)
}

func (apiHandler *APIHandlerV2) handleDeleteRoleBindingsWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	tenantName := request.PathParameter("tenant")
	namespaceName := request.PathParameter("namespace")
	rolebindingName := request.PathParameter("rolebinding")
	if err := rolebinding.DeleteRoleBindingsWithMultiTenancy(tenantName, namespaceName, rolebindingName, k8sClient); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeader(http.StatusOK)
}

func (apiHandler *APIHandlerV2) handleCreateClusterRoleBindings(request *restful.Request, response *restful.Response) {

	clusterRoleBindingSpec := new(clusterrolebinding.ClusterRoleBindingSpec)
	if err := request.ReadEntity(clusterRoleBindingSpec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	client := ResourceAllocator(clusterRoleBindingSpec.Tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	if err := clusterrolebinding.CreateClusterRoleBindings(clusterRoleBindingSpec, k8sClient); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusCreated, clusterRoleBindingSpec)
}

func (apiHandler *APIHandlerV2) handleDeleteClusterRoleBindings(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	clusterrolebindingName := request.PathParameter("clusterrolebinding")
	if err := clusterrolebinding.DeleteClusterRoleBindings(clusterrolebindingName, k8sClient); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeader(http.StatusOK)
}

func (apiHandler *APIHandlerV2) handleDeleteClusterRole(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	clusterroleName := request.PathParameter("clusterrole")
	if err := clusterrole.DeleteClusterRole(clusterroleName, k8sClient); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeader(http.StatusOK)
}

func (apiHandler *APIHandlerV2) handleCreateClusterRoleBindingsWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	clusterRoleBindingSpec := new(clusterrolebinding.ClusterRoleBindingSpec)
	if err := request.ReadEntity(clusterRoleBindingSpec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	if err := clusterrolebinding.CreateClusterRoleBindingsWithMultiTenancy(clusterRoleBindingSpec, k8sClient); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusCreated, clusterRoleBindingSpec)
}

func (apiHandler *APIHandlerV2) handleDeleteClusterRoleBindingsWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	tenantName := request.PathParameter("tenant")
	clusterrolebindingName := request.PathParameter("clusterrolebinding")
	if err := clusterrolebinding.DeleteClusterRoleBindingsWithMultiTenancy(tenantName, clusterrolebindingName, k8sClient); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeader(http.StatusOK)
}

func (apiHandler *APIHandlerV2) handleGetRoles(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	Namespace := request.PathParameter("namespace")
	var namespaces []string
	namespaces = append(namespaces, Namespace)
	namespace := common.NewNamespaceQuery(namespaces)
	dataSelect := parseDataSelectPathParameter(request)

	result, err := role.GetRoleList(k8sClient, namespace, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetRoleDetail(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	name := request.PathParameter("name")
	namespace := request.PathParameter("namespace")
	result, err := role.GetRoleDetail(k8sClient, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleCreateRole(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	roleSpec := new(role.RoleSpec)
	if err := request.ReadEntity(roleSpec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	if err := role.CreateRole(roleSpec, k8sClient); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusCreated, roleSpec)
}

func (apiHandler *APIHandlerV2) handleDeleteRole(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	roleName := request.PathParameter("role")
	if err := role.DeleteRole(namespace, roleName, k8sClient); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeader(http.StatusOK)
}

func (apiHandler *APIHandlerV2) handleGetRolesWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	result, err := role.GetRolesWithMultiTenancy(k8sClient, tenant, namespace)
	//role.GetRoleList(k8sClient, namespace, result)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetRoleDetailWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")
	result, err := role.GetRoleDetailWithMultiTenancy(k8sClient, tenant, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleCreateRolesWithMultiTenancy(request *restful.Request, response *restful.Response) {

	roleSpec := new(role.RoleSpec)
	if err := request.ReadEntity(roleSpec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	client := ResourceAllocator(roleSpec.Tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	log.Printf("potnt1")
	if err := role.CreateRolesWithMultiTenancy(roleSpec, k8sClient); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			msg := "roles '" + roleSpec.Name + "' already exists"
			err = er.New(msg)
		}
		errorMsg := Error{Msg: err.Error(), StatusCode: http.StatusConflict}
		response.WriteHeaderAndEntity(http.StatusConflict, errorMsg)
		return
	}
	response.WriteHeaderAndEntity(http.StatusCreated, roleSpec)
}

func (apiHandler *APIHandlerV2) handleDeleteRolesWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	roleName := request.PathParameter("role")
	if err := role.DeleteRolesWithMultiTenancy(tenant, namespace, roleName, k8sClient); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeader(http.StatusOK)
}

func (apiHandler *APIHandlerV2) handleAddResourceQuota(request *restful.Request, response *restful.Response) {
	log.Printf("Adding Quota")
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	resourceQuotaSpec := new(resourcequota.ResourceQuotaSpec)
	if err := request.ReadEntity(resourceQuotaSpec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	//tenant := request.PathParameter("tenant")
	//namespace := request.PathParameter("namespace")
	result, err := resourcequota.AddResourceQuotas(k8sClient, resourceQuotaSpec.NameSpace, resourceQuotaSpec.Tenant, resourceQuotaSpec)
	if err != nil {
		errorMsg := Error{Msg: err.Error(), StatusCode: http.StatusConflict}
		response.WriteHeaderAndEntity(http.StatusConflict, errorMsg)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetResourceQuotaList(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	c, err := request.Request.Cookie("tenant")
	var CookieTenant string
	if err != nil {
		log.Printf("Cookie error: %v", err)
		CookieTenant = tenant
	} else {
		CookieTenant = c.Value
	}
	log.Printf("cookie_tenant is: %s", CookieTenant)
	var k8sClient kubernetes.Interface
	if tenant != CookieTenant {
		k8sClient = client.InsecureClient()
	} else {
		k8sClient, err = client.Client(request)
		if err != nil {
			errors.HandleInternalError(response, err)
			return
		}
	}

	//Namespace := request.PathParameter("namespace")
	var namespaces []string
	//namespaces = append(namespaces, Namespace)
	namespace := common.NewNamespaceQuery(namespaces)
	dataSelect := parseDataSelectPathParameter(request)
	result, err := resourcequota.GetResourceQuotaList(k8sClient, namespace, tenant, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetResourceQuotaListWithMultiTenancy(request *restful.Request, response *restful.Response) {
	log.Printf("Get Quota List")
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	result, err := resourcequota.GetResourceQuotaListsWithMultiTenancy(k8sClient, namespace, tenant)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetResourceQuotaDetails(request *restful.Request, response *restful.Response) {
	log.Printf("Get Quota List calling details")
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")

	result, err := resourcequota.GetResourceQuotaDetails(k8sClient, namespace, tenant, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleDeleteResourceQuota(request *restful.Request, response *restful.Response) {
	log.Printf("Deleting Quota")
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	resourceQuotaSpec := new(resourcequota.ResourceQuotaSpec)
	if err := request.ReadEntity(resourceQuotaSpec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")
	err = resourcequota.DeleteResourceQuota(k8sClient, namespace, tenant, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeader(http.StatusOK)
}

func (apiHandler *APIHandlerV2) handleCreateNamespace(request *restful.Request, response *restful.Response) {
	namespaceSpec := new(ns.NamespaceSpec)
	if err := request.ReadEntity(namespaceSpec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	if len(apiHandler.tpManager) == 0 {
		apiHandler.tpManager = append(apiHandler.tpManager, apiHandler.defaultClientmanager)
	}
	client := ResourceAllocator(namespaceSpec.Tenant, apiHandler.tpManager)
	k8sClient := client.InsecureClient()
	//k8sClient, err := client.Client(request)
	//if err != nil {
	//	errors.HandleInternalError(response, err)
	//	return
	//}

	if err := ns.CreateNamespace(namespaceSpec, namespaceSpec.Tenant, k8sClient); err != nil {
		errorMsg := Error{Msg: err.Error(), StatusCode: http.StatusConflict}
		response.WriteHeaderAndEntity(http.StatusConflict, errorMsg)
		return
	}
	response.WriteHeaderAndEntity(http.StatusCreated, namespaceSpec)
}

func (apiHandler *APIHandlerV2) handleGetServiceAccountList(request *restful.Request, response *restful.Response) {

	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	Namespace := request.PathParameter("namespace")
	var namespaces []string
	namespaces = append(namespaces, Namespace)
	namespace := common.NewNamespaceQuery(namespaces)
	dataSelect := parseDataSelectPathParameter(request)

	result, err := serviceaccount.GetServiceAccountList(k8sClient, namespace, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetServiceAccountDetail(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	name := request.PathParameter("serviceaccount")
	namespace := request.PathParameter("namespace")
	result, err := serviceaccount.GetServiceAccountDetail(k8sClient, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleCreateServiceAccount(request *restful.Request, response *restful.Response) {
	//tenant := request.PathParameter("tenant")

	serviceaccountSpec := new(serviceaccount.ServiceAccountSpec)
	if err := request.ReadEntity(serviceaccountSpec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	client := ResourceAllocator(serviceaccountSpec.Tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	if err := serviceaccount.CreateServiceAccount(serviceaccountSpec, k8sClient); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusCreated, serviceaccountSpec)
}

func (apiHandler *APIHandlerV2) handleDeleteServiceAccount(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	serviceaccountName := request.PathParameter("serviceaccount")
	namespaceName := request.PathParameter("namespace")
	if err := serviceaccount.DeleteServiceAccount(namespaceName, serviceaccountName, k8sClient); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeader(http.StatusOK)
}

func (apiHandler *APIHandlerV2) handleGetServiceAccountListWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	Namespace := request.PathParameter("namespace")
	var namespaces []string
	namespaces = append(namespaces, Namespace)
	namespace := common.NewNamespaceQuery(namespaces)
	dataSelect := parseDataSelectPathParameter(request)

	result, err := serviceaccount.GetServiceAccountListWithMultiTenancy(k8sClient, tenant, namespace, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetServiceAccountDetailWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	name := request.PathParameter("name")
	namespace := request.PathParameter("namespace")
	result, err := serviceaccount.GetServiceAccountDetailWithMultiTenancy(k8sClient, tenant, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleCreateServiceAccountsWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	serviceaccountSpec := new(serviceaccount.ServiceAccountSpec)
	if err := request.ReadEntity(serviceaccountSpec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	if err := serviceaccount.CreateServiceAccountsWithMultiTenancy(serviceaccountSpec, k8sClient); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusCreated, serviceaccountSpec)
}

func (apiHandler *APIHandlerV2) handleDeleteServiceAccountsWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	serviceaccountName := request.PathParameter("serviceaccount")
	tenantName := request.PathParameter("tenant")
	namespaceName := request.PathParameter("namespace")
	if err := serviceaccount.DeleteServiceAccountsWithMultiTenancy(tenantName, namespaceName, serviceaccountName, k8sClient); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeader(http.StatusOK)
}

func (apiHandler *APIHandlerV2) handleGetNamespaces(request *restful.Request, response *restful.Response) {
	var namespacesList ns.NamespaceList
	if len(apiHandler.tpManager) == 0 {
		apiHandler.tpManager = append(apiHandler.tpManager, apiHandler.defaultClientmanager)
	}
	for _, tpManager := range apiHandler.tpManager {
		k8sClient := tpManager.InsecureClient()

		dataSelect := parseDataSelectPathParameter(request)
		result, err := ns.GetNamespaceList(k8sClient, dataSelect)
		if err != nil {
			errors.HandleInternalError(response, err)
			return
		}
		for _, x := range result.Namespaces {
			namespacesList.Namespaces = append(namespacesList.Namespaces, x)
			namespacesList.ListMeta.TotalItems++
		}

	}
	response.WriteHeaderAndEntity(http.StatusOK, namespacesList)
}

func (apiHandler *APIHandlerV2) handleGetNamespacesWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	c, err := request.Request.Cookie("tenant")
	var CookieTenant string
	if err != nil {
		log.Printf("Cookie error: %v", err)
		CookieTenant = tenant
	} else {
		CookieTenant = c.Value
	}
	log.Printf("cookie_tenant is: %s", CookieTenant)
	var k8sClient kubernetes.Interface
	if tenant != CookieTenant {
		k8sClient = client.InsecureClient()
	} else {
		k8sClient, err = client.Client(request)
		if err != nil {
			errors.HandleInternalError(response, err)
			return
		}
	}
	dataSelect := parseDataSelectPathParameter(request)
	result, err := ns.GetNamespaceListWithMultiTenancy(k8sClient, tenant, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetNamespaceDetail(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	name := request.PathParameter("name")
	result, err := ns.GetNamespaceDetail(k8sClient, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetNamespaceDetailWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tnt := request.PathParameter("tenant")
	var result *ns.NamespaceDetail
	if len(apiHandler.tpManager) == 0 {
		apiHandler.tpManager = append(apiHandler.tpManager, apiHandler.defaultClientmanager)
	}
	for _, tpManager := range apiHandler.tpManager {
		k8sClient := tpManager.InsecureClient()

		//k8sClient, err := tpManager.Client(request)
		//if err != nil {
		//	errors.HandleInternalError(response, err)
		//	return
		//}

		dataSelect := parseDataSelectPathParameter(request)
		tenantList, err := tenant.GetTenantList(k8sClient, dataSelect, tpManager.GetClusterName(), "system")
		if err != nil {
			errors.HandleInternalError(response, err)
			return
		}
		for _, tnts := range tenantList.Tenants {
			if tnt == tnts.ObjectMeta.Name {
				tnts.ClusterName = tpManager.GetClusterName()
				name := request.PathParameter("name")
				result, err = ns.GetNamespaceDetailWithMultiTenancy(k8sClient, tnt, name)
				if err != nil {
					errors.HandleInternalError(response, err)
					return
				}
			}
		}
	}

	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetNamespaceEvents(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	name := request.PathParameter("name")
	dataSelect := parseDataSelectPathParameter(request)
	result, err := event.GetNamespaceEvents(k8sClient, dataSelect, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetNamespaceEventsWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	name := request.PathParameter("name")
	dataSelect := parseDataSelectPathParameter(request)
	result, err := event.GetNamespaceEventsWithMultiTenancy(k8sClient, dataSelect, tenant, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleCreateImagePullSecret(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	spec := new(secret.ImagePullSecretSpec)
	if err := request.ReadEntity(spec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	result, err := secret.CreateSecret(k8sClient, spec)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusCreated, result)
}

func (apiHandler *APIHandlerV2) handleCreateImagePullSecretWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	spec := new(secret.ImagePullSecretSpec)
	if err := request.ReadEntity(spec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	result, err := secret.CreateSecretWithMultiTenancy(k8sClient, tenant, spec)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusCreated, result)
}

func (apiHandler *APIHandlerV2) handleGetSecretDetail(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")
	result, err := secret.GetSecretDetail(k8sClient, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetSecretDetailWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")
	result, err := secret.GetSecretDetailWithMultiTenancy(k8sClient, tenant, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetSecretList(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	dataSelect := parseDataSelectPathParameter(request)
	namespace := parseNamespacePathParameter(request)
	result, err := secret.GetSecretList(k8sClient, namespace, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetSecretListWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	dataSelect := parseDataSelectPathParameter(request)
	namespace := parseNamespacePathParameter(request)
	result, err := secret.GetSecretListWithMultiTenancy(k8sClient, tenant, namespace, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetConfigMapList(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := parseNamespacePathParameter(request)
	dataSelect := parseDataSelectPathParameter(request)
	result, err := configmap.GetConfigMapList(k8sClient, namespace, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetConfigMapDetail(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("configmap")
	result, err := configmap.GetConfigMapDetail(k8sClient, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetPersistentVolumeList(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	dataSelect := parseDataSelectPathParameter(request)
	result, err := persistentvolume.GetPersistentVolumeList(k8sClient, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetPersistentVolumeListWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	dataSelect := parseDataSelectPathParameter(request)
	result, err := persistentvolume.GetPersistentVolumeListWithMultiTenancy(k8sClient, dataSelect, tenant)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetPersistentVolumeDetail(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	name := request.PathParameter("persistentvolume")
	result, err := persistentvolume.GetPersistentVolumeDetail(k8sClient, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetPersistentVolumeDetailWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	name := request.PathParameter("persistentvolume")
	result, err := persistentvolume.GetPersistentVolumeDetailWithMultiTenancy(k8sClient, tenant, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetPersistentVolumeClaimList(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := parseNamespacePathParameter(request)
	dataSelect := parseDataSelectPathParameter(request)
	result, err := persistentvolumeclaim.GetPersistentVolumeClaimList(k8sClient, namespace, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetPersistentVolumeClaimDetail(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")
	result, err := persistentvolumeclaim.GetPersistentVolumeClaimDetail(k8sClient, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetPodContainers(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("pod")
	result, err := container.GetPodContainers(k8sClient, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetPodContainersWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("pod")
	result, err := container.GetPodContainersWithMultiTenancy(k8sClient, tenant, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetReplicationControllerEvents(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("replicationController")
	dataSelect := parseDataSelectPathParameter(request)
	result, err := event.GetResourceEvents(k8sClient, dataSelect, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetReplicationControllerEventsWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("replicationController")
	dataSelect := parseDataSelectPathParameter(request)
	result, err := event.GetResourceEventsWithMultiTenancy(k8sClient, dataSelect, tenant, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetReplicationControllerServices(request *restful.Request,
	response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("replicationController")
	dataSelect := parseDataSelectPathParameter(request)
	result, err := replicationcontroller.GetReplicationControllerServices(k8sClient, dataSelect, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetReplicationControllerServicesWithMultiTenancy(request *restful.Request,
	response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("replicationController")
	dataSelect := parseDataSelectPathParameter(request)
	result, err := replicationcontroller.GetReplicationControllerServicesWithMultiTenancy(k8sClient, dataSelect, tenant, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetDaemonSetList(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := parseNamespacePathParameter(request)
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := daemonset.GetDaemonSetList(k8sClient, namespace, dataSelect, apiHandler.iManager.Metric().Client())
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetDaemonSetListWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := parseNamespacePathParameter(request)
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := daemonset.GetDaemonSetListWithMultiTenancy(k8sClient, tenant, namespace, dataSelect, apiHandler.iManager.Metric().Client())
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetDaemonSetDetail(
	request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("daemonSet")
	result, err := daemonset.GetDaemonSetDetail(k8sClient, apiHandler.iManager.Metric().Client(), namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetDaemonSetDetailWithMultiTenancy(
	request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("daemonSet")
	result, err := daemonset.GetDaemonSetDetailWithMultiTenancy(k8sClient, apiHandler.iManager.Metric().Client(), tenant, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetDaemonSetPods(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("daemonSet")
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := daemonset.GetDaemonSetPods(k8sClient, apiHandler.iManager.Metric().Client(), dataSelect, name, namespace)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetDaemonSetPodsWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("daemonSet")
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := daemonset.GetDaemonSetPodsWithMultiTenancy(k8sClient, apiHandler.iManager.Metric().Client(), dataSelect, tenant, name, namespace)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetDaemonSetServices(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	daemonSet := request.PathParameter("daemonSet")
	dataSelect := parseDataSelectPathParameter(request)
	result, err := daemonset.GetDaemonSetServices(k8sClient, dataSelect, namespace, daemonSet)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetDaemonSetServicesWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	daemonSet := request.PathParameter("daemonSet")
	dataSelect := parseDataSelectPathParameter(request)
	result, err := daemonset.GetDaemonSetServicesWithMultiTenancy(k8sClient, dataSelect, tenant, namespace, daemonSet)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetDaemonSetEvents(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("daemonSet")
	dataSelect := parseDataSelectPathParameter(request)
	result, err := event.GetResourceEvents(k8sClient, dataSelect, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetDaemonSetEventsWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("daemonSet")
	dataSelect := parseDataSelectPathParameter(request)
	result, err := event.GetResourceEventsWithMultiTenancy(k8sClient, dataSelect, tenant, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetHorizontalPodAutoscalerList(request *restful.Request,
	response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := parseNamespacePathParameter(request)
	dataSelect := parseDataSelectPathParameter(request)
	result, err := horizontalpodautoscaler.GetHorizontalPodAutoscalerList(k8sClient, namespace, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetHorizontalPodAutoscalerDetail(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("horizontalpodautoscaler")
	result, err := horizontalpodautoscaler.GetHorizontalPodAutoscalerDetail(k8sClient, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetJobList(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := parseNamespacePathParameter(request)
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := job.GetJobList(k8sClient, namespace, dataSelect, apiHandler.iManager.Metric().Client())
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetJobListWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := parseNamespacePathParameter(request)
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := job.GetJobListWithMultiTenancy(k8sClient, tenant, namespace, dataSelect, apiHandler.iManager.Metric().Client())
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetJobDetail(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := job.GetJobDetail(k8sClient, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetJobDetailWithMultitenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := job.GetJobDetailWithMultiTenancy(k8sClient, tenant, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetJobPods(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := job.GetJobPods(k8sClient, apiHandler.iManager.Metric().Client(), dataSelect, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetJobPodsWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := job.GetJobPodsWithMultiTenancy(k8sClient, apiHandler.iManager.Metric().Client(), dataSelect, tenant, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetJobEvents(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")
	dataSelect := parseDataSelectPathParameter(request)
	result, err := job.GetJobEvents(k8sClient, dataSelect, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetJobEventsWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")
	dataSelect := parseDataSelectPathParameter(request)
	result, err := job.GetJobEventsWithMultiTenancy(k8sClient, dataSelect, tenant, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetCronJobList(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := parseNamespacePathParameter(request)
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := cronjob.GetCronJobList(k8sClient, namespace, dataSelect, apiHandler.iManager.Metric().Client())
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetCronJobListWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := parseNamespacePathParameter(request)
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := cronjob.GetCronJobListWithMultiTenancy(k8sClient, tenant, namespace, dataSelect, apiHandler.iManager.Metric().Client())
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetCronJobDetail(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := cronjob.GetCronJobDetail(k8sClient, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetCronJobDetailWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := cronjob.GetCronJobDetailWithMultiTenancy(k8sClient, tenant, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetCronJobJobs(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")
	active := true
	if request.QueryParameter("active") == "false" {
		active = false
	}

	dataSelect := parseDataSelectPathParameter(request)
	result, err := cronjob.GetCronJobJobs(k8sClient, apiHandler.iManager.Metric().Client(), dataSelect, namespace, name, active)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetCronJobJobsWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")
	active := true
	if request.QueryParameter("active") == "false" {
		active = false
	}

	dataSelect := parseDataSelectPathParameter(request)
	result, err := cronjob.GetCronJobJobsWithMultiTenancy(k8sClient, apiHandler.iManager.Metric().Client(), dataSelect, tenant, namespace, name, active)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetCronJobEvents(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")
	dataSelect := parseDataSelectPathParameter(request)
	result, err := cronjob.GetCronJobEvents(k8sClient, dataSelect, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetCronJobEventsWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")
	dataSelect := parseDataSelectPathParameter(request)
	result, err := cronjob.GetCronJobEventsWithMultiTenancy(k8sClient, dataSelect, tenant, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleTriggerCronJob(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")
	err = cronjob.TriggerCronJob(k8sClient, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeader(http.StatusOK)
}

func (apiHandler *APIHandlerV2) handleTriggerCronJobWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")
	err = cronjob.TriggerCronJobWithMultiTenancy(k8sClient, tenant, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeader(http.StatusOK)
}

func (apiHandler *APIHandlerV2) handleGetStorageClassList(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	dataSelect := parseDataSelectPathParameter(request)
	result, err := storageclass.GetStorageClassList(k8sClient, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetStorageClassListWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	dataSelect := parseDataSelectPathParameter(request)
	result, err := storageclass.GetStorageClassListWithMultiTenancy(k8sClient, dataSelect, tenant)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetStorageClass(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	name := request.PathParameter("storageclass")
	result, err := storageclass.GetStorageClass(k8sClient, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetStorageClassWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	name := request.PathParameter("storageclass")
	result, err := storageclass.GetStorageClassWithMultiTenancy(k8sClient, tenant, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetStorageClassPersistentVolumes(request *restful.Request,
	response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	name := request.PathParameter("storageclass")
	dataSelect := parseDataSelectPathParameter(request)
	result, err := persistentvolume.GetStorageClassPersistentVolumes(k8sClient,
		name, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetStorageClassPersistentVolumesWithMultiTenancy(request *restful.Request,
	response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	name := request.PathParameter("storageclass")
	dataSelect := parseDataSelectPathParameter(request)
	result, err := persistentvolume.GetStorageClassPersistentVolumesWithMultiTenancy(k8sClient, tenant,
		name, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetPodPersistentVolumeClaims(request *restful.Request,
	response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	name := request.PathParameter("pod")
	namespace := request.PathParameter("namespace")
	dataSelect := parseDataSelectPathParameter(request)
	result, err := persistentvolumeclaim.GetPodPersistentVolumeClaims(k8sClient,
		namespace, name, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetCustomResourceDefinitionList(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	apiextensionsclient, err := client.APIExtensionsClient(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	dataSelect := parseDataSelectPathParameter(request)
	result, err := customresourcedefinition.GetCustomResourceDefinitionList(apiextensionsclient, dataSelect, client.GetClusterName())
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetCustomResourceDefinitionListWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	newrequest := restful.NewRequest(&http.Request{})

	result := new(customresourcedefinition.CustomResourceDefinitionList)
	if tenant != "system" {
		client := ResourceAllocator(tenant, apiHandler.tpManager)
		apiextensionsclient, err := client.APIExtensionsClient(request)
		if err != nil {
			errors.HandleInternalError(response, err)
			return
		}

		//dataSelect := parseDataSelectPathParameter(request)
		dataSelect := dataselect.NewDataSelectQuery(dataselect.NoPagination, dataselect.NoSort, dataselect.NoFilter, dataselect.NoMetrics)
		result, err = customresourcedefinition.GetCustomResourceDefinitionListWithMultiTenancy(apiextensionsclient, dataSelect, tenant, client.GetClusterName())
		if err != nil {
			errors.HandleInternalError(response, err)
			return
		}
	} else {
		for _, client := range apiHandler.tpManager {
			apiextensionsclient, err := client.APIExtensionsClient(newrequest)
			if err != nil {
				errors.HandleInternalError(response, err)
				return
			}

			dataSelect := dataselect.NewDataSelectQuery(dataselect.NoPagination, dataselect.NoSort, dataselect.NoFilter, dataselect.NoMetrics)
			nresult, err := customresourcedefinition.GetCustomResourceDefinitionList(apiextensionsclient, dataSelect, client.GetClusterName())
			if err != nil {
				errors.HandleInternalError(response, err)
				return
			}
			result.Items = append(result.Items, nresult.Items...)
			result.ListMeta.TotalItems += nresult.ListMeta.TotalItems
			result.Errors = append(result.Errors, nresult.Errors...)
		}
	}
	dataselect2 := parseDataSelectPathParameter(request)
	crdCells, filteredTotal := dataselect.GenericDataSelectWithFilter(toCellsCRD(result.Items), dataselect2)
	result.Items = fromCellsCRD(crdCells)
	result.ListMeta = api.ListMeta{TotalItems: filteredTotal}

	response.WriteHeaderAndEntity(http.StatusOK, result)
}

type CrdCell customresourcedefinition.CustomResourceDefinition

func (self CrdCell) GetProperty(name dataselect.PropertyName) dataselect.ComparableValue {
	switch name {
	case dataselect.NameProperty:
		return dataselect.StdComparableString(self.ObjectMeta.Name)
	case dataselect.CreationTimestampProperty:
		return dataselect.StdComparableTime(self.ObjectMeta.CreationTimestamp.Time)
	default:
		return nil
	}
}
func toCellsCRD(std []customresourcedefinition.CustomResourceDefinition) []dataselect.DataCell {
	cells := make([]dataselect.DataCell, len(std))
	for i := range std {
		cells[i] = CrdCell(std[i])
	}
	return cells
}
func fromCellsCRD(cells []dataselect.DataCell) []customresourcedefinition.CustomResourceDefinition {
	std := make([]customresourcedefinition.CustomResourceDefinition, len(cells))
	for i := range std {
		std[i] = customresourcedefinition.CustomResourceDefinition(cells[i].(CrdCell))
	}
	return std
}

func (apiHandler *APIHandlerV2) handleGetCustomResourceDefinitionDetail(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	config, err := client.Config(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	apiextensionsclient, err := client.APIExtensionsClient(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	name := request.PathParameter("crd")
	result, err := customresourcedefinition.GetCustomResourceDefinitionDetail(apiextensionsclient, config, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetCustomResourceDefinitionDetailWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)

	config, err := client.Config(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	apiextensionsclient, err := client.APIExtensionsClient(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	name := request.PathParameter("crd")
	result, err := customresourcedefinition.GetCustomResourceDefinitionDetailWithMultiTenancy(apiextensionsclient, config, tenant, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetCustomResourceObjectList(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	config, err := client.Config(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	apiextensionsclient, err := client.APIExtensionsClient(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	crdName := request.PathParameter("crd")
	namespace := parseNamespacePathParameter(request)
	dataSelect := parseDataSelectPathParameter(request)
	result, err := customresourcedefinition.GetCustomResourceObjectList(apiextensionsclient, config, namespace, dataSelect, crdName)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetCustomResourceObjectListWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	config, err := client.Config(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	apiextensionsclient, err := client.APIExtensionsClient(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	crdName := request.PathParameter("crd")
	namespace := parseNamespacePathParameter(request)
	dataSelect := parseDataSelectPathParameter(request)
	result, err := customresourcedefinition.GetCustomResourceObjectListWithMultiTenancy(apiextensionsclient, config, tenant, namespace, dataSelect, crdName)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetCustomResourceObjectDetail(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	config, err := client.Config(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	apiextensionsclient, err := client.APIExtensionsClient(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	name := request.PathParameter("object")
	crdName := request.PathParameter("crd")
	namespace := parseNamespacePathParameter(request)
	result, err := customresourcedefinition.GetCustomResourceObjectDetail(apiextensionsclient, namespace, config, crdName, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetCustomResourceObjectDetailWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	config, err := client.Config(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	apiextensionsclient, err := client.APIExtensionsClient(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	name := request.PathParameter("object")
	crdName := request.PathParameter("crd")
	namespace := parseNamespacePathParameter(request)
	result, err := customresourcedefinition.GetCustomResourceObjectDetailWithMultiTenancy(apiextensionsclient, tenant, namespace, config, crdName, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetCustomResourceObjectEvents(request *restful.Request, response *restful.Response) {
	log.Println("Getting events related to a custom resource object in namespace")

	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	name := request.PathParameter("object")
	namespace := request.PathParameter("namespace")
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := customresourcedefinition.GetEventsForCustomResourceObject(k8sClient, dataSelect, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleGetCustomResourceObjectEventsWithMultiTenancy(request *restful.Request, response *restful.Response) {
	log.Println("Getting events related to a custom resource object in namespace")

	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	name := request.PathParameter("object")
	namespace := request.PathParameter("namespace")
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := customresourcedefinition.GetEventsForCustomResourceObjectWithMultiTenancy(k8sClient, dataSelect, tenant, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleLogSource(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	resourceName := request.PathParameter("resourceName")
	resourceType := request.PathParameter("resourceType")
	namespace := request.PathParameter("namespace")
	logSources, err := logs.GetLogSources(k8sClient, namespace, resourceName, resourceType)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, logSources)
}

func (apiHandler *APIHandlerV2) handleLogSourceWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	resourceName := request.PathParameter("resourceName")
	resourceType := request.PathParameter("resourceType")
	namespace := request.PathParameter("namespace")
	logSources, err := logs.GetLogSourcesWithMultiTenancy(k8sClient, tenant, namespace, resourceName, resourceType)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, logSources)
}

func (apiHandler *APIHandlerV2) handleLogs(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	podID := request.PathParameter("pod")
	containerID := request.PathParameter("container")

	refTimestamp := request.QueryParameter("referenceTimestamp")
	if refTimestamp == "" {
		refTimestamp = logs.NewestTimestamp
	}

	refLineNum, err := strconv.Atoi(request.QueryParameter("referenceLineNum"))
	if err != nil {
		refLineNum = 0
	}
	usePreviousLogs := request.QueryParameter("previous") == "true"
	offsetFrom, err1 := strconv.Atoi(request.QueryParameter("offsetFrom"))
	offsetTo, err2 := strconv.Atoi(request.QueryParameter("offsetTo"))
	logFilePosition := request.QueryParameter("logFilePosition")

	logSelector := logs.DefaultSelection
	if err1 == nil && err2 == nil {
		logSelector = &logs.Selection{
			ReferencePoint: logs.LogLineId{
				LogTimestamp: logs.LogTimestamp(refTimestamp),
				LineNum:      refLineNum,
			},
			OffsetFrom:      offsetFrom,
			OffsetTo:        offsetTo,
			LogFilePosition: logFilePosition,
		}
	}

	result, err := container.GetLogDetails(k8sClient, namespace, podID, containerID, logSelector, usePreviousLogs)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleLogsWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	podID := request.PathParameter("pod")
	containerID := request.PathParameter("container")

	refTimestamp := request.QueryParameter("referenceTimestamp")
	if refTimestamp == "" {
		refTimestamp = logs.NewestTimestamp
	}

	refLineNum, err := strconv.Atoi(request.QueryParameter("referenceLineNum"))
	if err != nil {
		refLineNum = 0
	}
	usePreviousLogs := request.QueryParameter("previous") == "true"
	offsetFrom, err1 := strconv.Atoi(request.QueryParameter("offsetFrom"))
	offsetTo, err2 := strconv.Atoi(request.QueryParameter("offsetTo"))
	logFilePosition := request.QueryParameter("logFilePosition")

	logSelector := logs.DefaultSelection
	if err1 == nil && err2 == nil {
		logSelector = &logs.Selection{
			ReferencePoint: logs.LogLineId{
				LogTimestamp: logs.LogTimestamp(refTimestamp),
				LineNum:      refLineNum,
			},
			OffsetFrom:      offsetFrom,
			OffsetTo:        offsetTo,
			LogFilePosition: logFilePosition,
		}
	}

	result, err := container.GetLogDetailsWithMultiTenancy(k8sClient, tenant, namespace, podID, containerID, logSelector, usePreviousLogs)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandlerV2) handleLogFile(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	podID := request.PathParameter("pod")
	containerID := request.PathParameter("container")
	usePreviousLogs := request.QueryParameter("previous") == "true"

	logStream, err := container.GetLogFile(k8sClient, namespace, podID, containerID, usePreviousLogs)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	handleDownload(response, logStream)
}

func (apiHandler *APIHandlerV2) handleLogFileWithMultiTenancy(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	k8sClient, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	podID := request.PathParameter("pod")
	containerID := request.PathParameter("container")
	usePreviousLogs := request.QueryParameter("previous") == "true"

	logStream, err := container.GetLogFileWithMultiTenancy(k8sClient, tenant, namespace, podID, containerID, usePreviousLogs)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	handleDownload(response, logStream)
}

// parseNamespacePathParameter parses namespace selector for list pages in path parameter.
// The namespace selector is a comma separated list of namespaces that are trimmed.
// No namespaces means "view all user namespaces", i.e., everything except kube-system.
func parseNamespacePathParameter(request *restful.Request) *common.NamespaceQuery {
	namespace := request.PathParameter("namespace")

	namespaces := strings.Split(namespace, ",")
	var nonEmptyNamespaces []string
	for _, n := range namespaces {
		n = strings.Trim(n, " ")
		if len(n) > 0 {
			nonEmptyNamespaces = append(nonEmptyNamespaces, n)
			//nonEmptyNamespaces = append(nonEmptyNamespaces, n)
		}
	}
	return common.NewNamespaceQuery(nonEmptyNamespaces)
}

func parsePaginationPathParameter(request *restful.Request) *dataselect.PaginationQuery {
	itemsPerPage, err := strconv.ParseInt(request.QueryParameter("itemsPerPage"), 10, 0)
	if err != nil {
		return dataselect.NoPagination
	}

	page, err := strconv.ParseInt(request.QueryParameter("page"), 10, 0)
	if err != nil {
		return dataselect.NoPagination
	}

	// Frontend pages start from 1 and backend starts from 0
	return dataselect.NewPaginationQuery(int(itemsPerPage), int(page-1))
}

func parseFilterPathParameter(request *restful.Request) *dataselect.FilterQuery {
	return dataselect.NewFilterQuery(strings.Split(request.QueryParameter("filterBy"), ","))
}

// Parses query parameters of the request and returns a SortQuery object
func parseSortPathParameter(request *restful.Request) *dataselect.SortQuery {
	return dataselect.NewSortQuery(strings.Split(request.QueryParameter("sortBy"), ","))
}

// Parses query parameters of the request and returns a MetricQuery object
func parseMetricPathParameter(request *restful.Request) *dataselect.MetricQuery {
	metricNamesParam := request.QueryParameter("metricNames")
	var metricNames []string
	if metricNamesParam != "" {
		metricNames = strings.Split(metricNamesParam, ",")
	} else {
		metricNames = nil
	}
	aggregationsParam := request.QueryParameter("aggregations")
	var rawAggregations []string
	if aggregationsParam != "" {
		rawAggregations = strings.Split(aggregationsParam, ",")
	} else {
		rawAggregations = nil
	}
	aggregationModes := metricapi.AggregationModes{}
	for _, e := range rawAggregations {
		aggregationModes = append(aggregationModes, metricapi.AggregationMode(e))
	}
	return dataselect.NewMetricQuery(metricNames, aggregationModes)

}

// Parses query parameters of the request and returns a DataSelectQuery object
func parseDataSelectPathParameter(request *restful.Request) *dataselect.DataSelectQuery {
	paginationQuery := parsePaginationPathParameter(request)
	sortQuery := parseSortPathParameter(request)
	filterQuery := parseFilterPathParameter(request)
	metricQuery := parseMetricPathParameter(request)
	return dataselect.NewDataSelectQuery(paginationQuery, sortQuery, filterQuery, metricQuery)
}

// IAM Service related functions
type response struct {
	ID      int64  `json:"id,omitempty"`
	Message string `json:"message,omitempty"`
}

func (apiHandler *APIHandlerV2) handleCreateUser(w *restful.Request, r *restful.Response) {
	//_, error := apiHandler.tpManager.Client(w)
	//if error != nil {
	//	ErrMsg := ErrorMsg{Msg: error.Error()}
	//	r.WriteHeaderAndEntity(http.StatusUnauthorized, ErrMsg)
	//	return
	//}

	var user model.User
	err := w.ReadEntity(&user)
	if err != nil {
		log.Fatalf("Unable to decode the request body.  %v", err)
	}
	if user.NameSpace == "" {
		user.NameSpace = "default"
	}
	if user.Type == "tenant-admin" {
		client := ResourceAllocator(user.Tenant, apiHandler.tpManager)

		user, err = iam.TenantAdmin(user, client)
		if err != nil {
			ErrMsg := ErrorMsg{Msg: err.Error()}
			r.WriteHeaderAndEntity(http.StatusConflict, ErrMsg)
		}
	}

	user.CreationTimestamp = time.Now().Truncate(time.Second)
	insertID := db.InsertUser(user)
	res := response{
		ID:      insertID,
		Message: "User created successfully",
	}

	r.WriteHeaderAndEntity(http.StatusCreated, res)
}

func (apiHandler *APIHandlerV2) handleGetUser(w *restful.Request, r *restful.Response) {
	username := w.PathParameter("username")
	decode, err := base64.StdEncoding.DecodeString(username)
	if err != nil {
		ErrMsg := ErrorMsg{Msg: err.Error()}
		r.WriteHeaderAndEntity(http.StatusUnauthorized, ErrMsg)
		return
	}

	substrings := strings.Split(string(decode), "+")
	user, err := db.GetUser(substrings[0])

	if err != nil {
		log.Printf("Unable to get user. %v", err)
		r.WriteHeaderAndEntity(http.StatusUnauthorized, err.Error())
		return
	}

	r.WriteHeaderAndEntity(http.StatusOK, user)
}

func (apiHandler *APIHandlerV2) handleGetUserDetail(w *restful.Request, r *restful.Response) {
	username := w.PathParameter("username")
	user, err := db.GetUser(username)

	if err != nil {
		log.Printf("Unable to get user. %v", err)
		r.WriteHeaderAndEntity(http.StatusUnauthorized, err.Error())
		return
	}
	user.ObjectMeta.Password = "***********"
	user.ObjectMeta.Token = "***********"

	r.WriteHeaderAndEntity(http.StatusOK, user)
}

func (apiHandler *APIHandlerV2) handleGetAllUser(request *restful.Request, response *restful.Response) {
	tenant := request.PathParameter("tenant")
	client := ResourceAllocator(tenant, apiHandler.tpManager)
	_, err := client.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	users, err := db.GetAllUsers(tenant)

	if err != nil {
		log.Fatalf("Unable to get all user. %v", err)
	}
	dataSelect := parseDataSelectPathParameter(request)
	userCells, filteredTotal := dataselect.GenericDataSelectWithFilter(toCellsUser(users.Users), dataSelect)
	users.Users = fromCellsUser(userCells)

	users.ListMeta = api.ListMeta{TotalItems: filteredTotal}
	response.WriteHeaderAndEntity(http.StatusOK, users)
}

type UserCell model.UserDetails

func (self UserCell) GetProperty(name dataselect.PropertyName) dataselect.ComparableValue {
	switch name {
	case dataselect.NameProperty:
		return dataselect.StdComparableString(self.ObjectMeta.Username)
	case dataselect.CreationTimestampProperty:
		return dataselect.StdComparableTime(self.ObjectMeta.CreationTimestamp)
	default:
		return nil
	}
}
func toCellsUser(std []model.UserDetails) []dataselect.DataCell {
	cells := make([]dataselect.DataCell, len(std))
	for i := range std {
		cells[i] = UserCell(std[i])
	}
	return cells
}
func fromCellsUser(cells []dataselect.DataCell) []model.UserDetails {
	std := make([]model.UserDetails, len(cells))
	for i := range std {
		std[i] = model.UserDetails(cells[i].(UserCell))
	}
	return std
}

func (apiHandler *APIHandlerV2) handleDeleteUser(w *restful.Request, r *restful.Response) {
	var k8sClient kubernetes.Interface
	var err error
	for _, cManager := range apiHandler.tpManager {
		k8sClient, err = cManager.Client(w)
		if err == nil {
			break
		}
	}
	if err != nil {
		errors.HandleInternalError(r, err)
		return
	}
	tenantName := w.PathParameter("tenant")

	userName := w.PathParameter("username")
	userid := w.PathParameter("userid")
	userDetail, _ := db.GetUser(userName)
	if userDetail.ObjectMeta.Username == "" {
		errors.HandleInternalError(r, errors.NewInternal("User do not exists"))
		return
	}
	client := ResourceAllocator(userDetail.ObjectMeta.Tenant, apiHandler.tpManager)
	k8sClient = client.InsecureClient()
	if tenantName == userDetail.ObjectMeta.Tenant && userDetail.ObjectMeta.Type == `tenant-admin` {
		errors.HandleInternalError(r, er.New("Cannot delete admin users"))
		return
	}

	if userDetail.ObjectMeta.Type == `tenant-admin` {
		var clusterRoleName = userDetail.ObjectMeta.Username + "-" + userDetail.ObjectMeta.Tenant + "-" + "role"
		var saName = userDetail.ObjectMeta.Tenant + "-" + userDetail.ObjectMeta.Tenant + "-sa"
		var clusterRoleBinding = userDetail.ObjectMeta.Username + "-" + userDetail.ObjectMeta.Tenant + "-" + "rb"

		if err := clusterrolebinding.DeleteClusterRoleBindings(clusterRoleBinding, k8sClient); err != nil {
		}
		if err := serviceaccount.DeleteServiceAccount(userDetail.ObjectMeta.NameSpace, saName, k8sClient); err != nil {
		}
		if err := clusterrole.DeleteClusterRole(clusterRoleName, k8sClient); err != nil {
		}
	} else if userDetail.ObjectMeta.Type == `cluster-admin` {
		if err := serviceaccount.DeleteServiceAccount(userDetail.ObjectMeta.NameSpace, userName, k8sClient); err != nil {
		}
		if err := clusterrolebinding.DeleteClusterRoleBindings(userName, k8sClient); err != nil {
		}
	} else {
		if userDetail.ObjectMeta.Type == `tenant-user` {
			if err := serviceaccount.DeleteServiceAccountsWithMultiTenancy(userDetail.ObjectMeta.Tenant, userDetail.ObjectMeta.NameSpace, userName, k8sClient); err != nil {
			}
			if err := rolebinding.DeleteRoleBindingsWithMultiTenancy(userDetail.ObjectMeta.Tenant, userDetail.ObjectMeta.NameSpace, userName, k8sClient); err != nil {
			}
		}
	}
	msg := "User deleted successfully"
	id, err := strconv.Atoi(userid)
	//if userDetail.ObjectMeta.Type != `tenant-admin` {

	deletedRows := db.DeleteUser(int64(id))

	if err != nil {
		log.Printf("Unable to get user. %v", err)
		errors.HandleInternalError(r, err)
		return
	}
	msg = fmt.Sprintf("User deleted successfully. Total rows/record affected %v", deletedRows)
	//}
	res := response{
		ID:      int64(id),
		Message: msg,
	}
	r.WriteHeaderAndEntity(http.StatusOK, res)
}
