import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {DEFAULT_ACTIONBAR} from '../../../common/components/actionbars/routing';
import {NodeDetailComponent} from './detail/component';
import {NodeListComponent} from './list/component';
import {PARTITION_LIST_ROUTE} from "../partition/routing";

const NODE_LIST_ROUTE: Route = {
  path: '',
  component: NodeListComponent,
  data: {
    breadcrumb: '{{ clusterName }}',
    parent: PARTITION_LIST_ROUTE,
  },
};

const NODE_DETAIL_ROUTE: Route = {
  path: ':resourceName',
  component: NodeDetailComponent,
  data: {
    breadcrumb: '{{ resourceName }}',
    parent: NODE_LIST_ROUTE,
  },
};

@NgModule({
  imports: [RouterModule.forChild([NODE_LIST_ROUTE, NODE_DETAIL_ROUTE, DEFAULT_ACTIONBAR])],
  exports: [RouterModule],
})
export class NodeRoutingModule {}
