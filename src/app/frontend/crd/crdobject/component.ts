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

import {Component, ElementRef, OnDestroy, OnInit, Renderer2, ViewChild} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {MatButtonToggleGroup} from '@angular/material';
import {HttpClient} from '@angular/common/http';
import {dump as toYaml} from 'js-yaml';
import {Subscription} from 'rxjs';
import {CRDObjectDetail} from '@api/backendapi';
import {highlightAuto} from 'highlight.js';
import {ActionbarService, ResourceMeta} from '../../common/services/global/actionbar';
import {NamespacedResourceService} from '../../common/services/resource/resource';
import {EndpointManager, Resource} from '../../common/services/resource/endpoint';
import {NotificationsService} from '../../common/services/global/notifications';
import {RawResource} from '../../common/resources/rawresource';
import {TenantService} from 'common/services/global/tenant';

enum Modes {
  JSON = 'json',
  YAML = 'yaml',
}

@Component({
  selector: 'kd-crd-object-detail',
  templateUrl: './template.html',
})

export class CRDObjectDetailComponent implements OnInit, OnDestroy {
  @ViewChild('group', {static: true}) buttonToggleGroup: MatButtonToggleGroup;
  @ViewChild('code', {static: true}) codeRef: ElementRef;

  private objectSubscription_: Subscription;
  private readonly endpoint_ = EndpointManager.resource(Resource.crd, false, true);
  object: CRDObjectDetail;
  modes = Modes;
  isInitialized = false;
  selectedMode = Modes.YAML;
  objectRaw: {[s: string]: string} = {};
  eventListEndpoint: string;
  tenantName: string;
  partitionName: string;
  partition: string;

  constructor(
    private readonly object_: NamespacedResourceService<CRDObjectDetail>,
    private readonly actionbar_: ActionbarService,
    private readonly activatedRoute_: ActivatedRoute,
    private readonly notifications_: NotificationsService,
    private readonly http_: HttpClient,
    private readonly renderer_: Renderer2,
    private readonly tenant_: TenantService,
  ) {
    this.tenantName = this.tenant_.current() === 'system' && sessionStorage.getItem('crdPartition') !== null ?
      this.tenant_.current() : this.tenant_.resourceTenant()
    this.partitionName = this.tenantName === 'system' ? sessionStorage.getItem('crdPartition') : ''
    this.partition = this.tenantName === 'system' ? 'partition/' + sessionStorage.getItem('crdPartition') + '/' : ''
  }

  ngOnInit(): void {

    const {crdNamespace, crdName, objectName} = this.activatedRoute_.snapshot.params;
    this.eventListEndpoint = this.endpoint_.child(
      `${crdName}/${objectName}`,
      Resource.event,
      crdNamespace,
    );

    let endpoint = ''
    if (sessionStorage.getItem('userType') === 'cluster-admin' && crdNamespace === undefined) {
      endpoint = `api/v1/cluster/${this.partitionName}/tenants/${this.tenantName}/crd/${crdName}/${objectName}`
    } else if (sessionStorage.getItem('userType') === 'cluster-admin' && crdNamespace !== undefined) {
      endpoint = `api/v1/${this.partition}tenants/${this.tenantName}/crd/${crdNamespace}/${crdName}/${objectName}`
    } else {
      endpoint = this.endpoint_.child(crdName, objectName, crdNamespace, this.tenantName)
    }

    this.objectSubscription_ = this.object_
      .get(endpoint)
      .subscribe((d: CRDObjectDetail) => {
        this.object = d;
        this.notifications_.pushErrors(d.errors);
        this.actionbar_.onInit.emit(new ResourceMeta(d.typeMeta.kind, d.objectMeta, d.typeMeta));
        this.isInitialized = true;

        // Get raw resource
        let url = RawResource.getUrl(
          this.tenantName,
          this.object.typeMeta,
          this.object.objectMeta,
          this.partitionName,
        );
        if (crdNamespace === undefined) {
          url = url.replace('/partition/', '/cluster/')
        }
        this.http_
          .get(url)
          .toPromise()
          .then(response => {
            this.objectRaw = {
              json: highlightAuto(`${this.toRawJSON(response)}`, ['json']).value,
              yaml: highlightAuto(`${toYaml(response)}`, ['yaml']).value,
            };
            this.updateText();
          });
      });

    this.buttonToggleGroup.valueChange.subscribe((selectedMode: Modes) => {
      this.selectedMode = selectedMode;

      if (Object.keys(this.objectRaw).length > 0) {
        this.updateText();
      }
    });
  }

  ngOnDestroy(): void {
    this.objectSubscription_.unsubscribe();
    this.actionbar_.onDetailsLeave.emit();
  }

  private updateText(): void {
    this.renderer_.setProperty(
      this.codeRef.nativeElement,
      'innerHTML',
      this.objectRaw[this.selectedMode],
    );
  }

  private toRawJSON(object: {}): string {
    return JSON.stringify(object, null, 2);
  }
}
