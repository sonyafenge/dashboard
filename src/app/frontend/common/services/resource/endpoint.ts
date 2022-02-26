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
  resourcequota = 'resourcequota',
  resourcePartition = 'resourcepartition',
  role = 'role',
  tenantPartition = 'tenantpartition',
  secret = 'secret',
  ingress = 'ingress',
  service = 'service',
  event = 'event',
  container = 'container',
  tenant = 'tenant',
  partition = 'partition',
}

export enum Utility {
  shell = 'shell',
}

class ResourceEndpoint {
  constructor(
    private readonly resource_: Resource,
    private readonly namespaced_ = false,
    private readonly tenanted_ = false,
  ) {}

  list(): string {
    return `${baseHref}${this.tenanted_ ? '/tenants/:tenant' : ''}/${this.resource_}${
      this.namespaced_ ? '/:namespace' : ''
    }`;
  }

  detail(): string {
    return `${baseHref}${this.tenanted_ ? '/tenants/:tenant' : ''}/${this.resource_}${
      this.namespaced_ ? '/:namespace' : ''
    }/:name`;
  }

  child(resourceName: string, relatedResource: Resource, resourceNamespace?: string): string {
    if (!resourceNamespace) {
      resourceNamespace = ':namespace';
    }

    return `${baseHref}${this.tenanted_ ? '/tenants/:tenant' : ''}/${this.resource_}${
      this.namespaced_ ? `/${resourceNamespace}` : ''
    }/${resourceName}/${relatedResource}`;
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
  static resource(resource: Resource, namespaced?: boolean, tenanted?: boolean): ResourceEndpoint {
    return new ResourceEndpoint(resource, namespaced, tenanted);
  }

  static utility(utility: Utility): UtilityEndpoint {
    return new UtilityEndpoint(utility);
  }
}
