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

import {Component, OnInit, Inject,NgZone} from '@angular/core';
import {MatDialog} from '@angular/material/dialog';
import {HttpClient, HttpHeaders} from '@angular/common/http';
import {ActivatedRoute} from '@angular/router';
import {MAT_DIALOG_DATA, MatDialogRef} from '@angular/material';
import {AbstractControl, Validators,FormBuilder} from '@angular/forms';
import {FormGroup} from '@angular/forms';
import {CONFIG} from "../../../index.config";
import {CsrfTokenService} from "../../services/global/csrftoken";
import {NamespacedResourceService} from '../../services/resource/resource';
import {TenantService} from "../../services/global/tenant";
import {
  SecretDetail,
  Role,
  RoleList,
  NamespaceList,
  Namespace, TenantList, Tenant,
} from '../../../typings/backendapi';
import {validateUniqueName} from "../../../create/from/form/validator/uniquename.validator";
import {TenantDetail} from "@api/backendapi";
import {NamespaceService} from "../../services/global/namespace";
// @ts-ignore
import Swal from "sweetalert2/dist/sweetalert2.js";

export interface UserToken {
  token: string;
}

export interface CreateUserDialogMeta {
  tenants: string;
}

@Component({
  selector: 'kd-create-tenant-dialog',
  templateUrl: 'template.html',
})

export class CreateUserDialog implements OnInit {
  form1: FormGroup;
  tenants: string[];
  secrets: string[];
  roles: string[];
  namespaces: string[];
  namespaceUsed = "default"
  adminroleUsed = "admin-role";
  tenantUsed = ""
  apiGroups : string [] = ["*"]
  resources : string [] = ["*"]
  verbs :string [] = ["*"]
  serviceAccountCreated:any[] = [];
  secretDetails:any[] = [];
  selected = '';
  selectednamespace = '';
  userType = '';
  message = false;
  success: string;

  private readonly config_ = CONFIG

  //validation
  usernameMaxLength = 24;
  usernamePattern: RegExp = new RegExp('^[a-z0-9]([-a-z-0-9]*[a-z0-9])?$');

  passwordMaxLength = 20;
  passwordPattern: RegExp = new RegExp('^[a-z\\A-Z\\0-9\\d_@.#$=!%^~)(\\]:\\*;\\?\\/\\,}{\'\\|<>\\[&\\+-]*$');

  tenantMaxLength = 24;
  tenantPattern: RegExp = new RegExp('^[a-z0-9]([-a-z0-9]*[a-z0-9])?$');

  usertypeMaxLength = 24;
  usertypePattern: RegExp = new RegExp('^[a-z0-9]([-a-z0-9]*[a-z0-9])?$');

  secret: SecretDetail;
  secretName =""

  private currentTenant: string;
  private tenant_: string;

  constructor(

    private readonly secret_: NamespacedResourceService<SecretDetail>,
    public dialogRef: MatDialogRef<CreateUserDialog>,
    @Inject(MAT_DIALOG_DATA) public data: CreateUserDialogMeta,
    private readonly http_: HttpClient,
    private readonly csrfToken_: CsrfTokenService,
    private readonly matDialog_: MatDialog,
    private readonly fb_: FormBuilder,
    private readonly tenantService_ : TenantService,
    private readonly dialog_: MatDialog,
    private readonly route_: ActivatedRoute,
    private readonly ngZone_: NgZone,
    private readonly tenants_: NamespacedResourceService<TenantDetail>,
    private readonly namespace_: NamespaceService,
  ) {}

  ngOnInit(): void {

    this.passwordTimeout();
    this.currentTenant = this.tenants_['tenant_']['currentTenant_']
    this.form1 = this.fb_.group({
        role: [this.route_.snapshot.params.role || '', Validators.required],
        namespace: [this.route_.snapshot.params.namespace || '', Validators.required],
        usertype: [
          '',
          Validators.compose([
            Validators.maxLength(this.usertypeMaxLength),
            Validators.pattern(this.usertypePattern),
          ]),
        ],
        tenant: [
          '',
          Validators.compose([
            Validators.maxLength(this.tenantMaxLength),
            Validators.pattern(this.tenantPattern),
          ]),
        ],
        username: [
          '',
          Validators.compose([
            Validators.maxLength(this.usernameMaxLength),
            Validators.pattern(this.usernamePattern),
          ]),
        ],
        password: [
          '',
          Validators.compose([
            Validators.maxLength(this.passwordMaxLength),
            Validators.pattern(this.passwordPattern),
          ]),
        ],
      },
    );
    this.namespace.valueChanges.subscribe((namespace: string) => {
      if (this.name !== null) {
        this.name.clearAsyncValidators();
        this.name.setAsyncValidators(validateUniqueName(this.http_, namespace));
        this.name.updateValueAndValidity();
      }
    });
    this.http_.get(`api/v1/tenants/${this.currentTenant}/namespace`).subscribe((result: NamespaceList) => {
      this.namespaces = result.namespaces.map((namespace: Namespace) => namespace.objectMeta.name);
      this.namespace.patchValue(
        !this.namespace_.areMultipleNamespacesSelected()
          ? this.route_.snapshot.params.namespace || this.namespaces
          : this.namespaces,
      );
    });
    this.tenant.valueChanges.subscribe((tenant: string) => {
      if (this.name !== null) {
        this.name.clearAsyncValidators();
        this.name.setAsyncValidators(validateUniqueName(this.http_, tenant));
        this.name.updateValueAndValidity();
      }
    });
    this.http_.get(`api/v1/tenants/${this.currentTenant}/tenant`).subscribe((result: TenantList) => {
      this.tenants = result.tenants.map((tenant: Tenant) => tenant.objectMeta.name);
      console.log("list: ", this.tenants)
      this.tenant.patchValue(
        this.tenant.value,
      );
    });

    this.ngZone_.run(() => {
      const usertype = sessionStorage.getItem('userType');
      this.userType = usertype
    });
  }

  passwordTimeout() {
    this.message = true;
    this.success = 'Password must contain uppercase,lowercase,number,special characters only!'
    setTimeout(()=>{
      this.message = false;
    }, 10000);
  }

  selectUserType(event:any)
  {
    this.selected=event;
  }
  selectNamespace(event:any)
  {
    this.selectednamespace=event;
    this.getRole()
  }

  //getting roles for a particular tenant
  getRole(){
    this.role.valueChanges.subscribe((role: string) => {
      if (this.name !== null) {
        this.name.clearAsyncValidators();
        this.name.setAsyncValidators(validateUniqueName(this.http_, role));
        this.name.updateValueAndValidity();
      }
    });

    this.http_.get(`api/v1/tenants/${this.currentTenant}/role/${this.selectednamespace}`).subscribe((result: RoleList) => {
      this.roles = result.items.map((role: Role) => role.objectMeta.name);
      this.role.patchValue(
        !this.tenantService_.isCurrentSystem()
          ? this.route_.snapshot.params.role || this.roles
          : this.roles,
      );
    });
  }


  get name(): AbstractControl {
    return this.form1.get('name');
  }
  get role(): any {
    return this.form1.get('role');
  }
  get username(): AbstractControl {
    return this.form1.get('username');
  }
  get password(): AbstractControl {
    return this.form1.get('password');
  }
  get usertype(): any {
    return this.form1.get('usertype');
  }
  get namespace(): AbstractControl {
    return this.form1.get('namespace');
  }
  get tenant(): any {
    return this.form1.get('tenant');
  }

  //function to create a user
  createUser() {
    const currentType = sessionStorage.getItem('userType')
    if (this.usertype.value === 'tenant-admin' && currentType === 'cluster-admin') {
      this.tenant_ = this.username.value
    } else if (this.usertype.value === 'tenant-admin' && currentType === 'tenant-admin') {
      this.tenant_ = this.tenant
    } else if (this.usertype.value === 'tenant-user') {
      this.tenant_ = this.tenant
    } else {
      this.tenant_ = 'system'
    }
    if( this.selected == "cluster-admin")
    {
      this.tenant_ = "system"
    } else if(this.selected == "tenant-admin")
    {
      this.tenant_ = this.currentTenant
    } else
    {
      this.tenant_ = this.currentTenant
      this.namespaceUsed = this.selectednamespace
    }

    this.getToken(async (token_:any)=>{
      const userSpec= {name: this.username.value, password:this.password.value, token:token_,namespace:this.namespaceUsed, type:this.usertype.value,tenant:this.tenant_,role:this.role.value};
      if (this.selected === "tenant-user") {
        userSpec.role = this.role.value;
      }
      else {
        userSpec.role = '';
      }

      if (this.currentTenant==='') {
        this.currentTenant=this.tenantService_.current()
      }
      const userTokenPromise = await this.csrfToken_.getTokenForAction(this.currentTenant,'users');
      userTokenPromise.subscribe(csrfToken => {
        return this.http_
          .post<{valid: boolean}>(
            'api/v1/users',
            {...userSpec},
            {
              headers: new HttpHeaders().set(this.config_.csrfHeaderName, csrfToken.token),
            },
          )
          .subscribe(
            () => {
            },
            () => {},
          );
      });
    })
  }

  //function to create a tenant-admin
  createTenantAdmin() {
    const currentType = sessionStorage.getItem('userType')
    {
      const userSpec = {
        name: this.username.value,
        password: this.password.value,
        type: this.usertype.value,
        tenant: this.tenant.value,
      };
      const userTokenPromise = this.csrfToken_.getTokenForAction(this.currentTenant, 'users');
      userTokenPromise.subscribe(csrfToken => {
        return this.http_
          .post<{ valid: boolean }>(
            'api/v1/users',
            {...userSpec},
            {
              headers: new HttpHeaders().set(this.config_.csrfHeaderName, csrfToken.token),
            },
          )
          .subscribe(
            (data: any) => {
              Swal.fire({
                type: 'success',
                title: this.username.value,
                text: 'user successfully created!',
                imageUrl: '/assets/images/tick-circle.svg',
              });
              this.dialogRef.close(this.username.value);
              this.serviceAccountCreated.push(Object.entries(data))
            },
            () =>{
              Swal.fire({
                type: 'error',
                title: this.username.value,
                text: 'user already exists!',
                imageUrl: '/assets/images/close-circle.svg',
              });
            },
          );
      });
    }
  }

  //function to create a service-account
  createServiceAccount() {
    if( this.selected == "cluster-admin")
    {
      this.tenantUsed = "system"
    }else if (this.selected == "tenant-admin")
    {
      this.tenantUsed = "system"
    }else
    {
      this.tenantUsed = this.currentTenant
      this.namespaceUsed = this.selectednamespace
    }

    const serviceAccountSpec= {name: this.username.value,namespace: this.namespaceUsed,tenant: this.tenantUsed};
    const tokenPromise = this.csrfToken_.getTokenForAction(this.tenantUsed,'serviceaccounts');
    tokenPromise.subscribe(csrfToken => {
      return this.http_
        .post<{valid: boolean}>(
          'api/v1/serviceaccounts',
          {...serviceAccountSpec},
          {
            headers: new HttpHeaders().set(this.config_.csrfHeaderName, csrfToken.token),
          },
        )
        .subscribe(
          (data: any) => {
            Swal.fire({
              type: 'success',
              title: this.username.value,
              text: 'user successfully created!',
              imageUrl: '/assets/images/tick-circle.svg',
            });
            this.dialogRef.close(this.username.value);
            this.serviceAccountCreated.push(Object.entries(data))
          },
          () =>{
            Swal.fire({
              type: 'error',
              title: this.username.value,
              text: 'user already exists!',
              imageUrl: '/assets/images/close-circle.svg',
            });
          },
        );
    })
  }

  //function to create a cluster-role binding
  createClusterRoleBinding(): void{
    if( this.usertype.value === "tenant-admin")
    {
      this.adminroleUsed = this.username.value
    }
    const crbSpec= {name: this.username.value,namespace: this.namespaceUsed, subject: { kind: "ServiceAccount", name: this.username.value,  namespace : this.namespaceUsed, apiGroup : ""},role_ref:{kind: "ClusterRole",name: this.adminroleUsed,apiGroup: "rbac.authorization.k8s.io"}};
    const tokenPromise = this.csrfToken_.getTokenForAction('system','clusterrolebinding');
    tokenPromise.subscribe(csrfToken => {
      return this.http_
        .post<{valid: boolean}>(
          'api/v1/clusterrolebinding',
          {...crbSpec},
          {
            headers: new HttpHeaders().set(this.config_.csrfHeaderName, csrfToken.token),
          },
        )
        .subscribe(
          () => {
          },
          () => {},
        );
    })
  }

  //function to create a role binding
  createRoleBinding(): void{
    if(this.selected == "tenant-user"){
      this.tenantUsed = this.currentTenant
      this.namespaceUsed = this.selectednamespace
    }
    const roleBindingsSpec= {name: this.username.value,namespace: this.namespaceUsed,tenant:this.tenantUsed, subject: { kind: "ServiceAccount", name: this.username.value,  namespace : this.namespaceUsed, apiGroup : ""},role_ref:{kind: "Role",name: this.role.value,apiGroup: "rbac.authorization.k8s.io"}};
    const tokenPromise = this.csrfToken_.getTokenForAction(this.tenantUsed,'rolebindings');
    tokenPromise.subscribe(csrfToken => {
      return this.http_
        .post<{valid: boolean}>(
          'api/v1/rolebindings',
          {...roleBindingsSpec},
          {
            headers: new HttpHeaders().set(this.config_.csrfHeaderName, csrfToken.token),
          },
        )
        .subscribe(
          () => {
          },
          () => {},
        );
    })
  }

  getToken(callback: any): any {
    if( this.selected == "cluster-admin")
    {
      this.tenantUsed = "system"
    }else if (this.selected == "tenant-admin")
    {
      this.tenantUsed = "system"
    }else
    {
      this.tenantUsed = this.currentTenant
    }
    const interval = setInterval(() => {
      this.http_.get(`api/v1/tenants/${this.tenantUsed}/secret/`+ this.namespaceUsed ).subscribe((data:any)=>{
        data.secrets.map((elem: any) => {
          if(elem.objectMeta.name.includes(this.username.value + '-token')){
            clearInterval(interval);
            this.http_.get(`api/v1/tenants/${this.tenantUsed}/secret/` + this.namespaceUsed + "/" + elem.objectMeta.name).subscribe((data: any) => {
              callback(this.decode(data.data.token));
            })
          }
        });
      });
    }, 3000);
  }

  //main user creating function
  createTenantUser() {
    if(this.usertype.value === "tenant-user"){
      this.createServiceAccount()
      this.createRoleBinding()
      this.createUser()
    } else if(this.usertype.value === "cluster-admin") {
      this.createServiceAccount()
      this.createClusterRoleBinding()
      this.createUser()
    }
    else{
      this.createTenantAdmin()
    }
  }

  decode(s: string): string {
    return atob(s);
  }

  isCreateDisabled(): boolean {
    return !this.username.value || !this.password.value || !this.usertype.value;
  }

  isDisabled(): boolean {
    return this.data.tenants.indexOf(this.tenant.value) >= 0;
  }

  cancel(): void {
    this.dialogRef.close();
  }
}
