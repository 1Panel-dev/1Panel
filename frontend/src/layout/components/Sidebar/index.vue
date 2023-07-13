<template>
    <div
        class="sidebar-container"
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
                popper-class="sidebar-container-popper"
            >
                <SubItem :menuList="routerMenus"></SubItem>
                <el-menu-item :index="''">
                    <el-icon @click="logout">
                        <SvgIcon :iconName="'p-logout'" :className="'svg-icon'"></SvgIcon>
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
import { MenuStore } from '@/store/modules/menu';
import { loadingSvg } from '@/utils/svg';
import Logo from './components/Logo.vue';
import Collapse from './components/Collapse.vue';
import SubItem from './components/SubItem.vue';
import router, { menuList } from '@/routers/router';
import { logOutApi } from '@/api/modules/auth';
import i18n from '@/lang';
import { ElMessageBox } from 'element-plus';
import { GlobalStore } from '@/store';
import { MsgSuccess } from '@/utils/message';
const route = useRoute();
const menuStore = MenuStore();
const globalStore = GlobalStore();
const activeMenu = computed((): string => {
    const { meta, path } = route;
    if (typeof meta.activeMenu === 'string') {
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
onMounted(async () => {
    menuStore.setMenuList(menuList);
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
    transition: all 0.3s ease;

    .el-scrollbar {
        height: calc(100% - 55px);
        .el-menu {
            flex: 1;
            overflow: auto;
            overflow-x: hidden;
            border-right: none;
        }
    }
    .sidebar-container-footer {
        height: 30px;
        background-color: #c0c0c0;
        text-align: center;
    }
}
</style>
