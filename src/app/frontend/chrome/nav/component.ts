import {Component, OnInit, ViewChild} from '@angular/core';
import {MatDrawer} from '@angular/material';
import {NavService} from '../../common/services/nav/service';
import {TenantService} from 'common/services/global/tenant';

@Component({
  selector: 'kd-nav',
  templateUrl: './template.html',
  styleUrls: ['./style.scss'],
})
export class NavComponent implements OnInit {
  @ViewChild(MatDrawer, {static: true}) private readonly nav_: MatDrawer;
  showCluster:boolean;
  showTenant:boolean;
  showUser:boolean;
  showMenu:boolean;

  constructor(
    private readonly navService_: NavService,
    private readonly tenantService_: TenantService,
  ) {

    this.showCluster = true;
    this.showTenant = true;
    this.showUser = true;
    this.showMenu = true;

    const usertype = sessionStorage.getItem('userType');//added

    if(usertype==='cluster-admin'){
      this.showCluster=this.showCluster;
      this.showTenant=!this.showTenant;
      this.showUser=this.showUser;
      this.showMenu=!this.showMenu;
    }

    else if(usertype==='tenant-admin'){
      this.showCluster=!this.showCluster;
      this.showTenant=this.showTenant;
      this.showUser=this.showUser;
      this.showMenu=this.showMenu;
    }

    else{
      this.showCluster=!this.showCluster;
      this.showTenant=!this.showTenant;
      this.showUser=!this.showUser;
      this.showMenu=this.showMenu;
    }
  }

  get isSystem(): boolean {
    return this.tenantService_.isSystem();
  }

  get isCurrentSystem(): boolean {
    return this.tenantService_.isCurrentSystem();
  }

  ngOnInit(): void {
    this.navService_.setNav(this.nav_);
    this.navService_.setVisibility(true);
  }
}
