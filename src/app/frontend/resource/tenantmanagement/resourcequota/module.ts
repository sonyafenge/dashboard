import {NgModule} from '@angular/core';
import {ComponentsModule} from '../../../common/components/module';
import {SharedModule} from '../../../shared.module';
import {ResourceQuotaRoutingModule} from './routing';
import {ResourceQuotasListComponent} from './list/component';
import {ResourceQuotaDetailComponent} from './detail/component';

@NgModule({
  imports: [SharedModule, ComponentsModule, ResourceQuotaRoutingModule],
  declarations: [ResourceQuotasListComponent, ResourceQuotaDetailComponent],
})
export class ResourceQuotaModule {}
