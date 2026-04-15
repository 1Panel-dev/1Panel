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
        currentMongodbDB: '',
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
        setScreenFull() {
            this.isFullScreen = !this.isFullScreen;
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
        isMasterPro() {
            return this.isMasterProductPro;
        },
    },
    persist: piniaPersistConfig('GlobalState'),
});

export default GlobalStore;
