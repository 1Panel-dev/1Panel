<template>
    <div class="model-page">
        <RouterButton :buttons="[{ label: $t('aiTools.model.services'), path: '/ai/model-services' }]" />
        <LayoutContent v-loading="loading">
            <template #leftToolBar>
                <el-button v-permission type="primary" @click="createRef.open()">
                    {{ $t('aiTools.model.addService') }}
                </el-button>
            </template>
            <template #rightToolBar>
                <el-select
                    v-model="apiType"
                    clearable
                    filterable
                    placeholder=""
                    class="model-filter"
                    @change="resetPageAndSearch"
                >
                    <template #prefix>{{ 'API ' + $t('commons.table.type') }}</template>
                    <el-option v-for="item in apiTypeOptions" :key="item" :label="item" :value="item" />
                </el-select>
                <el-select
                    v-model="provider"
                    clearable
                    filterable
                    class="model-filter"
                    placeholder=""
                    @change="resetPageAndSearch"
                >
                    <template #prefix>{{ $t('aiTools.agents.provider') }}</template>
                    <el-option
                        v-for="item in providerOptions"
                        :key="item.value"
                        :label="item.label"
                        :value="item.value"
                    >
                        <ProviderLabel :provider="item.value" :display-name="item.label" />
                    </el-option>
                </el-select>
                <TableSearch v-model:searchName="searchName" class="model-search" @search="resetPageAndSearch" />
                <TableRefresh @search="search" />
                <fu-table-column-select
                    :columns="accountColumns"
                    trigger="hover"
                    :title="$t('commons.table.selectColumn')"
                    :only-icon="true"
                />
            </template>
            <template #prompt>
                <el-alert type="info" :closable="false" :title="$t('aiTools.model.serviceHelper')" />
            </template>
            <template #main>
                <ComplexTable
                    :data="connections"
                    row-key="id"
                    :pagination-config="pagination"
                    :columns="accountColumns"
                    local-key="modelServiceColumn"
                    @search="search"
                >
                    <el-table-column prop="name" :label="$t('commons.table.name')" min-width="220" fixed="left">
                        <template #default="{ row }">
                            <div class="service-name">
                                <el-tooltip
                                    :content="
                                        $t(row.appInstallId ? 'aiTools.model.localModel' : 'aiTools.model.cloudAccount')
                                    "
                                    placement="top"
                                >
                                    <el-icon
                                        class="service-source-icon"
                                        :aria-label="
                                            $t(
                                                row.appInstallId
                                                    ? 'aiTools.model.localModel'
                                                    : 'aiTools.model.cloudAccount',
                                            )
                                        "
                                    >
                                        <Monitor v-if="row.appInstallId" />
                                        <Cloudy v-else />
                                    </el-icon>
                                </el-tooltip>
                                <div class="service-title cell-ellipsis" :title="row.name">{{ row.name }}</div>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column prop="provider" :label="$t('aiTools.model.engineProvider')" min-width="180">
                        <template #default="{ row }">
                            <div class="cell-ellipsis">
                                <ProviderLabel :provider="row.provider" :display-name="row.providerName" />
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column
                        prop="apiType"
                        :label="'API ' + $t('commons.table.type')"
                        min-width="200"
                        show-overflow-tooltip
                    >
                        <template #default="{ row }">
                            {{ row.apiType || '—' }}
                        </template>
                    </el-table-column>
                    <el-table-column prop="models" :label="$t('aiTools.model.model')" min-width="200">
                        <template #default="{ row }">
                            <div v-if="row.models?.length" class="model-line">
                                <span class="cell-ellipsis" :title="row.models[0].id">{{ row.models[0].id }}</span>
                                <el-button
                                    v-if="!row.appInstallId && row.models.length > 1"
                                    size="small"
                                    class="model-more"
                                    @click="poolRef.open(row)"
                                >
                                    +{{ row.models.length - 1 }}
                                </el-button>
                                <el-tooltip
                                    v-if="row.appInstallId && row.models.length > 1"
                                    :content="
                                        row.models
                                            .slice(1)
                                            .map((item) => item.id)
                                            .join(', ')
                                    "
                                    placement="top"
                                >
                                    <el-button size="small" class="model-more">+{{ row.models.length - 1 }}</el-button>
                                </el-tooltip>
                            </div>
                            <span v-if="!row.models?.length">—</span>
                        </template>
                    </el-table-column>
                    <el-table-column prop="verified" :label="$t('aiTools.model.connectionStatus')" min-width="130">
                        <template #default="{ row }">
                            <el-tag :type="connectionStatus(row).type">
                                {{ connectionStatus(row).label }}
                            </el-tag>
                        </template>
                    </el-table-column>
                    <el-table-column
                        prop="baseUrl"
                        :label="$t('aiTools.model.endpoint')"
                        min-width="280"
                        show-overflow-tooltip
                    >
                        <template #default="{ row }">
                            {{ row.baseUrl || '—' }}
                        </template>
                    </el-table-column>
                    <el-table-column
                        prop="createdAt"
                        :label="$t('commons.table.createdAt')"
                        :formatter="dateFormat"
                        min-width="180"
                        show-overflow-tooltip
                    />
                    <fu-table-operations
                        :buttons="accountButtons"
                        :label="$t('commons.table.operate')"
                        align="left"
                        fixed="right"
                        :ellipsis="3"
                        width="220"
                    />
                </ComplexTable>
            </template>
        </LayoutContent>
        <CreateService ref="createRef" @search="search" />
        <ModelPool ref="poolRef" @updated="search" />
    </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, toRef, watch } from 'vue';
import { useRoute } from 'vue-router';
import { ElMessageBox } from 'element-plus';
import { Cloudy, Monitor } from '@element-plus/icons-vue';
import type { AI } from '@/api/interface/ai';
import type { FuTableColumnConfig } from '@/components/table/shared';
import { deleteAgentAccount, getAgentProviders, pageAgentAccounts } from '@/api/modules/ai';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { getAgentProviderDisplayName, isAgentAccountVerificationSkipped } from '@/utils/agent';
import i18n from '@/lang';
import { dateFormat } from '@/utils/date';
import { useSearchState } from '@/utils/search-reset';
import CreateService from './create/index.vue';
import ModelPool from './model-pool/index.vue';
import ProviderLabel from '@/components/agent-provider-label/index.vue';

const route = useRoute();
const pagePath = route.path;
const isCurrentPage = computed(() => route.path === pagePath);
const { isIntl, currentNode } = useGlobalStore();
const createRef = ref();
const poolRef = ref();
const loading = ref(false);
const connections = ref<AI.AgentAccountItem[]>([]);
const accountColumns = ref<FuTableColumnConfig[]>([]);
const providers = ref<AI.ProviderInfo[]>([]);
watch(
    currentNode,
    async (_, __, onCleanup) => {
        let cancelled = false;
        onCleanup(() => (cancelled = true));
        providers.value = [];
        try {
            const { data } = await getAgentProviders();
            if (!cancelled) providers.value = data || [];
        } catch {
            if (!cancelled) providers.value = [];
        }
    },
    { immediate: true },
);
const apiType = ref('');
const provider = ref('');
const searchName = ref('');
const pagination = reactive({ currentPage: 1, pageSize: 10, total: 0 });
useSearchState(
    'ai/agents/model/index',
    {
        search: searchName,
        apiType,
        provider,
        page: toRef(pagination, 'currentPage'),
        pageSize: toRef(pagination, 'pageSize'),
    },
    {
        onRevisit: () => {
            if (typeof route.query.name === 'string' && route.query.name !== searchName.value) {
                searchName.value = route.query.name;
                resetPage();
            }
            return search();
        },
    },
);
const apiTypeOptions = computed(() =>
    Array.from(new Set(providers.value.flatMap((item) => item.apiTypes.map((api) => api.apiType)))).sort(),
);
const providerOptions = computed(() =>
    providers.value
        .filter((item) => !isIntl.value || !['ark-coding-plan', 'bailian-coding-plan'].includes(item.provider))
        .map((item) => ({
            value: item.provider,
            label: getAgentProviderDisplayName(item.provider, item.displayName),
        })),
);
const resetPage = () => {
    pagination.currentPage = 1;
};
const resetPageAndSearch = () => {
    resetPage();
    search();
};
const resetPageIfNeeded = () => {
    pagination.currentPage = Math.min(
        pagination.currentPage,
        Math.max(1, Math.ceil(pagination.total / pagination.pageSize)),
    );
};
let searchVersion = 0;
const search = async () => {
    const version = ++searchVersion;
    if (!isCurrentPage.value) return;
    loading.value = true;
    try {
        const result = await pageAgentAccounts({
            page: pagination.currentPage,
            pageSize: pagination.pageSize,
            apiType: apiType.value,
            provider: provider.value,
            name: searchName.value,
        });
        if (version !== searchVersion) return;
        connections.value = result.data.items || [];
        pagination.total = result.data.total;
        const previousPage = pagination.currentPage;
        resetPageIfNeeded();
        if (previousPage !== pagination.currentPage) await search();
    } catch {
        if (version !== searchVersion) return;
        connections.value = [];
        pagination.total = 0;
    } finally {
        if (version === searchVersion) loading.value = false;
    }
};

const connectionStatus = (row: AI.AgentAccountItem) => {
    if (isAgentAccountVerificationSkipped(row.provider)) {
        return { type: 'success' as const, label: i18n.global.t('aiTools.agents.verifySkipped') };
    }
    return {
        type: row.verified ? ('success' as const) : ('warning' as const),
        label: i18n.global.t(row.verified ? 'aiTools.agents.verified' : 'aiTools.model.unverified'),
    };
};
const accountButtons = [
    {
        label: i18n.global.t('commons.button.edit'),
        permission: true,
        click: (row: AI.AgentAccountItem) => createRef.value.openAccount(row),
    },
    {
        label: i18n.global.t('aiTools.agents.modelPool'),
        disabled: (row: AI.AgentAccountItem) => row.provider === 'vllm',
        click: (row: AI.AgentAccountItem) => poolRef.value.open(row),
    },
    {
        label: i18n.global.t('commons.button.delete'),
        permission: true,
        click: async (row: AI.AgentAccountItem) => {
            await ElMessageBox.confirm(
                i18n.global.t('commons.msg.delete', [row.name]),
                i18n.global.t('commons.button.delete'),
                { type: 'warning' },
            );
            await deleteAgentAccount({ id: row.id });
            await search();
        },
    },
];
watch(
    [() => route.path, () => route.query.name, currentNode],
    async (_, previous) => {
        if (!isCurrentPage.value) {
            searchVersion++;
            loading.value = false;
            return;
        }
        if (previous.length && previous[0] !== pagePath) return;
        if (typeof route.query.name === 'string' && route.query.name !== searchName.value) {
            searchName.value = route.query.name;
            resetPage();
        }
        await search();
    },
    { immediate: true },
);
</script>

<style scoped lang="scss">
.service-name {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
}
.service-title {
    font-weight: 600;
}
.service-source-icon {
    flex-shrink: 0;
    font-size: 16px;
    color: var(--el-color-primary);
}
.cell-ellipsis {
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
}
.model-line {
    display: flex;
    align-items: center;
}
.model-more {
    flex-shrink: 0;
    margin-left: 8px;
}
.model-filter {
    width: 220px;
}
.model-search {
    width: 220px;
    :deep(.search-button) {
        width: 100%;
    }
}
@media (max-width: 768px) {
    .model-filter,
    .model-search {
        width: 100%;
    }
}
</style>
