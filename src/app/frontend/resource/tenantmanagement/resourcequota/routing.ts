import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {TENANTMANAGEMENT_ROUTE} from '../routing';
import {ResourceQuotasListComponent} from './list/component';
import {ResourceQuotaDetailComponent} from './detail/component';

const TENANTQUOTA_LIST_ROUTE: Route = {
  path: '',
  component: ResourceQuotasListComponent,
  data: {
    breadcrumb: 'Quota',
    parent: TENANTMANAGEMENT_ROUTE,
  },
};

const TENANTQUOTA_DETAIL_ROUTE: Route = {
  path: ':resourceName',
  component: ResourceQuotaDetailComponent,
  data: {
    breadcrumb: '{{ resourceName }}',
    parent: TENANTQUOTA_LIST_ROUTE,
  },
};

@NgModule({
  imports: [RouterModule.forChild([TENANTQUOTA_LIST_ROUTE, TENANTQUOTA_DETAIL_ROUTE])],
  exports: [RouterModule],
})
export class ResourceQuotaRoutingModule {

}
