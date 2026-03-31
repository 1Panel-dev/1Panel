import { defineStore } from 'pinia';
import piniaPersistConfig from '@/config/pinia-persist';
import { GlobalState, ThemeConfigProp } from '../interface';
import { DeviceType } from '@/enums/app';
import i18n, { setActiveLocale } from '@/lang';

const CN_DOCS_URL = 'https://1panel.cn/docs/v2';
const INTL_DOCS_URL = 'https://docs.1panel.pro/v2';

const GlobalStore = defineStore({
    id: 'GlobalState',
    state: (): GlobalState => ({
        isLoading: false,
        loadingText: '',
        isLogin: false,
        csrfToken: '',
        entrance: '',
        language: i18n.global.locale.value,
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
        watermark: null,
        watermarkShow: false,
        openMenuTabs: false,
        isFullScreen: false,
        isOnRestart: false,
        agreeLicense: false,
        hasNewVersion: false,
        ignoreCaptcha: true,
        device: DeviceType.Desktop,
        lastFilePath: '',
        currentDB: '',
        currentPgDB: '',
        currentRedisDB: '',
        showEntranceWarn: true,
        defaultNetwork: 'all',
        defaultIO: 'all',
        isFxplay: false,

        isProductPro: false,
        isIntl: false,
        docWithRegion: true,
        productProExpires: 0,
        isMasterProductPro: false,
        isOffLine: false,

        masterAlias: '',
        currentNode: 'local',
        currentNodeAddr: '',
    }),
    getters: {
        isDarkTheme: (state) =>
            state.themeConfig.theme === 'dark' ||
            (state.themeConfig.theme === 'auto' && window.matchMedia('(prefers-color-scheme: dark)').matches),
        isDarkGoldTheme: (state) => state.themeConfig.primary === '#F0BE96' && state.isProductPro,
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
        setOpenMenuTabs(openMenuTabs: boolean) {
            this.openMenuTabs = openMenuTabs;
        },
        setScreenFull() {
            this.isFullScreen = !this.isFullScreen;
        },
        setLogStatus(login: boolean) {
            this.isLogin = login;
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
        setThemeConfig(themeConfig: ThemeConfigProp) {
            this.themeConfig = themeConfig;
        },
        setAgreeLicense(agree: boolean) {
            this.agreeLicense = agree;
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
        isMasterPro() {
            return this.isMasterProductPro;
        },
        setLastFilePath(path: string) {
            this.lastFilePath = path;
        },
        setCurrentDB(name: string) {
            this.currentDB = name;
        },
        setCurrentPgDB(name: string) {
            this.currentPgDB = name;
        },
        setCurrentRedisDB(name: string) {
            this.currentRedisDB = name;
        },
        setShowEntranceWarn(show: boolean) {
            this.showEntranceWarn = show;
        },
        setDefaultNetwork(net: string) {
            this.defaultNetwork = net;
        },
        setDefaultIO(net: string) {
            this.defaultIO = net;
        },
    },
    persist: piniaPersistConfig('GlobalState'),
});

export default GlobalStore;
