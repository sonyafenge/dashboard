import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {NetworkDetailComponent} from './detail/component';
import {NetworkListComponent} from './list/component';
import {PIN_DEFAULT_ACTIONBAR} from '../common/components/actionbars/routing';
import {NetworkObjectDetailComponent} from './networkobject/component';

const NETWORK_LIST_ROUTE: Route = {
  path: '',
  component: NetworkListComponent,
  data: {breadcrumb: 'Custom Resource Definitions'},
};

const NETWORK_DETAIL_ROUTE: Route = {
  path: ':networkName',
  component: NetworkDetailComponent,
  data: {breadcrumb: '{{ networkName }}', parent: NETWORK_LIST_ROUTE},
};

const NETWORK_NAMESPACED_OBJECT_DETAIL_ROUTE: Route = {
  path: ':networkName/:objectName',
  component: NetworkObjectDetailComponent,
  data: {breadcrumb: '{{ objectName }}', routeParamsCount: 2, parent: NETWORK_DETAIL_ROUTE},
};

const NETWORK_CLUSTER_OBJECT_DETAIL_ROUTE: Route = {
  path: ':networkName/:objectName',
  component: NetworkObjectDetailComponent,
  data: {breadcrumb: '{{ objectName }}', routeParamsCount: 1, parent: NETWORK_DETAIL_ROUTE},
};

@NgModule({
  imports: [
    RouterModule.forChild([
      NETWORK_LIST_ROUTE,
      NETWORK_DETAIL_ROUTE,
      NETWORK_NAMESPACED_OBJECT_DETAIL_ROUTE,
      NETWORK_CLUSTER_OBJECT_DETAIL_ROUTE,
      PIN_DEFAULT_ACTIONBAR,
    ]),
  ],
})
export class NetworkRoutingModule {}
