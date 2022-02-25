import {NgModule} from '@angular/core';
import {ComponentsModule} from '../../../common/components/module';
import {SharedModule} from '../../../shared.module';
import {PartitionDetailComponent} from './detail/component';
import {PartitionListComponent} from './list/component';
import {PartitionRoutingModule} from './routing';

@NgModule({
  imports: [SharedModule, ComponentsModule, PartitionRoutingModule],
  declarations: [PartitionListComponent, PartitionDetailComponent],
})
export class PartitionModule {}
