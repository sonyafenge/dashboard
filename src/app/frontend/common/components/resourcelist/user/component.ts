import {HttpParams} from '@angular/common/http';
import {Component, Input, ViewChild} from '@angular/core';
import {User, UserList} from '@api/backendapi';
import {Observable} from 'rxjs/Observable';
import {MatDrawer} from '@angular/material';
import {ResourceListWithStatuses} from '../../../resources/list';
import {EndpointManager, Resource} from '../../../services/resource/endpoint';
import {ResourceService} from '../../../services/resource/resource';
import {NotificationsService} from '../../../services/global/notifications';
import {ListGroupIdentifier, ListIdentifier} from '../groupids';
import {MenuComponent} from '../../list/column/menu/component';
import {UserApi} from "../../../services/global/userapi"
import {VerberService} from "../../../services/global/verber"
import {NamespacedResourceService} from "../../../services/resource/resource";
import {TenantDetail} from "@api/backendapi";

@Component({
  selector: 'kd-user-list',
  templateUrl: './template.html',

})

export class UserListComponent extends ResourceListWithStatuses<UserList, User> {
  @Input() endpoint = EndpointManager.resource(Resource.user).list();
  @ViewChild(MatDrawer, {static: true}) private readonly nav_: MatDrawer;

  displayName:any;
  typeMeta:any;
  objectMeta:any;

  private currentTenant: string;

  constructor(
    public readonly verber_: VerberService,
    private readonly user_: ResourceService<UserList>,
    private userAPI_:UserApi,
    private readonly tenant_: NamespacedResourceService<TenantDetail>,
    notifications: NotificationsService,

  ) {
    super('user', notifications);
    this.id = ListIdentifier.user;
    this.groupId = ListGroupIdentifier.cluster;

    // Register status icon handlers
    this.registerBinding(this.icon.checkCircle, 'kd-success', this.isInSuccessState);
    this.registerBinding(this.icon.error, 'kd-error', this.isInErrorState);

    // Register action columns.
    this.registerActionColumn<MenuComponent>('menu', MenuComponent);
    this.currentTenant = this.tenant_['tenant_']['currentTenant_']

  }

  getResourceObservable(params?: HttpParams): Observable<UserList> {
    return this.user_.get(this.endpoint, undefined, params);
  }

  map(userList: UserList): User[] {
    const userType=sessionStorage.getItem('userType');
    const data=userList.users
    const userdata:any=[];
    data.map((user)=>{
      if(userType.includes('tenant'))
      {
        const parentTenant = sessionStorage.getItem('parentTenant');
        //@ts-ignore
        if(user.objectMeta.tenant === this.currentTenant || user.objectMeta.tenant === parentTenant || user.objectMeta.username === this.currentTenant) {
          return userdata.push(user)
        }
      }
      else {
        return userdata.push(user)
      }
    })
    this.totalItems=userdata.length
    return userdata
  }

  isInErrorState(resource: User): boolean {
    return resource.phase === 'Terminating';
  }

  isInSuccessState(resource: User): boolean {
    return resource.phase === 'Active';
  }

  getDisplayColumns(): string[] {
    return ['statusicon', 'username', 'phase', 'type','age','role'];
  }

  getDisplayColumns2(): string[] {
    return ['statusicon', 'username', 'phase', 'type','age','role'];
  }

  onClick(): void {
    this.verber_.showUserCreateDialog(this.displayName, this.typeMeta, this.objectMeta);
  }
}
