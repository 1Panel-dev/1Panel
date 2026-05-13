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
        <div class="app-sidebar" v-if="!isFullScreen">
            <Sidebar @menu-click="handleMenuClick" :menu-router="!classObj.openMenuTabs" @open-task="openTask" />
        </div>

        <el-watermark
            v-if="isXpackOrEE && watermarkShow && watermark"
            :content="loadContent()"
            :font="{
                fontSize: watermark.fontSize,
                color: isDarkTheme ? watermark.darkColor : watermark.lightColor,
                textBaseline: 'top',
            }"
            :rotate="watermark.rotate"
            :gap="[watermark.gap, watermark.gap]"
        >
            <div class="main-container">
                <mobile-header v-if="classObj.mobile" />
                <Tabs v-if="classObj.openMenuTabs" />
                <app-main :keep-alive="classObj.openMenuTabs ? tabsStore.cachedTabs : null" class="app-main" />
                <Footer class="app-footer" v-if="!isFullScreen" />
            </div>
        </el-watermark>
        <div class="main-container" v-else>
            <mobile-header v-if="classObj.mobile" />
            <Tabs v-if="classObj.openMenuTabs" />
            <app-main :keep-alive="classObj.openMenuTabs ? tabsStore.cachedTabs : null" class="app-main" />
            <Footer class="app-footer" v-if="!isFullScreen" />
        </div>
        <TaskList ref="taskListRef" />
    </div>
</template>

<script setup lang="ts">
import { onMounted, computed, ref, watch, onBeforeUnmount } from 'vue';
import { Sidebar, Footer, AppMain, MobileHeader, Tabs } from './components';
import useResize from './hooks/useResize';
import { MenuStore, TabsStore } from '@/store';
import { getSystemAvailable } from '@/api/modules/setting';
import { useRoute, useRouter } from 'vue-router';
import { loadMasterProductProFromDB, loadProductProFromDB } from '@/utils/xpack';
import { useTheme } from '@/global/use-theme';
import TaskList from '@/components/task-list/index.vue';
import i18n from '@/lang';
import { useGlobalStore } from '@/composables/useGlobalStore';

const {
    globalStore,
    currentNode,
    currentNodeAddr,
    entrance,
    isDarkTheme,
    isFullScreen,
    isLoading,
    isMobile,
    isXpackOrEE,
    loadingText: globalLoadingText,
    openMenuTabs,
    watermark,
    watermarkShow,
} = useGlobalStore();
const { switchTheme } = useTheme();

useResize();

const taskListRef = ref();
const openTask = () => {
    taskListRef.value.acceptParams();
    if (isMobile.value) {
        menuStore.setCollapse();
    }
};
const router = useRouter();
const route = useRoute();
const menuStore = MenuStore();

const tabsStore = TabsStore();

const loading = ref(false);
const loadingText = computed(() =>
    globalLoadingText.value ? i18n.global.t(`commons.loadingText.${globalLoadingText.value}`) : '',
);

let timer: NodeJS.Timer | null = null;

const classObj = computed(() => {
    return {
        fullScreen: isFullScreen.value,
        hideSidebar: menuStore.isCollapse,
        openSidebar: !menuStore.isCollapse,
        mobile: isMobile.value,
        openMenuTabs: openMenuTabs.value,
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
    if (!watermark.value) {
        return '';
    }

    let itemName = watermark.value.content.replaceAll(
        '${nodeName}',
        currentNode.value === 'local' ? globalStore.getMasterAlias() : currentNode.value,
    );
    itemName = itemName.replaceAll('${nodeAddr}', currentNodeAddr.value || '127.0.0.1');
    return itemName;
};

watch(
    () => isLoading.value,
    () => {
        if (isLoading.value) {
            loadStatus();
        } else {
            loading.value = isLoading.value;
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
    if (entrance.value) {
        newUrl = baseUrl + '/' + entrance.value;
    } else {
        newUrl = baseUrl + '/login';
    }
    window.open(newUrl, '_self');
};

const loadStatus = async () => {
    loading.value = isLoading.value;
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
    if (openMenuTabs.value && !tabsStore.activeTabPath) {
        handleMenuClick('/');
    }

    loadStatus();
    loadProductProFromDB();
    loadMasterProductProFromDB();
    isFullScreen.value = false;

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
