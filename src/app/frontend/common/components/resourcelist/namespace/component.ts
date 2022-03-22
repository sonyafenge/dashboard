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
import {Component, Input} from '@angular/core';
import {Namespace, NamespaceList, ObjectMeta, TypeMeta} from '@api/backendapi';
import {Observable} from 'rxjs/Observable';
import {ResourceListWithStatuses} from '../../../resources/list';
import {NotificationsService} from '../../../services/global/notifications';
import {EndpointManager, Resource} from '../../../services/resource/endpoint';
import {ResourceService} from '../../../services/resource/resource';
import {MenuComponent} from '../../list/column/menu/component';
import {ListGroupIdentifier, ListIdentifier} from '../groupids';
import {VerberService} from '../../../services/global/verber';
import {ActivatedRoute} from "@angular/router";
import {TenantService} from "../../../services/global/tenant";

@Component({
  selector: 'kd-namespace-list',
  templateUrl: './template.html',
})
export class NamespaceListComponent extends ResourceListWithStatuses<NamespaceList, Namespace> {
  @Input() endpoint = EndpointManager.resource(Resource.namespace, false, true).list();

  displayName: string;
  typeMeta: TypeMeta;
  objectMeta: ObjectMeta;
  tenantName: string;

  constructor(
    private readonly verber_: VerberService,
    private readonly namespace_: ResourceService<NamespaceList>,
    private readonly activatedRoute_: ActivatedRoute,
    private readonly tenant_: TenantService,
    notifications: NotificationsService,
  ) {
    super('namespace', notifications);
    this.id = ListIdentifier.namespace;
    this.groupId = ListGroupIdentifier.cluster;

    // Register status icon handlers
    this.registerBinding(this.icon.checkCircle, 'kd-success', this.isInSuccessState);
    this.registerBinding(this.icon.error, 'kd-error', this.isInErrorState);

    // Register action columns.
    this.registerActionColumn<MenuComponent>('menu', MenuComponent);

    this.tenantName = this.activatedRoute_.snapshot.params.resourceName === undefined ?
      this.tenant_.current() : this.activatedRoute_.snapshot.params.resourceName
    sessionStorage.setItem('tenantName',this.tenantName);
  }

  getResourceObservable(params?: HttpParams): Observable<NamespaceList> {
    const partition = this.tenantName === 'system' ? 'partition/' + sessionStorage.getItem(`${this.tenantName}`) + '/' : ''
    let endpoint = ''
    if (sessionStorage.getItem('userType') === 'cluster-admin') {
      endpoint = `api/v1/${partition}tenants/${this.tenantName}/namespace`
    } else {
      endpoint = this.endpoint
    }
    return this.namespace_.get(endpoint, undefined, params, this.tenantName);
  }

  map(namespaceList: NamespaceList): Namespace[] {
    return namespaceList.namespaces;
  }

  isInErrorState(resource: Namespace): boolean {
    return resource.phase === 'Terminating';
  }

  isInSuccessState(resource: Namespace): boolean {
    return resource.phase === 'Active';
  }

  getDisplayColumns(): string[] {
    return ['statusicon', 'name', 'labels', 'phase', 'age'];
  }

  onClick(): void {
    this.verber_.showNamespaceCreateDialog(this.displayName, this.typeMeta, this.objectMeta); //added showNamespaceCreateDialog
  }
}
