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

const baseHref = 'api/v1';

export enum Resource {
  job = 'job',
  cronJob = 'cronjob',
  crd = 'crd',
  crdFull = 'customresourcedefinition',
  crdObject = 'object',
  daemonSet = 'daemonset',
  deployment = 'deployment',
  pod = 'pod',
  virtualMachine = 'virtualmachine',
  replicaSet = 'replicaset',
  oldReplicaSet = 'oldreplicaset',
  newReplicaSet = 'newreplicaset',
  replicationController = 'replicationcontroller',
  statefulSet = 'statefulset',
  node = 'node',
  namespace = 'namespace',
  persistentVolume = 'persistentvolume',
  storageClass = 'storageclass',
  clusterRole = 'clusterrole',
  configMap = 'configmap',
  persistentVolumeClaim = 'persistentvolumeclaim',
  secret = 'secret',
  ingress = 'ingress',
  service = 'service',
  event = 'event',
  container = 'container',
  tenant = 'tenant',
  partition = 'partition',
  resourcequota = 'resourcequota',
  role = 'role',
  user = 'users',
  users = 'user',
  serviceaccount= 'serviceaccount',
  resourcePartition = 'resourcepartition',
  tenantPartition = 'tenantpartition',
  network = 'crd',
  networkFull = 'network',
  networkObject = 'object',
}

export enum Utility {
  shell = 'shell',
}

class ResourceEndpoint {
  constructor(
    private readonly resource_: Resource,
    private readonly namespaced_ = false,
    private readonly tenanted_ = false,
    private readonly partitioned_ = false,
  ) {
  }

  list(): string {
    return `${baseHref}${this.partitioned_ ? '/partition/:partition' : ''}${this.tenanted_ ? '/tenants/:tenant' : ''}/${this.resource_}${
      this.namespaced_ ? '/:namespace' : ''
    }`;
  }

  detail(): string {
    return `${baseHref}${this.partitioned_ ? '/partition/:partition' : ''}${this.tenanted_ ? '/tenants/:tenant' : ''}/${this.resource_}${
      this.namespaced_ ? '/:namespace' : ''
    }/:name`;
  }

  child(resourceName: string, relatedResource: Resource, resourceNamespace?: string, tenant?: string): string {
    if (!resourceNamespace) {
      resourceNamespace = ':namespace';
    }
    let url = ''
    if (tenant) {
      url = `${baseHref}${this.tenanted_ ? `/tenants/${tenant}` : ''}/${this.resource_}${
        this.namespaced_ ? `/${resourceNamespace}` : ''
      }/${resourceName}/${relatedResource}`
    } else {
      url = `${baseHref}${this.tenanted_ ? '/tenants/:tenant' : ''}/${this.resource_}${
        this.namespaced_ ? `/${resourceNamespace}` : ''
      }/${resourceName}/${relatedResource}`
    }
    return url;
  }
}

class UtilityEndpoint {
  constructor(private readonly utility_: Utility) {}

  shell(namespace: string, resourceName: string, tenant?: string): string {
    return (
      baseHref +
      (tenant ? `/tenants/${tenant}` : '') +
      `/${Resource.pod}/${namespace}/${resourceName}/${this.utility_}`
    );
  }
}

export class EndpointManager {
  static resource(resource: Resource, namespaced?: boolean, tenanted?: boolean, partitioned?: boolean): ResourceEndpoint {
    return new ResourceEndpoint(resource, namespaced, tenanted, partitioned);
  }

  static utility(utility: Utility): UtilityEndpoint {
    return new UtilityEndpoint(utility);
  }
}
