import { defineStore } from 'pinia';
import piniaPersistConfig from '@/config/pinia-persist';
import { GlobalState, ThemeConfigProp } from '../interface';
import { DeviceType } from '@/enums/app';
import i18n from '@/lang';

const GlobalStore = defineStore({
    id: 'GlobalState',
    state: (): GlobalState => ({
        isLoading: false,
        loadingText: '',
        isLogin: false,
        entrance: '',
        language: '',
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
        },
        openMenuTabs: false,
        isFullScreen: false,
        isOnRestart: false,
        agreeLicense: false,
        hasNewVersion: false,
        ignoreCaptcha: true,
        device: DeviceType.Desktop,
        lastFilePath: '',
        currentDB: '',
        currentRedisDB: '',
        showEntranceWarn: true,
        defaultNetwork: 'all',

        isProductPro: false,
        isIntl: false,
        isTrial: false,
        productProExpires: 0,

        errStatus: '',
    }),
    getters: {
        isDarkTheme: (state) =>
            state.themeConfig.theme === 'dark' ||
            (state.themeConfig.theme === 'auto' && window.matchMedia('(prefers-color-scheme: dark)').matches),
        isDarkGoldTheme: (state) => state.themeConfig.primary === '#F0BE96' && state.isProductPro,
        docsUrl: (state) => (state.isIntl ? 'https://docs.1panel.hk' : 'https://1panel.cn/docs'),
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
            this.loadingText = i18n.global.t('commons.loadingText.' + text);
        },
        setCsrfToken(token: string) {
            this.csrfToken = token;
        },
        updateLanguage(language: any) {
            this.language = language;
            localStorage.setItem('lang', language);
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
        setLastFilePath(path: string) {
            this.lastFilePath = path;
        },
        setCurrentDB(name: string) {
            this.currentDB = name;
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
    },
    persist: piniaPersistConfig('GlobalState'),
});

export default GlobalStore;
