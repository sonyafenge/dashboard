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

import {Component, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {DeploymentDetail, ReplicaSet} from '@api/backendapi';
import {Subject} from 'rxjs';
import {takeUntil} from 'rxjs/operators';

import {ActionbarService, ResourceMeta} from '../../../../common/services/global/actionbar';
import {NotificationsService} from '../../../../common/services/global/notifications';
import {KdStateService} from '../../../../common/services/global/state';
import {EndpointManager, Resource} from '../../../../common/services/resource/endpoint';
import {NamespacedResourceService} from '../../../../common/services/resource/resource';
import {TenantService} from "../../../../common/services/global/tenant";

@Component({
  selector: 'kd-deployment-detail',
  templateUrl: './template.html',
})
export class DeploymentDetailComponent implements OnInit, OnDestroy {
  private unsubscribe_ = new Subject<void>();
  private readonly endpoint_ = EndpointManager.resource(Resource.deployment, true, true);
  deployment: DeploymentDetail;
  newReplicaSet: ReplicaSet;
  isInitialized = false;
  eventListEndpoint: string;
  oldReplicaSetsEndpoint: string;
  newReplicaSetEndpoint: string;

  constructor(
    private readonly deployment_: NamespacedResourceService<DeploymentDetail>,
    private readonly replicaSet_: NamespacedResourceService<ReplicaSet>,
    private readonly activatedRoute_: ActivatedRoute,
    private readonly tenant_: TenantService,
    private readonly actionbar_: ActionbarService,
    private readonly kdState_: KdStateService,
    private readonly notifications_: NotificationsService,
  ) {}

  ngOnInit(): void {
    const resourceName = this.activatedRoute_.snapshot.params.resourceName;
    const resourceNamespace = this.activatedRoute_.snapshot.params.resourceNamespace === undefined ?
      window.history.state.namespace : this.activatedRoute_.snapshot.params.resourceNamespace;
    const resourceTenant = this.tenant_.current() === 'system' ?
      sessionStorage.getItem('deploymentTenant') : this.tenant_.current()
    const partition = resourceTenant === 'system' ? 'partition/' + this.tenant_.tenantPartition() + '/' : ''
    let endpoint = ''
    if (sessionStorage.getItem('userType') === 'cluster-admin') {
      endpoint = `api/v1/${partition}tenants/${resourceTenant}/deployment/${resourceNamespace}/${resourceName}`
    } else {
      endpoint = this.endpoint_.detail()
    }
    const resourcePartition = resourceTenant === 'system' ? this.tenant_.tenantPartition() : ''
    this.eventListEndpoint = this.endpoint_.child(resourceName, Resource.event, resourceNamespace, resourceTenant, resourcePartition);
    this.oldReplicaSetsEndpoint = this.endpoint_.child(
      resourceName,
      Resource.oldReplicaSet,
      resourceNamespace,
      resourceTenant,
      resourcePartition,
    );
    this.newReplicaSetEndpoint = this.endpoint_.child(
      resourceName,
      Resource.newReplicaSet,
      resourceNamespace,
      resourceTenant,
      resourcePartition,
    );

    this.deployment_
      .get(endpoint, resourceName, resourceNamespace, undefined, resourceTenant)
      .pipe(takeUntil(this.unsubscribe_))
      .subscribe((d: DeploymentDetail) => {
        this.deployment = d;
        this.notifications_.pushErrors(d.errors);
        this.actionbar_.onInit.emit(new ResourceMeta('Deployment', d.objectMeta, d.typeMeta));
        this.isInitialized = true;
      });

    this.replicaSet_
      .get(this.newReplicaSetEndpoint)
      .pipe(takeUntil(this.unsubscribe_))
      .subscribe((rs: ReplicaSet) => {
        this.newReplicaSet = rs;
      });
  }

  getNewReplicaSetHref(): string {
    return this.kdState_.href(
      this.newReplicaSet.typeMeta.kind,
      this.newReplicaSet.objectMeta.name,
      this.newReplicaSet.objectMeta.namespace,
    );
  }

  ngOnDestroy(): void {
    this.unsubscribe_.next();
    this.unsubscribe_.complete();
    this.actionbar_.onDetailsLeave.emit();
  }
}
