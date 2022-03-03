import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {WorkloadOverviewComponent} from './component';
import {WORKLOADS_ROUTE} from "../routing";

export const OVERVIEW_ROUTE: Route = {
  path: '',
  component: WorkloadOverviewComponent,
  data: {
    breadcrumb: 'Overview',
    parent: WORKLOADS_ROUTE,
  },
};

@NgModule({
  imports: [RouterModule.forChild([OVERVIEW_ROUTE])],
  exports: [RouterModule],
})

export class WorkloadOverviewRoutingModule {}
