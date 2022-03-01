import {HttpParams} from '@angular/common/http';
import {Component, Input} from '@angular/core';
import {CRD, CRDList} from '@api/backendapi';
import {Observable} from 'rxjs';
import {ResourceListWithStatuses} from '../../../resources/list';
import {NotificationsService} from '../../../services/global/notifications';
import {EndpointManager, Resource} from '../../../services/resource/endpoint';
import {ResourceService} from '../../../services/resource/resource';
import {MenuComponent} from '../../list/column/menu/component';
import {ListGroupIdentifier, ListIdentifier} from '../groupids';

@Component({
  selector: 'kd-crd-list',
  templateUrl: './template.html',
})
export class CRDListComponent extends ResourceListWithStatuses<CRDList, CRD> {
  @Input() endpoint = EndpointManager.resource(Resource.crd, false, true).list();

  constructor(
    private readonly crd_: ResourceService<CRDList>,
    notifications: NotificationsService,
  ) {
    super(Resource.crdFull, notifications);
    this.id = ListIdentifier.crd;
    this.groupId = ListGroupIdentifier.none;

    // Register action columns.
    this.registerActionColumn<MenuComponent>('menu', MenuComponent);

    // Register status icon handlers
    this.registerBinding(this.icon.checkCircle, 'kd-success', this.isInSuccessState);
    this.registerBinding(this.icon.help, 'kd-muted', this.isInUnknownState);
    this.registerBinding(this.icon.error, 'kd-error', this.isInErrorState);
  }

  isNamespaced(crd: CRD): string {
    return crd.scope === 'Namespaced' ? 'True' : 'False';
  }

  getResourceObservable(params?: HttpParams): Observable<CRDList> {
    return this.crd_.get(this.endpoint, undefined, params);
  }

  map(crdList: CRDList): CRD[] {
    const crdLists:CRD[] = [];
    crdList.items.map((crd)=>{
      if (crd.names.kind !== 'Network' && crd.group !== 'mizar.com') {
        crdLists.push(crd)
      }
    })
    this.totalItems = crdLists.length
    return crdLists;
  }

  isInErrorState(resource: CRD): boolean {
    return resource.established === 'False';
  }

  isInUnknownState(resource: CRD): boolean {
    return resource.established === 'Unknown';
  }

  isInSuccessState(resource: CRD): boolean {
    return resource.established === 'True';
  }

  getDisplayColumns(): string[] {
    return ['statusicon', 'name', 'group', 'fullName', 'namespaced', 'age'];
  }

  getDisplayColumns2(): string[] {
    return ['statusicon', 'name', 'group', 'fullName', 'namespaced', 'age'];
  }
}
