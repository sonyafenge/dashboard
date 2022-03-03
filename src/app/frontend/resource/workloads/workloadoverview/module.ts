import {NgModule} from '@angular/core';
import {ComponentsModule} from '../../../common/components/module';
import {SharedModule} from '../../../shared.module';
import {WorkloadOverviewComponent} from './component';
import {WorkloadOverviewRoutingModule} from './routing';

@NgModule({
  imports: [SharedModule, ComponentsModule, WorkloadOverviewRoutingModule],
  declarations: [WorkloadOverviewComponent],
})
export class WorkloadOverviewModule {}
