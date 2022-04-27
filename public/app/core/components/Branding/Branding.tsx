import { css, cx } from '@emotion/css';
import React, { FC } from 'react';

import { colorManipulator } from '@grafana/data';
import { useTheme2, styleMixins } from '@grafana/ui';

export interface BrandComponentProps {
  className?: string;
  children?: JSX.Element | JSX.Element[];
}

const LoginLogo: FC<BrandComponentProps> = ({ className }) => {
  return (
    <img
      className={className}
      src="https://cdn-icons.flaticon.com/png/512/5680/premium/5680565.png?token=exp=1651057403~hmac=abbd4d5adf25bdeafcbc246238887e74"
      alt="Grafana"
    />
  );
};

const LoginBackground: FC<BrandComponentProps> = ({ className, children }) => {
  const theme = useTheme2();

  const background = css`
    &:before {
      content: '';
      position: absolute;
      left: 0;
      right: 0;
      bottom: 0;
      top: 0;
      background: url(public/img/g8_login_${theme.isDark ? 'dark' : 'light'}.svg);
      background-position: top center;
      background-size: auto;
      background-repeat: no-repeat;

      opacity: 0;
      transition: opacity 3s ease-in-out;

      @media ${styleMixins.mediaUp(theme.v1.breakpoints.md)} {
        background-position: center;
        background-size: cover;
      }
    }
  `;

  return <div className={cx(background, className)}>{children}</div>;
};

const MenuLogo: FC<BrandComponentProps> = ({ className }) => {
  return (
    <img
      className={className}
      src="https://cdn-icons.flaticon.com/png/512/5680/premium/5680565.png?token=exp=1651057403~hmac=abbd4d5adf25bdeafcbc246238887e74"
      alt="Grafana"
    />
  );
};

const LoginBoxBackground = () => {
  const theme = useTheme2();
  return css`
    background: ${colorManipulator.alpha(theme.colors.background.primary, 0.7)};
    background-size: cover;
  `;
};

export class Branding {
  static LoginLogo = LoginLogo;
  static LoginBackground = LoginBackground;
  static MenuLogo = MenuLogo;
  static LoginBoxBackground = LoginBoxBackground;
  static AppTitle = 'Soi Devops';
  static LoginTitle = 'Welcome to Soi Devops';
  static GetLoginSubTitle = (): null | string => {
    return null;
  };
}
