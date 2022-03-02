import {Component, Input} from '@angular/core';
import {ResourcesRatio} from '@api/frontendapi';

export const emptyResourcesRatio: ResourcesRatio = {
  cronJobRatio: [],
  daemonSetRatio: [],
  deploymentRatio: [],
  jobRatio: [],
  podRatio: [],
  replicaSetRatio: [],
  statefulSetRatio: [],
};

@Component({
  selector: 'kd-workload-statuses',
  templateUrl: './template.html',
  styleUrls: ['./style.scss'],
})
export class WorkloadStatusComponent {
  @Input() resourcesRatio: ResourcesRatio;
  colors: string[] = ['#00c752', '#f00', '#ffad20', '#006028'];

  constructor() {
    if (!this.resourcesRatio) {
      this.resourcesRatio = emptyResourcesRatio;
    }
  }
}
