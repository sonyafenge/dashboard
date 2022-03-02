import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {DEFAULT_ACTIONBAR} from '../../../common/components/actionbars/routing';
import {WORKLOADS_ROUTE} from '../routing';
import {ServiceAccountDetailComponent} from './detail/component';
import {ServiceAccountListComponent} from './list/component';

const SERVICEACCOUNT_LIST_ROUTE: Route = {
  path: '',
  component: ServiceAccountListComponent,
  data: {
    breadcrumb: 'Service Accounts',
    parent: WORKLOADS_ROUTE,
  },
};

const SERVICEACCOUNT_DETAIL_ROUTE: Route = {
  path: ':resourceNamespace/:resourceName',
  component: ServiceAccountDetailComponent,
  data: {
    breadcrumb: '{{ resourceName }}',
    parent: SERVICEACCOUNT_LIST_ROUTE,
  },
};

@NgModule({
  imports: [RouterModule.forChild([SERVICEACCOUNT_LIST_ROUTE, SERVICEACCOUNT_DETAIL_ROUTE, DEFAULT_ACTIONBAR])],
  exports: [RouterModule],
})
export class ServiceAccountRoutingModule {}
