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

// tenat dialog
import { CreateTenantDialog } from './../../dialogs/createTenant/dialog';
// user dialog
import {CreateUserDialog} from './../../dialogs/createUser/dialog';
// namespace dialog
import {CreateNamespaceDialog} from '../../dialogs/createNamespace/dialog'; // namespace dialog
// role dialog
import {CreateRoleDialog} from '../../dialogs/createRole/dialog'; // role dialog
// clusterrole dialog
import {CreateClusterroleDialog} from '../../dialogs/createClusterrole/dialog';
import { assignQuotaDialog } from '../../dialogs/assignQuota/dialog';
import {ResourceMeta} from './actionbar';
import {TenantService} from './tenant';
import {CreateNodeDialog} from "../../dialogs/createNode/dialog";

@Injectable()
export class VerberService {
  onCreate = new EventEmitter<boolean>();
  onCreateTenant = new EventEmitter<boolean>(); //added
  onCreateNode = new EventEmitter<boolean>(); //added
  onDelete = new EventEmitter<boolean>();
  onEdit = new EventEmitter<boolean>();
  onScale = new EventEmitter<boolean>();
  onTrigger = new EventEmitter<boolean>();
  onCreateQuota = new EventEmitter<boolean>();

  constructor(
    private readonly dialog_: MatDialog,
    private readonly http_: HttpClient,
    private tenant_: TenantService,

  ) {}

  // create tenant
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
  // create user
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

  // create Namespace
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

  //Create Quota
  showResourceQuotaCreateDialog(displayName: string, typeMeta: TypeMeta, objectMeta: ObjectMeta): void {
    const dialogConfig = this.getDialogConfig_(displayName, typeMeta, objectMeta);
    this.dialog_
      .open(assignQuotaDialog, dialogConfig)
      .afterClosed()
      .subscribe(result => {
        if (result) {
        }
      });
  }

  // create Role
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
  // create Clusterrole
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

  // create node
  showNodeCreateDialog(displayName: string, typeMeta: TypeMeta, objectMeta: ObjectMeta): void {
    const dialogConfig = this.getDialogConfig_(displayName, typeMeta, objectMeta);
    this.dialog_
      .open(CreateNodeDialog, dialogConfig)
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
