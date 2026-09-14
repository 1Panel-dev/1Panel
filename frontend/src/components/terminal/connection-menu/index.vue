<template>
    <el-popover
        v-model:visible="visible"
        trigger="click"
        placement="bottom-start"
        :width="400"
        popper-class="terminal-connection-popover"
        @before-enter="loadConnections"
    >
        <template #reference>
            <el-button
                class="terminal-connection-add"
                @click.stop
                @keydown.stop
                icon="Plus"
                text
                :aria-label="$t('terminal.createConn')"
            />
        </template>
        <div class="terminal-connection-menu">
            <div class="terminal-connection-actions">
                <el-button
                    v-if="!isNodeAdmin"
                    text
                    class="terminal-connection-action"
                    :disabled="connecting"
                    @click="onNewSsh"
                >
                    <el-icon><Plus /></el-icon>
                    {{ $t('terminal.createConn') }}
                </el-button>
                <el-button text class="terminal-connection-action" :disabled="connecting" @click="connectLocal()">
                    <el-icon><House /></el-icon>
                    {{ $t('terminal.localhost') }}
                </el-button>
            </div>
            <template v-if="connectionTree.length > 0 || !isNodeAdmin || loadingConnections">
                <el-input
                    v-model="connectionFilter"
                    size="small"
                    clearable
                    prefix-icon="Search"
                    :placeholder="$t('commons.button.search')"
                    :aria-label="$t('terminal.createConn')"
                />
                <el-tree
                    ref="treeRef"
                    v-loading="loadingConnections"
                    node-key="id"
                    default-expand-all
                    :expand-on-click-node="false"
                    :data="connectionTree"
                    :filter-node-method="filterConnection"
                    :empty-text="$t('commons.msg.noneData')"
                    class="terminal-connection-tree"
                >
                    <template #default="{ data }">
                        <span v-if="data.kind === 'group'" class="terminal-connection-group">
                            {{ data.label }}
                        </span>
                        <el-button
                            v-else
                            text
                            class="terminal-connection-item"
                            :disabled="connecting"
                            :title="data.node ? `${data.label} (${data.node.addr})` : data.label"
                            @click.stop="connectItem(data)"
                        >
                            <span class="terminal-connection-label">{{ data.label }}</span>
                            <span v-if="data.node" class="terminal-connection-address">
                                {{ data.node.addr }}
                            </span>
                        </el-button>
                    </template>
                </el-tree>
            </template>
        </div>
    </el-popover>
    <HostDialog
        ref="hostDialogRef"
        @on-conn-terminal="onHostCreated"
        @on-new-local="onLocalConfigured"
        @load-host-tree="loadHosts"
    />
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue';
import { ElTree } from 'element-plus';
import i18n from '@/lang';
import { listNodeOptions } from '@/api/modules/setting';
import { getHostTree, testByID, testLocalConn } from '@/api/modules/terminal';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { MsgError } from '@/utils/message';
import type { Host } from '@/api/interface/host';
import type { Setting } from '@/api/interface/setting';
import HostDialog from '@/components/terminal/host-create.vue';
import type { TerminalConnectionOptions } from './types';

const props = defineProps<{
    openSession: (options: TerminalConnectionOptions) => Promise<void>;
}>();
const visible = defineModel<boolean>({ default: false });
const { isNodeAdmin, isXpackOrEE, currentNode, globalStore } = useGlobalStore();
const hostDialogRef = ref<InstanceType<typeof HostDialog>>();
const onNewSsh = () => {
    if (isNodeAdmin.value || connecting.value) return;
    visible.value = false;
    hostDialogRef.value?.acceptParams({ isLocal: false });
};
const onHostCreated = (title: string, wsID: number) => connect(wsID, title);
const connectLocal = () => connect(0, i18n.global.t('terminal.localhost'));
const onLocalConfigured = (nodeName: string) =>
    connect(
        0,
        `${i18n.global.t('terminal.localhost')} (${nodeName === 'local' ? globalStore.getMasterAlias() : nodeName})`,
        nodeName,
    );

interface ConnectionTreeItem {
    id: string;
    label: string;
    kind: 'group' | 'node' | 'host';
    children?: ConnectionTreeItem[];
    node?: Setting.NodeItem;
    wsID?: number;
}

const hostTree = ref<Array<Host.HostTree>>([]);
const treeRef = ref<InstanceType<typeof ElTree>>();
const connectionFilter = ref('');
const nodes = ref<Setting.NodeItem[]>([]);
const loadingConnections = ref(false);
const connecting = ref(false);
const connectionTree = computed<ConnectionTreeItem[]>(() => {
    const groups: ConnectionTreeItem[] = [];
    const childNodes = isXpackOrEE.value
        ? nodes.value.filter((node) => node.name !== 'local' && node.status !== 'Deleted')
        : [];
    if (childNodes.length > 0) {
        groups.push({
            id: 'panel-nodes',
            label: i18n.global.t('xpack.node.node'),
            kind: 'group',
            children: childNodes.map((node) => ({
                id: `node-${node.id}`,
                label: node.name,
                kind: 'node',
                node,
            })),
        });
    }
    if (!isNodeAdmin.value) {
        groups.push(
            ...hostTree.value.map((group): ConnectionTreeItem => ({
                id: `host-group-${group.id}`,
                label: group.label === 'Default' ? i18n.global.t('commons.table.default') : group.label,
                kind: 'group',
                children: (group.children || []).map((host) => ({
                    id: `host-${host.id}`,
                    label: host.label,
                    kind: 'host',
                    wsID: host.id,
                })),
            })),
        );
    }
    return groups;
});
const loadNodes = async () => {
    nodes.value = [];
    if (!isXpackOrEE.value) return;
    const res = await listNodeOptions('all');
    nodes.value = res.data || [];
};
const loadHosts = async () => {
    hostTree.value = [];
    if (isNodeAdmin.value) return;
    const res = await getHostTree({});
    hostTree.value = res.data || [];
};
const loadConnections = async () => {
    loadingConnections.value = true;
    try {
        await Promise.allSettled([loadHosts(), loadNodes()]);
    } finally {
        loadingConnections.value = false;
    }
};
watch([connectionFilter, connectionTree], async () => {
    await nextTick();
    treeRef.value?.filter(connectionFilter.value);
});
const filterConnection = (value: string, data: ConnectionTreeItem) => {
    const filter = value.trim().toLowerCase();
    return !filter || [data.label, data.node?.addr || ''].some((text) => text.toLowerCase().includes(filter));
};
const connectItem = (item: ConnectionTreeItem) => {
    if (item.kind === 'node' && item.node) return connect(0, item.label, item.node.name);
    if (item.kind === 'host' && item.wsID !== undefined) return connect(item.wsID, item.label);
};

const connect = async (wsID: number, title: string, nodeName?: string) => {
    if (connecting.value || (wsID > 0 && isNodeAdmin.value)) return;
    connecting.value = true;
    visible.value = false;
    const targetNode = nodeName || currentNode.value || 'local';
    try {
        if (wsID === 0) {
            const res = await testLocalConn(targetNode);
            if (!res.data) {
                if (!nodeName) {
                    hostDialogRef.value?.acceptParams({ isLocal: true, nodeName: targetNode });
                } else {
                    MsgError(`${title}: ${i18n.global.t('terminal.connLocalErr')}`);
                }
                return;
            }
            await props.openSession({
                title: nodeName
                    ? title
                    : `${title} (${targetNode === 'local' ? globalStore.getMasterAlias() : targetNode})`,
                wsID,
                nodeName: targetNode,
            });
            return;
        }
        const res = await testByID(wsID);
        await props.openSession({
            title,
            wsID,
            error: res.data ? '' : 'Authentication failed. Please check the host information!',
        });
    } finally {
        connecting.value = false;
    }
};

defineExpose({ connectLocal });
</script>

<style lang="scss">
.terminal-connection-popover {
    max-width: calc(100vw - 24px);
    box-sizing: border-box;
}
</style>

<style scoped lang="scss">
.terminal-connection-add {
    width: 32px;
    height: 32px;
    margin: 0 4px;
    padding: 0;
    border-radius: 6px;
    color: var(--el-text-color-regular);

    &:hover {
        color: var(--el-color-primary);
    }
}

.terminal-connection-menu {
    display: flex;
    flex-direction: column;
    gap: 8px;
}

.terminal-connection-actions {
    display: flex;
    gap: 8px;
}

.terminal-connection-action {
    flex: 1;
    min-width: 0;
    height: 30px;
    margin: 0;
    padding: 0 6px;
    font-size: 13px;
    background-color: var(--el-fill-color-light);

    .el-icon {
        margin-right: 8px;
        color: var(--el-text-color-secondary);
    }
}

.terminal-connection-tree {
    max-height: min(192px, 35vh);
    min-height: 32px;
    overflow: auto;

    :deep(.el-tree-node__content) {
        height: 32px;
        border-radius: 4px;
    }
}

.terminal-connection-group {
    font-size: 12px;
    font-weight: 500;
    color: var(--el-text-color-secondary);
}

.terminal-connection-item {
    flex: 1;
    min-width: 0;
    height: 30px;
    margin: 0;
    padding: 0 8px 0 0;
    font-size: 13px;
    font-weight: normal;

    :deep(> span) {
        display: flex;
        align-items: center;
        gap: 12px;
        min-width: 0;
        width: 100%;
    }
}

.terminal-connection-label {
    flex: 1;
    text-align: left;
}

.terminal-connection-address {
    flex: 0 1 45%;
    font-size: 12px;
    text-align: right;
    color: var(--el-text-color-secondary);
}

.terminal-connection-label,
.terminal-connection-address {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
</style>
