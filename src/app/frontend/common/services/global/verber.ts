// Copyright 2017 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import {HttpClient, HttpErrorResponse, HttpHeaders} from '@angular/common/http';
import {EventEmitter, Injectable} from '@angular/core';
import {MatDialog, MatDialogConfig} from '@angular/material';
import {ObjectMeta, TypeMeta} from '@api/backendapi';
import {AlertDialog, AlertDialogConfig} from '../../dialogs/alert/dialog';
import {DeleteResourceDialog} from '../../dialogs/deleteresource/dialog';
import {EditResourceDialog} from '../../dialogs/editresource/dialog';
import {ScaleResourceDialog} from '../../dialogs/scaleresource/dialog';
import {TriggerResourceDialog} from '../../dialogs/triggerresource/dialog';
import {RawResource} from '../../resources/rawresource';
import {CreateTenantDialog } from '../../dialogs/createTenant/dialog';
import {CreateUserDialog} from '../../dialogs/createUser/dialog';
import {CreateNamespaceDialog} from '../../dialogs/createNamespace/dialog';
import {CreateRoleDialog} from '../../dialogs/createRole/dialog';
import {CreateClusterroleDialog} from '../../dialogs/createClusterrole/dialog';
import {CreateAssignQuotaDialog } from '../../dialogs/createAssignQuota/dialog';
import {ResourceMeta} from './actionbar';
import {TenantService} from './tenant';

@Injectable()
export class VerberService {
  onCreate = new EventEmitter<boolean>();
  onCreateTenant = new EventEmitter<boolean>();
  onDelete = new EventEmitter<boolean>();
  onEdit = new EventEmitter<boolean>();
  onScale = new EventEmitter<boolean>();
  onTrigger = new EventEmitter<boolean>();

  constructor(
    private readonly dialog_: MatDialog,
    private readonly http_: HttpClient,
    private tenant_: TenantService,

  ) {}

  // create Tenant dialog
  showTenantCreateDialog(displayName: string, typeMeta: TypeMeta, objectMeta: ObjectMeta): void {
    const dialogConfig = this.getDialogConfig_(displayName, typeMeta, objectMeta);
    this.dialog_
      .open(CreateTenantDialog, dialogConfig)
      .afterClosed()
      .subscribe(result => {
        if (result) {
        }
      });
  }

  // create User dialog
  showUserCreateDialog(displayName: string, typeMeta: TypeMeta, objectMeta: ObjectMeta): void {
    const dialogConfig = this.getDialogConfig_(displayName, typeMeta, objectMeta);
    this.dialog_
      .open(CreateUserDialog, dialogConfig)
      .afterClosed()
      .subscribe(result => {
        if (result) {
        }
      });
  }

  // create Namespace  dialog
  showNamespaceCreateDialog(displayName: string, typeMeta: TypeMeta, objectMeta: ObjectMeta): void {
    const dialogConfig = this.getDialogConfig_(displayName, typeMeta, objectMeta);
    this.dialog_
      .open(CreateNamespaceDialog, dialogConfig)
      .afterClosed()
      .subscribe(result => {
        if (result) {
        }
      });

  }

  //Create Quota dialog
  showResourceQuotaCreateDialog(displayName: string, typeMeta: TypeMeta, objectMeta: ObjectMeta): void {
    const dialogConfig = this.getDialogConfig_(displayName, typeMeta, objectMeta);
    this.dialog_
      .open(CreateAssignQuotaDialog, dialogConfig)
      .afterClosed()
      .subscribe(result => {
        if (result) {
        }
      });
  }

  // create Role dialog
  showRoleCreateDialog(displayName: string, typeMeta: TypeMeta, objectMeta: ObjectMeta): void {
    const dialogConfig = this.getDialogConfig_(displayName, typeMeta, objectMeta);
    this.dialog_
      .open(CreateRoleDialog, dialogConfig)
      .afterClosed()
      .subscribe(result => {
        if (result) {
        }
      });
  }

  // create Clusterrole dialog
  showClusterroleCreateDialog(displayName: string, typeMeta: TypeMeta, objectMeta: ObjectMeta): void {
    const dialogConfig = this.getDialogConfig_(displayName, typeMeta, objectMeta);
    this.dialog_
      .open(CreateClusterroleDialog, dialogConfig)
      .afterClosed()
      .subscribe(result => {
        if (result) {
        }
      });
  }

  showDeleteDialog(displayName: string, typeMeta: TypeMeta, objectMeta: ObjectMeta): void {
    const dialogConfig = this.getDialogConfig_(displayName, typeMeta, objectMeta);
    this.dialog_
      .open(DeleteResourceDialog, dialogConfig)
      .afterClosed()
      .subscribe(doDelete => {
        if (doDelete) {
          const url = RawResource.getUrl(this.tenant_.current(), typeMeta, objectMeta);
          this.http_
            .delete(url)
            .subscribe(() => this.onDelete.emit(true), this.handleErrorResponse_.bind(this));
        }
      });
  }

  showEditDialog(displayName: string, typeMeta: TypeMeta, objectMeta: ObjectMeta): void {
    const dialogConfig = this.getDialogConfig_(displayName, typeMeta, objectMeta);
    this.dialog_
      .open(EditResourceDialog, dialogConfig)
      .afterClosed()
      .subscribe(result => {
        if (result) {
          const url = RawResource.getUrl(this.tenant_.current(), typeMeta, objectMeta);
          this.http_
            .put(url, JSON.parse(result), {headers: this.getHttpHeaders_()})
            .subscribe(() => this.onEdit.emit(true), this.handleErrorResponse_.bind(this));
        }
      });
  }

  showScaleDialog(displayName: string, typeMeta: TypeMeta, objectMeta: ObjectMeta): void {
    const dialogConfig = this.getDialogConfig_(displayName, typeMeta, objectMeta);
    const current = this.tenant_.current();
    this.dialog_
      .open(ScaleResourceDialog, dialogConfig)
      .afterClosed()
      .subscribe(result => {
        if (Number.isInteger(result)) {
          const url =
            'api/v1' +
            (current ? `/tenants/${current}` : '') +
            `/scale/${typeMeta.kind}/${objectMeta.namespace}/${objectMeta.name}/`;
          this.http_
            .put(url, result, {
              params: {
                scaleBy: result,
              },
            })
            .subscribe(() => this.onScale.emit(true), this.handleErrorResponse_.bind(this));
        }
      });
  }

  showTriggerDialog(displayName: string, typeMeta: TypeMeta, objectMeta: ObjectMeta): void {
    const dialogConfig = this.getDialogConfig_(displayName, typeMeta, objectMeta);
    this.dialog_
      .open(TriggerResourceDialog, dialogConfig)
      .afterClosed()
      .subscribe(result => {
        if (result) {
          const url = `api/v1/cronjob/${objectMeta.namespace}/${objectMeta.name}/trigger`;
          this.http_
            .put(url, {})
            .subscribe(() => this.onTrigger.emit(true), this.handleErrorResponse_.bind(this));
        }
      });
  }

  getDialogConfig_(
    displayName: string,
    typeMeta: TypeMeta,
    objectMeta: ObjectMeta,
  ): MatDialogConfig<ResourceMeta> {
    return {width: '630px', data: {displayName, typeMeta, objectMeta}};
  }

  handleErrorResponse_(err: HttpErrorResponse): void {
    if (err) {
      const alertDialogConfig: MatDialogConfig<AlertDialogConfig> = {
        width: '630px',
        data: {
          title: err.statusText === 'OK' ? 'Internal server error' : err.statusText,
          message: err.error || 'Could not perform the operation.',
          confirmLabel: 'OK',
        },
      };
      this.dialog_.open(AlertDialog, alertDialogConfig);
    }
  }

  getHttpHeaders_(): HttpHeaders {
    const headers = new HttpHeaders();
    headers.set('Content-Type', 'application/json');
    headers.set('Accept', 'application/json');
    return headers;
  }
}
