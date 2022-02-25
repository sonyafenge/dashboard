import {Component, OnInit, Inject} from '@angular/core';
import {MatDialog} from '@angular/material/dialog';
import {HttpClient, HttpHeaders} from '@angular/common/http';
import {MAT_DIALOG_DATA, MatDialogRef} from '@angular/material';
import {AbstractControl, Validators,FormBuilder} from '@angular/forms';
import {FormGroup} from '@angular/forms';
import {CONFIG} from "../../../index.config";
import {CsrfTokenService} from "../../services/global/csrftoken";
export interface CreateTenantDialogMeta {
  tenants : string [];
  StorageClusterId: string []
}
@Component({
  selector: 'kd-create-tenant-dialog',
  templateUrl: 'template.html',
})
export class CreateTenantDialog implements OnInit {
  form1: FormGroup;
  tenants: any;
  private readonly config_ = CONFIG;
  tenantMaxLength = 10;
  storageidMaxLength =2;
  tenantPattern: RegExp = new RegExp('^[a-z0-9]([-a-z0-9]*[a-z0-9])?$');
  storageidPattern: RegExp = new RegExp('^[0-9]*$');
  constructor(
    public dialogRef: MatDialogRef<CreateTenantDialog>,
    @Inject(MAT_DIALOG_DATA) public data: CreateTenantDialogMeta,
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
            Validators.required,
            Validators.maxLength(this.tenantMaxLength),
            Validators.pattern(this.tenantPattern),
          ]),
        ],
        StorageClusterId :[
          '',
          Validators.compose([
            Validators.required,
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
  get StorageClusterId(): AbstractControl {
    return this.form1.get('StorageClusterId')
  }
  createTenant(): void{
    if (!this.form1.valid) return;
    const tenantSpec= {name: this.tenant.value,StorageClusterId: this.StorageClusterId.value};
    console.log(this.tenant.value)
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
          () => {
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
