import {Component} from '@angular/core';
import {GroupedResourceList} from '../../common/resources/groupedlist';
import {CONFIG} from '../../index.config';
import {TenantService} from '../../common/services/global/tenant';

@Component({
  selector: 'kd-cluster',
  templateUrl: './template.html',
})
export class ClusterComponent extends GroupedResourceList {
  constructor(private readonly tenantService_: TenantService) {
    super();
  }

  get isCurrentSystem(): boolean {
    return this.tenantService_.current() === CONFIG.systemTenantName;
  }
}
