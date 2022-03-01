
import {Inject, Injectable} from '@angular/core';
import {MatIconRegistry} from '@angular/material';
import {DomSanitizer} from '@angular/platform-browser';

@Injectable()
export class AssetsService {
  private readonly assetsPath_ = 'assets/images';
  private readonly appLogoSvg_ = 'centaurus-logo.svg';
  private readonly appLogoTextSvg_ = 'centaurus-logo-text.svg';
  private readonly appLogoIcon_ = 'kd-logo';
  private readonly appLogoTextIcon_ = 'kd-logo-text';

  constructor(
    @Inject(MatIconRegistry) private readonly iconRegistry_: MatIconRegistry,
    @Inject(DomSanitizer) private readonly sanitizer_: DomSanitizer,
  ) {
    iconRegistry_.addSvgIcon(
      this.appLogoIcon_,
      sanitizer_.bypassSecurityTrustResourceUrl(`${this.assetsPath_}/${this.appLogoSvg_}`),
    );
    iconRegistry_.addSvgIcon(
      this.appLogoTextIcon_,
      sanitizer_.bypassSecurityTrustResourceUrl(`${this.assetsPath_}/${this.appLogoTextSvg_}`),
    );
    iconRegistry_.addSvgIcon(
      'pin',
      sanitizer_.bypassSecurityTrustResourceUrl(`${this.assetsPath_}/pin.svg`),
    );
    iconRegistry_.addSvgIcon(
      'pin-crossed',
      sanitizer_.bypassSecurityTrustResourceUrl(`${this.assetsPath_}/pin-crossed.svg`),
    );
  }

  getAppLogo(): string {
    return this.appLogoIcon_;
  }

  getAppLogoText(): string {
    return this.appLogoTextIcon_;
  }
}
