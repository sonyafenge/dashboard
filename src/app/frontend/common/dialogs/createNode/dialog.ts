import {Component, OnInit, Inject} from '@angular/core';
import {MatDialog} from '@angular/material/dialog';
import {HttpClient, HttpHeaders} from '@angular/common/http';
import {MAT_DIALOG_DATA, MatDialogRef} from '@angular/material';
import {AbstractControl, Validators,FormBuilder} from '@angular/forms';
import { FormGroup } from '@angular/forms';
import {CONFIG} from "../../../index.config";
import {CsrfTokenService} from "../../services/global/csrftoken";
import {AlertDialog, AlertDialogConfig} from "../alert/dialog";

export interface CreateNodeDialogMeta {
  tenants: string[];
  StorageClusterId: string []
  data : string[]
}
@Component({
  selector: 'kd-delete-resource-dialog',
  templateUrl: 'template.html',
})

export class CreateNodeDialog implements OnInit {
  form1: FormGroup;

  private readonly config_ = CONFIG;

  tenantMaxLength = 63;
  storageidMaxLength =24;

  tenantPattern: RegExp = new RegExp('^[a-z0-9]([-a-z0-9]*[a-z0-9])?$');
  storageidPattern: RegExp = new RegExp('^[0-9]$');


  constructor(
    public dialogRef: MatDialogRef<CreateNodeDialog>,
    @Inject(MAT_DIALOG_DATA) public data: CreateNodeDialogMeta,
    private readonly http_: HttpClient,
    private readonly csrfToken_: CsrfTokenService,
    private readonly matDialog_: MatDialog,
    private readonly fb_: FormBuilder,
  ) {}

  ngOnInit(): void {
    this.form1 = this.fb_.group({
        tenant: [
          '',
          Validators.compose([
            Validators.maxLength(this.tenantMaxLength),
            Validators.pattern(this.tenantPattern),
          ]),
        ],
        StorageClusterId :[
          '',
          Validators.compose([
            Validators.maxLength(this.storageidMaxLength),
            Validators.pattern(this.storageidPattern),
          ]),
        ],
      }
    );

  }

  get tenant(): AbstractControl {
    return this.form1.get('tenant');
  }

  createNode(): void {
    if (!this.form1.valid) return;
    const tenantSpec= {name: this.tenant.value,StorageClusterId: this.tenant.value};
    const tokenPromise = this.csrfToken_.getTokenForAction('tenant');
    tokenPromise.subscribe(csrfToken => {
      return this.http_
        .post<{valid: boolean}>(
          'api/v1/tenant',
          {...tenantSpec},
          {
            headers: new HttpHeaders().set(this.config_.csrfHeaderName, csrfToken.token),
          },
        )
        .subscribe(
          () => {
            this.dialogRef.close(this.tenant.value);
          },
          (error: any) => {
            this.dialogRef.close();
            const configData: AlertDialogConfig = {
              title: 'Error creating node',
              message: error.data,
              confirmLabel: 'OK',
            };
            this.matDialog_.open(AlertDialog, {data: configData});
          },
        );
    });
  }

  isDisabled(): boolean {
    return this.data.tenants.indexOf(this.tenant.value) >= 0;
  }

  cancel(): void {
    this.dialogRef.close();
  }

}




