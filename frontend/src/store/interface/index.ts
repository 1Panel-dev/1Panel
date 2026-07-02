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
    loginImage: string;
    loginBgType: string;
    loginBackground: string;
    loginBtnLinkColor: string;
    themeColor: string;
}

export interface Watermark {
    lightColor: string;
    darkColor: string;
    fontSize: number;
    content: string;
    rotate: number;
    gap: number;
}

export interface GlobalState {
    language: string; // zh | en | tw
    device: DeviceType;
    themeConfig: ThemeConfigProp;
    // ui
    isFullScreen: boolean;
    openMenuTabs: boolean;
    menuAccordion: boolean;
    watermark: Watermark | null;
    watermarkShow: boolean;
    isLoading: boolean;
    loadingText: string;
    // auth
    ignoreCaptcha: boolean;
    agreeLicense: boolean;
    isLogin: boolean;
    entrance: string;
    csrfToken: string;
    // context
    hasNewVersion: boolean;
    lastFilePath: string;
    currentDB: string;
    currentPgDB: string;
    currentRedisDB: string;
    currentMongodbDB: string;
    showEntranceWarn: boolean;
    defaultNetwork: string;
    defaultIO: string;
    isOnRestart: boolean;
    // tags
    isAdmin: boolean;
    permissions: string[];
    masterOnlyPermissions: string[];
    nodeRoles: Array<{ nodeId: number; nodeName: string; roleId: number; roleName: string }>;
    isEnterprise: boolean;
    isIntl: boolean;
    docWithRegion: boolean;
    isFxplay: boolean;
    isOffline: boolean;
    // license
    isProductPro: boolean;
    productProExpires: number;
    isMasterProductPro: boolean;
    isEnterpriseLicensed: boolean;
    isEnterpriseLicenseLoaded: boolean;
    // multi-node
    masterAlias: string;
    currentNode: string;
    currentNodeAddr: string;
}

export interface MenuState {
    isCollapse: boolean;
    menuList: RouteRecordRaw[];
    withoutAnimation: boolean;
}

export interface TerminalState {
    lineHeight: number;
    letterSpacing: number;
    fontSize: number;
    fontFamily: string;
    backgroundColor: string;
    foregroundColor: string;
    cursorBlink: string;
    cursorStyle: string;
    scrollback: number;
    scrollSensitivity: number;
}
