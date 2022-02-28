import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {UserManagementComponent} from './component';

export const USERMANAGEMENT_ROUTE: Route = {
  path: '',
  component:UserManagementComponent,
  data: {
    breadcrumb: 'User Management',
    link: ['', 'usermanagement'],
  },
};

@NgModule({
  imports: [RouterModule.forChild([USERMANAGEMENT_ROUTE])],
  exports: [RouterModule],
})
export class UserRoutingModule {}
