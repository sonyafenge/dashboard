import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {CLUSTER_ROUTE} from '../routing';
import {PartitionListComponent} from './list/component';

export const PARTITION_LIST_ROUTE: Route = {
  path: '',
  component: PartitionListComponent,
  data: {
    breadcrumb: 'Partition',
    parent: CLUSTER_ROUTE,
  },
};

@NgModule({
  imports: [RouterModule.forChild([PARTITION_LIST_ROUTE])],
  exports: [RouterModule],
})
export class PartitionRoutingModule {}
