import {Component, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {NamespaceDetail} from '@api/backendapi';
import {Subscription} from 'rxjs/Subscription';
import {ActionbarService, ResourceMeta} from '../../../../common/services/global/actionbar';
import {NotificationsService} from '../../../../common/services/global/notifications';
import {EndpointManager, Resource} from '../../../../common/services/resource/endpoint';
import {ResourceService} from '../../../../common/services/resource/resource';
import {TenantService} from "../../../../common/services/global/tenant";

@Component({
  selector: 'kd-clusternamespace-detail',
  templateUrl: './template.html',
})
export class NamespaceDetailComponent implements OnInit, OnDestroy {
  private namespaceSubscription_: Subscription;
  private readonly endpoint_ = EndpointManager.resource(Resource.namespace, false, true);
  namespace: NamespaceDetail;
  isInitialized = false;
  eventListEndpoint: string;

  constructor(
    private readonly namespace_: ResourceService<NamespaceDetail>,
    private readonly actionbar_: ActionbarService,
    private readonly activatedRoute_: ActivatedRoute,
    private readonly tenant_: TenantService,
    private readonly notifications_: NotificationsService,
  ) {}

  ngOnInit(): void {
    const resourceName = this.activatedRoute_.snapshot.params.resourceName;

    const resourceTenant:any = this.tenant_.current() === 'system' ?
      sessionStorage.getItem('tenantName') : this.tenant_.current()

    let endpoint = ''
    if (sessionStorage.getItem('userType') === 'cluster-admin') {
      endpoint = `api/v1/tenants/${resourceTenant}/namespace/${resourceName}`
    } else {
      endpoint = this.endpoint_.detail()
    }

    // @ts-ignore
    this.eventListEndpoint = this.endpoint_.child(resourceName, Resource.event, undefined, resourceTenant);

    this.namespaceSubscription_ = this.namespace_
      .get(endpoint, resourceName, undefined, resourceTenant)
      .subscribe((d: NamespaceDetail) => {
        this.namespace = d;
        this.notifications_.pushErrors(d.errors);
        this.actionbar_.onInit.emit(new ResourceMeta('Namespace', d.objectMeta, d.typeMeta));
        this.isInitialized = true;
      });
  }

  ngOnDestroy(): void {
    this.namespaceSubscription_.unsubscribe();
    this.actionbar_.onDetailsLeave.emit();
  }
}
