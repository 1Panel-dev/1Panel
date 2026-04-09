<template>
    <div>
        <el-popover
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
            <div class="dropdown-menu" v-loading="loading">
                <div class="dropdown-item" v-if="currentUser" @click="changeUserInfo">
                    <SvgIcon class="icon" iconName="p-gerenzhongxin1" />
                    {{ currentUser.name }}
                </div>
                <el-divider class="divider" />

                <div class="dropdown-item" @click="openTask">
                    <div class="node">
                        <SvgIcon class="icon" iconName="p-renwuzhongxin1" />
                        {{ $t('menu.msgCenter') }}
                    </div>
                    <el-tag class="msg-tag" v-if="taskCount !== 0" size="small" round>{{ taskCount }}</el-tag>
                </div>
                <el-divider v-if="showNodes()" class="divider" />
                <div class="dropdown-item" @click="openNodeDashboard" v-if="isXpackOrEE">
                    <SvgIcon class="icon" iconName="p-gailan1" />
                    {{ $t('xpack.node.multiOverview') }}
                </div>
                <el-divider v-if="isXpackOrEE" class="divider" />

                <div v-if="showNodes()">
                    <el-scrollbar max-height="168px" :noresize="true">
                        <div
                            class="dropdown-item"
                            @click="changeNode(item.name)"
                            :disabled="item.status !== 'Healthy'"
                            v-for="item in nodeOptions"
                            :key="item.name"
                        >
                            <div class="node">
                                <SvgIcon class="icon" iconName="p-zhuji" />
                                {{ item.name === 'local' ? globalStore.getMasterAlias() : item.name }}
                                <el-tooltip
                                    v-if="item.status !== 'Healthy' || !item.isBound"
                                    :content="
                                        item.isBound ? $t('xpack.node.nodeUnhealthy') : $t('xpack.node.nodeUnbind')
                                    "
                                    placement="right"
                                >
                                    <el-icon class="icon-status" type="danger">
                                        <Warning />
                                    </el-icon>
                                </el-tooltip>
                            </div>
                        </div>
                    </el-scrollbar>
                </div>
                <el-input
                    v-if="showNodes() && nodes?.length > 5"
                    suffix-icon="Search"
                    v-model="filter"
                    @input="changeFilter"
                    class="w-full filter-input"
                    size="small"
                    clearable
                />
                <el-divider class="divider" />
                <div class="dropdown-item" @click="logout">
                    <SvgIcon class="icon" iconName="p-tuichudenglu3" />
                    {{ $t('commons.login.logout') }}
                </div>
            </div>
        </el-popover>
        <DrawerPro v-model="open" :title="$t('xpack.user.userInfo')">
            <el-form ref="userRef" label-position="top" :model="userForm" :rules="userRules" v-loading="loading">
                <el-form-item :label="$t('commons.login.username')" prop="name">
                    <el-tag type="primary">{{ userForm.name }}</el-tag>
                </el-form-item>
                <el-form-item :label="$t('setting.oldPassword')" prop="oldPassword">
                    <el-input type="password" show-password clearable v-model.trim="userForm.oldPassword" />
                </el-form-item>
                <el-form-item :label="$t('setting.newPassword')" prop="newPassword">
                    <el-input type="password" show-password clearable v-model.trim="userForm.newPassword" />
                </el-form-item>
                <el-form-item :label="$t('setting.retryPassword')" prop="retryPassword">
                    <el-input type="password" show-password clearable v-model.trim="userForm.retryPassword" />
                </el-form-item>
            </el-form>
            <template #footer>
                <el-button :disabled="loading" @click="open = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button :disabled="loading" type="primary" @click="onSubmit(userRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </template>
        </DrawerPro>
    </div>
</template>

<script setup lang="ts">
import { GlobalStore, MenuStore } from '@/store';
import { countExecutingTask } from '@/api/modules/log';
import { MsgError, MsgSuccess } from '@/utils/message';
import i18n from '@/lang';
import { getAgentSettingInfo } from '@/api/modules/setting';
import { ref } from 'vue';
import bus from '@/global/bus';
import { getAuthInfo, logOutApi, updateAuthInfo } from '@/api/modules/auth';
import router from '@/routers';
import { loadProductProFromDB } from '@/utils/xpack';
import { routerToNameWithQuery } from '@/utils/router';
import { changeToLocal, listNodes, setDefaultNodeInfo } from '@/utils/node';
import { Login } from '@/api/interface/auth';
import { Rules } from '@/global/form-rules';

const filter = ref();
const currentUser = ref<Login.AuthInfo>();
const globalStore = GlobalStore();
const menuStore = MenuStore();
const nodes = ref([]);
const nodeOptions = ref([]);
const loading = ref();
const props = defineProps({
    version: String,
});
const isXpackOrEE = computed(() => {
    return globalStore.isXpackOrEE();
});

const open = ref(false);
const userRef = ref();
const userForm = reactive({
    id: 0,
    name: '',
    oldPassword: '',
    newPassword: '',
    retryPassword: '',
});
const userRules = reactive({
    oldPassword: [Rules.requiredInput, Rules.noSpace],
    newPassword: [Rules.requiredInput, Rules.noSpace],
    retryPassword: [Rules.requiredInput, Rules.noSpace, { validator: checkPassword, trigger: 'blur' }],
});
function checkPassword(rule: any, value: any, callback: any) {
    let password = userForm.newPassword;
    if (password !== userForm.retryPassword) {
        return callback(new Error(i18n.global.t('commons.rule.rePassword')));
    }
    callback();
}

const emit = defineEmits(['openTask']);
bus.on('refreshTask', () => {
    checkTask();
});

const loadCurrentName = () => {
    if (globalStore.currentNode) {
        if (globalStore.currentNode === 'local') {
            return globalStore.getMasterAlias();
        }
        return globalStore.currentNode;
    }
    return globalStore.getMasterAlias();
};

const showPopover = () => {
    filter.value = '';
    loadNodes();
    changeFilter();
};

const changeFilter = () => {
    nodeOptions.value = [];
    for (const item of nodes.value) {
        if (item.name.indexOf(filter.value) !== -1) {
            nodeOptions.value.push(item);
        }
    }
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
            nodes.value.sort((a, b) => {
                if (a.name === 'local') return -1;
                if (b.name === 'local') return 1;
                return 0;
            });
            nodeOptions.value = nodes.value || [];
            loading.value = false;
        })
        .catch(() => {
            setDefaultNodeInfo();
            loading.value = false;
        });
};
const changeNode = (command: string) => {
    if (globalStore.currentNode === command) {
        return;
    }
    for (const item of nodes.value) {
        if (item.name == command) {
            if (command == 'local') {
                globalStore.currentNode = 'local';
                globalStore.currentNodeAddr = item.addr;
                loadGlobalSetting();
                localStorage.removeItem('dashboardCache');
                localStorage.removeItem('upgradeChecked');
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
            loadGlobalSetting();
            localStorage.removeItem('dashboardCache');
            localStorage.removeItem('upgradeChecked');
            globalStore.currentNode = command || 'local';
            globalStore.currentNodeAddr = item.addr;
            loadProductProFromDB();
            routerToNameWithQuery('home', { t: Date.now() });
        }
    }
};

const loadGlobalSetting = async () => {
    await getAgentSettingInfo().then((res) => {
        globalStore.defaultNetwork = res.data.defaultNetwork;
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
            router.push({ name: 'entrance', params: { code: globalStore.entrance } });
            globalStore.isLogin = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {});
};

const loadCurrentUser = async () => {
    await getAuthInfo().then((res) => {
        currentUser.value = res.data;
    });
};
const changeUserInfo = () => {
    if (currentUser.value.role === 'ADMIN') {
        return;
    }
    userForm.id = currentUser.value?.id || 0;
    userForm.name = currentUser.value?.name || '';
    open.value = true;
};
const onSubmit = async (formEl: any) => {
    if (!formEl) return;
    formEl.validate(async (valid: boolean) => {
        if (!valid) return;
        if (userForm.newPassword === userForm.oldPassword) {
            MsgError(i18n.global.t('setting.duplicatePassword'));
            return;
        }
        loading.value = true;
        await updateAuthInfo(userForm)
            .then(async () => {
                loading.value = false;
                open.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                await logOutApi();
                router.push({ name: 'entrance', params: { code: globalStore.entrance } });
                globalStore.setLogStatus(false);
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

onMounted(() => {
    loadNodes();
    checkTask();
    if (globalStore.isXpackEE) {
        loadCurrentUser();
    }
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
    padding: 2px 8px;
    cursor: pointer;
    line-height: 26px;
    transition: background 0.3s;
    .icon {
        font-size: 8px;
    }
    .icon-status {
        font-size: 16px;
        margin-left: auto;
    }
    .node {
        display: flex;
        align-items: center;
        gap: 6px;
        padding: 3px 0;
        width: 100%;
    }
    .node-name {
        flex: 1;
        min-width: 0;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
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
.filter-input {
    padding: 0 8px;
    margin-bottom: 4px;
}
.dropdown-item:hover {
    background: var(--el-menu-item-bg-color-active);
}
.ellipsis-text {
    display: inline-block;
    max-width: 120px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
</style>
