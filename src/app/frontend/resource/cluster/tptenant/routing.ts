import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {CLUSTER_ROUTE} from '../routing';
import {TpTenantListComponent} from './list/component';
import {TpTenantDetailComponent} from './detail/component';

const TPTENANT_LIST_ROUTE: Route = {
  path: '',
  component: TpTenantListComponent,
  data: {
    breadcrumb: 'Tenant',
    parent: CLUSTER_ROUTE,
  },
};

const TPTENANT_DETAIL_ROUTE: Route = {
  path: ':resourceName',
  component: TpTenantDetailComponent,
  data: {
    breadcrumb: '{{ resourceName }}',
    parent: TPTENANT_LIST_ROUTE,
  },
};

@NgModule({
  imports: [RouterModule.forChild([TPTENANT_LIST_ROUTE, TPTENANT_DETAIL_ROUTE])],
  exports: [RouterModule],
})
export class TpTenantRoutingModule {}
