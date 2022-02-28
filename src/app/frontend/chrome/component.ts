import {HttpClient} from '@angular/common/http';
import {Component, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {AssetsService} from '../common/services/global/assets';

class SystemBanner {
  message: string;
  severity: string;
}

@Component({
  selector: 'kd-chrome',
  templateUrl: './template.html',
  styleUrls: ['./style.scss'],
})
export class ChromeComponent implements OnInit {
  private static readonly systemBannerEndpoint = 'api/v1/systembanner';
  private systemBanner_: SystemBanner;
  loading = false;

  constructor(
    public assets: AssetsService,
    private readonly http_: HttpClient,
    private readonly router_: Router,
  ) {}

  ngOnInit(): void {
    this.http_
      .get<SystemBanner>(ChromeComponent.systemBannerEndpoint)
      .toPromise()
      .then(sb => {
        this.systemBanner_ = sb;
      });
  }

  getOverviewStateName(): string {
    return 'overview';
  }

  isSystemBannerVisible(): boolean {
    return this.systemBanner_ && this.systemBanner_.message.length > 0;
  }

  getSystemBannerClass(): string {
    const severity =
      this.systemBanner_ && this.systemBanner_.severity
        ? this.systemBanner_.severity.toLowerCase()
        : '';
    switch (severity) {
      case 'warning':
        return 'kd-bg-warning-light';
      case 'error':
        return 'kd-bg-error-light';
      default:
        return 'kd-bg-success-light';
    }
  }

  getSystemBannerMessage(): string {
    return this.systemBanner_ ? this.systemBanner_.message : '';
  }

  goToCreateState(): void {
    this.router_.navigate(['create'], {queryParamsHandling: 'preserve'});
  }
}
