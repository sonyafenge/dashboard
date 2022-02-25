
import {HttpParams} from '@angular/common/http';
import {Component, Input} from '@angular/core';
import {Role, RoleList} from '@api/backendapi';
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
  typeMeta:any;
  objectMeta:any;

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

  getDisplayColumns2(): string[] {
    return ['name', 'created'];
  }

  onClick(): void {
    this.verber_.showRoleCreateDialog('Role name',this.typeMeta,this.objectMeta);
  }
}
