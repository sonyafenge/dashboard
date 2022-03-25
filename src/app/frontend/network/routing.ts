// Copyright 2017 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
