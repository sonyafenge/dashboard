import {NgModule} from '@angular/core';
import {ComponentsModule} from '../../../common/components/module';
import {SharedModule} from '../../../shared.module';
import {UsersRoutingModule} from './routing';
import {UsersListComponent} from './list/component';
import {UsersDetailComponent} from './detail/component';

@NgModule({
  imports: [SharedModule, ComponentsModule, UsersRoutingModule],
  declarations: [UsersListComponent, UsersDetailComponent],
})

export class UsersModule {}
