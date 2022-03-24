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

import {HttpParams} from '@angular/common/http';
import {Component, ComponentFactoryResolver, Input} from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';
import {Event, ReplicaSet, ReplicaSetList} from '@api/backendapi';
import {Observable} from 'rxjs/Observable';

import {ResourceListWithStatuses} from '../../../resources/list';
import {NotificationsService} from '../../../services/global/notifications';
import {EndpointManager, Resource} from '../../../services/resource/endpoint';
import {NamespacedResourceService} from '../../../services/resource/resource';
import {MenuComponent} from '../../list/column/menu/component';
import {ListGroupIdentifier, ListIdentifier} from '../groupids';
import {TenantService} from "../../../services/global/tenant";

@Component({
  selector: 'kd-replica-set-list',
  templateUrl: './template.html',
})
export class ReplicaSetListComponent extends ResourceListWithStatuses<ReplicaSetList, ReplicaSet> {
  @Input() title: string;
  @Input() endpoint = EndpointManager.resource(Resource.replicaSet, true, true).list();

  tenantName: string;

  constructor(
    private readonly replicaSet_: NamespacedResourceService<ReplicaSetList>,
    private readonly activatedRoute_: ActivatedRoute,
    private readonly tenant_: TenantService,
    notifications: NotificationsService,
    resolver: ComponentFactoryResolver,
  ) {
    super('replicaset', notifications, resolver);
    this.id = ListIdentifier.replicaSet;
    this.groupId = ListGroupIdentifier.workloads;

    // Register status icon handlers
    this.registerBinding(this.icon.checkCircle, 'kd-success', this.isInSuccessState);
    this.registerBinding(this.icon.timelapse, 'kd-muted', this.isInPendingState);
    this.registerBinding(this.icon.error, 'kd-error', this.isInErrorState);

    // Register action columns.
    this.registerActionColumn<MenuComponent>('menu', MenuComponent);

    // Register dynamic columns.
    this.registerDynamicColumn('namespace', 'name', this.shouldShowNamespaceColumn_.bind(this));
    this.tenantName = this.activatedRoute_.snapshot.params.resourceName === undefined ?
      this.tenant_.current() : this.tenant_.resourceTenant()
    sessionStorage.setItem('replicaSetTenant', this.tenantName);
  }

  getResourceObservable(params?: HttpParams): Observable<ReplicaSetList> {
    const partition = this.tenantName === 'system' ? 'partition/' + this.tenant_.tenantPartition() + '/' : ''
    let endpoint = ''
    if (sessionStorage.getItem('userType') === 'cluster-admin' && !this.endpoint.includes('/deployment/')) {
      endpoint = `api/v1/${partition}tenants/${this.tenantName}/replicaset`
    } else {
      endpoint = this.endpoint
    }
    return this.replicaSet_.get(endpoint, undefined, undefined, params, this.tenantName);
  }

  map(rsList: ReplicaSetList): ReplicaSet[] {
    return rsList.replicaSets;
  }

  isInErrorState(resource: ReplicaSet): boolean {
    return resource.podInfo.warnings.length > 0;
  }

  isInPendingState(resource: ReplicaSet): boolean {
    return resource.podInfo.warnings.length === 0 && resource.podInfo.pending > 0;
  }

  isInSuccessState(resource: ReplicaSet): boolean {
    return resource.podInfo.warnings.length === 0 && resource.podInfo.pending === 0;
  }

  protected getDisplayColumns(): string[] {
    return ['statusicon', 'name', 'labels', 'pods', 'age', 'images'];
  }

  private shouldShowNamespaceColumn_(): boolean {
    return this.namespaceService_.areMultipleNamespacesSelected();
  }

  hasErrors(replicaSet: ReplicaSet): boolean {
    return replicaSet.podInfo.warnings.length > 0;
  }

  getEvents(replicaSet: ReplicaSet): Event[] {
    return replicaSet.podInfo.warnings;
  }
}
