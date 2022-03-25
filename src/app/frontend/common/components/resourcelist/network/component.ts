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
import {Network, NetworkList} from '@api/backendapi';
import {Observable} from 'rxjs';

import {ResourceListWithStatuses} from '../../../resources/list';
import {NotificationsService} from '../../../services/global/notifications';
import {EndpointManager, Resource} from '../../../services/resource/endpoint';
import {ResourceService} from '../../../services/resource/resource';
import {MenuComponent} from '../../list/column/menu/component';
import {ListGroupIdentifier, ListIdentifier} from '../groupids';

@Component({
  selector: 'kd-network-list',
  templateUrl: './template.html',
})
export class NetworkListComponent extends ResourceListWithStatuses<NetworkList, Network> {
  @Input() endpoint = EndpointManager.resource(Resource.network, false, true).list();

  constructor(
    private readonly network_: ResourceService<NetworkList>,
    notifications: NotificationsService,
  ) {
    super(Resource.networkFull, notifications);
    this.id = ListIdentifier.network;
    this.groupId = ListGroupIdentifier.none;

    // Register action columns.
    this.registerActionColumn<MenuComponent>('menu', MenuComponent);

    // Register status icon handlers
    this.registerBinding(this.icon.checkCircle, 'kd-success', this.isInSuccessState);
    this.registerBinding(this.icon.help, 'kd-muted', this.isInUnknownState);
    this.registerBinding(this.icon.error, 'kd-error', this.isInErrorState);
  }

  isNamespaced(network: Network): string {
    return network.scope === 'Namespaced' ? 'True' : 'False';
  }

  getResourceObservable(params?: HttpParams): Observable<NetworkList> {
    return this.network_.get(this.endpoint, undefined, params);
  }

  map(networkList: NetworkList): Network[] {
    const networkLists:Network[] = [];
    networkList.items.map((network)=>{
      if (network.names.kind === 'Network' || network.group === 'mizar.com') {
        networkLists.push(network)
      }
    })
    this.totalItems = networkLists.length
    return networkLists;
  }

  isInErrorState(resource: Network): boolean {
    return resource.established === 'False';
  }

  isInUnknownState(resource: Network): boolean {
    return resource.established === 'Unknown';
  }

  isInSuccessState(resource: Network): boolean {
    return resource.established === 'True';
  }

  getDisplayColumns(): string[] {
    return ['statusicon', 'name', 'group', 'fullName', 'namespaced', 'age'];
  }

}
