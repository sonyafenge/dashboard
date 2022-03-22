// Copyright 2020 Authors of Arktos.
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

import {HttpParams} from '@angular/common/http';
import {Component, Input} from '@angular/core';
import {ObjectMeta, Tenant, TenantList, TypeMeta} from '@api/backendapi';
import {Observable} from 'rxjs/Observable';

import {ResourceListWithStatuses} from '../../../resources/list';
import {EndpointManager, Resource} from '../../../services/resource/endpoint';
import {ResourceService} from '../../../services/resource/resource';
import {NotificationsService} from '../../../services/global/notifications';
import {ListGroupIdentifier, ListIdentifier} from '../groupids';
import {MenuComponent} from '../../list/column/menu/component';
import {VerberService} from '../../../services/global/verber';
import {ActivatedRoute, Router} from "@angular/router";
import {CookieService} from "ngx-cookie-service";

@Component({
  selector: 'kd-tptenant-list',
  templateUrl: './template.html',
})
//@ts-ignore
export class TpTenantListComponent extends ResourceListWithStatuses<TenantList, Tenant> {
  @Input() endpoint = EndpointManager.resource(Resource.tenant,false,true).list();

  displayName: string;
  typeMeta: TypeMeta;
  objectMeta: ObjectMeta;
  nodeName: string;
  clusterName: string;
  tenantList: Tenant[];
  tenantCount: number;

  constructor(
    readonly verber_: VerberService,
    private readonly tenant_: ResourceService<TenantList>,
    private readonly route_: ActivatedRoute,
    private readonly router_: Router,
    private readonly cookieService_: CookieService,
    notifications: NotificationsService,
  ) {
    super('tptenant', notifications);
    this.id = ListIdentifier.tenant;
    this.groupId = ListGroupIdentifier.cluster;

    this.nodeName = this.route_.snapshot.params.resourceName

    const routeInfo = this.router_.getCurrentNavigation();
    if ( routeInfo === null || routeInfo.extras.state === undefined ) {
      this.clusterName = sessionStorage.getItem(`${this.clusterName}`)
    } else {
      this.clusterName = (routeInfo.extras.state['clusterName']).toString();
      sessionStorage.setItem(`${this.clusterName}`, this.clusterName)
    }

    // Register status icon handlers
    this.registerBinding(this.icon.checkCircle, 'kd-success', this.isInSuccessState);
    this.registerBinding(this.icon.error, 'kd-error', this.isInErrorState);

    // Register action columns.
    this.registerActionColumn<MenuComponent>('menu', MenuComponent);
  }

  getResourceObservable(): Observable<TenantList> {
    return this.tenant_.get(this.endpoint, undefined);
  }

  map(tenantList: TenantList): Tenant[] {
    this.tenantList = []
    this.tenantCount = 0
    if (tenantList.tenants !== null) {
      const tenantsList: any = [];
      tenantList.tenants.map((tenant)=>{
        // @ts-ignore
        if(tenant['clusterName'].includes(this.clusterName))
        {
          tenantsList.push(tenant);
        }
      })
      this.tenantList = tenantsList
      this.totalItems = this.tenantList.length
    }
    return this.tenantList;
  }

  isInErrorState(resource: Tenant): boolean {
    return resource.phase === 'Terminating';
  }

  isInSuccessState(resource: Tenant): boolean {
    return resource.phase === 'Active';
  }

  getDisplayColumns(): string[] {
    return ['statusicon', 'name', 'phase', 'age'];
  }

  onClick(): void {
    this.verber_.showTenantCreateDialog(this.displayName, this.typeMeta, this.objectMeta);
  }

  setPartition(partitionName:string, $event:any) {
    const resourceName = $event.target.innerHTML.replace(/^\s+|\s+$/gm,'');
    sessionStorage.setItem(`${resourceName}`,partitionName);
  }
}
