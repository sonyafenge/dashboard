// Copyright 2020 Authors of Arktos.

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
import {TenantDetail} from '@api/backendapi';

import {ActionbarService, ResourceMeta} from '../../../../common/services/global/actionbar';
import {NotificationsService} from '../../../../common/services/global/notifications';
import {ResourceService} from 'common/services/resource/resource';
import {EndpointManager, Resource} from 'common/services/resource/endpoint';
import {Subscription} from 'rxjs/Subscription';

@Component({
  selector: 'kd-tenant-detail',
  templateUrl: './template.html',
})
export class TenantDetailComponent implements OnInit, OnDestroy {
  private tenantSubscription_: Subscription;
  private readonly endpoint_ = EndpointManager.resource(Resource.tenant);
  tenant: TenantDetail;
  isInitialized = false;

  constructor(
    private readonly tenant_: ResourceService<TenantDetail>,
    private readonly actionbar_: ActionbarService,
    private readonly activatedRoute_: ActivatedRoute,
    private readonly notifications_: NotificationsService,
  ) {}

  ngOnInit(): void {
    const resourceName = this.activatedRoute_.snapshot.params.resourceName;

    this.tenantSubscription_ = this.tenant_
      .get(this.endpoint_.detail(), resourceName)
      .subscribe((d: TenantDetail) => {
        this.tenant = d;
        this.notifications_.pushErrors(d.errors);
        this.actionbar_.onInit.emit(new ResourceMeta('Tenant', d.objectMeta, d.typeMeta));
        this.isInitialized = true;
      });
  }

  ngOnDestroy(): void {
    this.tenantSubscription_.unsubscribe();
    this.actionbar_.onDetailsLeave.emit();
  }
}
