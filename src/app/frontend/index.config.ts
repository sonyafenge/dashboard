import {InjectionToken} from '@angular/core';
import {MatTooltipDefaultOptions} from '@angular/material';

const username = sessionStorage.getItem('parentTenant');
const userType = sessionStorage.getItem('userType');
const defaultnamespace = sessionStorage.getItem('namespace');
let tenant = '';
if (userType === 'cluster-admin'){
  tenant = 'system'
} else {
  tenant = username
}

let namespace = '';

if (defaultnamespace !== '') {
  namespace = defaultnamespace
} else {
  namespace = 'default'
}

export let CONFIG_DI_TOKEN = new InjectionToken<Config>('kd.config');

export interface Config {
  authTokenCookieName: string;
  authTenantCookieName: string;
  skipLoginPageCookieName: string;
  csrfHeaderName: string;
  authTokenHeaderName: string;
  defaultNamespace: string;
  systemTenantName: string;
}

export const CONFIG: Config = {
  authTokenCookieName: 'jweToken',
  authTokenHeaderName: 'jweToken',
  authTenantCookieName: 'tenant',
  csrfHeaderName: 'X-CSRF-TOKEN',
  skipLoginPageCookieName: 'skipLoginPage',
  defaultNamespace: namespace,//'default',  //default
  systemTenantName: tenant,  //system
};

// Override default material tooltip values.
export const KD_TOOLTIP_DEFAULT_OPTIONS: MatTooltipDefaultOptions = {
  showDelay: 500,
  hideDelay: 0,
  touchendHideDelay: 0,
};
