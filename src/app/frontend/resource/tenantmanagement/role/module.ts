import {NgModule} from '@angular/core';
import {ComponentsModule} from '../../../common/components/module';
import {SharedModule} from '../../../shared.module';
import {RoleDetailComponent} from './detail/component';
import {RoleListComponent} from './list/component';
import {RoleRoutingModule} from 'resource/tenantmanagement/role/routing';

@NgModule({
  imports: [SharedModule, ComponentsModule, RoleRoutingModule],
  declarations: [RoleListComponent, RoleDetailComponent],
})
export class RoleModule {}
