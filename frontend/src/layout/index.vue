<template>
    <div :class="classObj" class="app-wrapper" v-loading="loading" :element-loading-text="loadinText" fullscreen>
        <div v-if="classObj.mobile && classObj.openSidebar" class="drawer-bg" @click="handleClickOutside" />
        <div class="app-sidebar" v-if="!globalStore.isFullScreen">
            <Sidebar />
        </div>

        <div class="main-container">
            <mobile-header v-if="classObj.mobile" />
            <app-main class="app-main" />

            <Footer class="app-footer" v-if="!globalStore.isFullScreen" />
        </div>
    </div>
</template>

<script setup lang="ts">
import { onMounted, computed, ref, watch, onBeforeUnmount } from 'vue';
import { Sidebar, Footer, AppMain, MobileHeader } from './components';
import useResize from './hooks/useResize';
import { GlobalStore } from '@/store';
import { MenuStore } from '@/store/modules/menu';
import { DeviceType } from '@/enums/app';
import { useI18n } from 'vue-i18n';
import { useTheme } from '@/hooks/use-theme';
import { getSettingInfo, getSystemAvailable } from '@/api/modules/setting';
useResize();

const menuStore = MenuStore();
const globalStore = GlobalStore();

const i18n = useI18n();
const loading = ref(false);
const loadinText = ref();
const themeConfig = computed(() => globalStore.themeConfig);
const { switchDark } = useTheme();

let timer: NodeJS.Timer | null = null;

const classObj = computed(() => {
    return {
        fullScreen: globalStore.isFullScreen,
        hideSidebar: menuStore.isCollapse,
        openSidebar: !menuStore.isCollapse,
        mobile: globalStore.device === DeviceType.Mobile,
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

const loadDataFromDB = async () => {
    const res = await getSettingInfo();
    document.title = res.data.panelName;
    i18n.locale.value = res.data.language;
    i18n.warnHtmlMessage = false;
    globalStore.entrance = res.data.securityEntrance;
    globalStore.updateLanguage(res.data.language);
    globalStore.setThemeConfig({ ...themeConfig.value, theme: res.data.theme });
    globalStore.setThemeConfig({ ...themeConfig.value, panelName: res.data.panelName });
    switchDark();
};

const updateDarkMode = async (event: MediaQueryListEvent) => {
    const res = await getSettingInfo();
    if (res.data.theme !== 'auto') {
        return;
    }
    globalStore.setThemeConfig({ ...themeConfig.value, theme: event.matches ? 'dark' : 'light' });
    switchDark();
};

const loadStatus = async () => {
    loading.value = globalStore.isLoading;
    loadinText.value = globalStore.loadingText;
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
    loadStatus();
    loadDataFromDB();

    const mqList = window.matchMedia('(prefers-color-scheme: dark)');
    if (mqList.addEventListener) {
        mqList.addEventListener('change', (e) => {
            updateDarkMode(e);
        });
    } else if (mqList.addListener) {
        mqList.addListener((e) => {
            updateDarkMode(e);
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
    background-color: #f4f4f4;
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
