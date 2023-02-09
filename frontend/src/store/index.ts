import { defineStore } from 'pinia';
import { GlobalState, ThemeConfigProp } from './interface';
import { createPinia } from 'pinia';
import piniaPersistConfig from '@/config/pinia-persist';
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate';
import i18n from '@/lang';

export const GlobalStore = defineStore({
    id: 'GlobalState',
    state: (): GlobalState => ({
        isLoading: false,
        loadingText: '',
        isLogin: false,
        csrfToken: '',
        language: '',
        themeConfig: {
            panelName: '',
            primary: '#005EEB',
            theme: 'bright',
            footer: true,
        },
        isFullScreen: false,
    }),
    getters: {},
    actions: {
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
        updateLanguage(language: string) {
            this.language = language;
        },
        setThemeConfig(themeConfig: ThemeConfigProp) {
            this.themeConfig = themeConfig;
        },
    },
    persist: piniaPersistConfig('GlobalState'),
});

const pinia = createPinia();
pinia.use(piniaPluginPersistedstate);

export default pinia;
