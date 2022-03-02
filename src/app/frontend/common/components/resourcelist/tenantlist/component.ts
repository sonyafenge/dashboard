import {Component, Input} from '@angular/core';
import {ObjectMeta, Tenant, TenantList, TypeMeta} from '@api/backendapi';
import {Observable} from 'rxjs/Observable';
import {ResourceListWithStatuses} from '../../../resources/list';
import {EndpointManager, Resource} from '../../../services/resource/endpoint';
import {ResourceService} from '../../../services/resource/resource';
import {NotificationsService} from '../../../services/global/notifications';
import {ListGroupIdentifier, ListIdentifier} from '../groupids';
import {MenuComponent} from '../../list/column/menu/component';
import {VerberService} from '../../../services/global/verber';
import {ActivatedRoute} from "@angular/router";

@Component({
  selector: 'kd-tenant-list',
  templateUrl: './template.html',
})
export class TenantListComponent extends ResourceListWithStatuses<TenantList, Tenant> {
  @Input() endpoint = EndpointManager.resource(Resource.tenant).list();

  displayName: string;
  typeMeta: TypeMeta;
  objectMeta: ObjectMeta;
  nodeName: string
  clusterName: string
  tenantList: Tenant[]
  tenantCount: number

  constructor(
    readonly verber_: VerberService,
    private readonly tenant_: ResourceService<TenantList>,
    private readonly activatedRoute_: ActivatedRoute,
    notifications: NotificationsService,
  ) {
    super('tenant', notifications);
    this.id = ListIdentifier.tenant;
    this.groupId = ListGroupIdentifier.cluster;

    this.nodeName = this.activatedRoute_.snapshot.params.resourceName

    // Register status icon handlers
    this.registerBinding(this.icon.checkCircle, 'kd-success', this.isInSuccessState);
    this.registerBinding(this.icon.error, 'kd-error', this.isInErrorState);

    // Register action columns.
    this.registerActionColumn<MenuComponent>('menu', MenuComponent);
  }

  getResourceObservable(): Observable<TenantList> {
    return this.tenant_.get(this.endpoint, undefined);
  }

  map(tenantList: TenantList): Tenant[] {
    this.tenantList = []
    this.tenantCount = 0
    if (tenantList.tenants !== null) {
      const tenantsList: any = [];
      tenantList.tenants.map((tenant)=>{
        tenantsList.push(tenant);
      })
      this.tenantList = tenantsList
      this.totalItems = this.tenantList.length
    }
    return this.tenantList;
  }

  isInErrorState(resource: Tenant): boolean {
    return resource.phase === 'Terminating';
  }

  isInSuccessState(resource: Tenant): boolean {
    return resource.phase === 'Active';
  }

  getDisplayColumns(): string[] {
    return ['statusicon', 'clusterName', 'name', 'phase', 'age'];
  }

  getDisplayColumns2(): string[] {
    return ['statusicon', 'clusterName', 'name', 'phase', 'age'];
  }

  //added the code
  onClick(): void {
    this.verber_.showTenantCreateDialog(this.displayName, this.typeMeta, this.objectMeta);  //changes needed
  }
}
