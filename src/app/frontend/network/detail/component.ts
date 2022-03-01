import {Component, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {NetworkDetail} from '@api/backendapi';
import {Subscription} from 'rxjs';
import {ActionbarService, ResourceMeta} from '../../common/services/global/actionbar';
import {NotificationsService} from '../../common/services/global/notifications';
import {ResourceService} from '../../common/services/resource/resource';
import {EndpointManager, Resource} from '../../common/services/resource/endpoint';

@Component({selector: 'kd-network-detail', templateUrl: './template.html'})
export class NetworkDetailComponent implements OnInit, OnDestroy {
  private networkSubscription_: Subscription;
  private readonly endpoint_ = EndpointManager.resource(Resource.network);
  network: NetworkDetail;
  networkObjectEndpoint: string;
  isInitialized = false;

  constructor(
    private readonly network_: ResourceService<NetworkDetail>,
    private readonly actionbar_: ActionbarService,
    private readonly activatedRoute_: ActivatedRoute,
    private readonly notifications_: NotificationsService,
  ) {}

  ngOnInit(): void {
    const {networkName} = this.activatedRoute_.snapshot.params;
    this.networkObjectEndpoint = EndpointManager.resource(Resource.network, false,true).child(
      networkName,
      Resource.networkObject,
    );

    this.networkSubscription_ = this.network_
      .get(this.endpoint_.detail(), networkName)
      .subscribe((d: NetworkDetail) => {
        this.network = d;
        this.notifications_.pushErrors(d.errors);
        this.actionbar_.onInit.emit(
          new ResourceMeta('Network', d.objectMeta, d.typeMeta),
        );
        this.isInitialized = true;
      });
  }

  ngOnDestroy(): void {
    this.networkSubscription_.unsubscribe();
    this.actionbar_.onDetailsLeave.emit();
  }
}
