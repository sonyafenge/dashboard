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

import {Component,Inject,OnInit} from '@angular/core';
import {MatDialog} from '@angular/material/dialog';
import {HttpClient, HttpHeaders} from '@angular/common/http';
import {MAT_DIALOG_DATA, MatDialogRef} from '@angular/material';
import {AbstractControl, Validators,FormBuilder} from '@angular/forms';
import { FormGroup } from '@angular/forms';
import {CONFIG} from "../../../index.config";
import {CsrfTokenService} from "../../services/global/csrftoken";

// @ts-ignore
import Swal from "sweetalert2/dist/sweetalert2.js";


export interface CreateClusterroleDialogMeta {
  name: string;
  apiGroups: string []
  resources: string[]
  verbs: string[]
}

@Component({
  selector: 'kd-create-clusterrole-dialog',
  templateUrl: 'template.html',
})

export class CreateClusterroleDialog implements OnInit {
  form1: FormGroup;
  private readonly config_ = CONFIG;
  name: string
  apigroup: string[]
  resource: string[]
  verb : string[]

  // validation
  clusterroleMaxLength = 24;
  clusterrolePattern: RegExp = new RegExp('^[a-z0-9]([-a-z0-9]*[a-z0-9])?$');

  apigroupsMaxLength = 63;
  apigroupsPattern:  RegExp = new RegExp('^[a-z\\a-z\\d_@.#$=!%^)(\\]:\\*;\\?\\/\\,}{\'\\|<>\\[&\\+-]*$');

  resourceMaxLength = 63;
  resourcePattern: RegExp = new RegExp('^^[a-z\\a-z\\d_@.#$=!%^)(\\]:\\*;\\?\\/\\,}{\'\\|<>\\[&\\+-]*$');

  verbsMaxLength = 63;
  verbsPattern: RegExp = new RegExp('^^[a-z\\a-z\\d_@.#$=!%^)(\\]:\\*;\\?\\/\\,}{\'\\|<>\\[&\\+-]*$');

  constructor(
    public dialogRef: MatDialogRef<CreateClusterroleDialog>,
    @Inject(MAT_DIALOG_DATA) public data: CreateClusterroleDialogMeta,
    private readonly http_: HttpClient,
    private readonly csrfToken_: CsrfTokenService,
    private readonly matDialog_: MatDialog,
    private readonly fb_: FormBuilder,
  ) {}

  ngOnInit(): void {
    this.form1 = this.fb_.group({
      clusterrole: [
        '',
        Validators.compose([
          Validators.maxLength(this.clusterroleMaxLength),
          Validators.pattern(this.clusterrolePattern),
        ]),
      ],
      apigroups: [
        '',
        Validators.compose([
          Validators.maxLength(this.apigroupsMaxLength),
          Validators.pattern(this.apigroupsPattern),
        ]),
      ],
      resources: [
        '',
        Validators.compose([
          Validators.maxLength(this.resourceMaxLength),
          Validators.pattern(this.resourcePattern),
        ]),
      ],
      verbs: [
        '',
        Validators.compose([
          Validators.maxLength(this.verbsMaxLength),
          Validators.pattern(this.verbsPattern),
        ]),
      ],
    });
  }

  get clusterrole(): AbstractControl {
    return this.form1.get('clusterrole');
  }
  get apigroups(): AbstractControl {
    return this.form1.get('apigroups');
  }
  get verbs(): AbstractControl {
    return this.form1.get('verbs');
  }
  get resources(): AbstractControl {
    return this.form1.get('resources');
  }

  // To create clusterole specific tenant
  createClusterrole(): void {
    if (!this.form1.valid) return;
    this.apigroup = this.apigroups.value.split(',')
    this.resource = this.resources.value.split(',')
    this.verb = this.verbs.value.split(',')

    const clusterroleSpec= {name: this.clusterrole.value,apiGroups: this.apigroup,verbs: this.verb,resources: this.resource};
    const tokenPromise = this.csrfToken_.getTokenForAction('system','clusterrole');
    tokenPromise.subscribe(csrfToken => {
      return this.http_
        .post<{valid: boolean}>(
          'api/v1/clusterrole',
          {...clusterroleSpec},
          {
            headers: new HttpHeaders().set(this.config_.csrfHeaderName, csrfToken.token),
          },
        )
        .subscribe(
          () => {
            Swal.fire({
              type: 'success',
              title: this.clusterrole.value,
              text: 'clusterrole successfully created!',
              imageUrl: '/assets/images/tick-circle.svg',
            })
            this.dialogRef.close(this.clusterrole.value);

          },
          (error:any) => {
            if (error) {
              Swal.fire({
                type:'error',
                title: this.clusterrole.value,
                text: 'clusterrole already exists!',
                imageUrl: '/assets/images/close-circle.svg',
              })
            }
          },
        );
    });
  }

  isDisabled(): boolean {
    return this.data.name.indexOf(this.clusterrole.value) >= 0;
  }
  cancel(): void {
    this.dialogRef.close();
  }

}
