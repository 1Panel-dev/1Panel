import { defineStore } from 'pinia';
import { GlobalState, ThemeConfigProp } from './interface';
import { createPinia } from 'pinia';
import piniaPersistConfig from '@/config/piniaPersist';
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate';

// defineStore 调用后返回一个函数，调用该函数获得 Store 实体
export const GlobalStore = defineStore({
    // id: 必须的，在所有 Store 中唯一
    id: 'GlobalState',
    // state: 返回对象的函数
    state: (): GlobalState => ({
        // token
        token: '',
        // userInfo
        userInfo: '',
        // element组件大小
        assemblySize: 'default',
        // language
        language: '',
        // themeConfig
        themeConfig: {
            // 默认 primary 主题颜色
            primary: '#409EFF',
            // 深色模式
            isDark: false,
            // 灰色模式
            isGrey: false,
            // 色弱模式
            isWeak: false,
            // 面包屑导航
            breadcrumb: true,
            // 标签页
            tabs: false,
            // 页脚
            footer: true,
        },
    }),
    getters: {},
    actions: {
        // setToken
        setToken(token: string) {
            this.token = token;
        },
        // setUserInfo
        setUserInfo(userInfo: any) {
            this.userInfo = userInfo;
        },
        // setAssemblySizeSize
        setAssemblySizeSize(assemblySize: string) {
            this.assemblySize = assemblySize;
        },
        // updateLanguage
        updateLanguage(language: string) {
            this.language = language;
        },
        // setThemeConfig
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
