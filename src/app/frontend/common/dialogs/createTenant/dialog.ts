// Copyright 2020 Authors of Arktos - file modified.
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

import {Component, OnInit, Inject} from '@angular/core';
import {MatDialog} from '@angular/material/dialog';
import {HttpClient, HttpHeaders} from '@angular/common/http';
import {MAT_DIALOG_DATA, MatDialogRef} from '@angular/material';
import {AbstractControl, Validators,FormBuilder} from '@angular/forms';
import { FormGroup } from '@angular/forms';
import {CONFIG} from "../../../index.config";
import {CsrfTokenService} from "../../services/global/csrftoken";
// @ts-ignore
import Swal from "sweetalert2/dist/sweetalert2.js";

export interface CreateTenantDialogMeta {
  tenants: string[];
}

@Component({
  selector: 'kd-create-tenant-dialog',
  templateUrl: 'template.html',
})

export class CreateTenantDialog implements OnInit {
  form1: FormGroup;
  private readonly config_ = CONFIG;

  //validation
  tenantMaxLength = 24;
  tenantPattern: RegExp = new RegExp('^[a-z0-9]([-a-z0-9]*[a-z0-9])?$');

  usernameMaxLength = 24;
  usernamePattern: RegExp = new RegExp('^[a-z0-9]([-a-z-0-9]*[a-z0-9])?$');

  passwordMaxLength = 20;
  passwordPattern: RegExp = new RegExp('^[a-z\\A-Z\\0-9\\d_@.#$=!%^~)(\\]:\\*;\\?\\/\\,}{\'\\|<>\\[&\\+-]*$');

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
        username :[
          '',
          Validators.compose([
            Validators.required,
            Validators.maxLength(this.usernameMaxLength),
            Validators.pattern(this.usernamePattern),
          ]),
        ],
        password :[
          '',
          Validators.compose([
            Validators.required,
            Validators.maxLength(this.passwordMaxLength),
            Validators.pattern(this.passwordPattern),
          ]),
        ],
      },
    );
  }

  get tenant(): AbstractControl {
    return this.form1.get('tenant');
  }
  get username(): AbstractControl {
    return this.form1.get('username')
  }
  get password(): AbstractControl{
    return this.form1.get('password')
  }

  //creating tenant-admin with tenant
  createTenant(): void {
    if (!this.form1.valid) return;
    const tenantSpec= {name: this.tenant.value,username: this.username.value,password: this.password.value};
    const tokenPromise = this.csrfToken_.getTokenForAction(this.tenant.value,'tenant');
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
            Swal.fire({
              type: 'success',
              title: this.tenant.value,
              text: 'tenant successfully created!',
              imageUrl: '/assets/images/tick-circle.svg',
            })
            this.dialogRef.close(this.tenant.value);
          },
          (error:any) => {
            if (error) {
              Swal.fire({
                type:'error',
                title: this.tenant.value,
                text: 'tenant already exists!',
                imageUrl: '/assets/images/close-circle.svg',
              })
            }
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
