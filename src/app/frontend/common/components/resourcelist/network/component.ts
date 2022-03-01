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

  getDisplayColumns2(): string[] {
    return ['statusicon', 'name', 'group', 'fullName', 'namespaced', 'age'];
  }
}
