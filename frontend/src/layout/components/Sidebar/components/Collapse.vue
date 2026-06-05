<template>
    <div>
        <el-popover
            v-model:visible="popoverVisible"
            placement="right-end"
            :show-arrow="false"
            :offset="0"
            :width="200"
            trigger="click"
            @before-enter="showPopover"
            popper-class="custom-popover-dropdown"
        >
            <template #reference>
                <div class="el-dropdown-link" v-if="!menuStore.isCollapse">
                    <el-badge is-dot :value="taskCount" :show-zero="false" :offset="[5, 5]">
                        <el-button link>
                            <SvgIcon class="icon" iconName="p-pcm" />
                            <span class="ellipsis-text">{{ loadCurrentName() }}</span>
                        </el-button>
                    </el-badge>
                </div>
                <div v-else class="el-dropdown-link">
                    <el-badge is-dot :value="taskCount" :show-zero="false" :offset="[-5, 5]">
                        <SvgIcon class="icon" iconName="p-pcm" />
                    </el-badge>
                </div>
            </template>
            <div class="dropdown-menu" v-loading="loading || switchingNode">
                <div class="dropdown-item" v-if="currentUser" @click="changeUserInfo">
                    <SvgIcon class="icon" iconName="p-gerenzhongxin1" />
                    {{ currentUser.name }}
                </div>
                <el-divider class="divider" />

                <div class="dropdown-item" @click="openTask">
                    <SvgIcon class="icon" iconName="p-renwuzhongxin1" />
                    {{ $t('menu.msgCenter') }}
                    <el-tag class="msg-tag" v-if="taskCount !== 0" size="small" round>{{ taskCount }}</el-tag>
                </div>
                <el-divider v-if="showNodes()" class="divider" />
                <div class="dropdown-item" @click="openNodeDashboard" v-if="isXpackOrEE">
                    <SvgIcon class="icon" iconName="p-gailan1" />
                    {{ $t('xpack.node.multiOverview') }}
                </div>
                <el-divider v-if="isXpackOrEE" class="divider" />

                <div v-if="showNodes()">
                    <el-scrollbar max-height="288px" :noresize="true">
                        <div
                            class="dropdown-item"
                            @click="changeNode(item.name)"
                            :disabled="item.status !== 'Healthy'"
                            v-for="item in visibleNodeOptions"
                            :key="item.name"
                        >
                            <SvgIcon class="icon" iconName="p-zhuji" />
                            <span class="node-name">{{ displayNodeName(item) }}</span>
                            <el-tooltip
                                v-if="item.status !== 'Healthy' || !item.isBound"
                                :content="item.isBound ? $t('xpack.node.nodeUnhealthy') : $t('xpack.node.nodeUnbind')"
                                placement="right"
                            >
                                <el-icon class="icon-status" type="danger">
                                    <Warning />
                                </el-icon>
                            </el-tooltip>
                        </div>
                    </el-scrollbar>
                    <div v-if="showMoreNodes" class="dropdown-item more-node-button" @click.stop="openNodeDrawer">
                        <span class="more-node-label">{{ $t('tabs.more') }}</span>
                        <span class="more-node-count">+{{ nodeOptions.length - defaultNodeLimit }}</span>
                        <el-icon class="more-node-arrow">
                            <ArrowRight />
                        </el-icon>
                    </div>
                </div>
                <el-divider class="divider" />
                <div class="dropdown-item" @click="logout">
                    <SvgIcon class="icon" iconName="p-tuichudenglu3" />
                    {{ $t('commons.login.logout') }}
                </div>
            </div>
        </el-popover>
        <NodeDrawer
            v-model="nodeDrawerVisible"
            :nodes="nodeOptions"
            :master-alias="globalStore.getMasterAlias()"
            @change="changeNode"
            @favorite-change="handleFavoriteChange"
        />
        <UserInfo ref="userInfoRef" :currentUser="currentUser" @search="loadCurrentUser()" />
    </div>
</template>

<script setup lang="ts">
import { MenuStore } from '@/store';
import { countExecutingTask } from '@/api/modules/log';
import { MsgError, MsgSuccess } from '@/utils/message';
import i18n from '@/lang';
import { getAgentSettingInfo } from '@/api/modules/setting';
import { computed, onMounted, ref } from 'vue';
import bus from '@/global/bus';
import { logOutApi } from '@/api/modules/auth';
import router from '@/routers';
import { loadProductProFromDB } from '@/utils/xpack';
import { routerToNameWithQuery } from '@/utils/router';
import { changeToLocal, listNodes, setDefaultNodeInfo } from '@/utils/node';
import { Login } from '@/api/interface/auth';
import { syncAuthInfo } from '@/utils/rbac';
import UserInfo from './user-info/index.vue';
import NodeDrawer from './node-drawer/index.vue';
import { useGlobalStore } from '@/composables/useGlobalStore';

const currentUser = ref<Login.AuthInfo>();
const { globalStore, currentNode, currentNodeAddr, defaultNetwork, entrance, isEnterprise, isXpackOrEE } =
    useGlobalStore();
const menuStore = MenuStore();
const nodes = ref([]);
const nodeOptions = ref([]);
const loading = ref();
const switchingNode = ref(false);
const popoverVisible = ref(false);
const nodeDrawerVisible = ref(false);
const userInfoRef = ref();
const props = defineProps({
    version: String,
});

const defaultNodeLimit = 8;

const emit = defineEmits(['openTask', 'refresh']);
bus.on('refreshTask', () => {
    checkTask();
});

const loadCurrentName = () => {
    if (currentNode.value) {
        if (currentNode.value === 'local') {
            return globalStore.getMasterAlias();
        }
        return currentNode.value;
    }
    return globalStore.getMasterAlias();
};

const visibleNodeOptions = computed(() => {
    return nodeOptions.value.slice(0, defaultNodeLimit);
});
const showMoreNodes = computed(() => {
    return nodeOptions.value.length > defaultNodeLimit;
});

const showPopover = async () => {
    await loadNodes();
};

const displayNodeName = (item) => {
    return item.name === 'local' ? globalStore.getMasterAlias() : item.name;
};

const openNodeDrawer = () => {
    nodeDrawerVisible.value = true;
    popoverVisible.value = false;
};

const handleFavoriteChange = async () => {
    await loadNodes();
};

const loadNodes = async () => {
    loading.value = true;
    nodes.value = [];
    if (!isXpackOrEE.value) {
        changeToLocal();
        loading.value = false;
        return;
    }
    await listNodes('all')
        .then((res) => {
            nodes.value = res || [];
            if (nodes.value.length === 0) {
                setDefaultNodeInfo();
            }
            nodeOptions.value = nodes.value || [];
            loading.value = false;
        })
        .catch(() => {
            setDefaultNodeInfo();
            loading.value = false;
        });
};
const changeNode = async (command: string) => {
    if (currentNode.value === command || switchingNode.value) {
        return;
    }
    switchingNode.value = true;
    try {
        for (const item of nodes.value) {
            if (item.name == command) {
                if (command == 'local') {
                    if (isEnterprise.value) {
                        await loadCurrentUser('local');
                    }
                    await loadGlobalSetting('local');
                    currentNode.value = 'local';
                    currentNodeAddr.value = item.addr;
                    localStorage.removeItem('dashboardCache');
                    localStorage.removeItem('upgradeChecked');
                    menuStore.setMenuList([]);
                    emit('refresh');
                    loadProductProFromDB();
                    routerToNameWithQuery('home', { t: Date.now() });
                    return;
                }
                if (!item.isBound) {
                    MsgError(i18n.global.t('xpack.node.nodeUnbindHelper'));
                    return;
                }
                if (item.status !== 'Healthy') {
                    MsgError(i18n.global.t('xpack.node.nodeUnhealthyHelper'));
                    return;
                }
                if (props.version != item.version) {
                    MsgError(i18n.global.t('setting.versionNotSame'));
                    return;
                }
                await loadGlobalSetting(command);
                localStorage.removeItem('dashboardCache');
                localStorage.removeItem('upgradeChecked');
                currentNode.value = command;
                currentNodeAddr.value = item.addr;
                if (isEnterprise.value) {
                    await loadCurrentUser(command);
                }
                menuStore.setMenuList([]);
                emit('refresh');
                loadProductProFromDB();
                routerToNameWithQuery('home', { t: Date.now() });
                return;
            }
        }
    } finally {
        switchingNode.value = false;
    }
};

const loadGlobalSetting = async (currentNode?: string) => {
    await getAgentSettingInfo(currentNode).then((res) => {
        defaultNetwork.value = res.data.defaultNetwork;
    });
};

const showNodes = () => {
    return nodes.value.length > 0 && isXpackOrEE.value;
};

const taskCount = ref(0);
const checkTask = async () => {
    try {
        const res = await countExecutingTask();
        taskCount.value = res.data;
    } catch (error) {}
};

const openTask = () => {
    emit('openTask');
};

const openNodeDashboard = () => {
    routerToNameWithQuery('NodeDashboard', { uncached: 'true' });
};

const logout = () => {
    ElMessageBox.confirm(i18n.global.t('commons.msg.sureLogOut'), i18n.global.t('commons.msg.infoTitle'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'warning',
    })
        .then(async () => {
            await logOutApi();
            globalStore.setLogStatus(false);
            globalStore.clearAuthInfo();
            router.push({ name: 'entrance', params: { code: entrance.value } });
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {});
};

const loadCurrentUser = async (currentNode?: string) => {
    const authInfo = await syncAuthInfo(currentNode);
    if (authInfo) {
        currentUser.value = authInfo;
    }
};
const changeUserInfo = () => {
    loadCurrentUser();
    userInfoRef.value?.openDrawer();
};

onMounted(() => {
    loadNodes();
    checkTask();
    loadCurrentUser();
});
</script>

<style scoped lang="scss">
@use '../index';

.el-dropdown-link {
    display: flex;
    align-items: center;
    box-sizing: border-box;
    border-top: 1px solid var(--panel-footer-border);
    height: 48px;
    .icon {
        margin-left: 25px;
        font-size: 8px;
        margin-right: 7px;
        color: var(--panel-main-bg-color-1);
    }
    &:hover {
        .icon {
            color: var(--el-color-primary);
        }
        .el-button {
            color: var(--el-color-primary);
        }
    }
}
.custom-popover-dropdown {
    padding: 0 !important;
    border: 1px solid #e4e7ed !important;
    box-shadow: 0 2px 8px 0 rgba(0, 0, 0, 0.1) !important;
    background-color: var(--el-menu-item-bg-color);
    .divider {
        display: block;
        height: 1px;
        width: 91%;
        margin: 3px 8px;
        border-top: 1px var(--el-border-color) var(--el-border-style);
    }
}

.dropdown-menu {
    min-width: 120px;
}

.dropdown-item {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 5px 8px;
    cursor: pointer;
    line-height: 26px;
    transition: background 0.3s;
    .icon {
        font-size: 8px;
    }
    .icon-status {
        font-size: 16px;
    }
    .msg-tag {
        margin-top: 3px;
        float: right;
        background-color: transparent;
        color: var(--panel-main-bg-color-1);
    }
    &:hover {
        color: var(--el-color-primary);
        .icon {
            color: var(--el-color-primary);
        }
        .msg-tag {
            color: var(--el-color-primary);
        }
    }
}
.node-name {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
.more-node-button {
    margin: 4px 8px 0;
    padding: 5px 8px 5px 26px;
    border-radius: 4px;
    color: var(--el-color-primary);
    font-size: 12px;
    line-height: 24px;
}
.more-node-label {
    flex: 1;
    min-width: 0;
}
.more-node-count {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 24px;
    height: 18px;
    padding: 0 7px;
    border-radius: 9px;
    color: var(--el-color-primary);
    line-height: 18px;
    font-size: 12px;
    font-weight: 500;
}
.more-node-arrow {
    margin-left: 2px;
    font-size: 12px;
}
.ellipsis-text {
    display: inline-block;
    max-width: 120px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
</style>
