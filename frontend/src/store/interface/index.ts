import { RouteRecordRaw } from 'vue-router';
import { DeviceType } from '@/enums/app';
export interface ThemeConfigProp {
    panelName: string;
    primary: string;
    theme: string; // dark | bright ｜ auto
    footer: boolean;

    title: string;
    logo: string;
    logoWithText: string;
    favicon: string;
    themeColor: string;
}

export interface GlobalState {
    isLoading: boolean;
    loadingText: string;
    isLogin: boolean;
    entrance: string;
    language: string; // zh | en | tw
    themeConfig: ThemeConfigProp;
    isFullScreen: boolean;
    openMenuTabs: boolean;
    isOnRestart: boolean;
    agreeLicense: boolean;
    hasNewVersion: boolean;
    ignoreCaptcha: boolean;
    device: DeviceType;
    lastFilePath: string;
    currentDB: string;
    currentRedisDB: string;
    showEntranceWarn: boolean;
    defaultNetwork: string;

    isProductPro: boolean;
    isIntl: boolean;
    isTrial: boolean;
    productProExpires: number;

    errStatus: string;
}

export interface MenuState {
    isCollapse: boolean;
    menuList: RouteRecordRaw[];
    withoutAnimation: boolean;
}
