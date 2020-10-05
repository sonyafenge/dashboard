import {Injectable} from '@angular/core';
import {CONFIG} from 'index.config';
import {CookieService} from 'ngx-cookie-service';

@Injectable()
export class TenantService {
  private isSystemTenant_: boolean;
  private authTenant_: string;
  private currentTenant_ = '';

  private readonly tenantRegex = /^([a-z0-9]([-a-z0-9]*[a-z0-9])?)$/; // TODO: need to verify

  constructor(private readonly cookies_: CookieService) {
    const tenantCookie = this.cookies_.get(CONFIG.authTenantCookieName) || '';
    this.setAuthTenant(tenantCookie);
  }

  // Set to a differnet tenant when logged in as a system tenant.
  setCurrent(tenant: string) {
    if (this.isSystemTenant_) {
      this.currentTenant_ = tenant;
    }
  }

  // Only system can set and return current tenant , otherwise return empty.
  current(): string {
    return this.currentTenant_ || '';
  }

  setAuthTenant(tenant: string) {
    this.authTenant_ = tenant;
    this.currentTenant_ = tenant;
    this.isSystemTenant_ = tenant === CONFIG.systemTenantName;
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
