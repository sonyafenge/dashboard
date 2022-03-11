// Copyright 2017 The Kubernetes Authors.
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

  // validation
  nodeMaxLength = 63;
  nodePattern: RegExp = new RegExp('^[a-z0-9]([-a-z0-9]*[a-z0-9])?$');


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
        node: [
          '',
          Validators.compose([
            Validators.maxLength(this.nodeMaxLength),
            Validators.pattern(this.nodePattern),
          ]),
        ],
      }
    );
  }

  get node(): AbstractControl {
    return this.form1.get('node');
  }

  createNode(): void {
    if (!this.form1.valid) return;
    const tenantSpec= {name: this.node.value,StorageClusterId: this.node.value};
    const tokenPromise = this.csrfToken_.getTokenForAction(this.node.value,'node');
    tokenPromise.subscribe(csrfToken => {
      return this.http_
        .post<{valid: boolean}>(
          'api/v1/node',
          {...tenantSpec},
          {
            headers: new HttpHeaders().set(this.config_.csrfHeaderName, csrfToken.token),
          },
        )
        .subscribe(
          () => {
            this.dialogRef.close(this.node.value);
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
    return this.data.tenants.indexOf(this.node.value) >= 0;
  }

  cancel(): void {
    this.dialogRef.close();
  }

}




