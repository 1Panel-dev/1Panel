<template>
    <DrawerPro
        v-model="terminalVisible"
        :header="$t('cronjob.library.executeScript')"
        @close="handleClose"
        @before-close="beforeClose"
        :confirmBeforeClose="true"
        :resource="scriptName"
        size="60%"
        :autoClose="false"
        :fullScreen="true"
    >
        <template #content>
            <div class="executor-toolbar">
                <el-select
                    v-model="selectedNodes"
                    multiple
                    filterable
                    :loading="loading"
                    :disabled="sessions.length > 0 || loading"
                    :placeholder="$t('cronjob.library.targetNodes')"
                    :aria-label="$t('cronjob.library.targetNodes')"
                >
                    <el-option v-for="node in nodes" :key="node.name" :value="node.name" :label="node.label" />
                </el-select>
                <el-button
                    type="primary"
                    :disabled="loading || selectedNodes.length === 0 || sessions.length > 0"
                    @click="startExecution"
                >
                    {{ $t('commons.button.handle') }}
                </el-button>
            </div>
            <div v-if="sessions.length" class="executor-workspace">
                <aside class="executor-nodes">
                    <button
                        v-for="session in sessions"
                        :key="session.name"
                        type="button"
                        class="executor-node"
                        :class="{ active: activeNode === session.name }"
                        :aria-pressed="activeNode === session.name"
                        @click="selectNode(session.name)"
                    >
                        <span class="executor-node-name">{{ session.label }}</span>
                        <span class="executor-node-status" :class="session.status">
                            {{ $t(`cronjob.library.${session.status}`) }}
                        </span>
                        <span v-if="session.addr" class="executor-node-address">{{ session.addr }}</span>
                    </button>
                </aside>
                <div class="executor-terminals">
                    <div class="executor-hint">{{ $t('cronjob.library.switchNodeHelper') }}</div>
                    <Terminal
                        v-for="session in sessions"
                        v-show="activeNode === session.name"
                        :key="session.name"
                        :ref="(el) => setTerminalRef(session.name, el)"
                        class="executor-terminal"
                        @connection="(status) => (session.status = status)"
                    />
                </div>
            </div>
            <el-empty v-else :description="$t('cronjob.library.targetNodes')" />
        </template>
        <template #footer>
            <el-button @click="beforeClose(() => (terminalVisible = false))">
                {{ $t(sessions.length ? 'commons.button.disConn' : 'commons.button.close') }}
            </el-button>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { ref, nextTick } from 'vue';
import type { ComponentPublicInstance } from 'vue';
import { ElMessageBox } from 'element-plus';
import Terminal from '@/components/terminal/index.vue';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { listNodeOptions } from '@/api/modules/setting';
import i18n from '@/lang';

const { currentNode, globalStore, isXpackOrEE, isAdmin, nodeRoles } = useGlobalStore();
const terminalVisible = ref(false);
const scriptID = ref(0);
const scriptName = ref('');
const loading = ref(false);
const selectedNodes = ref<string[]>([]);
const activeNode = ref('');
interface ExecutorNode {
    name: string;
    label: string;
    addr: string;
}
interface ExecutorSession extends ExecutorNode {
    status: 'connecting' | 'connected' | 'disconnected';
}
const nodes = ref<ExecutorNode[]>([]);
const sessions = ref<ExecutorSession[]>([]);
const terminals = new Map<string, InstanceType<typeof Terminal>>();
let loadToken = 0;

const setTerminalRef = (name: string, el: Element | ComponentPublicInstance | null) => {
    if (el) {
        terminals.set(name, el as InstanceType<typeof Terminal>);
    } else {
        terminals.delete(name);
    }
};

interface DialogProps {
    scriptID: number;
    scriptName: string;
}
const acceptParams = async (params: DialogProps): Promise<void> => {
    const token = ++loadToken;
    scriptID.value = params.scriptID;
    scriptName.value = params.scriptName;
    const name = currentNode.value;
    nodes.value = [{ name, label: name === 'local' ? globalStore.getMasterAlias() : name, addr: '' }];
    selectedNodes.value = [name];
    terminalVisible.value = true;
    if (!isXpackOrEE.value) return;
    loading.value = true;
    try {
        const res = await listNodeOptions('all');
        if (token !== loadToken) return;
        nodes.value = (res.data || [])
            .filter(
                (node) =>
                    node.status !== 'Deleted' &&
                    (isAdmin.value ||
                        nodeRoles.value.some((role) => role.nodeName === node.name && role.roleName === 'Node Admin')),
            )
            .map((node) => ({
                name: node.name,
                label: node.name === 'local' ? globalStore.getMasterAlias() : node.name,
                addr: node.addr,
            }));
        if (isAdmin.value && !nodes.value.some((node) => node.name === 'local')) {
            nodes.value.unshift({ name: 'local', label: globalStore.getMasterAlias(), addr: '' });
        }
        selectedNodes.value = selectedNodes.value.filter((selected) =>
            nodes.value.some((node) => node.name === selected),
        );
    } finally {
        if (token === loadToken) loading.value = false;
    }
};

const startExecution = async () => {
    if (sessions.value.length || loading.value) return;
    sessions.value = nodes.value
        .filter((node) => selectedNodes.value.includes(node.name))
        .map((node) => ({ ...node, status: 'connecting' }));
    activeNode.value = sessions.value[0]?.name || '';
    await nextTick();
    for (const session of sessions.value) {
        terminals.get(session.name)?.acceptParams({
            endpoint: '/api/v2/core/script/run',
            args: new URLSearchParams({ script_id: String(scriptID.value), operateNode: session.name }).toString(),
            error: '',
            initCmd: '',
        });
    }
};

const selectNode = async (name: string) => {
    activeNode.value = name;
    await nextTick();
    terminals.get(name)?.refit();
    terminals.get(name)?.focus();
};

const beforeClose = async (done: () => void) => {
    if (sessions.value.some((session) => session.status !== 'disconnected')) {
        try {
            await ElMessageBox.confirm(i18n.global.t('cronjob.library.executorCloseHelper'), scriptName.value, {
                type: 'warning',
                confirmButtonText: i18n.global.t('commons.button.disConn'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
            });
        } catch {
            return;
        }
    }
    done();
};

const handleClose = () => {
    loadToken++;
    loading.value = false;
    terminals.forEach((terminal) => terminal.onClose());
    terminals.clear();
    sessions.value = [];
};

defineExpose({ acceptParams });
</script>

<style scoped lang="scss">
.executor-toolbar {
    display: flex;
    gap: 12px;
    margin-bottom: 16px;

    .el-select {
        flex: 1;
        min-width: 0;
    }
}
.executor-workspace {
    display: flex;
    height: calc(100vh - 240px);
    min-height: 320px;
    gap: 16px;
}
.executor-nodes {
    width: 210px;
    flex-shrink: 0;
    overflow-y: auto;
    border-right: 1px solid var(--el-border-color-light);
    padding-right: 12px;
}
.executor-node {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    gap: 8px;
    width: 100%;
    padding: 14px 12px;
    margin-bottom: 8px;
    border: 1px solid transparent;
    border-radius: 6px;
    background: var(--el-fill-color-light);
    color: var(--el-text-color-primary);
    text-align: left;
    cursor: pointer;

    &.active {
        border-color: var(--el-color-primary);
        background: var(--el-color-primary-light-9);
    }
}
.executor-node-name {
    overflow-wrap: anywhere;
}
.executor-node-status,
.executor-node-address {
    font-size: 12px;
    color: var(--el-text-color-secondary);
}
.executor-node-status.connected {
    color: var(--el-color-success);
}
.executor-node-address {
    grid-column: 1 / -1;
    overflow-wrap: anywhere;
}
.executor-terminals {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
}
.executor-hint {
    color: var(--el-text-color-secondary);
    font-size: 12px;
    margin-bottom: 12px;
}
.executor-terminal {
    flex: 1;
    min-height: 0;
}
@media (max-width: 768px) {
    .executor-workspace {
        flex-direction: column;
        height: calc(100vh - 260px);
    }
    .executor-nodes {
        width: 100%;
        max-height: 140px;
        border-right: 0;
        padding-right: 0;
    }
    .executor-terminals {
        min-height: 180px;
    }
}
</style>
