import { defineStore } from 'pinia';
import { GlobalState, ThemeConfigProp } from './interface';
import { createPinia } from 'pinia';
import piniaPersistConfig from '@/config/piniaPersist';
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate';

export const GlobalStore = defineStore({
    id: 'GlobalState',
    state: (): GlobalState => ({
        isLogin: false,
        userInfo: '',
        assemblySize: 'default',
        language: '',
        themeConfig: {
            primary: '#409EFF',
            isDark: false,
            isGrey: false,
            isWeak: false,
            breadcrumb: true,
            tabs: false,
            footer: true,
        },
    }),
    getters: {},
    actions: {
        setLogStatus(login: boolean) {
            this.isLogin = login;
        },
        setUserInfo(userInfo: any) {
            this.userInfo = userInfo;
        },
        setAssemblySizeSize(assemblySize: string) {
            this.assemblySize = assemblySize;
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

// piniaPersist(持久化)
const pinia = createPinia();
pinia.use(piniaPluginPersistedstate);

export default pinia;
