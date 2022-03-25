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

import {HttpParams} from '@angular/common/http';
import {Component, ComponentFactoryResolver, Input, OnDestroy} from '@angular/core';
import {Event, Pod, PodList} from '@api/backendapi';
import {Observable} from 'rxjs/Observable';

import {ResourceListWithStatuses} from '../../../resources/list';
import {NotificationsService} from '../../../services/global/notifications';
import {EndpointManager, Resource} from '../../../services/resource/endpoint';
import {NamespacedResourceService} from '../../../services/resource/resource';
import {MenuComponent} from '../../list/column/menu/component';
import {ListGroupIdentifier, ListIdentifier} from '../groupids';
import {ActivatedRoute} from "@angular/router";
import {TenantService} from "../../../services/global/tenant";

@Component({selector: 'kd-pod-list', templateUrl: './template.html'})
export class PodListComponent extends ResourceListWithStatuses<PodList, Pod> {
  @Input() endpoint = EndpointManager.resource(Resource.pod, true, true).list();

  tenantName: string;

  constructor(
    private readonly podList: NamespacedResourceService<PodList>,
    private readonly activatedRoute_: ActivatedRoute,
    private readonly tenant_: TenantService,
    resolver: ComponentFactoryResolver,
    notifications: NotificationsService,
  ) {
    super('pod', notifications, resolver);
    this.id = ListIdentifier.pod;
    this.groupId = ListGroupIdentifier.workloads;

    // Register status icon handlers
    this.registerBinding(this.icon.checkCircle, 'kd-success', this.isInSuccessState);
    this.registerBinding(this.icon.timelapse, 'kd-muted', this.isInPendingState);
    this.registerBinding(this.icon.error, 'kd-error', this.isInErrorState);

    // Register action columns.
    this.registerActionColumn<MenuComponent>('menu', MenuComponent);

    // Register dynamic columns.
    this.registerDynamicColumn('namespace', 'name', this.shouldShowNamespaceColumn_.bind(this));

    this.tenantName = this.activatedRoute_.snapshot.params.resourceName === undefined ?
      this.tenant_.current() : this.tenant_.resourceTenant()
    sessionStorage.setItem('podTenant', this.tenantName);
  }

  getResourceObservable(params?: HttpParams): Observable<PodList> {
    this.tenantName = this.tenantName === '' ? this.tenant_.current() : this.tenantName
    const partition = this.tenantName === 'system' ? 'partition/' + this.tenant_.tenantPartition() + '/' : ''
    let endpoint = ''
    if (sessionStorage.getItem('userType') === 'cluster-admin' && !(this.endpoint.includes('/replicaset/') || this.endpoint.includes('/service/'))) {
      endpoint = `api/v1/${partition}tenants/${this.tenantName}/pod`
    } else {
      endpoint = this.endpoint
    }
    return this.podList.get(endpoint, undefined, undefined, params, this.tenantName);
  }

  map(podList: PodList): Pod[] {
    return podList.pods;
  }

  isInErrorState(resource: Pod): boolean {
    return resource.podStatus.status === 'Failed';
  }

  isInPendingState(resource: Pod): boolean {
    return resource.podStatus.status === 'Pending';
  }

  isInSuccessState(resource: Pod): boolean {
    return resource.podStatus.status === 'Succeeded' || resource.podStatus.status === 'Running';
  }

  getDisplayColumns(): string[] {
    return ['statusicon', 'name', 'labels', 'node', 'status', 'restarts', 'cpu', 'mem', 'age'];
  }

  private shouldShowNamespaceColumn_(): boolean {
    return this.namespaceService_.areMultipleNamespacesSelected();
  }

  hasErrors(pod: Pod): boolean {
    return pod.warnings.length > 0;
  }

  getEvents(pod: Pod): Event[] {
    return pod.warnings;
  }

  getDisplayStatus(pod: Pod): string {
    // See kubectl printers.go for logic in kubectl:
    // https://github.com/kubernetes/kubernetes/blob/39857f486511bd8db81868185674e8b674b1aeb9/pkg/printers/internalversion/printers.go
    let msgState = 'running';
    let reason = undefined;

    // Init container statuses are currently not taken into account.
    // However, init containers with errors will still show as failed because of warnings.
    if (pod.podStatus.containerStates) {
      // Container states array may be null when no containers have started yet.
      for (let i = pod.podStatus.containerStates.length - 1; i >= 0; i--) {
        const state = pod.podStatus.containerStates[i];
        if (state.waiting) {
          msgState = 'waiting';
          reason = state.waiting.reason;
        }
        if (state.terminated) {
          msgState = 'terminated';
          reason = state.terminated.reason;
          if (!reason) {
            if (state.terminated.signal) {
              reason = `Signal:${state.terminated.signal}`;
            } else {
              reason = `ExitCode:${state.terminated.exitCode}`;
            }
          }
        }
      }
    }

    if (msgState === 'waiting') {
      return `Waiting: ${reason}`;
    }

    if (msgState === 'terminated') {
      return `Terminated: ${reason}`;
    }

    return pod.podStatus.podPhase;
  }
}
