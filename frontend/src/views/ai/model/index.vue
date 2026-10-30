<template>
    <div class="model-page">
        <RouterButton :buttons="tabs" />
        <component
            :is="downloaderView"
            v-if="isFiles && downloaderView"
            @deploy="(modelDir) => createRef.open('vllm', modelDir)"
        />
        <LayoutContent v-else-if="isFiles">
            <template #main><el-empty :description="$t('license.levelUpPro')" /></template>
        </LayoutContent>
        <LocalView v-else-if="isManagement" />
        <template v-else>
            <LayoutContent v-loading="loading">
                <template #leftToolBar>
                    <el-button v-permission type="primary" @click="createRef.open()">
                        {{ $t(isLocal ? 'aiTools.model.addEngine' : 'aiTools.model.addService') }}
                    </el-button>
                    <el-button
                        v-if="isLocal && isProductPro && vllmAvailable"
                        v-permission
                        @click="routerToName('VllmCommandTemplates')"
                    >
                        {{ $t('aiTools.model.deploymentTemplates') }}
                    </el-button>
                </template>
                <template #rightToolBar>
                    <el-select
                        v-if="!isLocal"
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
                        v-if="isLocal"
                        v-model="engine"
                        clearable
                        placeholder=""
                        class="model-filter"
                        @change="resetPageAndSearch"
                    >
                        <template #prefix>{{ $t('aiTools.model.engines') }}</template>
                        <el-option label="Ollama" value="ollama" />
                        <el-option label="vLLM" value="vllm" />
                    </el-select>
                    <el-select
                        v-else
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
                    <TableSetting v-if="isLocal" title="model-deployment-refresh" @search="!loading && search()" />
                    <fu-table-column-select
                        v-else
                        :columns="accountColumns"
                        trigger="hover"
                        :title="$t('commons.table.selectColumn')"
                        :only-icon="true"
                    />
                </template>
                <template #main>
                    <el-alert
                        v-if="loadFailed"
                        type="error"
                        :closable="false"
                        :title="$t('aiTools.model.loadFailed')"
                        class="mb-4"
                    />
                    <ComplexTable
                        v-if="!isLocal"
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
                                    <div class="service-title cell-ellipsis" :title="row.name">{{ row.name }}</div>
                                    <el-tag class="service-source-tag" type="info" size="small">
                                        {{
                                            $t(
                                                !row.appInstallId
                                                    ? 'aiTools.model.cloudAccount'
                                                    : 'aiTools.model.localModel',
                                            )
                                        }}
                                    </el-tag>
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
                                <div
                                    v-for="(model, index) in (row.models || []).slice(0, 2)"
                                    :key="model.id"
                                    class="model-line"
                                >
                                    <span class="cell-ellipsis" :title="model.id">{{ model.id }}</span>
                                    <el-button
                                        v-if="index === 1 && row.models.length > 2"
                                        size="small"
                                        class="model-more"
                                        @click="poolRef.open(row)"
                                    >
                                        +{{ row.models.length - 2 }}
                                    </el-button>
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
                            fixed="right"
                            :ellipsis="3"
                            width="220"
                        />
                    </ComplexTable>
                    <ComplexTable
                        v-else
                        :data="deploymentPage"
                        :pagination-config="pagination"
                        @search="resetPageIfNeeded"
                    >
                        <el-table-column :label="$t('commons.table.name')" min-width="190">
                            <template #default="{ row }">
                                <el-button link type="primary" @click="manageRef.open(row, 'connection', true)">
                                    {{ row.name }}
                                </el-button>
                            </template>
                        </el-table-column>
                        <el-table-column :label="$t('app.version')" min-width="150">
                            <template #default="{ row }">
                                <div class="flex items-center gap-2">
                                    <span>{{ row.version || '—' }}</span>
                                    <el-button
                                        v-if="row.upgradable && hasPermission('app_manage')"
                                        v-permission
                                        link
                                        type="primary"
                                        class="shrink-0"
                                        :disabled="operationDisabled(row)"
                                        @click="manageRef.open(row, 'upgrade', true)"
                                    >
                                        {{ $t('commons.button.upgrade') }}
                                    </el-button>
                                </div>
                            </template>
                        </el-table-column>
                        <el-table-column :label="$t('aiTools.model.model')" min-width="200">
                            <template #default="{ row }">
                                <div v-for="(model, index) in row.models.slice(0, 2)" :key="model" class="model-line">
                                    <span class="cell-ellipsis" :title="model">{{ model }}</span>
                                    <el-button
                                        v-if="row.engine === 'ollama' && index === 1 && row.models.length > 2"
                                        size="small"
                                        class="model-more"
                                        @click="manageDeployment(row)"
                                    >
                                        +{{ row.models.length - 2 }}
                                    </el-button>
                                </div>
                                <span v-if="!row.models.length">—</span>
                            </template>
                        </el-table-column>
                        <el-table-column :label="$t('commons.table.status')" min-width="150">
                            <template #default="{ row }">
                                <el-dropdown placement="bottom" :disabled="statusDisabled(row, 'stop')">
                                    <Status
                                        v-permission
                                        :status="row.status"
                                        operate
                                        :disabled="statusDisabled(row, 'stop')"
                                    />
                                    <template #dropdown>
                                        <el-dropdown-menu>
                                            <el-dropdown-item
                                                :disabled="statusDisabled(row) || row.status === 'Running'"
                                                @click="operate(row, 'start')"
                                            >
                                                {{ $t('commons.operate.start') }}
                                            </el-dropdown-item>
                                            <el-dropdown-item
                                                :disabled="
                                                    statusDisabled(row, 'stop') ||
                                                    ['Stopped', 'Exited'].includes(row.status)
                                                "
                                                @click="operate(row, 'stop')"
                                            >
                                                {{ $t('commons.operate.stop') }}
                                            </el-dropdown-item>
                                            <el-dropdown-item
                                                :disabled="statusDisabled(row)"
                                                @click="operate(row, 'restart')"
                                            >
                                                {{ $t('commons.button.restart') }}
                                            </el-dropdown-item>
                                        </el-dropdown-menu>
                                    </template>
                                </el-dropdown>
                            </template>
                        </el-table-column>
                        <el-table-column :label="$t('container.source')" min-width="170">
                            <template #default="{ row }">
                                <template v-if="row.status === 'Running' && row.containerName">
                                    <ContainerResource
                                        v-if="resourceStats[row.containerName]"
                                        :stats="resourceStats[row.containerName]"
                                    />
                                    <el-icon
                                        v-else-if="resourceStats[row.containerName] === undefined"
                                        class="is-loading"
                                    >
                                        <Loading />
                                    </el-icon>
                                    <el-tooltip v-else :content="$t('aiTools.model.loadFailed')">
                                        <span>—</span>
                                    </el-tooltip>
                                </template>
                                <span v-else>—</span>
                            </template>
                        </el-table-column>
                        <fu-table-operations
                            :buttons="deploymentButtons"
                            :label="$t('commons.table.operate')"
                            fixed="right"
                            :ellipsis="2"
                            width="240"
                        />
                    </ComplexTable>
                </template>
            </LayoutContent>
        </template>
        <CreateService
            ref="createRef"
            :engines-only="isLocal || isFiles"
            :vllm-only="isFiles"
            @search="search"
            @manage="manageDeployment"
        />
        <ManageDeployment
            ref="manageRef"
            @search="search"
            @upgrade="(app) => upgradeRef.acceptParams(app, 'upgrade')"
        />
        <AppUpgrade ref="upgradeRef" @close="search" />
        <AppDelete :key="`delete:${currentNode}:${route.fullPath}`" ref="appDeleteRef" @close="search" />
        <AppResources :key="`resources:${currentNode}:${route.fullPath}`" ref="appResourcesRef" @close="search" />
        <TaskLog ref="taskRef" width="70%" @close="search" />
        <ModelPool ref="poolRef" @updated="search" />
        <TerminalDialog :key="`terminal:${currentNode}:${route.fullPath}`" ref="terminalRef" />
        <SyncService ref="syncRef" @search="search" />
    </div>
</template>

<script setup lang="ts">
import { computed, defineAsyncComponent, nextTick, reactive, ref, toRef, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ElMessageBox } from 'element-plus';
import type { AI } from '@/api/interface/ai';
import type { Container } from '@/api/interface/container';
import { containerListStats } from '@/api/modules/container';
import ContainerResource from '@/components/container-resource/index.vue';
import type { FuTableColumnConfig } from '@/components/table/shared';
import { deleteAgentAccount, getAgentProviders, pageAgentAccounts, operateTensorRTLLM } from '@/api/modules/ai';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { getAgentProviderDisplayName, isAgentAccountVerificationSkipped } from '@/utils/agent';
import { loadOptionalComponent } from '@/extensions/optional';
import i18n from '@/lang';
import { appInstalledDeleteCheck, installedOp, searchAppInstalled } from '@/api/modules/app';
import { hasPermission } from '@/utils/rbac';
import { hasManagePermissionAccess } from '@/utils/permission';
import { newUUID } from '@/utils/id';
import { dateFormat } from '@/utils/date';
import { useSearchState } from '@/utils/search-reset';
import { MsgError } from '@/utils/message';
import { routerToName } from '@/utils/router';
import TaskLog from '@/components/log/task/index.vue';
import LocalView from './local/index.vue';
import CreateService from './components/create/index.vue';
import ManageDeployment from './components/deployment-manage/index.vue';
import AppUpgrade from '@/views/app-store/installed/upgrade/index.vue';
import AppDelete from '@/views/app-store/installed/delete/index.vue';
import AppResources from '@/views/app-store/installed/check/index.vue';
import ModelPool from './components/model-pool/index.vue';
import TerminalDialog from '@/views/container/container/terminal/index.vue';
import SyncService from './components/sync/index.vue';
import ProviderLabel from '@/components/agent-provider-label/index.vue';
import { loadModelOverview, loadVllmAPI, readPages, vllmAvailable, type Deployment } from './deployment';

const route = useRoute();
const router = useRouter();
const pagePath = route.path;
const isCurrentPage = computed(() => route.path === pagePath);
const { isProductPro, isFxplay, isIntl, currentNode, isAdminOrNodeAdmin } = useGlobalStore();
const isFiles = pagePath === '/ai/model/files';
const isLocal = pagePath === '/ai/model/local';
const isManagement = computed(
    () => isCurrentPage.value && isLocal && ['vllm', 'tensorrt'].includes(String(route.query.tab)),
);
const tabs = computed(() => [
    { label: i18n.global.t('aiTools.model.services'), path: '/ai/model/account' },
    { label: i18n.global.t('aiTools.model.files'), path: '/ai/model/files' },
    { label: i18n.global.t('aiTools.model.engines'), path: '/ai/model/local' },
]);
const downloaderModules = import.meta.glob('/src/xpack/views/model-downloader/index.vue');
const downloaderView = Object.keys(downloaderModules).length
    ? defineAsyncComponent(() => loadOptionalComponent('/src/xpack/views/model-downloader/index.vue'))
    : null;
const createRef = ref();
const manageRef = ref();
const upgradeRef = ref();
const appDeleteRef = ref();
const appResourcesRef = ref();
const poolRef = ref();
const terminalRef = ref<InstanceType<typeof TerminalDialog>>();
const syncRef = ref<InstanceType<typeof SyncService>>();
const loading = ref(false);
const loadFailed = ref(false);
const connections = ref<AI.AgentAccountItem[]>([]);
const accountColumns = ref<FuTableColumnConfig[]>([]);
const providers = ref<AI.ProviderInfo[]>([]);
watch(
    currentNode,
    async (_, __, onCleanup) => {
        if (isLocal || isFiles) return;
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
const deployments = ref<Deployment[]>([]);
const apiType = ref('');
const engine = ref('');
const provider = ref('');
const searchName = ref('');
const pagination = reactive({ currentPage: 1, pageSize: 10, total: 0 });
if (!isLocal && !isFiles) {
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
}
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
const filteredDeployments = computed(() => {
    const keyword = searchName.value.trim().toLowerCase();
    return deployments.value.filter(
        (row) =>
            (!engine.value || row.engine === engine.value) &&
            (!keyword || [row.name, row.modelDir, ...row.models].join(' ').toLowerCase().includes(keyword)),
    );
});
const deploymentPage = computed(() =>
    filteredDeployments.value.slice(
        (pagination.currentPage - 1) * pagination.pageSize,
        pagination.currentPage * pagination.pageSize,
    ),
);
const resourceStats = ref<Record<string, Container.ContainerListStats | null>>({});
watch(
    [deploymentPage, isCurrentPage, isManagement, currentNode],
    async ([rows, active, management, node], _, onCleanup) => {
        resourceStats.value = {};
        const controller = new AbortController();
        onCleanup(() => {
            controller.abort();
        });
        if (!active || !isLocal || management) return;
        const names = [
            ...new Set(
                rows.filter((row) => row.status === 'Running' && row.containerName).map((row) => row.containerName),
            ),
        ];
        if (!names.length) return;
        try {
            const { data } = await containerListStats(node, undefined, {
                signal: controller.signal,
                skipErrorMessage: true,
            });
            if (controller.signal.aborted) return;
            const statsByName = new Map((data || []).map((item) => [item.containerName, item]));
            resourceStats.value = Object.fromEntries(names.map((name) => [name, statsByName.get(name) || null]));
        } catch {
            if (!controller.signal.aborted) resourceStats.value = Object.fromEntries(names.map((name) => [name, null]));
        }
    },
    { immediate: true },
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
watch(filteredDeployments, () => {
    if (!isLocal) return;
    pagination.total = filteredDeployments.value.length;
    resetPageIfNeeded();
});

let searchVersion = 0;
const search = async () => {
    const version = ++searchVersion;
    if (!isCurrentPage.value || isFiles || isManagement.value) return;
    loading.value = true;
    loadFailed.value = false;
    try {
        if (isLocal) {
            const data = await loadModelOverview(isProductPro.value, isFxplay.value);
            if (version !== searchVersion) return;
            deployments.value = data;
        } else {
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
        }
    } catch {
        if (version !== searchVersion) return;
        connections.value = [];
        pagination.total = 0;
        deployments.value = [];
        loadFailed.value = true;
    } finally {
        if (version === searchVersion) loading.value = false;
    }
};

const connectionStatus = (row: AI.AgentAccountItem) => {
    if (isAgentAccountVerificationSkipped(row.provider)) {
        return { type: 'info' as const, label: i18n.global.t('aiTools.agents.verifySkipped') };
    }
    return {
        type: row.verified ? ('success' as const) : ('warning' as const),
        label: i18n.global.t(row.verified ? 'aiTools.agents.verified' : 'aiTools.model.unverified'),
    };
};
const manageDeployment = (row: Deployment) => manageRef.value.open(row);
const accountButtons = [
    {
        label: i18n.global.t('commons.button.edit'),
        permission: true,
        click: (row: AI.AgentAccountItem) => createRef.value.openAccount(row),
    },
    {
        label: i18n.global.t('aiTools.agents.modelPool'),
        click: (row: AI.AgentAccountItem) => poolRef.value.open(row),
    },
    {
        label: i18n.global.t('commons.button.delete'),
        permission: true,
        show: (row: AI.AgentAccountItem) => !row.appInstallId || row.provider === 'ollama',
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
const taskRef = ref();
const pending = ref('');
const operationDisabled = (row: Deployment, operation = '') =>
    Boolean(pending.value) ||
    ['Installing', 'Starting', 'Rebuilding', 'Waiting', 'Uninstalling'].includes(row.status) ||
    (['ReStarting', 'Restarting'].includes(row.status) && !['stop', 'delete'].includes(operation));
const statusDisabled = (row: Deployment, operation = '') =>
    operationDisabled(row, operation) ||
    !hasManagePermissionAccess() ||
    (row.engine === 'ollama' && !hasPermission('app_manage'));
const operate = async (row: Deployment, operation: 'start' | 'stop' | 'restart') => {
    if (statusDisabled(row, operation)) return;
    const node = currentNode.value;
    await ElMessageBox.confirm(
        i18n.global.t('app.operatorHelper', [i18n.global.t('commons.operate.' + operation)]),
        i18n.global.t('commons.operate.' + operation),
        { type: 'warning' },
    );
    if (node !== currentNode.value || statusDisabled(row, operation)) return;
    pending.value = `${row.engine}:${row.id}`;
    try {
        if (row.engine === 'ollama') {
            await installedOp({ installId: row.id, operate: operation });
        } else if (row.engine === 'tensorrt') {
            await operateTensorRTLLM({ id: row.id, operate: operation });
        } else {
            const taskID = newUUID();
            await (await loadVllmAPI()).operateVllm({ id: row.id, operate: operation, taskID });
            if (node !== currentNode.value) return;
            if (row.version.toLowerCase().startsWith('nvidia-gb10-dspark-')) {
                await search();
                taskRef.value.openWithTaskID(taskID);
                return;
            }
        }
        if (node !== currentNode.value) return;
        await search();
    } finally {
        pending.value = '';
    }
};
const deploymentButtons = [
    {
        label: i18n.global.t('commons.button.edit'),
        permission: true,
        show: (row: Deployment) => row.engine === 'vllm' || (row.engine === 'ollama' && hasPermission('app_manage')),
        disabled: operationDisabled,
        click: (row: Deployment) => manageRef.value.open(row, 'config', true),
    },
    {
        label: i18n.global.t('aiTools.model.manage'),
        show: (row: Deployment) => row.engine !== 'vllm',
        click: manageDeployment,
    },
    {
        label: i18n.global.t('menu.monitor'),
        show: (row: Deployment) => row.engine === 'vllm',
        click: (row: Deployment) => manageRef.value.open(row, 'monitor', true),
    },
    {
        label: i18n.global.t('menu.terminal'),
        show: (row: Deployment) => row.engine === 'ollama' || row.engine === 'vllm',
        disabled: (row: Deployment) =>
            operationDisabled(row) || row.status !== 'Running' || !row.containerName || !isAdminOrNodeAdmin.value,
        click: (row: Deployment) =>
            terminalRef.value?.acceptParams({
                containerID: row.containerName,
                title: row.name,
                node: currentNode.value,
            }),
    },
    {
        label: i18n.global.t('commons.button.sync'),
        permission: true,
        show: (row: Deployment) => row.engine === 'ollama' || row.engine === 'vllm',
        disabled: operationDisabled,
        click: (row: Deployment) => syncRef.value.open(row),
    },
    {
        label: i18n.global.t('commons.button.log'),
        click: (row: Deployment) => manageRef.value.open(row, 'log', true),
    },
    {
        label: i18n.global.t('commons.button.delete'),
        permission: true,
        show: (row: Deployment) => row.engine === 'ollama' && hasPermission('app_manage'),
        disabled: (row: Deployment) => operationDisabled(row, 'delete'),
        click: async (row: Deployment) => {
            const node = currentNode.value;
            const path = route.fullPath;
            pending.value = `${row.engine}:${row.id}`;
            try {
                const [apps, resources] = await Promise.all([
                    readPages((page) => searchAppInstalled({ ...page, name: row.name, update: false }, node)),
                    appInstalledDeleteCheck(row.id, node),
                ]);
                if (node !== currentNode.value || path !== route.fullPath) return;
                const app = apps.find((item) => item.id === row.id && item.appKey === 'ollama');
                if (!app) {
                    MsgError(i18n.global.t('commons.res.notFound'));
                    return;
                }
                if (resources.data?.length) {
                    appResourcesRef.value.acceptParams({ items: resources.data, key: app.appKey, installID: app.id });
                } else {
                    await appDeleteRef.value.acceptParams(app);
                }
            } finally {
                pending.value = '';
            }
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        permission: true,
        show: (row: Deployment) => row.engine === 'vllm',
        disabled: (row: Deployment) => operationDisabled(row, 'delete') && row.status !== 'Uninstalling',
        click: (row: Deployment) => manageRef.value.open(row, 'delete', true),
    },
];
watch(
    [
        () => route.path,
        () => route.query.tab,
        () => route.query.name,
        () => route.query.open,
        currentNode,
        isProductPro,
        isFxplay,
    ],
    async (_, previous, onCleanup) => {
        let cancelled = false;
        onCleanup(() => (cancelled = true));
        if (!isCurrentPage.value) {
            searchVersion++;
            loading.value = false;
            return;
        }
        if (isLocal && route.query.tab === 'downloader') {
            await router.replace('/ai/model/files');
            return;
        }
        if (isLocal || isFiles) {
            apiType.value = '';
            engine.value = '';
            provider.value = '';
            searchName.value = '';
            resetPage();
        } else {
            if (previous.length && previous[0] !== pagePath) return;
            if (typeof route.query.name === 'string' && route.query.name !== searchName.value) {
                searchName.value = route.query.name;
                resetPage();
            }
        }
        await search();
        if (cancelled) return;
        if (
            isCurrentPage.value &&
            isLocal &&
            route.query.open === 'create' &&
            (route.query.engine === 'ollama' || route.query.tab === 'ollama')
        ) {
            await nextTick();
            if (cancelled) return;
            await createRef.value.open('ollama');
            if (cancelled) return;
            const query = { ...route.query };
            delete query.open;
            delete query.tab;
            await router.replace({ path: route.path, query });
        }
    },
    { immediate: true },
);
</script>

<style scoped lang="scss">
.service-name {
    min-width: 0;
}
.service-title {
    font-weight: 600;
}
.service-source-tag {
    margin-top: 4px;
}
.cell-ellipsis {
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
}
.model-line {
    display: flex;
    align-items: center;
    & + & {
        margin-top: 6px;
    }
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
