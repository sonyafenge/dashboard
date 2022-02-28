import {Component, OnInit} from '@angular/core';
import {LoginStatus} from '@api/backendapi';
import {AuthService} from '../../common/services/global/authentication';
import {Router} from '@angular/router';

@Component({
  selector: 'kd-user-panel',
  templateUrl: './template.html',
  styleUrls: ['./style.scss'],
  host: {
    '[class.kd-hidden]': 'this.isAuthEnabled() === false',
  },
})

export class UserPanelComponent implements OnInit {
  loginStatus: LoginStatus;
  isLoginStatusInitialized = false;

  constructor(private readonly authService_: AuthService, private readonly router_: Router) {}

  ngOnInit(): void {
    this.authService_.getLoginStatus().subscribe(status => {
      this.loginStatus = status;
      this.isLoginStatusInitialized = true;
    });
  }

  isAuthSkipped(): boolean {
    return (
      this.loginStatus && !this.authService_.isLoginPageEnabled() && !this.loginStatus.headerPresent
    );
  }

  isLoggedIn(): boolean {
    return this.loginStatus && !this.loginStatus.headerPresent && this.loginStatus.tokenPresent;
  }

  isAuthEnabled(): boolean {
    return this.loginStatus ? this.loginStatus.httpsMode : false;
  }

  logout(): void {
    this.authService_.logout();
  }

  settings(): void {
    this.router_.navigate(['settings'], {queryParamsHandling: 'preserve'});
  }

  about(): void {
    this.router_.navigate(['about'], {queryParamsHandling: 'preserve'});
  }
}
