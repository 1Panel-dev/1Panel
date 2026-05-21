<template>
    <div v-loading="loading">
        <el-alert
            v-if="agentType !== 'openclaw'"
            :title="$t('aiTools.agents.agentRoleUnsupported')"
            type="info"
            :closable="false"
        />
        <template v-else>
            <div class="role-toolbar">
                <el-button v-permission type="primary" @click="openCreateDialog">
                    {{ $t('commons.button.add') }}
                </el-button>
            </div>
            <el-empty
                v-if="!configuredAgents.length"
                :description="$t('commons.msg.noneData')"
                :image-size="80"
                class="role-empty"
            />
            <div v-else>
                <el-card v-for="row in configuredAgents" :key="row.id || row.name" class="role-card">
                    <div class="role-card__header">
                        <div class="role-card__name">
                            <div class="role-card__field">
                                <el-button
                                    v-if="row.workspace"
                                    type="primary"
                                    link
                                    class="role-card__title"
                                    @click="openDetailDrawer(row)"
                                >
                                    {{ row.name || row.id || '-' }}
                                </el-button>
                                <span v-else class="role-card__title role-card__title--static">
                                    {{ row.name || row.id || '-' }}
                                </span>
                                <div class="role-card__model">{{ row.model || '-' }}</div>
                            </div>
                        </div>
                        <div class="role-card__path-actions">
                            <el-tooltip v-if="row.workspace" :content="row.workspace" placement="top">
                                <el-button plain size="small" round @click="routerToFileWithPath(row.workspace)">
                                    {{ $t('aiTools.agents.workspace') }}
                                </el-button>
                            </el-tooltip>
                            <el-tooltip v-if="row.agentDir" :content="row.agentDir" placement="top">
                                <el-button plain size="small" round @click="routerToFileWithPath(row.agentDir)">
                                    {{ $t('aiTools.agents.agentDir') }}
                                </el-button>
                            </el-tooltip>
                            <el-button plain size="small" round v-permission @click="handleDelete(row)">
                                {{ $t('commons.button.delete') }}
                            </el-button>
                        </div>
                    </div>
                    <div class="role-card__content">
                        <div class="role-card__field role-card__field--bindings">
                            <el-divider class="role-card__divider" border-style="dashed" />
                            <div class="role-card__section-title">
                                {{ $t('aiTools.agents.channelsTab') }}
                            </div>
                            <div v-if="row.bindings?.length" class="role-card__tags">
                                <el-tag
                                    v-for="item in row.bindings"
                                    :key="`${item.channel}-${item.accountId || 'default'}`"
                                    :closable="canManageCurrentAgent"
                                    @close="handleUnbind(row, item)"
                                >
                                    {{ item.accountId ? `${item.channel}:${item.accountId}` : item.channel }}
                                </el-tag>
                                <el-button v-permission size="small" @click="openBindingDialog(row)">
                                    + {{ $t('commons.button.add') }}
                                </el-button>
                            </div>
                            <div v-else class="role-card__tags">
                                <el-button v-permission size="small" @click="openBindingDialog(row)">
                                    + {{ $t('commons.button.add') }}
                                </el-button>
                            </div>
                        </div>
                    </div>
                </el-card>
            </div>

            <CreateDialog ref="createRef" @success="handleCreated" />
            <BindingDialog ref="bindingRef" @success="handleCreated" />
            <DetailDrawer ref="detailRef" />
            <OpDialog ref="opRef" @search="handleCreated" />
        </template>
    </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { deleteAgentRole, getConfiguredAgentRoles, unbindAgentRole } from '@/api/modules/ai';
import { AI } from '@/api/interface/ai';
import i18n from '@/lang';
import { routerToFileWithPath } from '@/utils/router';
import { hasManagePermissionAccess } from '@/utils/permission';
import { MsgSuccess } from '@/utils/message';
import CreateDialog from './create/index.vue';
import BindingDialog from './binding/index.vue';
import DetailDrawer from './detail/index.vue';

interface AgentRoleLoadParams {
    agentId: number;
    agentType: AI.AgentType;
    accountId: number;
    model: string;
    configPath: string;
}

const loading = ref(false);
const createRef = ref<InstanceType<typeof CreateDialog> | null>(null);
const bindingRef = ref<InstanceType<typeof BindingDialog> | null>(null);
const detailRef = ref<InstanceType<typeof DetailDrawer> | null>(null);
const opRef = ref();
const agentId = ref(0);
const agentType = ref<AI.AgentType>('openclaw');
const accountId = ref(0);
const currentModel = ref('');
const configuredAgents = ref<AI.AgentConfiguredAgentItem[]>([]);
const canManageCurrentAgent = computed(() => hasManagePermissionAccess());

const loadConfiguredAgents = async (id: number) => {
    const res = await getConfiguredAgentRoles({ agentId: id });
    configuredAgents.value = res.data || [];
};

const openCreateDialog = () => {
    createRef.value?.acceptParams({
        agentId: agentId.value,
        accountId: accountId.value,
        model: currentModel.value,
    });
};

const openBindingDialog = (row: AI.AgentConfiguredAgentItem) => {
    bindingRef.value?.acceptParams({
        agentId: agentId.value,
        role: row,
    });
};

const openDetailDrawer = (row: AI.AgentConfiguredAgentItem) => {
    if (!row.workspace) {
        return;
    }
    detailRef.value?.acceptParams({
        name: row.name || row.id || '-',
        agentId: agentId.value,
        workspace: row.workspace,
    });
};

const handleDelete = async (row: AI.AgentConfiguredAgentItem) => {
    opRef.value?.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('aiTools.agents.agentRoleTab'),
            i18n.global.t('commons.button.delete'),
        ]),
        names: [row.name],
        api: deleteAgentRole,
        params: {
            agentId: agentId.value,
            id: row.id,
        },
        successMsg: i18n.global.t('commons.msg.operationSuccess'),
        noMsg: false,
    });
};

const handleUnbind = async (row: AI.AgentConfiguredAgentItem, item: AI.AgentRoleBinding) => {
    loading.value = true;
    try {
        await unbindAgentRole({
            agentId: agentId.value,
            id: row.id,
            channel: item.channel,
            accountId: item.accountId || '',
        });
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        await handleCreated();
    } finally {
        loading.value = false;
    }
};

const handleCreated = async () => {
    if (!agentId.value) {
        return;
    }
    await loadConfiguredAgents(agentId.value);
};

const load = async (params: AgentRoleLoadParams) => {
    loading.value = true;
    try {
        agentId.value = params.agentId;
        agentType.value = params.agentType;
        accountId.value = params.accountId;
        currentModel.value = params.model;
        configuredAgents.value = [];
        if (agentType.value === 'openclaw') {
            await loadConfiguredAgents(params.agentId);
        }
    } finally {
        loading.value = false;
    }
};

defineExpose({
    load,
});
</script>

<style scoped lang="scss">
.role-toolbar {
    display: flex;
    margin-bottom: 16px;
}

.role-empty {
    padding: 24px 0;
}

.role-card {
    --el-card-border-color: var(--el-border-color);
    border-radius: 18px;
    margin-top: 7px;
}

.role-card__header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
    margin-bottom: 14px;
}

.role-card__name {
    min-width: 0;
    flex: 1;
}

.role-card__title {
    justify-content: flex-start;
    padding: 0;
    font-size: 16px;
    font-weight: 600;
}

.role-card__title--static {
    color: var(--el-text-color-primary);
}

.role-card__model {
    margin-top: 6px;
    color: var(--el-text-color-secondary);
    font-size: 13px;
    line-height: 1.5;
    word-break: break-word;
}

.role-card__field {
    min-width: 0;
}

.role-card__field--bindings {
    word-break: break-word;
}

.role-card__path-actions {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;
    justify-content: flex-end;
}

.role-card__content {
    display: block;
}

.role-card__tags {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    align-items: center;
}

.role-card__divider {
    margin: 0 0 8px;
}

.role-card__section-title {
    margin-bottom: 10px;
    font-size: 13px;
    line-height: 1.4;
    color: var(--el-text-color-secondary);
}

@media (max-width: 1200px) {
    .role-card__header {
        flex-direction: column;
        align-items: stretch;
    }

    .role-card__path-actions {
        justify-content: flex-start;
    }
}
</style>
