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

import {Component, Input} from '@angular/core';
import {HttpParams} from '@angular/common/http';
import {Observable} from 'rxjs';
import {NetworkObject, NetworkObjectList} from '@api/backendapi';
import {ResourceListBase} from '../../../resources/list';
import {NamespacedResourceService} from '../../../services/resource/resource';
import {NotificationsService} from '../../../services/global/notifications';
import {ActivatedRoute} from '@angular/router';
import {ListGroupIdentifier, ListIdentifier} from '../groupids';
import {MenuComponent} from '../../list/column/menu/component';

@Component({
  selector: 'kd-network-object-list',
  templateUrl: './template.html',
})
export class NetworkObjectListComponent extends ResourceListBase<NetworkObjectList, NetworkObject> {
  @Input() endpoint: string;
  @Input() networkName: string;
  networkObjectKind: string;

  constructor(
    private readonly networkObject_: NamespacedResourceService<NetworkObjectList>,
    notifications: NotificationsService,
    private readonly activatedRoute_: ActivatedRoute,
  ) {
    super(`network/${activatedRoute_.snapshot.params.networkName}`, notifications);
    this.id = ListIdentifier.networkObject;
    this.groupId = ListGroupIdentifier.none;

    // Register action columns.
    this.registerActionColumn<MenuComponent>('menu', MenuComponent);
  }

  getResourceObservable(params?: HttpParams): Observable<NetworkObjectList> {
    return this.networkObject_.get(this.endpoint, undefined, undefined, params);
  }

  map(networkObjectList: NetworkObjectList): NetworkObject[] {
    this.networkObjectKind = networkObjectList.typeMeta.kind.split("List")[0]+"s"
    return networkObjectList.items;
  }

  getDisplayColumns(): string[] {
    return ['name', 'namespace', 'age'];
  }

}
