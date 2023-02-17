import { defineStore } from 'pinia';
import { MenuState } from '../interface';
import piniaPersistConfig from '@/config/pinia-persist';
import { RouteRecordRaw } from 'vue-router';
const whiteList = ['/', '/error'];

export const MenuStore = defineStore({
    id: 'MenuState',
    state: (): MenuState => ({
        isCollapse: false,
        menuList: [],
    }),
    getters: {},
    actions: {
        async setCollapse() {
            this.isCollapse = !this.isCollapse;
        },
        async setMenuList(menuList: RouteRecordRaw[]) {
            const menus = menuList.filter((item) => {
                return whiteList.indexOf(item.path) < 0;
            });
            this.menuList = menus;
        },
    },
    persist: piniaPersistConfig('MenuStore'),
});
