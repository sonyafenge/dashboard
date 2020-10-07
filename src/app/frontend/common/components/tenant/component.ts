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

import {Component, OnInit, ViewChild, ElementRef} from '@angular/core';
import {Router, ActivatedRoute, NavigationEnd} from '@angular/router';
import {Subject} from 'rxjs';
import {takeUntil, startWith, switchMap} from 'rxjs/operators';
import {MatSelect} from '@angular/material';

import {TenantList} from '@api/backendapi';
import {TENANT_STATE_PARAM, NAMESPACE_STATE_PARAM} from '../../params/params';
import {TenantService} from '../../services/global/tenant';
import {ResourceService} from 'common/services/resource/resource';
import {EndpointManager, Resource} from 'common/services/resource/endpoint';
import {NotificationsService, NotificationSeverity} from 'common/services/global/notifications';
import {CONFIG} from 'index.config';

@Component({
  selector: 'kd-tenant-selector',
  templateUrl: './template.html',
  styleUrls: ['style.scss'],
})
export class TenantSelectorComponent implements OnInit {
  private tenantUpdate_ = new Subject();
  private unsubscribe_ = new Subject();
  private readonly endpoint_ = EndpointManager.resource(Resource.tenant);

  tenants: string[] = [];
  selectedTenant: string;
  resourceNameParam: string;
  selectTenantInput = '';
  systemTenantName = CONFIG.systemTenantName;

  @ViewChild(MatSelect, {static: true}) private readonly select_: MatSelect;
  @ViewChild('tenantInput', {static: true}) private readonly tenantInputEl_: ElementRef;

  constructor(
    private readonly router_: Router,
    private readonly tenantService_: TenantService,
    private readonly tenant_: ResourceService<TenantList>,
    private readonly notifications_: NotificationsService,
    private readonly _activeRoute: ActivatedRoute,
  ) {}

  ngOnInit(): void {
    this._activeRoute.queryParams.pipe(takeUntil(this.unsubscribe_)).subscribe(params => {
      const tenant = params.tenant;

      if (!tenant) {
        this.setDefaultQueryParams_();
        return;
      }

      if (this.tenantService_.current() === tenant) {
        return;
      }

      this.tenantService_.setCurrent(tenant);
      this.selectedTenant = tenant;
    });

    this.selectedTenant = this.tenantService_.current();
    this.select_.value = this.selectTenant;
    this.loadTenants_();
  }

  ngOnDestroy(): void {
    this.unsubscribe_.next();
    this.unsubscribe_.complete();
  }

  loadTenants_(): void {
    this.tenantUpdate_
      .pipe(takeUntil(this.unsubscribe_))
      .pipe(startWith({}))
      .pipe(switchMap(() => this.tenant_.get(this.endpoint_.list())))
      .subscribe(
        tenantList => {
          this.tenants = tenantList.tenants
            .map(t => t.objectMeta.name)
            .filter(t => t !== this.systemTenantName);

          if (tenantList.errors.length > 0) {
            for (const err of tenantList.errors) {
              this.notifications_.push(err.ErrStatus.message, NotificationSeverity.error);
            }
          }
        },
        () => {},
        () => {
          this.onTenantLoaded_();
        },
      );
  }

  selectTenant(): void {
    this.changeTenant_(this.selectedTenant);
  }

  onTenantToggle(opened: boolean): void {
    if (opened) {
      this.tenantUpdate_.next();
      this.focusTenantInput_();
    } else {
      this.changeTenant_(this.selectedTenant);
    }
  }

  /**
   * When state is loaded and tenants are fetched, perform basic validation.
   */
  private onTenantLoaded_(): void {
    let newTenant = this.tenantService_.getAuthTenant();
    const targetTenant = this.selectedTenant;

    if (
      targetTenant &&
      (this.tenants.indexOf(targetTenant) >= 0 || this.tenantService_.isTenantValid(targetTenant))
    ) {
      newTenant = targetTenant;
    }

    if (newTenant !== this.selectedTenant) {
      this.changeTenant_(newTenant);
    }
  }

  /**
   * Focuses tenant input field after clicking on tenant selector menu.
   */
  private focusTenantInput_(): void {
    // Wrap in a timeout to make sure that element is rendered before looking for it.
    setTimeout(() => {
      this.tenantInputEl_.nativeElement.focus();
    }, 150);
  }

  private clearTenantInput_(): void {
    this.selectTenantInput = '';
  }

  private changeTenant_(tenant: string): void {
    this.clearTenantInput_();

    this.router_.navigate(['overview'], {
      queryParams: {
        [TENANT_STATE_PARAM]: tenant,
        [NAMESPACE_STATE_PARAM]: '',
      },
      queryParamsHandling: 'merge',
    });
  }

  setDefaultQueryParams_(): void {
    this.router_.navigate([this._activeRoute.snapshot.url], {
      queryParams: {[TENANT_STATE_PARAM]: 'system'},
      queryParamsHandling: 'merge',
    });
  }
}
