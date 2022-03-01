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
    return networkObjectList.items;
  }

  getDisplayColumns(): string[] {
    return ['name', 'namespace', 'age'];
  }

  getDisplayColumns2(): string[] {
    return ['name', 'namespace', 'age'];
  }
}
