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
import {Component, ComponentFactoryResolver, Input} from '@angular/core';
import {Deployment, DeploymentList, Event} from '@api/backendapi';
import {Observable} from 'rxjs/Observable';
import {ResourceListWithStatuses} from '../../../resources/list';
import {NotificationsService} from '../../../services/global/notifications';
import {EndpointManager, Resource} from '../../../services/resource/endpoint';
import {NamespacedResourceService} from '../../../services/resource/resource';
import {MenuComponent} from '../../list/column/menu/component';
import {ListGroupIdentifier, ListIdentifier} from '../groupids';
import {ActivatedRoute} from "@angular/router";
import {TenantService} from "../../../services/global/tenant";

@Component({
  selector: 'kd-deployment-list',
  templateUrl: './template.html',
})
export class DeploymentListComponent extends ResourceListWithStatuses<DeploymentList, Deployment> {
  @Input() endpoint = EndpointManager.resource(Resource.deployment, true, true).list();

  tenantName: string;

  constructor(
    private readonly deployment_: NamespacedResourceService<DeploymentList>,
    private readonly activatedRoute_: ActivatedRoute,
    private readonly tenant_: TenantService,
    notifications: NotificationsService,
    resolver: ComponentFactoryResolver,
  ) {
    super('deployment', notifications, resolver);
    this.id = ListIdentifier.deployment;
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
      this.tenant_.current() : this.activatedRoute_.snapshot.params.resourceName
    sessionStorage.setItem('tenant', this.tenantName);
  }

  getResourceObservable(params?: HttpParams): Observable<DeploymentList> {
    const partition = this.tenantName === 'system' ? 'partition/' + sessionStorage.getItem(`${this.tenantName}`) + '/' : ''
    let endpoint = ''
    if (sessionStorage.getItem('userType') === 'cluster-admin') {
      endpoint = `api/v1/${partition}tenants/${this.tenantName}/deployment`
    } else {
      endpoint = this.endpoint
    }

    return this.deployment_.get(endpoint, undefined, undefined, params, this.tenantName);
  }

  map(deploymentList: DeploymentList): Deployment[] {
    return deploymentList.deployments;
  }

  isInErrorState(resource: Deployment): boolean {
    return resource.pods.warnings.length > 0;
  }

  isInPendingState(resource: Deployment): boolean {
    return resource.pods.warnings.length === 0 && resource.pods.pending > 0;
  }

  isInSuccessState(resource: Deployment): boolean {
    return resource.pods.warnings.length === 0 && resource.pods.pending === 0;
  }

  getDisplayColumns(): string[] {
    return ['statusicon', 'name', 'labels', 'pods', 'age', 'images'];
  }

  hasErrors(deployment: Deployment): boolean {
    return deployment.pods.warnings.length > 0;
  }

  getEvents(deployment: Deployment): Event[] {
    return deployment.pods.warnings;
  }

  private shouldShowNamespaceColumn_(): boolean {
    return this.namespaceService_.areMultipleNamespacesSelected();
  }
}
