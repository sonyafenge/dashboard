import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {TenantManagementComponent} from './component';

export const TENANTMANAGEMENT_ROUTE: Route = {
  path: '',
  component:TenantManagementComponent,
  data: {
    breadcrumb: 'Tenant Management',
    link: ['', 'tenantmanagement'],
  },
};

@NgModule({
  imports: [RouterModule.forChild([TENANTMANAGEMENT_ROUTE])],
  exports: [RouterModule],
})
export class TenantManagementRoutingModule {}
