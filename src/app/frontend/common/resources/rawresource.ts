import {ObjectMeta, TypeMeta} from '@api/backendapi';

export class RawResource {
  static getUrl(tenant: string, typeMeta: TypeMeta, objectMeta: ObjectMeta): string {
    let resourceUrl = 'api/v1' + (tenant ? `/tenants/${tenant}` : '') + `/_raw/${typeMeta.kind}`;
    if (objectMeta.namespace !== undefined) {
      resourceUrl += `/namespace/${objectMeta.namespace}`;
    }
    resourceUrl += `/name/${objectMeta.name}`;

    if (resourceUrl.includes('/tenant/name')) {
      resourceUrl = 'api/v1/tenants' + `/${objectMeta.name}`
    }
    if (resourceUrl.includes('/role/')) {
      resourceUrl = `api/v1/tenants/${tenant}/namespaces/${objectMeta.namespace}/${typeMeta.kind}/${objectMeta.name}`;
    }
    return resourceUrl;
  }

}
