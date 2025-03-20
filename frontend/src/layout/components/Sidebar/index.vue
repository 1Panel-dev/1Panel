<template>
    <div
        class="sidebar-container"
        element-loading-text="Loading..."
        :element-loading-spinner="loadingSvg"
        element-loading-svg-view-box="-10, -10, 50, 50"
        element-loading-background="rgba(122, 122, 122, 0.01)"
    >
        <div class="fixed">
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
                <SubItem :menuList="routerMenus" />
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
    const menuItem = JSON.parse(res.data.hideMenu);
    const showMap = new Map();
    getCheckedLabels(menuItem, showMap);
    let rstMenuList: RouteRecordRaw[] = [];
    for (const menu of menuStore.menuList) {
        let menuItem = JSON.parse(JSON.stringify(menu));
        if (!showMap[menuItem.name]) {
            continue;
        } else if (menuItem.name === 'Xpack-Menu') {
            menuItem.meta.hideInSidebar = false;
        }
        let itemChildren = [];
        for (const item of menuItem.children) {
            if (item.name === 'XAlertDashboard' && globalStore.isIntl) {
                continue;
            }
            if (showMap[item.name]) {
                itemChildren.push(item);
            }
        }
        if (itemChildren.length === 1) {
            menuItem.meta.title = itemChildren[0].meta.title;
        }
        menuItem.children = itemChildren;
        rstMenuList.push(menuItem);
    }
    menuStore.menuList = rstMenuList;
};

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
