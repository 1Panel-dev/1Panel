<template>
    <div
        class="sidebar-container"
        element-loading-text="Loading..."
        :element-loading-spinner="loadingSvg"
        element-loading-svg-view-box="-10, -10, 50, 50"
        element-loading-background="rgba(122, 122, 122, 0.01)"
    >
        <Logo :isCollapse="isCollapse" />
        <el-scrollbar>
            <el-menu
                :default-active="activeMenu"
                :router="true"
                :collapse="isCollapse"
                :collapse-transition="false"
                :unique-opened="true"
            >
                <SubItem :menuList="routerMenus" />
                <el-menu-item :index="''">
                    <el-icon @click="logout">
                        <SvgIcon :iconName="'p-logout'" />
                    </el-icon>
                    <template #title>
                        <span @click="logout">{{ $t('commons.login.logout') }}</span>
                    </template>
                </el-menu-item>
            </el-menu>
        </el-scrollbar>
        <Collapse />
    </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { RouteRecordRaw, useRoute } from 'vue-router';
import { loadingSvg } from '@/utils/svg';
import Logo from './components/Logo.vue';
import Collapse from './components/Collapse.vue';
import SubItem from './components/SubItem.vue';
import router, { menuList } from '@/routers/router';
import { logOutApi } from '@/api/modules/auth';
import i18n from '@/lang';
import { ElMessageBox } from 'element-plus';
import { GlobalStore, MenuStore } from '@/store';
import { MsgSuccess } from '@/utils/message';
import { isString } from '@vueuse/core';
import { getSettingInfo } from '@/api/modules/setting';

const route = useRoute();
const menuStore = MenuStore();
const globalStore = GlobalStore();
const activeMenu = computed(() => {
    const { meta, path } = route;
    return isString(meta.activeMenu) ? meta.activeMenu : path;
});
const isCollapse = computed((): boolean => menuStore.isCollapse);

let routerMenus = computed((): RouteRecordRaw[] => {
    return menuStore.menuList.filter((route) => route.meta && !route.meta.hideInSidebar);
});

const screenWidth = ref(0);

interface Node {
    id: string;
    title: string;
    path?: string;
    label: string;
    isCheck: boolean;
    children?: Node[];
}
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

const logout = () => {
    ElMessageBox.confirm(i18n.global.t('commons.msg.sureLogOut'), i18n.global.t('commons.msg.infoTitle'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'warning',
    })
        .then(() => {
            systemLogOut();
            router.push({ name: 'entrance', params: { code: globalStore.entrance } });
            globalStore.setLogStatus(false);
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {});
};

const systemLogOut = async () => {
    await logOutApi();
};

function extractLabels(node: Node, result: string[]): void {
    if (node.isCheck) {
        result.push(node.label);
    }
    if (node.children) {
        for (const childNode of node.children) {
            extractLabels(childNode, result);
        }
    }
}

function getCheckedLabels(json: Node): string[] {
    let result: string[] = [];
    extractLabels(json, result);
    return result;
}

const search = async () => {
    const res = await getSettingInfo();
    const json: Node = JSON.parse(res.data.xpackHideMenu);
    const checkedLabels = getCheckedLabels(json);
    let rstMenuList: RouteRecordRaw[] = [];
    menuStore.menuList.forEach((item) => {
        let menuItem = JSON.parse(JSON.stringify(item));
        let menuChildren: RouteRecordRaw[] = [];
        if (menuItem.path === '/xpack') {
            if (checkedLabels.length) {
                menuItem.children.forEach((child: any) => {
                    for (const str of checkedLabels) {
                        if (child.name === str) {
                            child.hidden = false;
                        }
                    }
                    if (child.hidden === false) {
                        menuChildren.push(child);
                        if (checkedLabels.length === 2) {
                            menuItem.meta.title = child.meta.title;
                        } else {
                            menuItem.meta.title = 'xpack.menu';
                        }
                    }
                });
                menuItem.meta.hideInSidebar = false;
            }
            menuItem.children = menuChildren as RouteRecordRaw[];
            rstMenuList.push(menuItem);
        } else {
            menuItem.children.forEach((child: any) => {
                if (child.hidden == undefined || child.hidden == false) {
                    menuChildren.push(child);
                }
            });
            menuItem.children = menuChildren as RouteRecordRaw[];
            rstMenuList.push(menuItem);
        }
    });
    menuStore.menuList = rstMenuList;
};

onMounted(() => {
    menuStore.setMenuList(menuList);
    search();
});
</script>

<style lang="scss">
@import './index.scss';

.sidebar-container {
    position: relative;
    display: flex;
    flex-direction: column;
    height: 100%;
    background: url(@/assets/images/menu-bg.png) var(--el-menu-bg-color) no-repeat top;

    .el-scrollbar {
        flex: 1;
        .el-menu {
            overflow: auto;
            overflow-x: hidden;
            border-right: none;
        }
    }
}
</style>
