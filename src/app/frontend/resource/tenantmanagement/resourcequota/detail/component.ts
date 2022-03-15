// Copyright 2017 The Kubernetes Authors.
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
import {QuotaAllocationStatus, ResourceQuotaDetail} from '@api/backendapi';
import {Subject} from 'rxjs';
import {first} from 'rxjs/operators';
import {ActionbarService, ResourceMeta} from '../../../../common/services/global/actionbar';
import {NotificationsService} from '../../../../common/services/global/notifications';
import {EndpointManager, Resource} from '../../../../common/services/resource/endpoint';
import {NamespacedResourceService} from '../../../../common/services/resource/resource';
import {MatTableDataSource} from "@angular/material";
import {TenantService} from "../../../../common/services/global/tenant";

@Component({
  selector: 'kd-resourcequota-detail',
  templateUrl: './template.html',
})
export class ResourceQuotaDetailComponent implements OnInit, OnDestroy {
  private readonly endpoint_ = EndpointManager.resource(Resource.resourcequota, true,true);
  private readonly unsubscribe_ = new Subject<void>();

  resourceQuota: ResourceQuotaDetail;
  isInitialized = false;
  statusList: any;
  allocationData: any;

  constructor(
    private readonly resourceQuota_: NamespacedResourceService<ResourceQuotaDetail>,
    private readonly actionbar_: ActionbarService,
    private readonly route_: ActivatedRoute,
    private readonly tenant_: TenantService,
    private readonly notifications_: NotificationsService,
  ) {}


  ngOnInit(): void {
    const resourceName = this.route_.snapshot.params.resourceName;
    const resourceNamespace = this.route_.snapshot.params.resourceNamespace === undefined ?
      window.history.state.namespace : this.route_.snapshot.params.resourceNamespace;
    const resourceTenant:any = this.tenant_.current() === 'system' ?
      sessionStorage.getItem('tenantName') : this.tenant_.current()

    this.allocationData = [];
    let endpoint = ''
    if (sessionStorage.getItem('userType') === 'cluster-admin') {
      endpoint = `api/v1/tenants/${resourceTenant}/resourcequota/${resourceNamespace}/${resourceName}`
    } else {
      endpoint = this.endpoint_.detail()
    }

    this.resourceQuota_
      .get(endpoint, resourceName, resourceNamespace, resourceTenant)
      .pipe(first())
      .subscribe((d: ResourceQuotaDetail) => {
        this.resourceQuota = d;
        this.statusList = d.statusList
        this.notifications_.pushErrors(d.errors);
        this.actionbar_.onInit.emit(new ResourceMeta('ResourceQuota', d.objectMeta, d.typeMeta));
        this.isInitialized = true;
        for(const key in this.statusList){
          if(this.statusList[key]){
            this.allocationData.push({name: key, used: this.statusList[key].used, hard: this.statusList[key].hard})
          }
        }
      });
  }

  ngOnDestroy(): void {
    this.unsubscribe_.next();
    this.unsubscribe_.complete();
    this.actionbar_.onDetailsLeave.emit();
  }

  getDataSource(): MatTableDataSource<QuotaAllocationStatus> {
    const tableData = new MatTableDataSource<QuotaAllocationStatus>();
    tableData.data = this.allocationData;
    return tableData;
  }

  getAllocationColumns(): string[] {
    return ['resources', 'used', 'hard'];
  }
}
