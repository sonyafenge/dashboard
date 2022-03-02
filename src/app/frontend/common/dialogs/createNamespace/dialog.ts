import {HttpClient, HttpHeaders} from '@angular/common/http';
import {Component, Inject, OnInit} from '@angular/core';
import {AbstractControl, FormBuilder, FormGroup, Validators} from '@angular/forms';
import {MAT_DIALOG_DATA, MatDialog, MatDialogRef} from '@angular/material';
import {CsrfTokenService} from '../../../common/services/global/csrftoken';
import {CONFIG} from '../../../index.config';
import {NamespacedResourceService} from "../../services/resource/resource";
import {TenantDetail} from "@api/backendapi";

// @ts-ignore
import Swal from "sweetalert2/dist/sweetalert2.js";

export interface CreateNamespaceDialogMeta {
  namespaces: string[];
}

@Component({
  selector: 'kd-create-namespace-dialog',
  templateUrl: 'template.html',
})
export class CreateNamespaceDialog implements OnInit {
  form1: FormGroup;
  private readonly config_ = CONFIG;
  private currentTenant:string

  namespaceMaxLength = 63;
  namespacePattern: RegExp = new RegExp('^[a-z0-9]([-a-z0-9]*[a-z0-9])?$');

  constructor(
    public dialogRef: MatDialogRef<CreateNamespaceDialog>,
    @Inject(MAT_DIALOG_DATA) public data: CreateNamespaceDialogMeta,
    private readonly http_: HttpClient,
    private readonly csrfToken_: CsrfTokenService,
    private readonly matDialog_: MatDialog,
    private readonly fb_: FormBuilder,
    private readonly tenant_: NamespacedResourceService<TenantDetail>,
  ) {}

  ngOnInit(): void {
    this.currentTenant = this.tenant_['tenant_']['currentTenant_']

    this.form1 = this.fb_.group({
      namespace: [
        '',
        Validators.compose([
          Validators.maxLength(this.namespaceMaxLength),
          Validators.pattern(this.namespacePattern),
        ]),
      ],
    });
  }

  get namespace(): AbstractControl {
    return this.form1.get('namespace');
  }

  createNamespace(): void {
    if (!this.form1.valid) return;
    const namespaceSpec = {name: this.namespace.value, tenant: this.currentTenant};
    const tokenPromise = this.csrfToken_.getTokenForAction('namespace');
    tokenPromise.subscribe(csrfToken => {
      return this.http_
        .post<{valid: boolean}>(
          'api/v1/namespace',
          {...namespaceSpec},
          {
            headers: new HttpHeaders().set(this.config_.csrfHeaderName, csrfToken.token),
          },
        )
        .subscribe(
          () => {
            Swal.fire({
              type: 'success',
              title: this.namespace.value,
              text: 'namespace successfully created!',
              imageUrl: '/assets/images/tick-circle.svg',
            })
            this.dialogRef.close(this.namespace.value);
          },
          (error:any) => {
            if (error) {
              Swal.fire({
                type:'error',
                title: this.namespace.value,
                text: 'namespace already exists!',
                imageUrl: '/assets/images/close-circle.svg',
              })
            }
          },
        );
    });
  }

  isDisabled(): boolean {
    return this.data.namespaces.indexOf(this.namespace.value) >= 0;
  }

  cancel(): void {
    this.dialogRef.close();
  }

}
