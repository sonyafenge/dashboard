import {HttpParams} from '@angular/common/http';
import {Component, Input} from '@angular/core';
import {TenantPartitionList, TenantPartition} from '@api/backendapi';
import {Observable} from 'rxjs/Observable';
import {ResourceListWithStatuses} from '../../../resources/list';
import {NotificationsService} from '../../../services/global/notifications';
import {EndpointManager, Resource} from '../../../services/resource/endpoint';
import {ResourceService} from '../../../services/resource/resource';
import {ListGroupIdentifier, ListIdentifier} from '../groupids';
import {MenuComponent} from '../../list/column/menu/component';
import {VerberService} from '../../../services/global/verber';
import {Router} from "@angular/router";

@Component({
  selector: 'kd-tenant-partition-list',
  templateUrl: './template.html',
})
// @ts-ignore
export class TenantPartitionListComponent extends ResourceListWithStatuses<TenantPartitionList, TenantPartition> {
  @Input() endpointTp = EndpointManager.resource(Resource.tenantPartition).list();
  displayName:any;
  typeMeta:any;
  objectMeta:any;

  constructor(
    readonly verber_: VerberService,
    private readonly router_: Router,
    private readonly tenantPartition_: ResourceService<TenantPartitionList>,
    notifications: NotificationsService,
  ) {
    super('tenantPartition', notifications);
    this.id = ListIdentifier.tenantPartition;
    this.groupId = ListGroupIdentifier.cluster;

    // Register action columns.
    this.registerActionColumn<MenuComponent>('menu', MenuComponent);

    // Register status icon handlers
    this.registerBinding(this.icon.checkCircle, 'kd-success', this.isInSuccessState);
  }

  setClusterName($event:any) {
    const clusterName = $event.target.innerHTML.replace(/^\s+|\s+$/gm,'');
    this.router_.navigateByUrl('/tptenant', {state: {clusterName}});
  }

  getResourceObservable(params?: HttpParams): Observable<TenantPartitionList> {
    return this.tenantPartition_.get(this.endpointTp, undefined, params);
  }

  map(tenantPartitionList: TenantPartitionList): TenantPartition[] {
    console.log("Value of it is: ",tenantPartitionList.tenantPartitions)
    return tenantPartitionList.tenantPartitions
  }

  isInSuccessState(): boolean {
    return true;
  }

  getDisplayColumns(): string[] {
    return ['statusicon', 'name', 'tenantcount','pods','cpu','memory','health','etcd'];
  }

  getDisplayColumns2(): string[] {
    return ['statusicon', 'name', 'tenantcount','pods','cpu','memory','health','etcd'];
  }

  //added the code
  onClick(): void {
    this.verber_.showNodeCreateDialog(this.displayName, this.typeMeta, this.objectMeta); //added
  }
}
