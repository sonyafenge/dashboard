import {NgModule} from '@angular/core';
import {ComponentsModule} from '../../../common/components/module';
import {SharedModule} from '../../../shared.module';
import {TpTenantRoutingModule} from './routing';
import {TpTenantListComponent} from './list/component';
import {TpTenantDetailComponent} from './detail/component';

@NgModule({
  imports: [SharedModule, ComponentsModule, TpTenantRoutingModule],
  declarations: [TpTenantListComponent, TpTenantDetailComponent],
})
export class TpTenantModule {}
