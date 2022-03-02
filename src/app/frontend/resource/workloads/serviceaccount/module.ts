import {NgModule} from '@angular/core';
import {ComponentsModule} from '../../../common/components/module';
import {SharedModule} from '../../../shared.module';
import {ServiceAccountDetailComponent} from './detail/component';
import {ServiceAccountListComponent} from './list/component';
import {ServiceAccountRoutingModule} from './routing';

@NgModule({
  imports: [SharedModule, ComponentsModule, ServiceAccountRoutingModule],
  declarations: [ServiceAccountListComponent, ServiceAccountDetailComponent],
})
export class ServiceAccountModule {}
