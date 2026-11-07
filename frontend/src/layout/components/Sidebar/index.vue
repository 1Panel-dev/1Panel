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
            <el-skeleton v-if="!menuStore.isMenuLoaded" :rows="8" animated class="p-4" />
            <el-menu
                v-else
                :default-active="activeMenu"
                :router="false"
                :collapse="isCollapse"
                :collapse-transition="false"
                :unique-opened="!menuAccordion"
                @select="handleMenuClick"
                class="custom-menu"
            >
                <SubItem :menuList="routerMenus" :level="0" />
            </el-menu>
        </el-scrollbar>
        <Collapse
            ref="nodeSelectorRef"
            class="node-selector"
            :version="menuStore.systemVersion"
            @open-task="openTask"
            @node-switching="handleNodeSwitching"
            @user-change="currentUser = $event"
        />
        <div class="user-actions" :class="{ 'is-collapsed': isCollapse }">
            <button
                v-if="currentUser"
                type="button"
                class="user-entry"
                :aria-label="$t('xpack.user.userInfo') + ' ' + currentUser.name"
                @click="nodeSelectorRef?.openUserInfo()"
            >
                <SvgIcon class="icon" iconName="p-gerenzhongxin1" />
                <span v-if="!isCollapse" class="user-name">{{ currentUser.name }}</span>
            </button>
            <button
                type="button"
                class="user-entry logout-entry"
                :aria-label="$t('header.exitSystem')"
                @click="nodeSelectorRef?.logout()"
            >
                <SvgIcon class="icon" iconName="p-tuichudenglu3" />
            </button>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, computed, onBeforeUnmount, onMounted, watch } from 'vue';
import { RouteRecordRaw, useRoute, useRouter } from 'vue-router';
import { loadingSvg } from '@/utils/svg';
import Logo from './components/Logo.vue';
import Collapse from './components/Collapse.vue';
import SubItem from './components/SubItem.vue';
import { MenuStore } from '@/store';
import { loadMenuList } from '@/utils/menu';
import PrimaryMenu from '@/assets/images/menu-bg.svg?component';
import { useGlobalStore } from '@/composables/useGlobalStore';
import type { Login } from '@/api/interface/auth';

const route = useRoute();
const router = useRouter();
const menuStore = MenuStore();
const { currentNode, isAdmin, isNodeAdmin, menuAccordion, permissions, masterOnlyPermissions } = useGlobalStore();
const nodeSelectorRef = ref<InstanceType<typeof Collapse>>();
const currentUser = ref<Login.AuthInfo>();

const activeMenu = computed(() => {
    const { meta, path } = route;
    return typeof meta.activeMenu === 'string' ? meta.activeMenu : path;
});
const isCollapse = computed((): boolean => menuStore.isCollapse);

let routerMenus = computed((): RouteRecordRaw[] => {
    return buildRegisteredMenuList(menuStore.menuList as RouteRecordRaw[]).filter(
        (route) => route.meta && !route.meta.hideInSidebar,
    );
});

function buildRegisteredMenuList(source: RouteRecordRaw[]): RouteRecordRaw[] {
    return source.reduce<RouteRecordRaw[]>((result, item) => {
        if (!item.name || !router.hasRoute(item.name)) {
            return result;
        }

        const menuItem = { ...item };
        if (Array.isArray(item.children)) {
            menuItem.children = buildRegisteredMenuList(item.children);
            if (item.children.length > 0 && menuItem.children.length === 0) {
                return result;
            }
        }
        result.push(menuItem);
        return result;
    }, []);
}

const screenWidth = ref(0);
const handleWindowResize = () => {
    screenWidth.value = document.body.clientWidth;
    if (!isCollapse.value && screenWidth.value < 1200) menuStore.setCollapse();
    if (isCollapse.value && screenWidth.value > 1200) menuStore.setCollapse();
};
window.addEventListener('resize', handleWindowResize);
const emit = defineEmits(['menuClick', 'openTask']);
const handleMenuClick = (path) => {
    emit('menuClick', path);
};

const openTask = () => {
    emit('openTask');
};

onMounted(() => {
    screenWidth.value = document.body.clientWidth;
    if (!isCollapse.value && screenWidth.value < 1200) {
        menuStore.setCollapse();
    }
    if (!menuStore.isMenuLoaded) {
        loadMenuList();
    }
});

onBeforeUnmount(() => {
    window.removeEventListener('resize', handleWindowResize);
});

const menuWatcher = watch(
    [
        () => currentNode.value,
        () => isAdmin.value,
        () => isNodeAdmin.value,
        () => [...permissions.value].sort().join('|'),
        () => [...masterOnlyPermissions.value].sort().join('|'),
    ],
    (values, previousValues) => {
        if (values[0] !== previousValues[0]) {
            menuStore.resetMenuList();
        }
        loadMenuList();
    },
);

const handleNodeSwitching = (switching: boolean) => {
    if (switching) {
        menuWatcher.pause();
    } else {
        menuWatcher.resume();
    }
};
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

    .logo,
    .node-selector {
        flex-shrink: 0;
    }

    .el-scrollbar {
        flex: 1;
        min-height: 0;
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

.user-actions {
    display: flex;
    flex-shrink: 0;
    border-top: 1px solid var(--panel-footer-border);

    &.is-collapsed .user-entry {
        justify-content: center;
        padding: 0;
    }

    &.is-collapsed .logout-entry {
        flex: 1;
    }
}

.user-entry {
    display: flex;
    align-items: center;
    gap: 7px;
    flex: 1;
    min-width: 0;
    position: relative;
    height: 48px;
    padding: 0 0 0 25px;
    border: 0;
    border-radius: 0;
    background: transparent;
    color: var(--panel-main-bg-color-1);
    font-family: inherit;
    font-size: var(--el-font-size-base);
    line-height: 1.5;
    text-align: left;
    cursor: pointer;

    .icon {
        flex-shrink: 0;
        font-size: 8px;
    }

    .user-name {
        min-width: 0;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }

    &:hover,
    &:focus-visible {
        color: var(--el-color-primary);
        background: var(--el-menu-item-bg-color);
    }
}
.logout-entry {
    flex: 0 0 42px;
    justify-content: center;
    padding: 0;
    border-left: 1px solid var(--panel-footer-border);
}
</style>
