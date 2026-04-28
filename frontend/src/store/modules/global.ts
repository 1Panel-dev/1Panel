import { defineStore } from 'pinia';
import piniaPersistConfig from '@/config/pinia-persist';
import { GlobalState } from '../interface';
import { DeviceType } from '@/enums/app';
import i18n, { setActiveLocale } from '@/lang';

const CN_DOCS_URL = 'https://1panel.cn/docs/v2';
const INTL_DOCS_URL = 'https://docs.1panel.pro/v2';

const GlobalStore = defineStore({
    id: 'GlobalState',
    state: (): GlobalState => ({
        language: i18n.global.locale.value,
        device: DeviceType.Desktop,
        themeConfig: {
            panelName: '',
            primary: '#005eeb',
            theme: 'auto',
            footer: true,
            themeColor: '',
            title: '',
            logo: '',
            logoWithText: '',
            favicon: '',
            loginImage: '',
            loginBackground: '',
            loginBgType: '',
            loginBtnLinkColor: '',
        },
        // ui
        isFullScreen: false,
        openMenuTabs: false,
        watermark: null,
        watermarkShow: false,
        isLoading: false,
        loadingText: '',
        csrfToken: '',
        // auth
        ignoreCaptcha: true,
        agreeLicense: false,
        isLogin: false,
        entrance: '',
        // context
        hasNewVersion: false,
        lastFilePath: '',
        currentDB: '',
        currentPgDB: '',
        currentRedisDB: '',
        currentMongodbDB: '',
        showEntranceWarn: true,
        defaultNetwork: 'all',
        defaultIO: 'all',
        isOnRestart: false,
        // tags
        isAdmin: false,
        permissions: [],
        nodeRoles: [],
        isEnterprise: false,
        isIntl: false,
        docWithRegion: true,
        isFxplay: false,
        isOffLine: false,
        // license
        isProductPro: false,
        productProExpires: 0,
        isMasterProductPro: false,
        isEnterpriseLicensed: false,
        isEnterpriseLicenseLoaded: false,
        // multi-node
        masterAlias: '',
        currentNode: 'local',
        currentNodeAddr: '',
    }),
    getters: {
        isDarkTheme: (state) =>
            state.themeConfig.theme === 'dark' ||
            (state.themeConfig.theme === 'auto' && window.matchMedia('(prefers-color-scheme: dark)').matches),
        isDarkGoldTheme: (state) => state.themeConfig.primary === '#F0BE96' && state.isProductPro,
        isNodeAdmin: (state) =>
            state.nodeRoles.some((item) => item.nodeName === state.currentNode && item.roleName === 'Node Admin'),
        docsUrl: (state) => {
            if (state.docWithRegion) {
                return state.isIntl ? INTL_DOCS_URL : CN_DOCS_URL;
            }
            const lang = state.language.toLowerCase();
            const isChinese = lang === 'zh';
            return isChinese ? CN_DOCS_URL : INTL_DOCS_URL;
        },
        isMaster: (state) => state.currentNode === 'local',
    },
    actions: {
        setScreenFull() {
            this.isFullScreen = !this.isFullScreen;
        },
        setLogStatus(login: boolean) {
            this.isLogin = login;
        },
        setAuthInfo(payload: {
            isAdmin: boolean;
            permissions: string[];
            nodeRoles?: Array<{ nodeId: number; nodeName: string; roleId: number; roleName: string }>;
        }) {
            this.isAdmin = !!payload.isAdmin;
            this.permissions = payload.permissions || [];
            this.nodeRoles = payload.nodeRoles || [];
        },
        clearAuthInfo() {
            this.permissions = [];
            this.nodeRoles = [];
            this.isAdmin = false;
        },
        hasPermission(permission: string) {
            if (this.isAdmin) {
                return true;
            }
            const normalizedPermission = permission.trim();
            if (!normalizedPermission) {
                return false;
            }
            if (this.permissions.includes(normalizedPermission)) {
                return true;
            }
            const managePermission = getManagePermission(normalizedPermission);
            return managePermission ? this.permissions.includes(managePermission) : false;
        },
        setGlobalLoading(loading: boolean) {
            this.isLoading = loading;
        },
        setLoadingText(text: string) {
            this.loadingText = text;
        },
        setCsrfToken(token: string) {
            this.csrfToken = token;
        },
        async updateLanguage(language: string) {
            const activeLocale = await setActiveLocale(language);
            this.language = activeLocale;
            return activeLocale;
        },
        toggleDevice(value: DeviceType) {
            this.device = value;
        },
        isMobile() {
            return this.device === DeviceType.Mobile;
        },
        getMasterAlias() {
            return this.masterAlias || i18n.global.t('xpack.node.master');
        },
        isEE() {
            return this.isEnterprise && this.isEnterpriseLicensed;
        },
        isXpackOrEE() {
            return (this.isEnterprise && this.isEnterpriseLicensed) || this.isMasterProductPro;
        },
        isXpackNodeOrEE() {
            return (this.isEnterprise && this.isEnterpriseLicensed) || this.isProductPro;
        },
        isMasterPro() {
            return this.isMasterProductPro;
        },
    },
    persist: piniaPersistConfig('GlobalState'),
});

export default GlobalStore;

function getManagePermission(permission: string) {
    if (permission.endsWith('_view')) {
        return permission.replace(/_view$/, '_manage');
    }
    return '';
}
