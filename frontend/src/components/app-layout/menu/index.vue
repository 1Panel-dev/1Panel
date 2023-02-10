<template>
    <div
        class="menu"
        :style="{ width: isCollapse ? '75px' : '180px' }"
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
                popper-class="menu-popper"
            >
                <SubItem :menuList="routerMenus"></SubItem>
                <el-menu-item>
                    <el-icon @click="logout">
                        <SvgIcon :iconName="'p-logout'" :className="'svg-icon'"></SvgIcon>
                    </el-icon>
                    <template #title>
                        <span @click="logout">{{ $t('commons.header.logout') }}</span>
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
import SubItem from './components/sub-item.vue';
import router, { menuList } from '@/routers/router';
import { logOutApi } from '@/api/modules/auth';
import i18n from '@/lang';
import { ElMessageBox, ElMessage } from 'element-plus';
import { GlobalStore } from '@/store';
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
    }).then(() => {
        systemLogOut();
        router.push({ name: 'login' });
        globalStore.setLogStatus(false);
        ElMessage({
            type: 'success',
            message: i18n.global.t('commons.msg.operationSuccess'),
        });
    });
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
.menu {
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
    .menu-footer {
        height: 30px;
        background-color: #c0c0c0;
        text-align: center;
    }
}
</style>
