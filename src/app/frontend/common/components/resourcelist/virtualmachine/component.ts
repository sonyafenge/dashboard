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
import {Event, VirtualMachine, VirtualMachineList} from '@api/backendapi';
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
  selector: 'kd-virtual-machine-list',
  templateUrl: './template.html'
})

export class VirtualMachineListComponent extends ResourceListWithStatuses<VirtualMachineList, VirtualMachine> {
  @Input() endpoint = EndpointManager.resource(Resource.virtualMachine, true, true).list();

  tenantName: string;

  constructor(
    private readonly virtualMachineList: NamespacedResourceService<VirtualMachineList>,
    private readonly activatedRoute_: ActivatedRoute,
    private readonly tenant_: TenantService,
    resolver: ComponentFactoryResolver,
    notifications: NotificationsService,
  ) {
    super('virtualmachine', notifications, resolver);
    this.id = ListIdentifier.virtualMachine;
    this.groupId = ListGroupIdentifier.workloads;

    // Register status icon handlers
    this.registerBinding(this.icon.checkCircle, 'kd-success', this.isInSuccessState);
    this.registerBinding(this.icon.timelapse, 'kd-muted', this.isInPendingState);
    this.registerBinding(this.icon.error, 'kd-error', this.isInErrorState);

    // Register action columns.
    this.registerActionColumn<MenuComponent>('menu', MenuComponent);

    this.tenantName = this.activatedRoute_.snapshot.params.resourceName === undefined ?
      this.tenant_.current() : this.activatedRoute_.snapshot.params.resourceName
    sessionStorage.setItem('tenantName',this.tenantName);
  }

  getResourceObservable(params?: HttpParams): Observable<VirtualMachineList> {
    let endpoint = ''
    if (sessionStorage.getItem('userType') === 'cluster-admin') {
      endpoint = `api/v1/tenants/${this.tenantName}/virtualmachine`
    } else {
      endpoint = this.endpoint
    }
    return this.virtualMachineList.get(endpoint, undefined, undefined, params,this.tenantName);
  }

  map(virtualMachineList: VirtualMachineList): VirtualMachine[] {
    return virtualMachineList.virtualMachines;
  }

  isInErrorState(resource: VirtualMachine): boolean {
    return resource.podStatus.status === 'Failed';
  }

  isInPendingState(resource: VirtualMachine): boolean {
    return resource.podStatus.status === 'Pending';
  }

  isInSuccessState(resource: VirtualMachine): boolean {
    return resource.podStatus.status === 'Succeeded' || resource.podStatus.status === 'Running';
  }

  getDisplayColumns(): string[] {
    return ['statusicon', 'name', 'image', 'keypair', 'status', 'phase', 'restart','age'];
  }

  hasErrors(virtualMachine: VirtualMachine): boolean {
    return virtualMachine.warnings.length > 0;
  }

  getEvents(virtualMachine: VirtualMachine): Event[] {
    return virtualMachine.warnings;
  }

  getDisplayStatus(virtualMachine: VirtualMachine): string {
    // See kubectl printers.go for logic in kubectl:
    // https://github.com/kubernetes/kubernetes/blob/39857f486511bd8db81868185674e8b674b1aeb9/pkg/printers/internalversion/printers.go
    let msgState = 'running';
    let reason = undefined;

    // Init container statuses are currently not taken into account.
    // However, init containers with errors will still show as failed because of warnings.


    if (msgState === 'waiting') {
      return `Waiting: ${reason}`;
    }

    if (msgState === 'terminated') {
      return `Terminated: ${reason}`;
    }

    return virtualMachine.podStatus.podPhase;
  }
}
