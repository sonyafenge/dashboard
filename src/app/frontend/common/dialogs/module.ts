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
  ],
})
export class DialogsModule {}
