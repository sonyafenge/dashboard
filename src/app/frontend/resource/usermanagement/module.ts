import {NgModule} from '@angular/core';
import {ComponentsModule} from '../../common/components/module';
import {SharedModule} from '../../shared.module';
import {UserManagementComponent} from './component';
import {UserRoutingModule} from './routing';

@NgModule({
  imports: [SharedModule, ComponentsModule, UserRoutingModule],
  declarations: [UserManagementComponent],
})
export class UserManagementModule {}
