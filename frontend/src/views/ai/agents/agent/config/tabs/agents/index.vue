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
                <el-button type="primary" @click="openCreateDialog">
                    {{ $t('commons.button.add') }}
                </el-button>
            </div>
            <ComplexTable :data="configuredAgents">
                <el-table-column :label="$t('commons.table.name')" min-width="140" show-overflow-tooltip>
                    <template #default="{ row }">
                        <el-button v-if="row.workspace" type="primary" link @click="openDetailDrawer(row)">
                            {{ row.name || row.id || '-' }}
                        </el-button>
                        <span v-else>{{ row.name || row.id || '-' }}</span>
                    </template>
                </el-table-column>
                <el-table-column :label="$t('aiTools.model.model')" min-width="120" show-overflow-tooltip>
                    <template #default="{ row }">
                        {{ row.model || '-' }}
                    </template>
                </el-table-column>
                <el-table-column :label="$t('aiTools.agents.workspace')" width="100" align="center">
                    <template #default="{ row }">
                        <el-tooltip v-if="row.workspace" :content="row.workspace" placement="top">
                            <el-button type="primary" link @click="routerToFileWithPath(row.workspace)">
                                <el-icon><FolderOpened /></el-icon>
                            </el-button>
                        </el-tooltip>
                        <span v-else>-</span>
                    </template>
                </el-table-column>
                <el-table-column :label="$t('aiTools.agents.agentDir')" width="120" align="center">
                    <template #default="{ row }">
                        <el-tooltip v-if="row.agentDir" :content="row.agentDir" placement="top">
                            <el-button type="primary" link @click="routerToFileWithPath(row.agentDir)">
                                <el-icon><FolderOpened /></el-icon>
                            </el-button>
                        </el-tooltip>
                        <span v-else>-</span>
                    </template>
                </el-table-column>
                <el-table-column :label="$t('aiTools.agents.bindings')" min-width="160" show-overflow-tooltip>
                    <template #default="{ row }">
                        {{ formatBindings(row.bindings) }}
                    </template>
                </el-table-column>
                <fu-table-operations :buttons="buttons" :label="$t('commons.table.operate')" width="90" fixed="right" />
            </ComplexTable>

            <CreateDialog ref="createRef" @success="handleCreated" />
            <DetailDrawer ref="detailRef" />
            <OpDialog ref="opRef" @search="handleCreated" />
        </template>
    </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { deleteAgentRole, getConfiguredAgentRoles } from '@/api/modules/ai';
import { AI } from '@/api/interface/ai';
import i18n from '@/lang';
import { routerToFileWithPath } from '@/utils/router';
import CreateDialog from './create/index.vue';
import DetailDrawer from './detail/index.vue';

interface AgentRoleLoadParams {
    agentId: number;
    agentType: 'openclaw' | 'copaw';
    accountId: number;
    model: string;
    configPath: string;
}

const loading = ref(false);
const createRef = ref<InstanceType<typeof CreateDialog> | null>(null);
const detailRef = ref<InstanceType<typeof DetailDrawer> | null>(null);
const opRef = ref();
const agentId = ref(0);
const agentType = ref<'openclaw' | 'copaw'>('openclaw');
const accountId = ref(0);
const currentModel = ref('');
const configuredAgents = ref<AI.AgentConfiguredAgentItem[]>([]);

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

const formatBindings = (bindings: AI.AgentRoleBinding[] = []) => {
    if (!bindings.length) {
        return '-';
    }
    return bindings
        .map((item) => (item.accountId ? `${item.channel}:${item.accountId}` : item.channel))
        .filter(Boolean)
        .join(', ');
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

const buttons = [
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: AI.AgentConfiguredAgentItem) => handleDelete(row),
    },
];

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
</style>
