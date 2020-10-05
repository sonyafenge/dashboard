import {Injectable} from '@angular/core';
import {CanActivate} from '@angular/router';
import {CONFIG} from 'index.config';
import {TenantService} from '../global/tenant';

@Injectable()
export class SystemGuard implements CanActivate {
  constructor(private readonly tenantService_: TenantService) {}

  canActivate(): boolean {
    return this.tenantService_.isSystem() && this.tenantService_.isCurrentSystem();
  }
}
