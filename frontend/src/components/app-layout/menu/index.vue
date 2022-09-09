<template>
    <div
        class="menu"
        :style="{ width: isCollapse ? '65px' : '220px' }"
        element-loading-text="Loading..."
        :element-loading-spinner="loadingSvg"
        element-loading-svg-view-box="-10, -10, 50, 50"
        element-loading-background="rgba(122, 122, 122, 0.01)"
    >
        <Logo :isCollapse="isCollapse"></Logo>
        <el-scrollbar>
            <el-menu
                :default-active="activeMenu"
                :router="true"
                :collapse="isCollapse"
                :collapse-transition="false"
                :unique-opened="true"
                background-color="#191a20"
                text-color="#bdbdc0"
                active-text-color="#fff"
            >
                <SubItem :menuList="routerMenus"></SubItem>
            </el-menu>
        </el-scrollbar>
    </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { RouteRecordRaw, useRoute } from 'vue-router';
import { MenuStore } from '@/store/modules/menu';
import { loadingSvg } from '@/utils/svg';
import Logo from './components/logo.vue';
import SubItem from './components/sub-item.vue';
import { menuList } from '@/routers/router';
const route = useRoute();
const menuStore = MenuStore();

onMounted(async () => {
    menuStore.setMenuList(menuList);
});

// const activeMenu = computed((): string => route.path);
const activeMenu = computed(() => {
    const { meta, path } = route;
    if (meta.activeMenu) {
        return meta.activeMenu;
    }
    return path;
});
const isCollapse = computed((): boolean => menuStore.isCollapse);
const routerMenus = computed((): RouteRecordRaw[] => menuStore.menuList);

const screenWidth = ref<number>(0);
const listeningWindow = () => {
    window.onresize = () => {
        return (() => {
            screenWidth.value = document.body.clientWidth;
            if (isCollapse.value === false && screenWidth.value < 1200) menuStore.setCollapse();
            if (isCollapse.value === true && screenWidth.value > 1200) menuStore.setCollapse();
        })();
    };
};
listeningWindow();
</script>

<style scoped lang="scss">
@import './index.scss';
</style>
