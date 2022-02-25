import {HttpParams} from '@angular/common/http';
import {Component, Input, OnInit} from '@angular/core';
import {ResourcePartition, ResourcePartitionList, TenantPartition, TenantPartitionList} from '@api/backendapi';
import {Observable} from 'rxjs/Observable';
import {ResourceListWithStatuses} from '../../../resources/list';
import {NotificationsService} from '../../../services/global/notifications';
import {EndpointManager, Resource} from '../../../services/resource/endpoint';
import {ResourceService} from '../../../services/resource/resource';
import {MenuComponent} from '../../list/column/menu/component';
import {ListGroupIdentifier, ListIdentifier} from '../groupids';
import {VerberService} from "../../../services/global/verber";

@Component({
  selector: 'kd-partition-list',
  templateUrl: './template.html',
})
export class PartitionListComponent implements  OnInit {
  displayName:any;
  typeMeta:any;
  objectMeta:any;

  constructor(
    readonly verber_: VerberService,
  ) {}

  //added the code
  onClick(): void {
    this.verber_.showNodeCreateDialog(this.displayName, this.typeMeta, this.objectMeta); //added
  }

  ngOnInit(): void {
  }
}

export class ResourcePartitionListComponent extends ResourceListWithStatuses<ResourcePartitionList, ResourcePartition> {
  @Input() endpointRp = EndpointManager.resource(Resource.resourcePartition).list();

  displayName:any;
  typeMeta:any;
  objectMeta:any;

  constructor(
    readonly verber_: VerberService,
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

  getResourceObservable(params?: HttpParams): Observable<ResourcePartitionList> {
    return this.resourcePartition_.get(this.endpointRp, undefined, params);
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

  //added the code
  onClick(): void {
    this.verber_.showNodeCreateDialog(this.displayName, this.typeMeta, this.objectMeta); //added
  }
}

export class TenantPartitionListComponent extends ResourceListWithStatuses<TenantPartitionList, TenantPartition> {
  @Input() endpointTp = EndpointManager.resource(Resource.tenantPartition).list();
  displayName:any;
  typeMeta:any;
  objectMeta:any;

  constructor(
    readonly verber_: VerberService,
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

  getResourceObservable(params?: HttpParams): Observable<TenantPartitionList> {
    return this.tenantPartition_.get(this.endpointTp, undefined, params);
  }

  map(tenantPartitionList: TenantPartitionList): TenantPartition[] {
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
