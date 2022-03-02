import {Component} from '@angular/core';
import {ListGroupIdentifier,} from '../../common/components/resourcelist/groupids';
import {GroupedResourceList} from '../../common/resources/groupedlist';

@Component({
  selector: 'kd-workloads',
  templateUrl: './template.html',
})
export class WorkloadsComponent extends GroupedResourceList {
  hasWorkloads(): boolean {
    return this.isGroupVisible(ListGroupIdentifier.workloads);
  }

  hasDiscovery(): boolean {
    return this.isGroupVisible(ListGroupIdentifier.discovery);
  }

  hasConfig(): boolean {
    return this.isGroupVisible(ListGroupIdentifier.config);
  }

  showWorkloadStatuses(): boolean {
    return (
      Object.values(this.resourcesRatio).reduce((sum, ratioItems) => sum + ratioItems.length, 0) !==
      0
    );
  }

  showGraphs(): boolean {
    return this.cumulativeMetrics.every(
      metrics => metrics.dataPoints && metrics.dataPoints.length > 1,
    );
  }
}
