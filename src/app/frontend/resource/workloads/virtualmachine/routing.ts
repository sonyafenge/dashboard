import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {LOGS_EXEC_DEFAULT_ACTIONBAR} from '../../../common/components/actionbars/routing';
import {WORKLOADS_ROUTE} from '../routing';
import {VirtualMachineDetailComponent} from './detail/component';
import {VirtualMachineListComponent} from './list/component';

const VIRTUAL_MACHINE_LIST_ROUTE: Route = {
  path: '',
  component: VirtualMachineListComponent,
  data: {
    breadcrumb: 'VirtualMachines',
    parent: WORKLOADS_ROUTE,
  },
};

export const VIRTUAL_MACHINE_DETAIL_ROUTE: Route = {
  path: ':resourceNamespace/:resourceName',
  component: VirtualMachineDetailComponent,
  data: {
    breadcrumb: '{{ resourceName }}',
    parent: VIRTUAL_MACHINE_LIST_ROUTE,
  },
};

@NgModule({
  imports: [RouterModule.forChild([VIRTUAL_MACHINE_LIST_ROUTE, VIRTUAL_MACHINE_DETAIL_ROUTE, LOGS_EXEC_DEFAULT_ACTIONBAR])],
  exports: [RouterModule],
})
export class VirtualMachineRoutingModule {}
