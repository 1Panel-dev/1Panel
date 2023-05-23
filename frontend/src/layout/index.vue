<template>
    <div :class="classObj" class="app-wrapper">
        <div v-if="classObj.mobile && classObj.openSidebar" class="drawer-bg" @click="handleClickOutside" />
        <div class="app-sidebar">
            <Sidebar />
        </div>

        <div class="main-container">
            <mobile-header v-if="classObj.mobile" />
            <app-main class="app-main" />

            <Footer class="app-footer" />
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { Sidebar, Footer, AppMain, MobileHeader } from './components';
import useResize from './hooks/useResize';
import { GlobalStore } from '@/store';
import { MenuStore } from '@/store/modules/menu';
import { DeviceType } from '@/enums/app';
useResize();

const menuStore = MenuStore();
const globalStore = GlobalStore();

const classObj = computed(() => {
    return {
        hideSidebar: menuStore.isCollapse,
        openSidebar: !menuStore.isCollapse,
        mobile: globalStore.device === DeviceType.Mobile,
        withoutAnimation: menuStore.withoutAnimation,
    };
});
const handleClickOutside = () => {
    menuStore.closeSidebar(false);
};
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
    flex: 1;
    flex-basis: auto;
    position: relative;
    min-height: 100%;
    height: calc(100vh);
    transition: margin-left 0.28s;
    margin-left: var(--panel-menu-width);
    background-color: #f4f4f4;
    overflow-x: hidden;
}
.app-main {
    padding: 20px;
    flex: 1;
    flex-basis: auto;
    overflow: auto;
}
.app-sidebar {
    transition: width 0.28s;
    width: var(--panel-menu-width) !important;
    height: 100%;
    position: fixed;
    font-size: 0px;
    top: 0;
    bottom: 0;
    left: 0;
    z-index: 1001;
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
// for mobile response 适配移动端
.mobile {
    .main-container {
        margin-left: 0px;
    }
    .app-sidebar {
        transition: transform 0.28s;
        width: var(--panel-menu-width) !important;
        background: #ffffff;
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
