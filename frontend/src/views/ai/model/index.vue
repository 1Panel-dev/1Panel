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
                    :cache-pagination="false"
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
                                        $t(
                                            row.appInstallId
                                                ? 'aiTools.model.localService'
                                                : 'aiTools.model.cloudService',
                                        )
                                    "
                                    v-bind="TABLE_TOOLTIP_OPTIONS"
                                >
                                    <el-icon
                                        class="service-source-icon"
                                        :aria-label="
                                            $t(
                                                row.appInstallId
                                                    ? 'aiTools.model.localService'
                                                    : 'aiTools.model.cloudService',
                                            )
                                        "
                                    >
                                        <Monitor v-if="row.appInstallId" />
                                        <Cloudy v-else />
                                    </el-icon>
                                </el-tooltip>
                                <el-button
                                    class="service-name-link"
                                    link
                                    type="primary"
                                    @click="detailAccountID = row.id"
                                >
                                    <span class="cell-ellipsis" :title="row.name">{{ row.name }}</span>
                                </el-button>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column prop="provider" :label="$t('aiTools.agents.provider')" min-width="180">
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
                            <AiModelList
                                v-if="row.models?.length"
                                :models="row.models.map((item) => item.id)"
                                :visible-count="1"
                            >
                                <template v-if="!row.appInstallId" #footer>
                                    <el-button link type="primary" @click="poolRef.open(row)">
                                        {{ $t('aiTools.agents.modelPool') }}
                                    </el-button>
                                </template>
                            </AiModelList>
                            <span v-if="!row.models?.length">—</span>
                        </template>
                    </el-table-column>
                    <el-table-column prop="verified" :label="$t('aiTools.model.connectionStatus')" min-width="130">
                        <template #default="{ row }">
                            <div class="verification-status">
                                <el-tooltip
                                    :disabled="!row.verifiedAt && !row.verifyError"
                                    v-bind="TABLE_TOOLTIP_OPTIONS"
                                >
                                    <template #content>
                                        <div v-if="row.verifiedAt">
                                            {{ $t('aiTools.model.verifiedAt') }}:
                                            {{ dateFormatSimpleWithSecond(row.verifiedAt) }}
                                        </div>
                                        <div v-if="row.verifyModel">
                                            {{ $t('aiTools.agents.verifyModel') }}: {{ row.verifyModel }}
                                        </div>
                                        <div v-if="row.verifyError" class="max-w-md break-words">
                                            {{ row.verifyError }}
                                        </div>
                                    </template>
                                    <el-tag :type="connectionStatus(row).type">
                                        {{ connectionStatus(row).label }}
                                    </el-tag>
                                </el-tooltip>
                                <el-tooltip
                                    v-if="row.validateAvailability"
                                    :content="$t('aiTools.model.reverify')"
                                    v-bind="TABLE_TOOLTIP_OPTIONS"
                                >
                                    <el-button
                                        v-permission
                                        link
                                        type="primary"
                                        :icon="RefreshRight"
                                        :aria-label="$t('aiTools.model.reverify')"
                                        :disabled="verifyingAccounts.has(`${currentNode}:${row.id}`)"
                                        @click="verifyTarget = { account: row, node: currentNode }"
                                    />
                                </el-tooltip>
                            </div>
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
                        :ellipsis="2"
                        width="220"
                    />
                </ComplexTable>
            </template>
        </LayoutContent>
        <DialogPro
            v-model="verifyDialogVisible"
            :title="$t('aiTools.model.verify')"
            :size="isMobile ? 'w-90' : 'normal'"
            class="verify-dialog"
        >
            <template #content>
                <div v-if="verifyTarget" class="verify-content">
                    <div class="verify-summary">
                        <div class="verify-icon">
                            <el-icon><Connection /></el-icon>
                        </div>
                        <div class="verify-heading">
                            <h3>{{ verifyTarget.account.name }}</h3>
                            <p>{{ $t('aiTools.model.verifyConfirm') }}</p>
                        </div>
                    </div>
                    <dl class="verify-fields">
                        <div class="verify-address">
                            <dt>{{ $t('aiTools.model.endpoint') }}</dt>
                            <dd>
                                <code>{{ verifyTarget.account.baseUrl || '—' }}</code>
                            </dd>
                        </div>
                        <div>
                            <dt>{{ $t('aiTools.aiProxy.apiType') }}</dt>
                            <dd>{{ verifyTarget.account.apiType || '—' }}</dd>
                        </div>
                        <div>
                            <dt>{{ $t('aiTools.agents.verifyModel') }}</dt>
                            <dd>
                                {{ verifyTarget.account.verifyModel || verifyTarget.account.models?.[0]?.id || '—' }}
                            </dd>
                        </div>
                    </dl>
                </div>
            </template>
            <template #footer>
                <el-button @click="verifyDialogVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button v-permission type="primary" @click="verifyAccount">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </template>
        </DialogPro>
        <ServiceDetails
            v-if="detailAccount"
            :model-value="true"
            :account="detailAccount"
            :status="connectionStatus(detailAccount)"
            :verifying="verifyingAccounts.has(`${currentNode}:${detailAccount.id}`)"
            @update:model-value="!$event && (detailAccountID = undefined)"
            @edit="createRef.openAccount(detailAccount)"
            @models="poolRef.open(detailAccount)"
            @verify="verifyTarget = { account: detailAccount, node: currentNode }"
        />
        <CreateService ref="createRef" @search="search" />
        <ModelPool ref="poolRef" @updated="search" />
    </div>
</template>

<script setup lang="ts">
import AiModelList from '@/components/ai-model-list/index.vue';
import { computed, reactive, ref, toRef, watch } from 'vue';
import { useRoute } from 'vue-router';
import { ElMessageBox } from 'element-plus';
import { Cloudy, Connection, Monitor, RefreshRight } from '@element-plus/icons-vue';
import type { AI } from '@/api/interface/ai';
import { TABLE_TOOLTIP_OPTIONS, type FuTableColumnConfig } from '@/components/table/shared';
import { deleteAgentAccount, getAgentProviders, pageAgentAccounts, verifyAgentAccount } from '@/api/modules/ai';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { getAgentProviderDisplayName } from '@/utils/agent';
import i18n from '@/lang';
import { dateFormat, dateFormatSimpleWithSecond } from '@/utils/date';
import { MsgError, MsgSuccess } from '@/utils/message';
import { useSearchState } from '@/utils/search-reset';
import CreateService from './create/index.vue';
import ModelPool from './model-pool/index.vue';
import ServiceDetails from './details/index.vue';
import ProviderLabel from '@/components/agent-provider-label/index.vue';

const route = useRoute();
const pagePath = route.path;
const isCurrentPage = computed(() => route.path === pagePath);
const { isIntl, isMobile, currentNode } = useGlobalStore();
const createRef = ref();
const poolRef = ref();
const loading = ref(false);
const verifyingAccounts = reactive(new Set<string>());
const verifyTarget = ref<{ account: AI.AgentAccountItem; node: string }>();
const verifyDialogVisible = computed({
    get: () => Boolean(verifyTarget.value),
    set: (visible: boolean) => {
        if (!visible) verifyTarget.value = undefined;
    },
});
const connections = ref<AI.AgentAccountItem[]>([]);
const detailAccountID = ref<number>();
const detailAccount = computed(() => connections.value.find((item) => item.id === detailAccountID.value));
watch(detailAccount, (account) => {
    if (!account) detailAccountID.value = undefined;
});
watch([currentNode, () => route.path], () => {
    detailAccountID.value = undefined;
});
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
    if (verifyingAccounts.has(`${currentNode.value}:${row.id}`)) {
        return { type: 'info' as const, label: i18n.global.t('aiTools.model.verifying') };
    }
    if (!row.validateAvailability) {
        return { type: 'success' as const, label: i18n.global.t('aiTools.agents.verifySkipped') };
    }
    if (row.verifyError) {
        return { type: 'danger' as const, label: i18n.global.t('aiTools.model.verifyFailed') };
    }
    return {
        type: row.verified ? ('success' as const) : ('warning' as const),
        label: i18n.global.t(row.verified ? 'aiTools.agents.verified' : 'aiTools.model.unverified'),
    };
};

const verifyAccount = async () => {
    if (!verifyTarget.value) return;
    const { account: row, node } = verifyTarget.value;
    verifyTarget.value = undefined;
    const key = `${node}:${row.id}`;
    if (!row.validateAvailability || currentNode.value !== node || !isCurrentPage.value || verifyingAccounts.has(key))
        return;
    verifyingAccounts.add(key);
    try {
        await verifyAgentAccount({ id: row.id }, node);
        MsgSuccess(i18n.global.t('aiTools.agents.verified'));
    } catch (error: any) {
        MsgError(String(error?.response?.data?.message || error?.message || i18n.global.t('commons.res.commonError')));
    } finally {
        verifyingAccounts.delete(key);
        if (currentNode.value === node) await search();
    }
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
        if (verifyTarget.value?.node !== currentNode.value || !isCurrentPage.value) {
            verifyTarget.value = undefined;
        }
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
.verification-status {
    display: inline-flex;
    align-items: center;
    gap: 8px;
}

.verify-dialog {
    max-width: 560px;
}
.verify-content {
    padding: 4px 0 8px;
}
.verify-summary {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    margin-bottom: 20px;
}
.verify-icon {
    display: grid;
    place-items: center;
    flex-shrink: 0;
    width: 40px;
    height: 40px;
    border-radius: 10px;
    background: var(--el-color-primary-light-9);
    color: var(--el-color-primary);
    font-size: 20px;
}
.verify-heading {
    min-width: 0;
    h3 {
        margin: 0 0 6px;
        font-size: 15px;
        color: var(--el-text-color-primary);
        overflow-wrap: anywhere;
    }
    p {
        margin: 0;
        color: var(--el-text-color-secondary);
        font-size: 13px;
        line-height: 1.6;
    }
}
.verify-fields {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 20px;
    margin: 0;
    padding: 16px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 8px;
    background: var(--el-fill-color-extra-light);
    dt {
        margin-bottom: 8px;
        color: var(--el-text-color-secondary);
        font-size: 12px;
    }
    dd {
        margin: 0;
        color: var(--el-text-color-primary);
        font-size: 13px;
        line-height: 1.6;
        overflow-wrap: anywhere;
    }
    code {
        font-family: var(--el-font-family-monospace, monospace);
    }
}
.verify-address {
    grid-column: 1 / -1;
    padding-bottom: 16px;
    border-bottom: 1px solid var(--el-border-color-lighter);
}
@media (max-width: 480px) {
    .verify-fields {
        grid-template-columns: minmax(0, 1fr);
    }
}

.service-name {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
}
.service-name-link {
    min-width: 0;
    font-weight: 400;
    :deep(> span) {
        min-width: 0;
    }
}
.service-source-icon {
    flex-shrink: 0;
    font-size: 16px;
    color: var(--el-text-color-secondary);
}
.cell-ellipsis {
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
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
