import {NgModule} from '@angular/core';
import {ComponentsModule} from '../../../common/components/module';
import {SharedModule} from '../../../shared.module';
import {VirtualMachineDetailComponent} from './detail/component';
import {VirtualMachineListComponent} from './list/component';
import {VirtualMachineRoutingModule} from './routing';

@NgModule({
  imports: [SharedModule, ComponentsModule, VirtualMachineRoutingModule],
  declarations: [VirtualMachineListComponent, VirtualMachineDetailComponent],
})
export class VirtualMachineModule {}
