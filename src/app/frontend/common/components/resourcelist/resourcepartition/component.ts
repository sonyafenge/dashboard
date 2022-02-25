import {HttpParams} from '@angular/common/http';
import {Component, Input} from '@angular/core';
import {ResourcePartitionList, ResourcePartition} from '@api/backendapi';
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
  selector: 'kd-resource-partition-list',
  templateUrl: './template.html',
})
// @ts-ignore
export class ResourcePartitionListComponent extends ResourceListWithStatuses<ResourcePartitionList, ResourcePartition> {
  @Input() endpointTp = EndpointManager.resource(Resource.resourcePartition).list();

  constructor(
    readonly verber_: VerberService,
    private router_: Router,
    private readonly resourcePartition_: ResourceService<ResourcePartitionList>,
    notifications: NotificationsService,
  ) {
    super('resourcePartition', notifications);
    this.id = ListIdentifier.resourcePartition;
    this.groupId = ListGroupIdentifier.cluster;

    // Register action columns.
    this.registerActionColumn<MenuComponent>('menu', MenuComponent);

    // Register status icon handlers
    this.registerBinding(this.icon.checkCircle, 'kd-success', this.isInSuccessState);
  }

  setClusterName($event:any) {
    const clusterName = $event.target.innerHTML.replace(/^\s+|\s+$/gm,'');
    this.router_.navigateByUrl('/node', {state: {clusterName}});
  }

  getResourceObservable(params?: HttpParams): Observable<ResourcePartitionList> {
    return this.resourcePartition_.get(this.endpointTp, undefined, params);
  }

  map(resourcePartitionList: ResourcePartitionList): ResourcePartition[] {
    return resourcePartitionList.resourcePartitions
  }

  isInSuccessState(): boolean {
    return true;
  }

  getDisplayColumns(): string[] {
    return ['statusicon', 'name', 'nodecount','cpu','memory','health','etcd'];
  }

  getDisplayColumns2(): string[] {
    return ['statusicon', 'name', 'nodecount','cpu','memory','health','etcd'];
  }

}
