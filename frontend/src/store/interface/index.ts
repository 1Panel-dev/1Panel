import { RouteRecordRaw } from 'vue-router';

/* themeConfigProp */
export interface ThemeConfigProp {
    primary: string;
    isDark: boolean;
    isGrey: boolean;
    isWeak: boolean;
    breadcrumb: boolean;
    tabs: boolean;
    footer: boolean;
}

/* GlobalState */
export interface GlobalState {
    token: string;
    userInfo: any;
    assemblySize: string;
    language: string;
    themeConfig: ThemeConfigProp;
}

export interface MenuState {
    isCollapse: boolean;
    menuList: RouteRecordRaw[];
}

/* AuthState */
export interface AuthState {
    authRouter: string[];
}
