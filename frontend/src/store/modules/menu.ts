import { defineStore } from 'pinia';
import type { StoreDefinition } from 'pinia';
import { MenuState } from '../interface';
import piniaPersistConfig from '@/config/pinia-persist';
import { RouteRecordRaw } from 'vue-router';
const whiteList = ['/login', '/error'];

export const MenuStore = defineStore('MenuState', {
    state: (): MenuState => ({
        isCollapse: false,
        menuList: [],
        isMenuLoaded: false,
        menuRequestId: 0,
        systemVersion: '',
        withoutAnimation: false,
    }),
    getters: {},
    actions: {
        async setCollapse() {
            this.isCollapse = !this.isCollapse;
            this.withoutAnimation = false;
        },
        async setMenuList(menuList: RouteRecordRaw[]) {
            const menus = menuList.filter((item) => {
                return whiteList.indexOf(item.path) < 0;
            });
            this.menuList = menus;
        },
        resetMenuList() {
            this.menuRequestId++;
            this.menuList = [];
            this.isMenuLoaded = false;
            this.systemVersion = '';
        },
        closeSidebar(withoutAnimation: boolean) {
            this.isCollapse = true;
            this.withoutAnimation = withoutAnimation;
        },
    },
    persist: {
        ...piniaPersistConfig('MenuStore'),
        pick: ['isCollapse', 'withoutAnimation'],
    },
}) as StoreDefinition<'MenuState', MenuState, any, any>;

export default MenuStore;
