import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {DEFAULT_ACTIONBAR} from '../../../common/components/actionbars/routing';
import {RoleDetailComponent} from './detail/component';
import {RoleListComponent} from './list/component';
import {TENANTMANAGEMENT_ROUTE} from "../routing";

const ROLE_LIST_ROUTE: Route = {
  path: '',
  component: RoleListComponent,
  data: {
    breadcrumb: 'Roles',
    parent: TENANTMANAGEMENT_ROUTE,
  },
};

const ROLE_DETAIL_ROUTE: Route = {
  path: ':resourceNamespace/:resourceName',
  component: RoleDetailComponent,
  data: {
    breadcrumb: '{{ resourceName }}',
    parent: ROLE_LIST_ROUTE,
  },
};

@NgModule({
  imports: [RouterModule.forChild([ROLE_LIST_ROUTE, ROLE_DETAIL_ROUTE, DEFAULT_ACTIONBAR])],
  exports: [RouterModule],
})
export class RoleRoutingModule {}
