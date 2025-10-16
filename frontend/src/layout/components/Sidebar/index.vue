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
import { sortMenu } from '@/utils/util';

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

function getCheckedLabels(menu: any, showMap: any) {
    for (const item of menu) {
        if (item.isShow) {
            showMap[item.label] = true;
        }
        if (item.children) {
            getCheckedLabels(item.children, showMap);
        }
    }
}

const openTask = () => {
    emit('openTask');
};

const search = async () => {
    const res = await getSettingInfo();
    version.value = res.data.systemVersion;
    let hideMenu = JSON.parse(res.data.hideMenu);
    sortMenu(hideMenu);
    const showMap = new Map();
    getCheckedLabels(hideMenu, showMap);
    const rootMap = new Map();
    hideMenu.forEach((m, index) => {
        rootMap.set(m.label, index);
    });
    let rstMenuList: RouteRecordRaw[] = [];
    let resMenuList: RouteRecordRaw[] = [];
    for (const menu of menuStore.menuList) {
        let menuItem = JSON.parse(JSON.stringify(menu));
        if (!showMap[menuItem.name]) {
            continue;
        } else if (menuItem.name === 'Xpack-Menu') {
            menuItem.meta.hideInSidebar = false;
        }
        const childMenu = hideMenu.find((item) => item.label == menu.name);
        const childMap = buildIndexMap(childMenu?.children || []);
        const itemChildren =
            (menuItem.children ?? [])
                .filter(
                    (item) =>
                        item.name &&
                        showMap[item.name as string] &&
                        !(item.name === 'XAlertDashboard' && globalStore.isIntl),
                )
                .sort(sortByMap(childMap)) || [];

        if (itemChildren.length === 1) {
            menuItem.meta.icon = itemChildren[0].meta.icon;
            menuItem.meta.title = itemChildren[0].meta.title;
        }
        menuItem.children = itemChildren;
        rstMenuList.push(menuItem);
    }
    rstMenuList.sort((a, b) => {
        const labelA = a.name;
        const labelB = b.name;
        const indexA = rootMap.get(labelA) ?? Infinity;
        const indexB = rootMap.get(labelB) ?? Infinity;
        return indexA - indexB;
    });
    resMenuList = adjustAndCleanMenu(hideMenu, rstMenuList);
    menuStore.menuList = resMenuList;
};

function buildIndexMap(list: any[]): Map<string, number> {
    const map = new Map<string, number>();
    list.forEach((m, i) => map.set(m.label, i));
    return map;
}

function sortByMap(map: Map<string, number>) {
    return (a: { name: string }, b: { name: string }) => {
        const indexA = map.get(a.name) ?? Infinity;
        const indexB = map.get(b.name) ?? Infinity;
        return indexA - indexB;
    };
}

function adjustAndCleanMenu(menuItem, list) {
    const menuList = JSON.parse(JSON.stringify(list));
    const orderMap = new Map();
    menuItem.forEach((item, index) => {
        orderMap.set(item.label, index);
    });
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
    newMenu.sort((a, b) => {
        const indexA = orderMap.get(a.name) ?? Infinity;
        const indexB = orderMap.get(b.name) ?? Infinity;
        return indexA - indexB;
    });
    for (const menu of newMenu) {
        if (menu.children?.length === 1) {
            menu.meta.icon = menu.children[0].meta.icon;
            menu.meta.title = menu.children[0].meta.title;
        }
    }

    return newMenu;
}

onMounted(() => {
    menuStore.setMenuList(menuList);
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
