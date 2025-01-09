<template>
    <div :class="classObj" class="app-wrapper" v-loading="loading" :element-loading-text="loadingText" fullscreen>
        <div v-if="classObj.mobile && classObj.openSidebar" class="drawer-bg" @click="handleClickOutside" />
        <div class="app-sidebar" v-if="!globalStore.isFullScreen">
            <Sidebar @menu-click="handleMenuClick" />
        </div>

        <div class="main-container">
            <mobile-header v-if="classObj.mobile" />
            <Tabs v-if="classObj.openMenuTabs" />
            <app-main :keep-alive="classObj.openMenuTabs ? tabsStore.cachedTabs : null" class="app-main" />
            <Footer class="app-footer" v-if="!globalStore.isFullScreen" />
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
import { loadProductProFromDB } from '@/utils/xpack';
import { useTheme } from '@/hooks/use-theme';
const { switchTheme } = useTheme();
useResize();

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

const loadStatus = async () => {
    loading.value = globalStore.isLoading;
    loadingText.value = globalStore.loadingText;
    if (loading.value) {
        timer = setInterval(async () => {
            await getSystemAvailable()
                .then((res) => {
                    if (res) {
                        location.reload();
                        clearInterval(Number(timer));
                        timer = null;
                    }
                })
                .catch(() => {
                    location.reload();
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
    padding: 20px;
    flex: 1;
    overflow: auto;
}
.app-sidebar {
    transition: width 0.3s;
    width: var(--panel-menu-width) !important;
    position: fixed;
    font-size: 0px;
    top: 0;
    bottom: 0;
    left: 0;
    overflow: hidden;
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
