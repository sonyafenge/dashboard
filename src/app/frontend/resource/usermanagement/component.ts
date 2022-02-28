import {Component} from '@angular/core';
import {GroupedResourceList} from "../../common/resources/groupedlist";
import {TenantService} from "../../common/services/global/tenant";
import {CONFIG} from "../../index.config";

@Component({
  selector: 'kd-usermanagement',
  templateUrl: './template.html'
})

export class UserManagementComponent extends GroupedResourceList {
  constructor(private readonly tenantService_: TenantService) {
    super();
  }

  get isCurrentSystem(): boolean {
    return this.tenantService_.current() === CONFIG.systemTenantName;
  }
}
