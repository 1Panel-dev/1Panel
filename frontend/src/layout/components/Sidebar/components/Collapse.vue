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
                        <el-button link @click="openChangeNode" @mouseenter="openChangeNode">
                            <SvgIcon class="icon" iconName="p-gerenzhongxin1" />
                            <span class="ellipsis-text">{{ loadCurrentName() }}</span>
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
                <el-divider v-if="showNodes()" class="divider" />

                <div v-if="showNodes()" class="mb-2">
                    <el-scrollbar max-height="168px" :noresize="true">
                        <div
                            class="dropdown-item mt-1"
                            @click="changeNode(item.name)"
                            :disabled="item.status !== 'Healthy'"
                            v-for="item in nodeOptions"
                            :key="item.name"
                        >
                            <div class="node">
                                {{ item.name }}
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
                    <div class="dropdown-item -mb-1" @click="changeNode('local')">
                        <div class="node">{{ $t('xpack.node.master') }}</div>
                    </div>
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
import { ref, watch } from 'vue';
import bus from '@/global/bus';
import { logOutApi } from '@/api/modules/auth';
import router from '@/routers';

const filter = ref();
const globalStore = GlobalStore();
const menuStore = MenuStore();
const nodes = ref([]);
const nodeOptions = ref([]);
const nodeChangeRef = ref<DropdownInstance>();
const props = defineProps({
    version: String,
});
const isMasterPro = computed(() => {
    return globalStore.isMasterPro();
});
watch(
    () => globalStore.isMasterPro(),
    () => {
        loadNodes();
    },
);

const emit = defineEmits(['openTask']);
bus.on('refreshTask', () => {
    checkTask();
});

const openChangeNode = () => {
    nodeChangeRef.value?.handleOpen();
};

const loadCurrentName = () => {
    if (globalStore.currentNode) {
        if (globalStore.currentNode === 'local') {
            return i18n.global.t('xpack.node.master');
        }
        return globalStore.currentNode;
    }
    return i18n.global.t('xpack.node.master');
};

const showPopover = () => {
    filter.value = '';
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
    nodes.value = [];
    if (!isMasterPro.value) {
        globalStore.currentNode = 'local';
        return;
    }
    await listNodeOptions('')
        .then((res) => {
            if (!res) {
                nodes.value = [];
                return;
            }
            nodes.value = res.data || [];
            if (nodes.value.length === 0) {
                globalStore.currentNode = 'local';
            }
            nodeOptions.value = nodes.value || [];
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
        globalStore.currentNodeAddr = '';
        router.push({ name: 'home' }).then(() => {
            window.location.reload();
        });
        return;
    }
    for (const item of nodes.value) {
        if (item.name == command) {
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
            globalStore.currentNode = command || 'local';
            globalStore.currentNodeAddr = item.addr;
            router.push({ name: 'home' }).then(() => {
                window.location.reload();
            });
        }
    }
};

const showNodes = () => {
    return nodes.value.length > 0 && isMasterPro;
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
        width: 91%;
        margin: 3px 8px;
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
    .node {
        padding: 3px 0;
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
