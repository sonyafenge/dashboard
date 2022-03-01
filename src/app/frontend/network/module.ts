import {NgModule} from '@angular/core';
import {ComponentsModule} from '../common/components/module';
import {SharedModule} from '../shared.module';
import {NetworkRoutingModule} from './routing';
import {NetworkDetailComponent} from './detail/component';
import {NetworkListComponent} from './list/component';
import {NetworkObjectDetailComponent} from './networkobject/component';

@NgModule({
  imports: [SharedModule, ComponentsModule, NetworkRoutingModule],
  declarations: [NetworkListComponent, NetworkDetailComponent, NetworkObjectDetailComponent],
})
export class NetworkModule {}
