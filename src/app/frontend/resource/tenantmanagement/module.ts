import {NgModule} from '@angular/core';
import {ComponentsModule} from '../../common/components/module';
import {SharedModule} from '../../shared.module';
import {TenantManagementComponent} from './component';
import {TenantManagementRoutingModule} from './routing';

@NgModule({
  imports: [SharedModule, ComponentsModule, TenantManagementRoutingModule],
  declarations: [TenantManagementComponent],
})
export class TenantManagementModule {}
