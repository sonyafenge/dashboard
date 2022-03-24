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
import {ServiceDetail} from '@api/backendapi';
import {Subscription} from 'rxjs/Subscription';

import {ActionbarService, ResourceMeta} from '../../../../common/services/global/actionbar';
import {NotificationsService} from '../../../../common/services/global/notifications';
import {EndpointManager, Resource} from '../../../../common/services/resource/endpoint';
import {NamespacedResourceService} from '../../../../common/services/resource/resource';
import {TenantService} from "../../../../common/services/global/tenant";

@Component({
  selector: 'kd-service-detail',
  templateUrl: './template.html',
})
export class ServiceDetailComponent implements OnInit, OnDestroy {
  private serviceSubscription_: Subscription;
  private readonly endpoint_ = EndpointManager.resource(Resource.service, true, true, true);
  service: ServiceDetail;
  isInitialized = false;
  podListEndpoint: string;
  eventListEndpoint: string;
  private tenantName: string;

  constructor(
    private readonly service_: NamespacedResourceService<ServiceDetail>,
    private readonly actionbar_: ActionbarService,
    private readonly activatedRoute_: ActivatedRoute,
    private readonly tenant_: TenantService,
    private readonly notifications_: NotificationsService,
  ) {}

  ngOnInit(): void {
    const resourceName = this.activatedRoute_.snapshot.params.resourceName;
    const resourceNamespace = this.activatedRoute_.snapshot.params.resourceNamespace;

    this.tenantName = this.tenant_.current() === 'system' ?
      sessionStorage.getItem('currentTenant') : this.tenant_.current()

    const partition = this.tenantName === 'system' ? this.tenant_.tenantPartition() : ''

    this.podListEndpoint = this.endpoint_.child(resourceName, Resource.pod, resourceNamespace, this.tenantName, partition);
    this.eventListEndpoint = this.endpoint_.child(resourceName, Resource.event, resourceNamespace, this.tenantName, partition);

    this.serviceSubscription_ = this.service_
      .get(this.endpoint_.detail(), resourceName, resourceNamespace, undefined, this.tenantName, partition)
      .subscribe((d: ServiceDetail) => {
        this.service = d;
        this.notifications_.pushErrors(d.errors);
        this.actionbar_.onInit.emit(new ResourceMeta('Service', d.objectMeta, d.typeMeta));
        this.isInitialized = true;
      });
  }

  ngOnDestroy(): void {
    this.serviceSubscription_.unsubscribe();
    this.actionbar_.onDetailsLeave.emit();
  }
}
