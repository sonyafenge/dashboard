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
import {ReplicaSetDetail} from '@api/backendapi';
import {Subscription} from 'rxjs/Subscription';

import {ActionbarService, ResourceMeta} from '../../../../common/services/global/actionbar';
import {NotificationsService} from '../../../../common/services/global/notifications';
import {EndpointManager, Resource} from '../../../../common/services/resource/endpoint';
import {NamespacedResourceService} from '../../../../common/services/resource/resource';
import {TenantService} from "../../../../common/services/global/tenant";

@Component({
  selector: 'kd-replica-set-detail',
  templateUrl: './template.html',
})
export class ReplicaSetDetailComponent implements OnInit, OnDestroy {
  private replicaSetSubscription_: Subscription;
  private readonly endpoint_ = EndpointManager.resource(Resource.replicaSet, true, true);
  replicaSet: ReplicaSetDetail;
  isInitialized = false;
  eventListEndpoint: string;
  podListEndpoint: string;
  serviceListEndpoint: string;

  constructor(
    private readonly replicaSet_: NamespacedResourceService<ReplicaSetDetail>,
    private readonly actionbar_: ActionbarService,
    private readonly activatedRoute_: ActivatedRoute,
    private readonly tenant_: TenantService,
    private readonly notifications_: NotificationsService,
  ) {}

  ngOnInit(): void {
    const resourceName = this.activatedRoute_.snapshot.params.resourceName;
    const resourceNamespace = this.activatedRoute_.snapshot.params.resourceNamespace === undefined ?
      window.history.state.namespace : this.activatedRoute_.snapshot.params.resourceNamespace;
    const resourceTenant = this.tenant_.current() === 'system' ?
      sessionStorage.getItem('replicaSetTenant') : this.tenant_.current()

    let endpoint = ''
    if (sessionStorage.getItem('userType') === 'cluster-admin') {
      endpoint = `api/v1/tenants/${resourceTenant}/replicaset/${resourceNamespace}/${resourceName}`
    } else {
      endpoint = this.endpoint_.detail()
    }

    this.eventListEndpoint = this.endpoint_.child(resourceName, Resource.event, resourceNamespace, resourceTenant);
    this.podListEndpoint = this.endpoint_.child(resourceName, Resource.pod, resourceNamespace, resourceTenant);
    this.serviceListEndpoint = this.endpoint_.child(resourceName, Resource.service, resourceNamespace, resourceTenant);

    this.replicaSetSubscription_ = this.replicaSet_
      .get(endpoint, resourceName, resourceNamespace, undefined, resourceTenant)
      .subscribe((d: ReplicaSetDetail) => {
        this.replicaSet = d;
        this.notifications_.pushErrors(d.errors);
        this.actionbar_.onInit.emit(new ResourceMeta('Replica Set', d.objectMeta, d.typeMeta));
        this.isInitialized = true;
      });
  }

  ngOnDestroy(): void {
    this.replicaSetSubscription_.unsubscribe();
    this.actionbar_.onDetailsLeave.emit();
  }
}
