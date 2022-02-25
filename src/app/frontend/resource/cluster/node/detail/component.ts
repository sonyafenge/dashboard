import {Component, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute,Router} from '@angular/router';
import {NodeAddress, NodeDetail, NodeTaint} from '@api/backendapi';
import {Subscription} from 'rxjs/Subscription';
import {ActionbarService, ResourceMeta} from '../../../../common/services/global/actionbar';
import {NotificationsService} from '../../../../common/services/global/notifications';
import {EndpointManager, Resource} from '../../../../common/services/resource/endpoint';
import {ResourceService} from '../../../../common/services/resource/resource';

@Component({
  selector: 'kd-node-detail',
  templateUrl: './template.html',
  styleUrls: ['./style.scss'],
})
export class NodeDetailComponent implements OnInit, OnDestroy {
  private nodeSubscription_: Subscription;
  private readonly endpoint_ = EndpointManager.resource(Resource.node);
  node: NodeDetail;
  isInitialized = false;
  podListEndpoint: string;
  eventListEndpoint: string;
  clusterName: string;
  showTenant: boolean;

  constructor(
    private readonly node_: ResourceService<NodeDetail>,
    private readonly actionbar_: ActionbarService,
    private readonly activatedRoute_: ActivatedRoute,
    private readonly router_: Router,
    private readonly notifications_: NotificationsService,
  ) {}

  ngOnInit(): void {
    this.showTenant = false;

    const resourceName = this.activatedRoute_.snapshot.params.resourceName;

    this.podListEndpoint = this.endpoint_.child(resourceName, Resource.pod);
    this.eventListEndpoint = this.endpoint_.child(resourceName, Resource.event);

    this.nodeSubscription_ = this.node_
      .get(this.endpoint_.detail(), resourceName)
      .subscribe((d: NodeDetail) => {
        this.node = d;
        this.notifications_.pushErrors(d.errors);
        this.actionbar_.onInit.emit(new ResourceMeta('Node', d.objectMeta, d.typeMeta));
        this.isInitialized = true;
        if (d.clusterName.includes('-tp')) {
          this.showTenant = true;
        }
      });
  }

  ngOnDestroy(): void {
    this.nodeSubscription_.unsubscribe();
    this.actionbar_.onDetailsLeave.emit();
  }

  getAddresses(): string[] {
    return this.node.addresses.map((address: NodeAddress) => `${address.type}: ${address.address}`);
  }

  getTaints(): string[] {
    return this.node.taints.map((taint: NodeTaint) => {
      return taint.value
        ? `${taint.key}=${taint.value}:${taint.effect}`
        : `${taint.key}=${taint.effect}`;
    });
  }
}
