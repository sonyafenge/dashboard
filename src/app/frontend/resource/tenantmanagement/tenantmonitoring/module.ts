import {NgModule} from '@angular/core';
import {ComponentsModule} from '../../../common/components/module';
import {SharedModule} from '../../../shared.module';
import {TenantMonitoringListComponent} from '../tenantmonitoring/list/component';
import {TenantMonitoringDetailComponent} from '../tenantmonitoring/detail/component';
import {TenantMonitoringRoutingModule} from './routing';

@NgModule({
  imports: [SharedModule, ComponentsModule, TenantMonitoringRoutingModule],
  declarations: [TenantMonitoringListComponent,TenantMonitoringDetailComponent],
})
export class TenantMonitoringModule {}
