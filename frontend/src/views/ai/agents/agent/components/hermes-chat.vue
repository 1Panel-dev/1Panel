<template>
    <DialogPro v-model="dialogVisible" :title="dialogTitle" size="w-90" @close="handleClose">
        <template #content>
            <div class="hermes-chat-dialog">
                <div class="hermes-chat-dialog__sidebar" v-loading="loading">
                    <div class="hermes-chat-dialog__toolbar">
                        <el-button v-permission="'ai_agent_manage'" type="primary" @click="openNewChat">
                            {{ t('aiTools.agents.hermesChatNewChat') }}
                        </el-button>
                        <el-button plain @click="loadSessions">{{ $t('commons.button.refresh') }}</el-button>
                        <el-button v-if="terminalOpen" plain @click="disconnectTerminal">
                            {{ $t('commons.button.disConn') }}
                        </el-button>
                    </div>
                    <div class="hermes-chat-dialog__list">
                        <el-empty
                            v-if="sessions.length === 0"
                            :image-size="64"
                            :description="t('aiTools.agents.hermesChatNoSessions')"
                        />
                        <div
                            v-for="item in sessions"
                            :key="item.id"
                            class="hermes-chat-dialog__item"
                            :class="{ 'is-active': activeSessionId === item.id }"
                            @click="openSession(item)"
                        >
                            <div
                                class="hermes-chat-dialog__item-header"
                                :class="{ 'is-editing': editingSessionId === item.id }"
                            >
                                <input
                                    v-if="editingSessionId === item.id"
                                    ref="editingInputRef"
                                    v-model="editingTitle"
                                    class="hermes-chat-dialog__item-input"
                                    :placeholder="t('aiTools.agents.hermesChatTitlePlaceholder')"
                                    @click.stop
                                    @mousedown.stop
                                    @keydown.enter.stop.prevent="saveSessionTitle(item)"
                                    @keydown.esc.stop.prevent="cancelEditSessionTitle"
                                />
                                <div v-else class="hermes-chat-dialog__item-title">{{ item.title || item.id }}</div>
                                <div class="hermes-chat-dialog__item-actions">
                                    <template v-if="editingSessionId === item.id">
                                        <el-button
                                            link
                                            v-permission="'ai_agent_manage'"
                                            type="primary"
                                            :disabled="!canSaveSessionTitle"
                                            @click.stop="saveSessionTitle(item)"
                                        >
                                            {{ $t('commons.button.save') }}
                                        </el-button>
                                        <el-button link @click.stop="cancelEditSessionTitle">
                                            {{ $t('commons.button.cancel') }}
                                        </el-button>
                                    </template>
                                    <el-button
                                        v-else-if="!terminalOpen"
                                        link
                                        v-permission="'ai_agent_manage'"
                                        icon="Edit"
                                        class="hermes-chat-dialog__edit-button"
                                        @mousedown.stop
                                        @click.stop="startEditSessionTitle(item)"
                                    />
                                    <el-button
                                        v-if="editingSessionId !== item.id"
                                        link
                                        v-permission="'ai_agent_manage'"
                                        type="danger"
                                        icon="Delete"
                                        :disabled="isDeleteDisabled(item)"
                                        :loading="deletingSessionId === item.id"
                                        @mousedown.stop
                                        @click.stop="deleteSession(item)"
                                    />
                                </div>
                            </div>
                            <div class="hermes-chat-dialog__item-model">{{ item.model || '-' }}</div>
                            <div class="hermes-chat-dialog__item-meta">
                                <span>{{ t('aiTools.agents.hermesChatMessageCount', [item.messageCount]) }}</span>
                                <span>{{ formatTime(item.lastActive) }}</span>
                            </div>
                        </div>
                    </div>
                </div>
                <div class="hermes-chat-dialog__content">
                    <div v-if="!terminalOpen" class="hermes-chat-dialog__empty">
                        {{ t('aiTools.agents.hermesChatEmptyHint') }}
                    </div>
                    <Terminal v-else ref="terminalRef" class="hermes-chat-dialog__terminal" />
                </div>
            </div>
        </template>
    </DialogPro>
</template>

<script setup lang="ts">
import { ElMessageBox } from 'element-plus';
import { computed, nextTick, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import DialogPro from '@/components/dialog-pro/index.vue';
import Terminal from '@/components/terminal/index.vue';
import {
    deleteAgentHermesChatSession,
    getAgentHermesChatSessions,
    renameAgentHermesChatSession,
} from '@/api/modules/ai';
import { AI } from '@/api/interface/ai';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { dateFormat } from '@/utils/date';
import { MsgError, MsgSuccess } from '@/utils/message';

const { currentNode } = useGlobalStore();
const { t } = useI18n();

interface HermesChatDialogParams {
    agentId: number;
    containerID: string;
    title: string;
    node?: string;
}

const dialogVisible = ref(false);
const dialogTitle = ref(t('aiTools.agents.hermesChatTitle'));
const terminalOpen = ref(false);
const loading = ref(false);
const terminalRef = ref<InstanceType<typeof Terminal> | null>(null);
const editingInputRef = ref<HTMLInputElement | null>(null);
const agentId = ref(0);
const containerID = ref('');
const node = ref('');
const sessions = ref<AI.AgentHermesChatSessionItem[]>([]);
const activeSessionId = ref('');
const editingSessionId = ref('');
const editingTitle = ref('');
const deletingSessionId = ref('');
const canSaveSessionTitle = computed(() => editingTitle.value.length > 0);

const formatTime = (value: string) => {
    if (!value) {
        return '-';
    }
    return dateFormat({}, {}, value);
};

const connectTerminal = async (initCmd: string) => {
    terminalRef.value?.onClose();
    terminalOpen.value = false;
    await nextTick();
    terminalOpen.value = true;
    await nextTick();
    let args = `source=container&containerid=${containerID.value}&user=hermes&command=/bin/bash`;
    if (node.value) {
        args += `&operateNode=${node.value}`;
    }
    terminalRef.value?.acceptParams({
        endpoint: '/api/v2/hosts/terminal/container',
        args,
        error: '',
        initCmd,
    });
};

const loadSessions = async () => {
    if (!agentId.value) {
        sessions.value = [];
        return;
    }
    loading.value = true;
    try {
        const res = await getAgentHermesChatSessions({ agentId: agentId.value });
        sessions.value = res.data || [];
    } catch (error) {
        sessions.value = [];
        MsgError(String(error?.message || t('commons.res.commonError')));
    } finally {
        loading.value = false;
    }
};

const openNewChat = async () => {
    cancelEditSessionTitle();
    activeSessionId.value = '';
    await connectTerminal('hermes\n');
};

const openSession = async (item: AI.AgentHermesChatSessionItem) => {
    if (editingSessionId.value === item.id) {
        return;
    }
    cancelEditSessionTitle();
    activeSessionId.value = item.id;
    await connectTerminal(`hermes --resume ${item.id}\n`);
};

const focusEditingInput = async () => {
    await nextTick();
    editingInputRef.value?.focus();
    editingInputRef.value?.select();
};

const startEditSessionTitle = async (item: AI.AgentHermesChatSessionItem) => {
    if (terminalOpen.value) {
        return;
    }
    editingSessionId.value = item.id;
    editingTitle.value = item.title || item.id;
    await focusEditingInput();
};

const cancelEditSessionTitle = () => {
    editingSessionId.value = '';
    editingTitle.value = '';
};

const saveSessionTitle = async (item: AI.AgentHermesChatSessionItem) => {
    if (!canSaveSessionTitle.value) {
        return;
    }
    await renameAgentHermesChatSession({
        agentId: agentId.value,
        id: item.id,
        title: editingTitle.value,
    });
    MsgSuccess(t('aiTools.agents.hermesChatRenameSuccess'));
    cancelEditSessionTitle();
    await loadSessions();
};

const isDeleteDisabled = (item: AI.AgentHermesChatSessionItem) => {
    return terminalOpen.value && activeSessionId.value === item.id;
};

const deleteSession = async (item: AI.AgentHermesChatSessionItem) => {
    if (isDeleteDisabled(item)) {
        return;
    }
    try {
        await ElMessageBox.confirm(
            t('aiTools.agents.hermesChatDeleteConfirm', [item.title || item.id]),
            t('commons.button.delete'),
            {
                confirmButtonText: t('commons.button.delete'),
                cancelButtonText: t('commons.button.cancel'),
                type: 'warning',
            },
        );
    } catch {
        return;
    }
    deletingSessionId.value = item.id;
    try {
        await deleteAgentHermesChatSession({
            agentId: agentId.value,
            id: item.id,
        });
        MsgSuccess(t('aiTools.agents.hermesChatDeleteSuccess'));
        cancelEditSessionTitle();
        if (activeSessionId.value === item.id) {
            activeSessionId.value = '';
        }
        await loadSessions();
    } finally {
        deletingSessionId.value = '';
    }
};

const disconnectTerminal = () => {
    terminalRef.value?.onClose();
    terminalOpen.value = false;
    cancelEditSessionTitle();
};

const acceptParams = async (params: HermesChatDialogParams) => {
    dialogVisible.value = true;
    dialogTitle.value = params.title;
    agentId.value = params.agentId;
    containerID.value = params.containerID;
    node.value = params.node || currentNode.value;
    activeSessionId.value = '';
    sessions.value = [];
    cancelEditSessionTitle();
    terminalRef.value?.onClose();
    terminalOpen.value = false;
    await loadSessions();
};

const handleClose = () => {
    terminalRef.value?.onClose();
    terminalOpen.value = false;
    dialogVisible.value = false;
    activeSessionId.value = '';
    cancelEditSessionTitle();
};

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss">
.hermes-chat-dialog {
    display: grid;
    grid-template-columns: 280px minmax(0, 1fr);
    gap: 16px;
    height: 68vh;
    min-height: 520px;
    max-height: 720px;
    padding-bottom: 2px;
    box-sizing: border-box;
    overflow: hidden;
}

.hermes-chat-dialog__sidebar {
    border: 1px solid var(--el-border-color);
    border-radius: 8px;
    padding: 12px;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    min-height: 0;
    height: 100%;
    overflow: hidden;
}

.hermes-chat-dialog__toolbar {
    display: flex;
    gap: 8px;
    margin-bottom: 12px;
}

.hermes-chat-dialog__list {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 8px;
    overflow: auto;
    min-height: 0;
}

.hermes-chat-dialog__item {
    border: 1px solid var(--el-border-color);
    border-radius: 8px;
    padding: 10px 12px;
    cursor: pointer;
}

.hermes-chat-dialog__item-header {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    align-items: start;
    gap: 8px;
}

.hermes-chat-dialog__item-header.is-editing {
    align-items: center;
}

.hermes-chat-dialog__item.is-active {
    border-color: var(--el-color-primary);
    background: var(--el-color-primary-light-9);
}

.hermes-chat-dialog__item-title {
    flex: 1;
    font-weight: 500;
    color: var(--el-text-color-primary);
    word-break: break-all;
    line-height: 20px;
}

.hermes-chat-dialog__item-actions {
    display: flex;
    align-items: flex-start;
    gap: 4px;
    flex-shrink: 0;
    min-height: 20px;
}

.hermes-chat-dialog__item-header.is-editing .hermes-chat-dialog__item-actions {
    align-items: center;
    min-height: 28px;
}

.hermes-chat-dialog__item-input {
    flex: 1;
    min-width: 0;
    height: 28px;
    padding: 0 8px;
    border: 1px solid var(--el-border-color);
    border-radius: 4px;
    outline: none;
    font-size: 13px;
    color: var(--el-text-color-primary);
    background: var(--el-bg-color);
    box-sizing: border-box;
}

.hermes-chat-dialog__item-input:focus {
    border-color: var(--el-color-primary);
}

.hermes-chat-dialog__item-model {
    margin-top: 6px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
    word-break: break-all;
}

.hermes-chat-dialog__item-meta {
    margin-top: 8px;
    display: flex;
    justify-content: space-between;
    gap: 8px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
}

.hermes-chat-dialog__content {
    border: 1px solid var(--el-border-color);
    border-radius: 8px;
    box-sizing: border-box;
    min-height: 0;
    height: 100%;
    overflow: hidden;
}

.hermes-chat-dialog__empty {
    height: 100%;
    min-height: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--el-text-color-secondary);
}

.hermes-chat-dialog__terminal {
    height: 100%;
}

.hermes-chat-dialog__list :deep(.el-empty) {
    margin: auto 0;
}
</style>
