import {HttpParams} from '@angular/common/http';
import {Component, Input} from '@angular/core';
import {ClusterRole, ClusterRoleList} from '@api/backendapi';
import {Observable} from 'rxjs/Observable';
import {ResourceListBase} from '../../../resources/list';
import {NotificationsService} from '../../../services/global/notifications';
import {EndpointManager, Resource} from '../../../services/resource/endpoint';
import {ResourceService} from '../../../services/resource/resource';
import {MenuComponent} from '../../list/column/menu/component';
import {ListGroupIdentifier, ListIdentifier} from '../groupids';
import {VerberService} from '../../../services/global/verber';

@Component({
  selector: 'kd-cluster-role-list',
  templateUrl: './template.html',
})
export class ClusterRoleListComponent extends ResourceListBase<ClusterRoleList, ClusterRole> {
  @Input() endpoint = EndpointManager.resource(Resource.clusterRole, false, true).list();
  typeMeta:any;
  objectMeta:any;
  constructor(
    private readonly verber_: VerberService,
    private readonly clusterRole_: ResourceService<ClusterRoleList>,
    notifications: NotificationsService,
  ) {
    super('clusterrole', notifications);
    this.id = ListIdentifier.clusterRole;
    this.groupId = ListGroupIdentifier.cluster;

    // Register action columns.
    this.registerActionColumn<MenuComponent>('menu', MenuComponent);
  }

  getResourceObservable(params?: HttpParams): Observable<ClusterRoleList> {
    return this.clusterRole_.get(this.endpoint, undefined, params);
  }

  map(clusterRoleList: ClusterRoleList): ClusterRole[] {
    return clusterRoleList.items;
  }

  getDisplayColumns(): string[] {
    return ['name', 'age'];
  }

  getDisplayColumns2(): string[] {
    return ['name', 'age'];
  }

  onClick(): void {
    this.verber_.showClusterroleCreateDialog('Cluster Role name',this.typeMeta,this.objectMeta);
  }
}
