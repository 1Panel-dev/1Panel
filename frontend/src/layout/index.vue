<template>
    <div
        :class="classObj"
        class="app-wrapper relative"
        v-loading="loading"
        :element-loading-text="loadingText"
        fullscreen
    >
        <div v-if="classObj.mobile && classObj.openSidebar" class="drawer-bg" @click="handleClickOutside" />
        <el-affix v-if="!classObj.mobile" :offset="classObj.openMenuTabs ? 8 : 15" class="affix">
            <el-tooltip :content="menuStore.isCollapse ? $t('commons.button.expand') : $t('commons.button.collapse')">
                <el-button
                    size="small"
                    circle
                    :style="{ 'margin-left': menuStore.isCollapse ? '63px' : '168px', position: 'absolute' }"
                    :icon="menuStore.isCollapse ? 'ArrowRight' : 'ArrowLeft'"
                    plain
                    @click="handleCollapse()"
                ></el-button>
            </el-tooltip>
        </el-affix>
        <div class="app-sidebar" v-if="!globalStore.isFullScreen">
            <Sidebar @menu-click="handleMenuClick" :menu-router="!classObj.openMenuTabs" @open-task="openTask" />
        </div>

        <div class="main-container">
            <mobile-header v-if="classObj.mobile" />
            <Tabs v-if="classObj.openMenuTabs" />
            <el-watermark
                v-if="globalStore.isMasterProductPro && globalStore.watermark"
                class="app-main"
                :content="loadContent()"
                :font="{
                    fontSize: globalStore.watermark.fontSize,
                    color: globalStore.watermark.color,
                }"
                :rotate="globalStore.watermark.rotate"
                :gap="[globalStore.watermark.gap, globalStore.watermark.gap]"
            >
                <app-main :keep-alive="classObj.openMenuTabs ? tabsStore.cachedTabs : null" />
            </el-watermark>
            <app-main class="app-main" v-else :keep-alive="classObj.openMenuTabs ? tabsStore.cachedTabs : null" />
            <Footer class="app-footer" v-if="!globalStore.isFullScreen" />
            <TaskList ref="taskListRef" />
        </div>
    </div>
</template>

<script setup lang="ts">
import { onMounted, computed, ref, watch, onBeforeUnmount } from 'vue';
import { Sidebar, Footer, AppMain, MobileHeader, Tabs } from './components';
import useResize from './hooks/useResize';
import { GlobalStore, MenuStore, TabsStore } from '@/store';
import { DeviceType } from '@/enums/app';
import { getSystemAvailable } from '@/api/modules/setting';
import { useRoute, useRouter } from 'vue-router';
import { loadMasterProductProFromDB, loadProductProFromDB } from '@/utils/xpack';
import { useTheme } from '@/global/use-theme';
import TaskList from '@/components/task-list/index.vue';
const { switchTheme } = useTheme();

useResize();

const taskListRef = ref();
const openTask = () => {
    taskListRef.value.acceptParams();
    if (globalStore.isMobile()) {
        menuStore.setCollapse();
    }
};
const router = useRouter();
const route = useRoute();
const menuStore = MenuStore();
const globalStore = GlobalStore();
const tabsStore = TabsStore();

const loading = ref(false);
const loadingText = ref();

let timer: NodeJS.Timer | null = null;

const classObj = computed(() => {
    return {
        fullScreen: globalStore.isFullScreen,
        hideSidebar: menuStore.isCollapse,
        openSidebar: !menuStore.isCollapse,
        mobile: globalStore.device === DeviceType.Mobile,
        openMenuTabs: globalStore.openMenuTabs,
        withoutAnimation: menuStore.withoutAnimation,
    };
});
const handleClickOutside = () => {
    menuStore.closeSidebar(false);
};

const handleCollapse = () => {
    menuStore.setCollapse();
};

const loadContent = () => {
    let itemName = globalStore.watermark.content.replaceAll(
        '${nodeName}',
        globalStore.currentNode === 'local' ? globalStore.masterAlias : globalStore.currentNode,
    );
    itemName = itemName.replaceAll('${nodeAddr}', globalStore.currentNodeAddr);
    return itemName;
};

watch(
    () => globalStore.isLoading,
    () => {
        if (globalStore.isLoading) {
            loadStatus();
        } else {
            loading.value = globalStore.isLoading;
        }
    },
);
const handleMenuClick = async (path) => {
    await router.push({ path: path });
    tabsStore.addTab(route);
    tabsStore.activeTabPath = route.path;
};

const toLogin = () => {
    let baseUrl = window.location.origin;
    let newUrl = '';
    if (globalStore.entrance) {
        newUrl = baseUrl + '/' + globalStore.entrance;
    } else {
        newUrl = baseUrl + '/login';
    }
    window.open(newUrl, '_self');
};

const loadStatus = async () => {
    loading.value = globalStore.isLoading;
    loadingText.value = globalStore.loadingText;
    if (loading.value) {
        timer = setInterval(async () => {
            await getSystemAvailable()
                .then((res) => {
                    if (res) {
                        toLogin();
                        clearInterval(Number(timer));
                        timer = null;
                    }
                })
                .catch(() => {
                    toLogin();
                    clearInterval(Number(timer));
                    timer = null;
                });
        }, 1000 * 10);
    }
};
onBeforeUnmount(() => {
    clearInterval(Number(timer));
    timer = null;
});
onMounted(() => {
    if (globalStore.openMenuTabs && !tabsStore.activeTabPath) {
        handleMenuClick('/');
    }

    loadStatus();
    loadProductProFromDB();
    loadMasterProductProFromDB();
    globalStore.isFullScreen = false;

    const mqList = window.matchMedia('(prefers-color-scheme: dark)');
    if (mqList.addEventListener) {
        mqList.addEventListener('change', () => {
            switchTheme();
        });
    } else if (mqList.addListener) {
        mqList.addListener(() => {
            switchTheme();
        });
    }
});
</script>

<style scoped lang="scss">
.app-wrapper {
    position: relative;
    width: 100%;
}

.drawer-bg {
    background-color: #000;
    opacity: 0.3;
    width: 100%;
    top: 0;
    height: 100%;
    position: absolute;
    z-index: 999;
}

.main-container {
    display: flex;
    flex-direction: column;
    position: relative;
    height: 100vh;
    transition: margin-left 0.3s;
    margin-left: var(--panel-menu-width);
    background-color: var(--panel-main-bg-color-9);
    overflow-x: hidden;
}
.app-main {
    padding: 7px 20px;
    flex: 1;
    overflow: auto;
}
.app-sidebar {
    z-index: 2;
    transition: width 0.3s;
    width: var(--panel-menu-width) !important;
    position: fixed;
    font-size: 0px;
    top: 0;
    bottom: 0;
    left: 0;
    overflow: hidden;
    .affix {
        z-index: 5;
    }
}

.hideSidebar {
    .main-container {
        margin-left: var(--panel-menu-hide-width);
    }
    .app-sidebar {
        width: var(--panel-menu-hide-width) !important;
    }
    .fixed-header {
        width: calc(100% - var(--panel-menu-hide-width));
    }
}

.fullScreen {
    .main-container {
        margin-left: 0px;
    }
}
// for mobile response 适配移动端
.mobile {
    .main-container {
        margin-left: 0px;
    }
    .app-sidebar {
        transition: transform 0.3s;
        width: var(--panel-menu-width) !important;
        background: #ffffff;
        z-index: 9999;
    }
    .app-footer {
        display: block;
        text-align: center;
    }
    &.openSidebar {
        position: fixed;
        top: 0;
    }
    &.hideSidebar {
        .app-sidebar {
            pointer-events: none;
            transition-duration: 0.3s;
            transform: translate3d(calc(0px - var(--panel-menu-width)), 0, 0);
        }
    }
}

.withoutAnimation {
    .main-container,
    .sidebar-container {
        transition: none;
    }
}
</style>
