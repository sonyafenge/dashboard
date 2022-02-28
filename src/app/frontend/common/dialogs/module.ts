import {NgModule} from '@angular/core';
import {SharedModule} from '../../shared.module';
import {ComponentsModule} from '../components/module';
import {AlertDialog} from './alert/dialog';
import {DeleteResourceDialog} from './deleteresource/dialog';
import {LogsDownloadDialog} from './download/dialog';
import {EditResourceDialog} from './editresource/dialog';
import {ScaleResourceDialog} from './scaleresource/dialog';
import {TriggerResourceDialog} from './triggerresource/dialog';
import {CreateTenantDialog} from "./createTenant/dialog";
import {CreateNamespaceDialog} from "./createNamespace/dialog";
import {CreateNodeDialog} from "./createNode/dialog";
import {CreateClusterroleDialog} from "./createClusterrole/dialog";
import {CreateRoleDialog} from "./createRole/dialog";
import {assignQuotaDialog} from "./assignQuota/dialog";
import {CreateUserDialog} from "./createUser/dialog";

@NgModule({
  imports: [SharedModule, ComponentsModule],
  declarations: [
    AlertDialog,
    EditResourceDialog,
    DeleteResourceDialog,
    LogsDownloadDialog,
    ScaleResourceDialog,
    TriggerResourceDialog,
    CreateTenantDialog,
    CreateNamespaceDialog,
    CreateNodeDialog,
    CreateClusterroleDialog,
    CreateRoleDialog,
    assignQuotaDialog,
    CreateUserDialog,
  ],
  exports: [
    AlertDialog,
    EditResourceDialog,
    DeleteResourceDialog,
    LogsDownloadDialog,
    ScaleResourceDialog,
    TriggerResourceDialog,
    CreateTenantDialog,
    CreateNamespaceDialog,
    CreateNodeDialog,
    CreateClusterroleDialog,
    CreateRoleDialog,
    assignQuotaDialog,
    CreateUserDialog,
  ],
  entryComponents: [
    AlertDialog,
    EditResourceDialog,
    DeleteResourceDialog,
    LogsDownloadDialog,
    ScaleResourceDialog,
    TriggerResourceDialog,
    CreateTenantDialog,
    CreateNamespaceDialog,
    CreateNodeDialog,
    CreateClusterroleDialog,
    CreateRoleDialog,
    assignQuotaDialog,
    CreateUserDialog,
  ],
})
export class DialogsModule {}
