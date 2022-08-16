import { defineStore } from 'pinia';
import { MenuState } from '../interface';
import piniaPersistConfig from '@/config/pinia-persist';
import { RouteRecordRaw } from 'vue-router';
const whiteList = ['/login', '/error'];

export const MenuStore = defineStore({
    id: 'MenuState',
    state: (): MenuState => ({
        // menu collapse
        isCollapse: false,
        // menu List
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
