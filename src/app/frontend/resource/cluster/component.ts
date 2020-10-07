// Copyright 2017 The Kubernetes Authors.
// Copyright 2020 Authors of Arktos - file modified.
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

import {Component} from '@angular/core';
import {GroupedResourceList} from '../../common/resources/groupedlist';

import {CONFIG} from '../../index.config';
import {TenantService} from '../../common/services/global/tenant';

@Component({
  selector: 'kd-cluster',
  templateUrl: './template.html',
})
export class ClusterComponent extends GroupedResourceList {
  constructor(private readonly tenantService_: TenantService) {
    super();
  }

  get isCurrentSystem(): boolean {
    return this.tenantService_.current() === CONFIG.systemTenantName;
  }
}
