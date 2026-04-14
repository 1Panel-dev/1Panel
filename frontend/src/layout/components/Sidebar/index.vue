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
import { ref, computed, onMounted, watch } from 'vue';
import { RouteRecordRaw, useRoute } from 'vue-router';
import { loadingSvg } from '@/utils/svg';
import Logo from './components/Logo.vue';
import Collapse from './components/Collapse.vue';
import SubItem from './components/SubItem.vue';
import { menuList } from '@/routers/router';
import { GlobalStore, MenuStore } from '@/store';
import { getSettingInfo } from '@/api/modules/setting';
import PrimaryMenu from '@/assets/images/menu-bg.svg?component';
import { hasPermission } from '@/utils/rbac';

const route = useRoute();
const menuStore = MenuStore();
const globalStore = GlobalStore();
const version = ref();

const activeMenu = computed(() => {
    const { meta, path } = route;
    return typeof meta.activeMenu === 'string' ? meta.activeMenu : path;
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
            const menuItem = buildVisibleMenu(menu, showSet);
            if (menuItem) {
                rstMenuList.push(menuItem);
            }
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

function allowMenuItem(item: RouteRecordRaw) {
    const permission = item.meta?.permission as string | undefined;
    if (!permission) {
        return true;
    }
    const allowed = hasPermission(permission);
    return allowed;
}

function buildVisibleMenu(menu: RouteRecordRaw, showSet: Set<string>): RouteRecordRaw | null {
    const menuItem = JSON.parse(JSON.stringify(menu));
    if (!menuItem?.name || !showSet.has(menuItem.name as string)) {
        return null;
    }
    if (!allowMenuItem(menuItem)) {
        return null;
    }

    const children = Array.isArray(menuItem.children) ? menuItem.children : [];
    if (children.length === 0) {
        return menuItem;
    }

    const visibleChildren = children
        .map((item) => {
            if (item.name === 'Upage' && globalStore.isIntl) {
                return null;
            }
            return buildVisibleMenu(item, showSet);
        })
        .filter(Boolean) as RouteRecordRaw[];

    menuItem.children = visibleChildren;
    if (menuItem.children.length === 0) {
        return null;
    }

    if (menuItem.children.length === 1) {
        menuItem.meta.icon = menuItem.children[0].meta.icon;
        menuItem.meta.title = menuItem.children[0].meta.title;
    }
    if (menuItem.name === 'Xpack-Menu') {
        menuItem.meta.hideInSidebar = false;
    }
    return menuItem;
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

watch(
    () => [globalStore.currentNode, globalStore.permissions.join('|')],
    () => {
        search();
    },
);
</script>

<style lang="scss" scoped>
@use 'index';

.background {
    z-index: 20;
}

.custom-menu :deep(.el-menu-item) {
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
