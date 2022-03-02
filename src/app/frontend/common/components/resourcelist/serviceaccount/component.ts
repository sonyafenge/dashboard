import {HttpParams} from '@angular/common/http';
import {ChangeDetectionStrategy, ChangeDetectorRef, Component, Input} from '@angular/core';
import {Observable} from 'rxjs';
import {ServiceAccount, ServiceAccountList} from 'typings/backendapi';
import {ResourceListBase} from '../../../resources/list';
import {NotificationsService} from '../../../services/global/notifications';
import {EndpointManager, Resource} from '../../../services/resource/endpoint';
import {NamespacedResourceService} from '../../../services/resource/resource';
import {MenuComponent} from '../../list/column/menu/component';
import {ListGroupIdentifier, ListIdentifier} from '../groupids';

@Component({
  selector: 'kd-service-account-list',
  templateUrl: './template.html',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class ServiceAccountListComponent extends ResourceListBase<ServiceAccountList, ServiceAccount> {
  protected getDisplayColumns2(): string[] {
      throw new Error('Method not implemented.');
  }
  @Input() endpoint = EndpointManager.resource(Resource.serviceaccount, false).list();

  constructor(
    private readonly serviceAccount_: NamespacedResourceService<ServiceAccountList>,
    notifications: NotificationsService,

  ) {
    super('serviceaccount', notifications,);
    this.id = ListIdentifier.serviceaccount;
    this.groupId = ListGroupIdentifier.config;

    // Register action columns.
    this.registerActionColumn<MenuComponent>('menu', MenuComponent);

    // Register dynamic columns.
    this.registerDynamicColumn('namespace', 'name', this.shouldShowNamespaceColumn_.bind(this));
  }

  getResourceObservable(params?: HttpParams): Observable<ServiceAccountList> {
    return this.serviceAccount_.get(this.endpoint, undefined, undefined, params);
  }

  map(serviceAccountList: ServiceAccountList): ServiceAccount[] {
    return serviceAccountList.items;
  }

  getDisplayColumns(): string[] {
    return ['name', 'labels', 'created'];
  }

  private shouldShowNamespaceColumn_(): boolean {
    return this.namespaceService_.areMultipleNamespacesSelected();
  }
}
