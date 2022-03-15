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

import {HttpParams} from '@angular/common/http';
import {Component, Input, OnInit} from '@angular/core';
import {
  ObjectMeta,
  ResourcePartition,
  ResourcePartitionList,
  TenantPartition,
  TenantPartitionList,
  TypeMeta
} from '@api/backendapi';
import {Observable} from 'rxjs/Observable';
import {ResourceListWithStatuses} from '../../../resources/list';
import {NotificationsService} from '../../../services/global/notifications';
import {EndpointManager, Resource} from '../../../services/resource/endpoint';
import {ResourceService} from '../../../services/resource/resource';
import {MenuComponent} from '../../list/column/menu/component';
import {ListGroupIdentifier, ListIdentifier} from '../groupids';
import {VerberService} from "../../../services/global/verber";

@Component({
  selector: 'kd-partition-list',
  templateUrl: './template.html',
})
export class PartitionListComponent implements  OnInit {

  displayName:string;
  typeMeta:TypeMeta;
  objectMeta:ObjectMeta;

  constructor(
    readonly verber_: VerberService,
  ) {}


  //added the code
  onClick(): void {
    this.verber_.showNodeCreateDialog(this.displayName, this.typeMeta, this.objectMeta); //added
  }

  ngOnInit(): void {
  }
}

export class ResourcePartitionListComponent extends ResourceListWithStatuses<ResourcePartitionList, ResourcePartition> {
  @Input() endpointRp = EndpointManager.resource(Resource.resourcePartition).list();

  displayName:string;
  typeMeta:TypeMeta;
  objectMeta:ObjectMeta;

  constructor(
    readonly verber_: VerberService,
    private readonly resourcePartition_: ResourceService<ResourcePartitionList>,
    notifications: NotificationsService,
  ) {
    super('resourcePartition', notifications);
    this.id = ListIdentifier.resourcePartition;
    this.groupId = ListGroupIdentifier.cluster;

    // Register action columns.
    this.registerActionColumn<MenuComponent>('menu', MenuComponent);

    // Register status icon handlers
    this.registerBinding(this.icon.checkCircle, 'kd-success', this.isInSuccessState);
  }

  getResourceObservable(params?: HttpParams): Observable<ResourcePartitionList> {
    return this.resourcePartition_.get(this.endpointRp, undefined, params);
  }

  map(resourcePartitionList: ResourcePartitionList): ResourcePartition[] {
    return resourcePartitionList.resourcePartitions
  }

  isInSuccessState(): boolean {
    return true;
  }

  getDisplayColumns(): string[] {
    return ['statusicon', 'name', 'nodecount','cpu','memory','health','etcd'];
  }

  onClick(): void {
    this.verber_.showNodeCreateDialog(this.displayName, this.typeMeta, this.objectMeta); //added
  }
}

export class TenantPartitionListComponent extends ResourceListWithStatuses<TenantPartitionList, TenantPartition> {
  @Input() endpointTp = EndpointManager.resource(Resource.tenantPartition).list();

  displayName:string;
  typeMeta:TypeMeta;
  objectMeta:ObjectMeta;

  constructor(
    readonly verber_: VerberService,
    private readonly tenantPartition_: ResourceService<TenantPartitionList>,
    notifications: NotificationsService,
  ) {
    super('tenantPartition', notifications);
    this.id = ListIdentifier.tenantPartition;
    this.groupId = ListGroupIdentifier.cluster;

    // Register action columns.
    this.registerActionColumn<MenuComponent>('menu', MenuComponent);

    // Register status icon handlers
    this.registerBinding(this.icon.checkCircle, 'kd-success', this.isInSuccessState);
  }

  getResourceObservable(params?: HttpParams): Observable<TenantPartitionList> {
    return this.tenantPartition_.get(this.endpointTp, undefined, params);
  }

  map(tenantPartitionList: TenantPartitionList): TenantPartition[] {
    return tenantPartitionList.tenantPartitions
  }

  isInSuccessState(): boolean {
    return true;
  }

  getDisplayColumns(): string[] {
    return ['statusicon', 'name', 'tenantcount','pods','cpu','memory','health','etcd'];
  }

  onClick(): void {
    this.verber_.showNodeCreateDialog(this.displayName, this.typeMeta, this.objectMeta); //added
  }
}
