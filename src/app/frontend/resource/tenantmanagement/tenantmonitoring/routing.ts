import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {TenantMonitoringListComponent} from "./list/component";
import {TENANTMANAGEMENT_ROUTE} from "../routing";
import {TenantMonitoringDetailComponent} from "./detail/component";
import {DEFAULT_ACTIONBAR} from "../../../common/components/actionbars/routing";

const TENANTMONITORING_LIST_ROUTE: Route = {
  path: '',
  component: TenantMonitoringListComponent,
  data: {
    breadcrumb: 'Tenant Monitoring',
    parent: TENANTMANAGEMENT_ROUTE,
  },
};


const TENANTMONITORING_DETAIL_ROUTE: Route = {
  path: ':resourceName',
  component: TenantMonitoringDetailComponent,
  data: {
    breadcrumb: ' {{ resourceName }} ',
    parent: TENANTMONITORING_LIST_ROUTE,
  },
};

@NgModule({
  imports: [RouterModule.forChild([TENANTMONITORING_LIST_ROUTE, TENANTMONITORING_DETAIL_ROUTE, DEFAULT_ACTIONBAR])],
  exports: [RouterModule],
})
export class TenantMonitoringRoutingModule {}
