import {Component, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {CRDDetail} from '@api/backendapi';
import {Subscription} from 'rxjs';
import {ActionbarService, ResourceMeta} from '../../common/services/global/actionbar';
import {NotificationsService} from '../../common/services/global/notifications';
import {ResourceService} from '../../common/services/resource/resource';
import {EndpointManager, Resource} from '../../common/services/resource/endpoint';

@Component({selector: 'kd-crd-detail', templateUrl: './template.html'})
export class CRDDetailComponent implements OnInit, OnDestroy {
  private crdSubscription_: Subscription;
  private readonly endpoint_ = EndpointManager.resource(Resource.crd,false,true);
  crd: CRDDetail;
  crdObjectEndpoint: string;
  isInitialized = false;

  constructor(
    private readonly crd_: ResourceService<CRDDetail>,
    private readonly actionbar_: ActionbarService,
    private readonly activatedRoute_: ActivatedRoute,
    private readonly notifications_: NotificationsService,
  ) {}

  ngOnInit(): void {
    const {crdName} = this.activatedRoute_.snapshot.params;
    this.crdObjectEndpoint = EndpointManager.resource(Resource.crd, false,true).child(
      crdName,
      Resource.crdObject,
    );
    this.crdSubscription_ = this.crd_
      .get(this.endpoint_.detail(), crdName)
      .subscribe((d: CRDDetail) => {
        this.crd = d;
        this.notifications_.pushErrors(d.errors);
        this.actionbar_.onInit.emit(
          new ResourceMeta('Custom Resource Definition', d.objectMeta, d.typeMeta),
        );
        this.isInitialized = true;
      });
  }

  ngOnDestroy(): void {
    this.crdSubscription_.unsubscribe();
    this.actionbar_.onDetailsLeave.emit();
  }
}
