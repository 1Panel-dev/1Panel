<template>
    <div
        class="sidebar-container"
        element-loading-text="Loading..."
        :element-loading-spinner="loadingSvg"
        element-loading-svg-view-box="-10, -10, 50, 50"
        element-loading-background="rgba(122, 122, 122, 0.01)"
    >
        <Logo :isCollapse="isCollapse" />
        <div class="el-dropdown-link flex justify-between items-center">
            <el-button link class="ml-4" @click="openChangeNode" @mouseenter="openChangeNode">
                {{ loadCurrentName() }}
            </el-button>
            <div>
                <el-dropdown
                    ref="nodeChangeRef"
                    trigger="contextmenu"
                    v-if="nodes.length > 0"
                    placement="right-start"
                    @command="changeNode"
                >
                    <span></span>
                    <template #dropdown>
                        <el-dropdown-menu>
                            <el-dropdown-item command="local">
                                <el-button link icon="CircleCheck" type="success" />
                                {{ $t('terminal.local') }}
                            </el-dropdown-item>
                            <el-dropdown-item v-for="item in nodes" :key="item.name" :command="item.name">
                                <el-button v-if="item.status === 'Healthy'" link icon="CircleCheck" type="success" />
                                <el-button v-else link icon="Warning" type="danger" />
                                {{ item.name }}
                            </el-dropdown-item>
                        </el-dropdown-menu>
                    </template>
                </el-dropdown>
            </div>
            <el-tag type="danger" size="small" effect="light" class="mr-2" @click="openTask">{{ taskCount }}</el-tag>
        </div>
        <el-scrollbar>
            <el-menu
                :default-active="activeMenu"
                :router="menuRouter"
                :collapse="isCollapse"
                :collapse-transition="false"
                :unique-opened="true"
                @select="handleMenuClick"
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
import { ref, computed, onMounted, defineEmits } from 'vue';
import { RouteRecordRaw, useRoute } from 'vue-router';
import { loadingSvg } from '@/utils/svg';
import Logo from './components/Logo.vue';
import Collapse from './components/Collapse.vue';
import SubItem from './components/SubItem.vue';
import router, { menuList } from '@/routers/router';
import { logOutApi } from '@/api/modules/auth';
import i18n from '@/lang';
import { DropdownInstance, ElMessageBox } from 'element-plus';
import { GlobalStore, MenuStore } from '@/store';
import { MsgError, MsgSuccess } from '@/utils/message';
import { isString } from '@vueuse/core';
import { getSettingInfo, listNodeOptions } from '@/api/modules/setting';
import { countExecutingTask } from '@/api/modules/log';
import { compareVersion } from '@/utils/version';
import bus from '@/global/bus';

const route = useRoute();
const menuStore = MenuStore();
const globalStore = GlobalStore();
const nodes = ref([]);
const nodeChangeRef = ref<DropdownInstance>();
const version = ref();

bus.on('refreshTask', () => {
    console.log('on bus message');
    checkTask();
});

defineProps({
    menuRouter: {
        type: Boolean,
        default: true,
        required: false,
    },
});
const activeMenu = computed(() => {
    const { meta, path } = route;
    return isString(meta.activeMenu) ? meta.activeMenu : path;
});
const isCollapse = computed((): boolean => menuStore.isCollapse);

let routerMenus = computed((): RouteRecordRaw[] => {
    return menuStore.menuList.filter((route) => route.meta && !route.meta.hideInSidebar);
});

const openChangeNode = () => {
    nodeChangeRef.value?.handleOpen();
};

const loadCurrentName = () => {
    if (globalStore.currentNode) {
        return globalStore.currentNode === 'local' ? i18n.global.t('terminal.local') : globalStore.currentNode;
    }
    return i18n.global.t('terminal.local');
};

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
const emit = defineEmits(['menuClick', 'openTask']);
const handleMenuClick = (path) => {
    emit('menuClick', path);
};
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

const loadNodes = async () => {
    await listNodeOptions()
        .then((res) => {
            if (!res) {
                nodes.value = [];
                return;
            }
            nodes.value = res.data || [];
            if (nodes.value.length === 0) {
                globalStore.currentNode = 'local';
            }
        })
        .catch(() => {
            nodes.value = [];
        });
};
const changeNode = (command: string) => {
    if (globalStore.currentNode === command) {
        return;
    }
    if (command == 'local') {
        globalStore.currentNode = command || 'local';
        location.reload();
        return;
    }
    for (const item of nodes.value) {
        if (item.name == command) {
            if (version.value == item.version) {
                globalStore.currentNode = command || 'local';
                location.reload();
                return;
            }
            let compareItem = compareVersion(item.version, version.value);
            if (compareItem) {
                MsgError(i18n.global.t('setting.versionHigher', [command]));
                return;
            }
            if (!compareItem) {
                MsgError(i18n.global.t('setting.versionLower', [command]));
                return;
            }
        }
    }
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
    version.value = res.data.systemVersion;
    const json: Node = JSON.parse(res.data.xpackHideMenu);
    if (json.isCheck === false) {
        json.children.forEach((child: any) => {
            if (child.isCheck === true) {
                child.isCheck = false;
            }
        });
    }
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

const taskCount = ref(0);
const checkTask = async () => {
    try {
        const res = await countExecutingTask();
        taskCount.value = res.data;
    } catch (error) {
        console.error(error);
    }
};

const openTask = () => {
    emit('openTask');
};

onMounted(() => {
    menuStore.setMenuList(menuList);
    search();
    loadNodes();
    checkTask();
});
</script>

<style lang="scss">
@use './index.scss';

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

.el-dropdown-link {
    margin-top: -5px;
    margin-left: 15px;
    font-size: 14px;
    font-weight: 500;
    color: var(--el-color-primary);
    height: 38px;
}
.ico {
    height: 20px !important;
}
</style>
