import {HttpParams} from '@angular/common/http';
import {Component, Input} from '@angular/core';
import {ObjectMeta, ResourceQuota, ResourceQuotaList, TypeMeta} from '@api/backendapi';
import {Observable} from 'rxjs/Observable';
import {ResourceListWithStatuses} from '../../../resources/list';
import {EndpointManager, Resource} from '../../../services/resource/endpoint';
import {NamespacedResourceService} from '../../../services/resource/resource';
import {NotificationsService} from '../../../services/global/notifications';
import {ListGroupIdentifier, ListIdentifier} from '../groupids';
import {MenuComponent} from '../../list/column/menu/component';
import {VerberService} from '../../../services/global/verber';
import {ActivatedRoute} from "@angular/router";
import {TenantService} from "../../../services/global/tenant";

@Component({
  selector: 'kd-resourcequota-list',
  templateUrl: './template.html',
})
export class ResourceQuotasListComponent extends ResourceListWithStatuses<ResourceQuotaList, ResourceQuota> {
  @Input() endpoint = EndpointManager.resource(Resource.resourcequota, true, true).list();

  displayName: string;
  typeMeta: TypeMeta;
  objectMeta: ObjectMeta;
  tenantName: string;

  constructor(
    public readonly verber_: VerberService,
    private readonly resourcequota_: NamespacedResourceService<ResourceQuotaList>,
    notifications: NotificationsService,
    private readonly tenant_: TenantService,
    private readonly activatedRoute_: ActivatedRoute,
  ) {

    super('resourcequota', notifications);
    this.id = ListIdentifier.resourcequota;
    this.groupId = ListGroupIdentifier.cluster;

    // Register action columns.
    this.registerActionColumn<MenuComponent>('menu', MenuComponent);

    this.registerBinding(this.icon.checkCircle, 'kd-success', this.isInSuccessState);
    this.tenantName = this.activatedRoute_.snapshot.params.resourceName === undefined ?
      this.tenant_.current() : this.activatedRoute_.snapshot.params.resourceName
    sessionStorage.setItem('tenantName',this.tenantName);
  }

  isInSuccessState(): boolean {
    return true;
  }

  getResourceObservable(params?: HttpParams): Observable<ResourceQuotaList> {
    let endpoint = ''
    if (sessionStorage.getItem('userType') === 'cluster-admin') {
      endpoint = `api/v1/tenants/${this.tenantName}/resourcequota`
    } else {
      endpoint = this.endpoint
    }

    return this.resourcequota_.get(endpoint, undefined, undefined, params, this.tenantName);
  }

  map(resourcequotaList: ResourceQuotaList): ResourceQuota[] {
    return resourcequotaList.items;
  }

  getDisplayColumns(): string[] {
    return ['statusicon', 'name', 'namespace', 'age'];
  }

  getDisplayColumns2(): string[] {
    return ['statusicon', 'name', 'namespace', 'age'];
  }

  onClick(): void {
    this.verber_.showResourceQuotaCreateDialog(this.displayName, this.typeMeta, this.objectMeta);
  }
}
