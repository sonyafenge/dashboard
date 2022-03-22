// Copyright 2020 Authors of Arktos.
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

import {Injectable} from '@angular/core';
import {CONFIG} from 'index.config';
import {CookieService} from 'ngx-cookie-service';

@Injectable()
export class TenantService {
  private isSystemTenant_: boolean;
  private authTenant_: string;
  private currentTenant_ = '';
  private resourceTenant_ = '';

  private readonly tenantRegex = /^([a-z0-9]([-a-z0-9]*[a-z0-9])?)$/; // TODO: need to verify

  constructor(private readonly cookies_: CookieService) {
    const tenantCookie = this.cookies_.get(CONFIG.authTenantCookieName) || '';
    this.setAuthTenant(tenantCookie);
  }

  setCurrent(tenant: string) {
    if (this.isSystemTenant_) {
      this.currentTenant_ = tenant;
    }

  }


  current(): string {
    const username = sessionStorage.getItem('parentTenant');
    const userType = sessionStorage.getItem('userType');
    if (userType === 'cluster-admin'){
      this.currentTenant_ = 'system'
    } else {
      this.currentTenant_ = username
    }
    return this.currentTenant_;
  }

  resourceTenant(): string {
    if (sessionStorage.getItem('currentTenant')) {
      this.resourceTenant_ = sessionStorage.getItem('currentTenant')
    } else if (sessionStorage.getItem('currentTpTenant')) {
      this.resourceTenant_ = sessionStorage.getItem('currentTpTenant')
    }
    return this.resourceTenant_
  }

  tenantPartition(): string {
    return sessionStorage.getItem(this.resourceTenant())
  }

  setAuthTenant(tenant: string) {
    this.authTenant_ = tenant;
    this.isSystemTenant_ = tenant === CONFIG.systemTenantName;
    this.setCurrent(tenant);
  }

  getAuthTenant(): string {
    return this.authTenant_;
  }

  isTenantValid(tenant: string): boolean {
    return this.tenantRegex.test(tenant);
  }

  isSystem(): boolean {
    return this.isSystemTenant_;
  }

  isCurrentSystem(): boolean {
    return this.currentTenant_ === CONFIG.systemTenantName;
  }
}
