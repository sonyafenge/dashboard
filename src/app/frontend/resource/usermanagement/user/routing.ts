import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {USERMANAGEMENT_ROUTE} from '../../usermanagement/routing';
import {UsersListComponent} from './list/component';
import {UsersDetailComponent} from './detail/component';
import {DEFAULT_ACTIONBAR} from "../../../common/components/actionbars/routing";

const USERS_LIST_ROUTE: Route = {
  path: '',
  component: UsersListComponent,
  data: {
    breadcrumb: 'Users',
    parent: USERMANAGEMENT_ROUTE,
  },
};

const USERS_DETAIL_ROUTE: Route = {
  path: ':resourceName',
  component: UsersDetailComponent,
  data: {
    breadcrumb: '{{ resourceName }}',
    parent: USERS_LIST_ROUTE,
  },
};

@NgModule({
  imports: [RouterModule.forChild([USERS_LIST_ROUTE, USERS_DETAIL_ROUTE, DEFAULT_ACTIONBAR])],
  exports: [RouterModule],
})
export class UsersRoutingModule {}
