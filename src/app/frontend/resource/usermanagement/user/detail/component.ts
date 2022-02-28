import {Component, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {UserDetail} from '@api/backendapi';
import {ActionbarService, ResourceMeta} from '../../../../common/services/global/actionbar';
import {NotificationsService} from '../../../../common/services/global/notifications';
import {ResourceService} from 'common/services/resource/resource';
import {EndpointManager, Resource} from 'common/services/resource/endpoint';
import {Subject} from "rxjs";
import {takeUntil} from "rxjs/operators";

@Component({
  selector: 'kd-user-detail',
  templateUrl: './template.html',
})

export class UsersDetailComponent implements OnInit, OnDestroy {
  private readonly endpoint_ = EndpointManager.resource(Resource.users);
  private readonly unsubscribe_ = new Subject<void>();

  user: UserDetail;
  isInitialized = false;

  constructor(
    private readonly user_: ResourceService<UserDetail>,
    private readonly actionbar_: ActionbarService,
    private readonly route_: ActivatedRoute,
    private readonly notifications_: NotificationsService
  ) {}

  ngOnInit(): void {
    const resourceName = this.route_.snapshot.params.resourceName;
    this.user_
      .get(this.endpoint_.detail(), resourceName)
      .pipe(takeUntil(this.unsubscribe_))
      .subscribe((d: UserDetail) => {
        this.user = d;
        this.notifications_.pushErrors(d.errors);
        this.actionbar_.onInit.emit(new ResourceMeta('User', d.objectMeta, d.typeMeta));
        this.isInitialized = true;
      });
  }

  ngOnDestroy(): void {
    this.unsubscribe_.next();
    this.unsubscribe_.complete();
    this.actionbar_.onDetailsLeave.emit();
  }
}
