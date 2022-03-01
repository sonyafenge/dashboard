import {Component, Input} from '@angular/core';
import {NetworkVersion} from '@api/backendapi';
import {MatTableDataSource} from '@angular/material';

@Component({
  selector: 'kd-network-versions-list',
  templateUrl: './template.html',
})
export class NetworkVersionListComponent {
  @Input() versions: NetworkVersion[];
  @Input() initialized: boolean;

  getDisplayColumns(): string[] {
    return ['name', 'served', 'storage'];
  }

  getDataSource(): MatTableDataSource<NetworkVersion> {
    const tableData = new MatTableDataSource<NetworkVersion>();
    tableData.data = this.versions;

    return tableData;
  }
}
