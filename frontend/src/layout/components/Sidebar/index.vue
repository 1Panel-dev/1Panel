<template>
    <div
        class="sidebar-container"
        element-loading-text="Loading..."
        :element-loading-spinner="loadingSvg"
        element-loading-svg-view-box="-10, -10, 50, 50"
        element-loading-background="rgba(122, 122, 122, 0.01)"
    >
        <div class="fixed" v-if="!isCollapse">
            <PrimaryMenu />
        </div>
        <Logo :isCollapse="isCollapse" />
        <el-scrollbar>
            <el-menu
                :default-active="activeMenu"
                :router="true"
                :collapse="isCollapse"
                :collapse-transition="false"
                :unique-opened="true"
                @select="handleMenuClick"
                class="custom-menu"
            >
                <SubItem :menuList="routerMenus" :level="0" />
            </el-menu>
        </el-scrollbar>
        <Collapse :version="version" @open-task="openTask" />
    </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { RouteRecordRaw, useRoute } from 'vue-router';
import { loadingSvg } from '@/utils/svg';
import Logo from './components/Logo.vue';
import Collapse from './components/Collapse.vue';
import SubItem from './components/SubItem.vue';
import { menuList } from '@/routers/router';
import { GlobalStore, MenuStore } from '@/store';
import { isString } from '@vueuse/core';
import { getSettingInfo } from '@/api/modules/setting';
import PrimaryMenu from '@/assets/images/menu-bg.svg?component';

const route = useRoute();
const menuStore = MenuStore();
const globalStore = GlobalStore();
const version = ref();

const activeMenu = computed(() => {
    const { meta, path } = route;
    return isString(meta.activeMenu) ? meta.activeMenu : path;
});
const isCollapse = computed((): boolean => menuStore.isCollapse);

let routerMenus = computed((): RouteRecordRaw[] => {
    return menuStore.menuList.filter((route) => route.meta && !route.meta.hideInSidebar) as RouteRecordRaw[];
});

const screenWidth = ref(0);
const listeningWindow = () => {
    window.onresize = () => {
        return (() => {
            screenWidth.value = document.body.clientWidth;
            if (!isCollapse.value && screenWidth.value < 1200) menuStore.setCollapse();
            if (isCollapse.value && screenWidth.value > 1200) menuStore.setCollapse();
        })();
    };
};
listeningWindow();
const emit = defineEmits(['menuClick', 'openTask']);
const handleMenuClick = (path) => {
    emit('menuClick', path);
};

function getCheckedLabels(menu: any, showSet: Set<string>) {
    for (const item of menu) {
        if (item.isShow) {
            showSet.add(item.label);
        }
        if (item.children) {
            getCheckedLabels(item.children, showSet);
        }
    }
}

const openTask = () => {
    emit('openTask');
};

const search = async () => {
    try {
        const res = await getSettingInfo();
        version.value = res.data.systemVersion;
        let hideMenu = JSON.parse(res.data.hideMenu);
        const showSet = new Set<string>();
        getCheckedLabels(hideMenu, showSet);
        const rstMenuList: RouteRecordRaw[] = [];
        const resMenuList = adjustAndCleanMenu(hideMenu, menuList);
        for (const menu of resMenuList) {
            let menuItem = JSON.parse(JSON.stringify(menu));
            if (!showSet.has(menuItem.name as string)) {
                continue;
            } else if (menuItem.name === 'Xpack-Menu') {
                menuItem.meta.hideInSidebar = false;
            }
            const itemChildren =
                (menuItem.children ?? []).filter(
                    (item) =>
                        item.name && showSet.has(item.name as string) && !(item.name === 'Upage' && globalStore.isIntl),
                ) || [];

            if (itemChildren.length === 1) {
                menuItem.meta.icon = itemChildren[0].meta.icon;
                menuItem.meta.title = itemChildren[0].meta.title;
            }
            menuItem.children = itemChildren;
            rstMenuList.push(menuItem);
        }
        if (!isSameMenuList(menuStore.menuList as RouteRecordRaw[], rstMenuList)) {
            menuStore.setMenuList(rstMenuList);
        }
    } catch (error) {
        if (!menuStore.menuList || menuStore.menuList.length === 0) {
            menuStore.setMenuList(menuList);
        }
    }
};

function isSameMenuList(source: RouteRecordRaw[], target: RouteRecordRaw[]) {
    return JSON.stringify(source) === JSON.stringify(target);
}

function adjustAndCleanMenu(menuItem, list) {
    const menuList = JSON.parse(JSON.stringify(list));
    const itemMap = new Map();
    for (const parent of menuList) {
        itemMap.set(parent.name, parent);
        if (Array.isArray(parent.children)) {
            for (const child of parent.children) {
                itemMap.set(child.name, child);
            }
        }
    }

    function buildTree(refList) {
        const result = [];

        for (const ref of refList) {
            const refName = ref.label;
            const matched = itemMap.get(refName);

            if (!matched) continue;

            if (Array.isArray(ref.children) && ref.children.length > 0) {
                matched.children = buildTree(ref.children);
            } else {
                delete matched.children;
            }

            result.push(matched);
        }

        return result;
    }

    const newMenu = buildTree(menuItem);
    for (const menu of newMenu) {
        if (menu.children?.length === 1) {
            menu.meta.icon = menu.children[0].meta.icon;
            menu.meta.title = menu.children[0].meta.title;
        }
    }

    return newMenu;
}

onMounted(() => {
    if (!menuStore.menuList || menuStore.menuList.length === 0) {
        menuStore.setMenuList(menuList);
    }
    search();
});
</script>

<style lang="scss">
@use 'index';

.background {
    z-index: 20;
}

.custom-menu .el-menu-item {
    white-space: normal !important;
    word-break: break-word;
    overflow-wrap: break-word;
    line-height: normal;
}

.sidebar-container {
    position: relative;
    display: flex;
    flex-direction: column;
    height: 100%;
    background: var(--panel-menu-bg-color) no-repeat top;

    .el-scrollbar {
        flex: 1;
        .el-menu {
            overflow: auto;
            overflow-x: hidden;
            border-right: none;
        }
    }
}

.ico {
    height: 20px !important;
}
</style>
