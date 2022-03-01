import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {CRDDetailComponent} from './detail/component';
import {CRDListComponent} from './list/component';
import {PIN_DEFAULT_ACTIONBAR} from '../common/components/actionbars/routing';
import {CRDObjectDetailComponent} from './crdobject/component';

const CRD_LIST_ROUTE: Route = {
  path: '',
  component: CRDListComponent,
  data: {breadcrumb: 'Custom Resource Definitions'},
};

const CRD_DETAIL_ROUTE: Route = {
  path: ':crdName',
  component: CRDDetailComponent,
  data: {breadcrumb: '{{ crdName }}', parent: CRD_LIST_ROUTE},
};

const CRD_NAMESPACED_OBJECT_DETAIL_ROUTE: Route = {
  path: ':crdName/:objectName',
  component: CRDObjectDetailComponent,
  data: {breadcrumb: '{{ objectName }}', routeParamsCount: 2, parent: CRD_DETAIL_ROUTE},
};

const CRD_CLUSTER_OBJECT_DETAIL_ROUTE: Route = {
  path: ':crdName/:objectName',
  component: CRDObjectDetailComponent,
  data: {breadcrumb: '{{ objectName }}', routeParamsCount: 1, parent: CRD_DETAIL_ROUTE},
};

@NgModule({
  imports: [
    RouterModule.forChild([
      CRD_LIST_ROUTE,
      CRD_DETAIL_ROUTE,
      CRD_NAMESPACED_OBJECT_DETAIL_ROUTE,
      CRD_CLUSTER_OBJECT_DETAIL_ROUTE,
      PIN_DEFAULT_ACTIONBAR,
    ]),
  ],
})
export class CRDRoutingModule {}
