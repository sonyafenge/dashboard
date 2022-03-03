
import {Status} from '@api/backendapi';
import {RatioItem} from '@api/frontendapi';

export enum ResourceRatioModes {
  Default = 'default',
  Suspendable = 'suspendable',
  Completable = 'completable',
}

export class Helper {
  static getResourceRatio(
    status: Status,
    totalItems: number,
    mode = ResourceRatioModes.Default,
  ): RatioItem[] {
    if (totalItems === 0) {
      return [];
    }

    let items = [
      {
        key: `Running: ${status.running}`,
        value: (status.running / totalItems) * 100,
      },
    ];

    switch (mode) {
      case ResourceRatioModes.Suspendable:
        items.push({
          key: `Suspended: ${status.failed}`,
          value: (status.failed / totalItems) * 100,
        });
        break;
      case ResourceRatioModes.Completable:
        items = items.concat([
          {
            key: `Failed: ${status.failed}`,
            value: (status.failed / totalItems) * 100,
          },
          {
            key: `Pending: ${status.pending}`,
            value: (status.pending / totalItems) * 100,
          },
          {
            key: `Succeeded: ${status.succeeded}`,
            value: (status.succeeded / totalItems) * 100,
          },
        ]);
        break;
      default:
        items = items.concat([
          {
            key: `Failed: ${status.failed}`,
            value: (status.failed / totalItems) * 100,
          },
          {
            key: `Pending: ${status.pending}`,
            value: (status.pending / totalItems) * 100,
          },
        ]);
        break;
    }

    return items;
  }
}
