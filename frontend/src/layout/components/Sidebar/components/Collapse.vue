<template>
    <div>
        <el-popover
            placement="right"
            :show-arrow="false"
            :offset="0"
            :width="200"
            trigger="hover"
            popper-class="custom-popover-dropdown"
        >
            <template #reference>
                <div class="el-dropdown-link" v-if="!menuStore.isCollapse">
                    <el-badge is-dot :value="taskCount" :show-zero="false" :offset="[5, 5]">
                        <el-button link @click="openChangeNode" @mouseenter="openChangeNode">
                            <SvgIcon class="icon" iconName="p-gerenzhongxin1" />
                            {{ loadCurrentName() }}
                        </el-button>
                    </el-badge>
                </div>
                <div v-else class="el-dropdown-link">
                    <el-badge is-dot :value="taskCount" :show-zero="false" :offset="[-5, 5]">
                        <SvgIcon class="icon" iconName="p-gerenzhongxin1" />
                    </el-badge>
                </div>
            </template>
            <div class="dropdown-menu">
                <div class="dropdown-item mb-2" @click="openTask">
                    {{ $t('menu.msgCenter') }}
                    <el-tag class="msg-tag" v-if="taskCount !== 0" size="small" round>{{ taskCount }}</el-tag>
                </div>
                <el-divider v-if="nodes.length > 0 && globalStore.isMasterProductPro" class="divider" />
                <div v-if="nodes.length > 0 && globalStore.isMasterProductPro" class="mt-2 mb-2">
                    <div class="dropdown-item" @click="changeNode('local')">
                        {{ $t('xpack.node.master') }}
                    </div>
                    <div
                        class="dropdown-item mt-1"
                        @click="changeNode(item.name)"
                        :disabled="item.status !== 'Healthy'"
                        v-for="item in nodes"
                        :key="item.name"
                    >
                        {{ item.name }}
                        <el-icon class="icon-status" v-if="item.status !== 'Healthy'" type="danger">
                            <Warning />
                        </el-icon>
                    </div>
                </div>
                <el-divider class="divider" />
                <div class="dropdown-item mt-2" @click="logout">{{ $t('commons.login.logout') }}</div>
            </div>
        </el-popover>
    </div>
</template>

<script setup lang="ts">
import { GlobalStore, MenuStore } from '@/store';
import { DropdownInstance } from 'element-plus';
import { countExecutingTask } from '@/api/modules/log';
import { MsgError, MsgSuccess } from '@/utils/message';
import i18n from '@/lang';
import { listNodeOptions } from '@/api/modules/setting';
import bus from '@/global/bus';
import { logOutApi } from '@/api/modules/auth';
import router from '@/routers';

const globalStore = GlobalStore();
const menuStore = MenuStore();
const nodes = ref([]);
const nodeChangeRef = ref<DropdownInstance>();
const props = defineProps({
    version: String,
});

const emit = defineEmits(['openTask']);
bus.on('refreshTask', () => {
    checkTask();
});

const openChangeNode = () => {
    nodeChangeRef.value?.handleOpen();
};

const loadCurrentName = () => {
    if (globalStore.currentNode) {
        return globalStore.currentNode === 'local' ? i18n.global.t('xpack.node.master') : globalStore.currentNode;
    }
    return i18n.global.t('xpack.node.master');
};

const loadNodes = async () => {
    if (!globalStore.isMasterProductPro) {
        globalStore.currentNode = 'local';
        return;
    }
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
        globalStore.isOffline = false;
        location.reload();
        return;
    }
    for (const item of nodes.value) {
        if (item.name == command) {
            if (props.version == item.version) {
                globalStore.currentNode = command || 'local';
                globalStore.isOffline = item.isOffline;
                location.reload();
                return;
            }
            if (item.version !== props.version) {
                MsgError(i18n.global.t('setting.versionNotSame'));
                return;
            }
        }
    }
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

const logout = () => {
    ElMessageBox.confirm(i18n.global.t('commons.msg.sureLogOut'), i18n.global.t('commons.msg.infoTitle'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'warning',
    })
        .then(async () => {
            await logOutApi();
            router.push({ name: 'entrance', params: { code: globalStore.entrance } });
            globalStore.setLogStatus(false);
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {});
};

onMounted(() => {
    loadNodes();
    checkTask();
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
        width: 100%;
        margin: 3px 0;
        border-top: 1px var(--el-border-color) var(--el-border-style);
    }
}

.dropdown-menu {
    min-width: 120px;
}

.dropdown-item {
    padding: 2px 8px;
    cursor: pointer;
    transition: background 0.3s;
    .icon {
        font-size: 6px;
    }
    .icon-status {
        float: right;
        font-size: 18px;
    }
    .msg-tag {
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

.dropdown-item:hover {
    background: var(--el-menu-item-bg-color-active);
}
</style>
