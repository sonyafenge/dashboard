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
import {ObjectMeta, Role, RoleList, TypeMeta} from '@api/backendapi';
import {Observable} from 'rxjs';
import {ResourceListBase} from '../../../resources/list';
import {NotificationsService} from '../../../services/global/notifications';
import {EndpointManager, Resource} from '../../../services/resource/endpoint';
import {NamespacedResourceService} from '../../../services/resource/resource';
import {MenuComponent} from '../../list/column/menu/component';
import {ListGroupIdentifier, ListIdentifier} from '../groupids';
import {VerberService} from '../../../services/global/verber';

@Component({
  selector: 'kd-role-list',
  templateUrl: './template.html',
})
export class RoleListComponent extends ResourceListBase<RoleList, Role> {
  @Input() endpoint = EndpointManager.resource(Resource.role, true,true).list();

  typeMeta:TypeMeta;
  objectMeta:ObjectMeta;

  constructor(
    private readonly role_: NamespacedResourceService<RoleList>,
    private readonly verber_: VerberService,
    notifications: NotificationsService,
  ) {
    super('role', notifications);
    this.id = ListIdentifier.role;
    this.groupId = ListGroupIdentifier.cluster;

    // Register action columns.
    this.registerActionColumn<MenuComponent>('menu', MenuComponent);

    // Register dynamic columns.
    this.registerDynamicColumn('namespace', 'name', this.shouldShowNamespaceColumn_.bind(this));
  }

  private shouldShowNamespaceColumn_(): boolean {
    return this.namespaceService_.areMultipleNamespacesSelected();
  }

  getResourceObservable(params?: HttpParams): Observable<RoleList> {
    return this.role_.get(this.endpoint, undefined, undefined, params);
  }

  map(roleList: RoleList): Role[] {
    return roleList.items;
  }

  getDisplayColumns(): string[] {
    return ['name', 'created'];
  }

  onClick(): void {
    this.verber_.showRoleCreateDialog('Role name',this.typeMeta,this.objectMeta);
  }
}
