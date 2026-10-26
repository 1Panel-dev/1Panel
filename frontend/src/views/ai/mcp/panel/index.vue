<template>
    <div class="min-w-0">
        <RouterMenu />
        <LayoutContent :title="$t('panelMCP.credentials')" v-loading="loading || !serviceChecked">
            <template #app>
                <MCPService
                    ref="service"
                    @endpoint="endpoint = $event"
                    @status="updateServiceStatus"
                    @connection="showConnection()"
                    @changed="refresh"
                />
            </template>
            <template v-if="!showInstallEmpty" #leftToolBar>
                <div class="flex flex-wrap gap-2">
                    <el-button
                        v-permission="'mcp_access_manage'"
                        type="primary"
                        :disabled="!serviceReady || !catalogAvailable || !tools.length || !nodes.length"
                        @click="editor?.open()"
                    >
                        {{ $t('panelMCP.create') }}
                    </el-button>
                    <el-button type="primary" plain @click="invocations?.open()">
                        {{ $t('panelMCP.invocations') }}
                    </el-button>
                </div>
            </template>
            <template v-if="!showInstallEmpty" #rightToolBar><TableRefresh @search="refresh" /></template>
            <template v-if="!showInstallEmpty" #main>
                <div :class="{ mask: !serviceReady }" :inert="!serviceReady">
                    <el-alert
                        v-if="serviceReady && !catalogAvailable && !catalogLoading"
                        :title="$t('panelMCP.catalogUnavailable')"
                        type="warning"
                        :closable="false"
                        class="!mb-4"
                    >
                        <el-button link type="primary" @click="loadCatalog">
                            {{ $t('commons.button.refresh') }}
                        </el-button>
                    </el-alert>
                    <el-alert
                        v-else-if="serviceReady && catalogAvailable && !tools.length"
                        :title="$t('panelMCP.catalogEmpty')"
                        type="info"
                        :closable="false"
                        class="!mb-4"
                    />
                    <div class="w-full min-w-0 overflow-x-auto">
                        <ComplexTable
                            :data="displayItems"
                            :pagination-config="pagination"
                            :page-sizes="[10, 20, 50, 100]"
                            :height-diff="410"
                            @search="search"
                        >
                            <el-table-column :label="$t('commons.table.name')" prop="name" min-width="180">
                                <template #default="{ row }">
                                    <el-button
                                        type="primary"
                                        link
                                        class="!whitespace-normal !text-left break-words"
                                        @click="showConnection(row)"
                                    >
                                        {{ row.name }}
                                    </el-button>
                                    <div class="break-all text-xs text-[var(--el-text-color-secondary)]">
                                        {{ row.keyHint }}
                                    </div>
                                </template>
                            </el-table-column>
                            <el-table-column :label="$t('commons.table.status')" min-width="110">
                                <template #default="{ row }">
                                    <el-tag
                                        size="small"
                                        round
                                        effect="light"
                                        :type="
                                            row.status === 'Enable'
                                                ? 'success'
                                                : row.status === 'Disable'
                                                  ? 'warning'
                                                  : 'info'
                                        "
                                    >
                                        {{ $t('apiKeyManagement.status' + row.status) }}
                                    </el-tag>
                                </template>
                            </el-table-column>
                            <el-table-column :label="$t('panelMCP.tools')" min-width="220">
                                <template #default="{ row }">
                                    <div class="flex max-h-28 flex-wrap gap-1 overflow-y-auto">
                                        <el-tag
                                            v-for="tool in row.tools"
                                            :key="tool"
                                            size="small"
                                            class="!h-auto min-h-5 max-w-full !whitespace-normal break-all"
                                            :type="isReadOnly(tool) ? 'info' : 'warning'"
                                        >
                                            {{ toolLabel(tool) }}
                                        </el-tag>
                                        <span v-if="!row.tools.length">{{ $t('panelMCP.noTools') }}</span>
                                    </div>
                                    <div v-if="needsReview(row)" class="mt-1 text-xs text-[var(--el-color-warning)]">
                                        {{ $t('panelMCP.grantsNeedReview') }}
                                    </div>
                                </template>
                            </el-table-column>
                            <el-table-column :label="$t('panelMCP.nodes')" min-width="160">
                                <template #default="{ row }">
                                    <div class="flex max-h-28 flex-wrap gap-1 overflow-y-auto">
                                        <el-tag
                                            v-for="id in row.nodeIds"
                                            :key="id"
                                            size="small"
                                            class="!h-auto min-h-5 max-w-full !whitespace-normal break-all"
                                        >
                                            {{ nodeLabel(id) }}
                                        </el-tag>
                                    </div>
                                </template>
                            </el-table-column>
                            <el-table-column :label="$t('apiKeyManagement.expiresAt')" min-width="180">
                                <template #default="{ row }">
                                    {{
                                        row.expiresAt
                                            ? dateFormat(null, null, row.expiresAt)
                                            : $t('apiKeyManagement.never')
                                    }}
                                </template>
                            </el-table-column>
                            <fu-table-operations
                                :buttons="buttons"
                                :ellipsis="isMobile ? 0 : 2"
                                :min-width="isMobile ? 'auto' : 230"
                                :label="$t('commons.table.operate')"
                                fixed="right"
                            />
                        </ComplexTable>
                    </div>
                </div>
                <el-card v-if="serviceChecked && !serviceReady" class="mask-prompt" role="status">
                    <div class="flex flex-wrap items-center justify-center gap-x-3 gap-y-2">
                        <span>{{ servicePrompt }}</span>
                        <el-button
                            v-if="canInstallService"
                            type="primary"
                            link
                            icon="Position"
                            :disabled="service?.busy"
                            @click="service?.openInstall()"
                        >
                            {{ $t('commons.button.install') }}
                        </el-button>
                        <el-button v-else-if="serviceStatus" type="primary" link @click="service?.openStatus()">
                            {{ $t('commons.table.status') }}
                        </el-button>
                        <el-button type="primary" link :disabled="service?.busy" @click="service?.refreshService()">
                            {{ $t('commons.button.refresh') }}
                        </el-button>
                    </div>
                    <p v-if="serviceStatus && !serviceStatus.canManageService" class="input-help">
                        {{ $t('panelMCP.serviceAdminOnly') }}
                    </p>
                </el-card>
            </template>
            <NoSuchService v-if="showInstallEmpty" :name="$t('panelMCP.title')">
                <span>{{ $t('panelMCP.installEmpty') }}</span>
                <el-button
                    v-if="canInstallService"
                    type="primary"
                    link
                    icon="Position"
                    :disabled="service?.busy"
                    @click="service?.openInstall()"
                >
                    {{ $t('firewall.quickJump') }}
                </el-button>
                <p v-else class="input-help !m-0">{{ $t('panelMCP.serviceAdminOnly') }}</p>
            </NoSuchService>
        </LayoutContent>
        <CredentialEditor
            ref="editor"
            :tools="tools"
            :tool-node-ids="toolNodeIds"
            :nodes="nodes"
            :catalog-available="catalogAvailable"
            :endpoint="endpoint"
            @changed="refresh"
        />
        <Invocations ref="invocations" />
        <DrawerPro
            v-model="connectionVisible"
            :header="$t('panelMCP.connection')"
            :resource="detail?.name || ''"
            size="min(720px, 100vw)"
        >
            <p v-if="!detail" class="input-help mb-4 mt-0 leading-5">
                {{ $t('panelMCP.connectionHelp') }}
            </p>
            <el-descriptions :column="1" :label-width="130" border size="small" class="connection-details">
                <el-descriptions-item :label="$t('panelMCP.endpoint')">
                    <div class="flex min-w-0 items-start gap-2">
                        <span class="min-w-0 break-all">{{ endpoint || $t('panelMCP.endpointUnavailable') }}</span>
                        <CopyButton v-if="endpoint" :content="endpoint" class="shrink-0" />
                    </div>
                </el-descriptions-item>
                <template v-if="detail">
                    <el-descriptions-item :label="$t('commons.table.name')">
                        <span class="break-words">{{ detail.name }}</span>
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('commons.table.description')">
                        <span class="break-words">{{ detail.description || '—' }}</span>
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('commons.table.status')">
                        <el-tag
                            size="small"
                            round
                            effect="light"
                            :type="
                                detailStatus === 'Enable' ? 'success' : detailStatus === 'Disable' ? 'warning' : 'info'
                            "
                        >
                            {{ $t('apiKeyManagement.status' + detailStatus) }}
                        </el-tag>
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('panelMCP.tools')">
                        <div class="flex flex-wrap gap-1">
                            <el-tag
                                v-for="tool in detail.tools"
                                :key="tool"
                                size="small"
                                class="!h-auto min-h-5 max-w-full !whitespace-normal break-all"
                                :type="isReadOnly(tool) ? 'info' : 'warning'"
                            >
                                {{ toolLabel(tool) }}
                            </el-tag>
                            <span v-if="!detail.tools.length">{{ $t('panelMCP.noTools') }}</span>
                        </div>
                        <p v-if="needsReview(detail)" class="input-help mt-2 leading-5">
                            {{ $t('panelMCP.grantsNeedReview') }}
                        </p>
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('panelMCP.nodes')">
                        <div class="flex flex-wrap gap-1">
                            <el-tag
                                v-for="id in detail.nodeIds"
                                :key="id"
                                size="small"
                                class="!h-auto min-h-5 max-w-full !whitespace-normal break-all"
                            >
                                {{ nodeLabel(id) }}
                            </el-tag>
                        </div>
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('apiKeyManagement.expiresAt')">
                        {{ detail.expiresAt ? dateFormat(null, null, detail.expiresAt) : $t('apiKeyManagement.never') }}
                    </el-descriptions-item>
                </template>
            </el-descriptions>
            <template v-if="detail">
                <el-alert
                    v-if="connectionError"
                    :title="connectionError"
                    type="warning"
                    :closable="false"
                    class="!mt-5"
                />
                <el-form label-position="top" class="mt-5 min-w-0">
                    <el-form-item class="[&>label]:!w-full">
                        <template #label>
                            <div class="flex w-full min-w-0 flex-wrap items-center justify-between gap-x-3 gap-y-1">
                                <span>Token</span>
                                <el-button
                                    v-if="canRotate"
                                    link
                                    type="info"
                                    class="!ml-0 max-w-full !whitespace-normal text-right"
                                    :disabled="connectionBusy"
                                    @click="rotateToken"
                                >
                                    {{ $t('panelMCP.rotateToken') }}
                                </el-button>
                            </div>
                        </template>
                        <el-input
                            :model-value="connectionToken || detail.keyHint || '••••••••'"
                            :type="connectionToken ? 'text' : 'password'"
                            readonly
                            autocomplete="off"
                            class="[&_input]:!font-mono"
                        >
                            <template v-if="canManageCredential" #suffix>
                                <div class="flex items-center gap-2 [&>.el-button]:!ml-0">
                                    <el-button
                                        link
                                        class="!h-7 !w-7 !p-0"
                                        :icon="connectionToken ? 'Hide' : 'View'"
                                        :title="$t(connectionToken ? 'panelMCP.hideToken' : 'panelMCP.revealToken')"
                                        :aria-label="
                                            $t(connectionToken ? 'panelMCP.hideToken' : 'panelMCP.revealToken')
                                        "
                                        :aria-pressed="!!connectionToken"
                                        :disabled="!canReveal || connectionBusy"
                                        @click="toggleToken"
                                    />
                                    <el-button
                                        link
                                        class="!h-7 !w-7 !p-0"
                                        icon="DocumentCopy"
                                        :title="$t('panelMCP.copyToken')"
                                        :aria-label="$t('panelMCP.copyToken')"
                                        :disabled="!canReveal || connectionBusy"
                                        @click="revealToken(true)"
                                    />
                                </div>
                            </template>
                        </el-input>
                        <span class="input-help mt-2 leading-5">{{ $t('panelMCP.tokenMemoryHelp') }}</span>
                    </el-form-item>
                </el-form>
            </template>
            <el-form label-position="top" class="mt-5 min-w-0">
                <el-form-item :label="$t('panelMCP.clientType')">
                    <el-select v-model="selectedClient" class="w-full" :disabled="connectionBusy">
                        <el-option :label="$t('panelMCP.genericClient')" value="generic" />
                        <el-option label="Codex" value="codex" />
                        <el-option label="Claude" value="claude" />
                    </el-select>
                    <span v-if="detail && !connectionToken" class="input-help mt-2 leading-5">
                        {{ $t('panelMCP.revealConfigHelp') }}
                    </span>
                </el-form-item>
            </el-form>
            <div v-if="clientConfig" class="mt-5 flex flex-wrap items-center justify-between gap-2">
                <span>{{ $t('panelMCP.clientConfig') }}</span>
                <CopyButton :content="clientConfig" :is-icon="false" />
            </div>
            <pre
                v-if="clientConfig"
                class="mt-2 max-h-80 overflow-auto whitespace-pre-wrap break-all rounded border border-solid border-[var(--el-border-color)] bg-[var(--el-fill-color-light)] p-3 font-mono text-xs leading-5"
                >{{ clientConfig }}</pre>
            <template #footer>
                <div class="flex flex-wrap items-center justify-end gap-2 [&>.el-button]:!ml-0">
                    <el-button
                        v-if="detail && canManageCredential"
                        :disabled="!canReveal || connectionBusy || !endpoint"
                        @click="exportConnection"
                    >
                        {{ $t('panelMCP.downloadGuide') }}
                    </el-button>
                    <el-button v-if="detail" type="primary" plain @click="invocations?.open(detail)">
                        {{ $t('panelMCP.invocations') }}
                    </el-button>
                    <el-button @click="connectionVisible = false">{{ $t('commons.button.close') }}</el-button>
                </div>
            </template>
        </DrawerPro>
    </div>
</template>

<script setup lang="ts">
import { computed, onActivated, onBeforeUnmount, onDeactivated, onMounted, reactive, ref, watch } from 'vue';
import { onBeforeRouteLeave } from 'vue-router';
import { ElMessageBox } from 'element-plus';
import type { PanelMCP } from '@/api/interface/panel-mcp';
import {
    getPanelMCPCatalog,
    getPanelMCPNodes,
    revokePanelMCPCredential,
    searchPanelMCPCredentials,
    setPanelMCPCredentialStatus,
    revealPanelMCPCredential,
    rotatePanelMCPCredential,
    exportPanelMCPCredential,
} from '@/api/modules/panel-mcp';
import type { FuTableOperationButton } from '@/components/table/shared';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { dateFormat } from '@/utils/date';
import { MsgSuccess } from '@/utils/message';
import useClipboard from 'vue-clipboard3';
import { downloadBlob } from '@/utils/file';
import { hasManagePermissionAccess } from '@/utils/permission';
import {
    buildMCPClientConfig,
    getMCPCredentialStatus,
    getMCPStaleTools,
    isMCPMutationUncertain,
    isUsableMCPCatalog,
    isUsableMCPNodes,
} from '@/utils/panel-mcp';
import i18n from '@/lang';
import NoSuchService from '@/components/layout-content/no-such-service.vue';
import RouterMenu from '@/views/ai/mcp/index.vue';
import MCPService from './service.vue';
import CredentialEditor from './credential-editor.vue';
import Invocations from './invocations.vue';

const { isMobile, globalStore } = useGlobalStore();
const { toClipboard } = useClipboard();
const loading = ref(false);
const credentialsLoaded = ref(false);
const mutating = ref(false);
const items = ref<PanelMCP.Credential[]>([]);
const tools = ref<PanelMCP.Tool[]>([]);
const toolNodeIds = ref<Record<string, string[]>>({});
const nodes = ref<PanelMCP.Node[]>([]);
const catalogAvailable = ref(false);
const catalogLoading = ref(true);
const endpoint = ref('');
const service = ref<InstanceType<typeof MCPService>>();
const serviceStatus = ref<PanelMCP.ServiceStatus>();
const serviceChecked = ref(false);
const serviceReady = computed(
    () =>
        !!serviceStatus.value?.installed &&
        serviceStatus.value.enabled &&
        serviceStatus.value.status === 'running' &&
        serviceStatus.value.ready,
);
const canInstallService = computed(
    () =>
        serviceStatus.value?.status === 'not_installed' &&
        !serviceStatus.value.installed &&
        serviceStatus.value.canManageService,
);
const servicePrompt = computed(() => {
    const current = serviceStatus.value;
    if (!current) return i18n.global.t('panelMCP.statusUnavailable');
    if (current.status === 'not_installed' && !current.installed) {
        return i18n.global.t('panelMCP.installEmpty');
    }
    if (current.status === 'unsupported') return i18n.global.t('panelMCP.serviceUnsupported');
    if (current.installed && (current.status === 'stopped' || !current.enabled)) {
        return i18n.global.t('commons.service.serviceNotStarted', [i18n.global.t('panelMCP.title')]);
    }
    if (current.status === 'incompatible') return i18n.global.t('panelMCP.statusIncompatible');
    return i18n.global.t('panelMCP.serviceUnhealthy');
});
const updateServiceStatus = (value: PanelMCP.ServiceStatus | undefined) => {
    const wasReady = serviceReady.value;
    serviceStatus.value = value;
    serviceChecked.value = true;
    if (!serviceReady.value) {
        catalogVersion++;
        catalogAvailable.value = false;
        catalogLoading.value = false;
    } else if (!wasReady) {
        void loadCatalog();
    }
};
const connectionVisible = ref(false);
const detail = ref<PanelMCP.Credential>();
const connectionToken = ref('');
const connectionBusy = ref(false);
const connectionError = ref('');
const selectedClient = ref<PanelMCP.Client>('generic');
const canManageCredential = computed(() => hasManagePermissionAccess('mcp_access_manage'));
const canReveal = computed(() => canManageCredential.value && detailStatus.value === 'Enable');
const canRotate = computed(
    () => canManageCredential.value && !!detail.value && ['Enable', 'Disable'].includes(detailStatus.value),
);
let connectionVersion = 0;
let connectionController: AbortController | undefined;
const clearConnectionSecret = () => {
    connectionVersion++;
    connectionController?.abort();
    connectionController = undefined;
    connectionToken.value = '';
    connectionError.value = '';
    connectionBusy.value = false;
};
const editor = ref<InstanceType<typeof CredentialEditor>>();
const invocations = ref<InstanceType<typeof Invocations>>();
const now = ref(Date.now());
const pagination = reactive({ currentPage: 1, pageSize: 20, total: 0, cacheSizeKey: 'panel-mcp-credentials-page' });
const showInstallEmpty = computed(
    () =>
        serviceChecked.value &&
        serviceStatus.value?.status === 'not_installed' &&
        !serviceStatus.value.installed &&
        credentialsLoaded.value &&
        !loading.value &&
        !items.value.length &&
        pagination.total === 0,
);
const clientConfig = computed(() =>
    detail.value && !connectionToken.value
        ? ''
        : buildMCPClientConfig(
              endpoint.value,
              connectionToken.value || '<MCP_TOKEN>',
              selectedClient.value,
              detail.value?.id,
          ),
);
const displayItems = computed(() =>
    items.value.map((item) => ({ ...item, status: getMCPCredentialStatus(item, now.value) })),
);
const detailStatus = computed(() => (detail.value ? getMCPCredentialStatus(detail.value, now.value) : 'Disable'));
const isReadOnly = (id: string) => tools.value.find((tool) => tool.toolId === id)?.readOnly === true;
const toolLabel = (id: string) => tools.value.find((tool) => tool.toolId === id)?.name || id;
const nodeLabel = (id: string) =>
    id === 'local'
        ? i18n.global.t('panelMCP.primaryNode')
        : `${nodes.value.find((node) => node.id === id)?.name || 'ID'} (${id})`;
const needsReview = (credential: PanelMCP.Credential) =>
    catalogAvailable.value && getMCPStaleTools(credential, tools.value).length > 0;
let active = false;
let pageVersion = 0;
let requestVersion = 0;
let catalogVersion = 0;
let timer: ReturnType<typeof setInterval> | undefined;
const isCurrentPage = (version: number) => active && version === pageVersion;
const search = async () => {
    if (!active) return;
    const ticket = ++requestVersion;
    loading.value = true;
    try {
        const response = await searchPanelMCPCredentials({
            page: pagination.currentPage,
            pageSize: pagination.pageSize,
        });
        if (!active || ticket !== requestVersion) return;
        items.value = response.data.items || [];
        pagination.total = response.data.total;
        credentialsLoaded.value = true;
        now.value = Date.now();
        if (detail.value) {
            const updated = items.value.find((item) => item.id === detail.value?.id);
            if (updated) detail.value = updated;
        }
    } catch {
        if (active && ticket === requestVersion) {
            items.value = [];
            pagination.total = 0;
            credentialsLoaded.value = false;
        }
    } finally {
        if (ticket === requestVersion) loading.value = false;
    }
};
const loadCatalog = async () => {
    const ticket = ++catalogVersion;
    catalogAvailable.value = false;
    catalogLoading.value = true;
    try {
        const [response, nodeResponse] = await Promise.all([getPanelMCPCatalog(), getPanelMCPNodes()]);
        if (!active || ticket !== catalogVersion) return;
        if (!isUsableMCPCatalog(response.data) || !isUsableMCPNodes(nodeResponse.data)) return;
        tools.value = response.data.tools;
        toolNodeIds.value = response.data.toolNodeIds;
        nodes.value = nodeResponse.data;
        catalogAvailable.value = true;
    } catch {
    } finally {
        if (active && ticket === catalogVersion) catalogLoading.value = false;
    }
};
const refresh = () => {
    void search();
    if (serviceReady.value) void loadCatalog();
};
const showConnection = (entry?: PanelMCP.Credential) => {
    clearConnectionSecret();
    detail.value = entry;
    selectedClient.value = 'generic';
    connectionVisible.value = true;
};
const currentConnection = (ticket: number, reference: PanelMCP.Reference) =>
    ticket === connectionVersion &&
    connectionVisible.value &&
    globalStore.isLogin &&
    canManageCredential.value &&
    detail.value?.id === reference.id &&
    detail.value?.revision === reference.revision;
const isRotationRejected = (error: unknown) => {
    const failure = error as {
        code?: number | string;
        errorCode?: string;
        response?: { status?: number; data?: { code?: number; errorCode?: string } };
    };
    const status = failure?.response?.status;
    if (typeof status === 'number') {
        if (status === 408 || status >= 500) return false;
        if (status >= 400) return true;
        if (status !== 200) return false;
    }
    const business = failure?.response?.data || failure;
    if (isMCPMutationUncertain(business)) return false;
    if (
        (typeof business?.code === 'number' && business.code >= 400) ||
        (typeof business?.errorCode === 'string' && business.errorCode.startsWith('MCP_'))
    ) {
        return true;
    }
    return false;
};
const revealToken = async (copy: boolean) => {
    if (!canReveal.value || connectionBusy.value || !detail.value || !connectionVisible.value || !globalStore.isLogin)
        return;
    const reference = { id: detail.value.id, revision: detail.value.revision };
    const ticket = ++connectionVersion;
    connectionController = new AbortController();
    connectionBusy.value = true;
    connectionError.value = '';
    try {
        const response = await revealPanelMCPCredential(reference, connectionController.signal);
        if (!currentConnection(ticket, reference) || !canReveal.value) return;
        if (response.data?.id !== reference.id || typeof response.data.token !== 'string' || !response.data.token) {
            throw new Error('Invalid MCP token response');
        }
        if (copy) {
            await toClipboard(response.data.token);
            if (!currentConnection(ticket, reference) || !canReveal.value) return;
            MsgSuccess(i18n.global.t('commons.msg.copySuccess'));
        } else {
            connectionToken.value = response.data.token;
        }
        connectionError.value = '';
    } catch {
        if (currentConnection(ticket, reference)) {
            connectionError.value = copy ? i18n.global.t('panelMCP.copyTokenFailed') : '';
        }
    } finally {
        if (ticket === connectionVersion) {
            connectionBusy.value = false;
            connectionController = undefined;
        }
    }
};
const toggleToken = async () => {
    if (!canReveal.value || connectionBusy.value || !connectionVisible.value || !globalStore.isLogin) return;
    if (connectionToken.value) clearConnectionSecret();
    else await revealToken(false);
};
const rotateToken = async () => {
    if (!canRotate.value || connectionBusy.value || !detail.value || !connectionVisible.value || !globalStore.isLogin)
        return;
    const reference = { id: detail.value.id, revision: detail.value.revision };
    const opening = connectionVersion;
    connectionError.value = '';
    try {
        await ElMessageBox.confirm(i18n.global.t('panelMCP.rotateTokenConfirm'), i18n.global.t('panelMCP.rotateToken'));
    } catch {
        return;
    }
    if (!currentConnection(opening, reference) || !canRotate.value || connectionBusy.value) return;
    const ticket = ++connectionVersion;
    connectionController = new AbortController();
    connectionBusy.value = true;
    connectionToken.value = '';
    connectionError.value = '';
    try {
        const response = await rotatePanelMCPCredential(reference, connectionController.signal);
        if (!currentConnection(ticket, reference)) return;
        if (
            response.data?.item?.id !== reference.id ||
            !Number.isSafeInteger(response.data.item.revision) ||
            response.data.item.revision <= reference.revision ||
            typeof response.data.token !== 'string' ||
            !response.data.token
        ) {
            throw new Error('Invalid MCP token rotation response');
        }
        detail.value = response.data.item;
        items.value = items.value.map((item) => (item.id === reference.id ? response.data.item : item));
        connectionError.value = '';
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        void search();
    } catch (error) {
        if (currentConnection(ticket, reference)) {
            connectionError.value = i18n.global.t(
                isRotationRejected(error) ? 'panelMCP.rotateTokenFailed' : 'panelMCP.rotateTokenUncertain',
            );
        }
    } finally {
        if (ticket === connectionVersion) {
            connectionBusy.value = false;
            connectionController = undefined;
        }
    }
};
const exportConnection = async () => {
    if (
        !canReveal.value ||
        connectionBusy.value ||
        !detail.value ||
        !endpoint.value ||
        !connectionVisible.value ||
        !globalStore.isLogin
    )
        return;
    const reference = { id: detail.value.id, revision: detail.value.revision };
    const client = selectedClient.value;
    const ticket = ++connectionVersion;
    connectionController = new AbortController();
    connectionBusy.value = true;
    connectionError.value = '';
    try {
        const response = await exportPanelMCPCredential({ ...reference, client }, connectionController.signal);
        if (!currentConnection(ticket, reference) || !canReveal.value || selectedClient.value !== client) return;
        downloadBlob(response.data.content, response.data.filename, 'text/markdown;charset=utf-8');
        connectionError.value = '';
    } catch {
        if (currentConnection(ticket, reference)) connectionError.value = i18n.global.t('panelMCP.exportFailed');
    } finally {
        if (ticket === connectionVersion) {
            connectionBusy.value = false;
            connectionController = undefined;
        }
    }
};
const operate = async (entry: PanelMCP.Credential, revoke = false) => {
    if (mutating.value) return;
    const version = pageVersion;
    const currentStatus = getMCPCredentialStatus(entry);
    if (!revoke && currentStatus !== 'Enable' && currentStatus !== 'Disable') {
        await search();
        return;
    }
    if (!revoke && currentStatus === 'Disable' && (!catalogAvailable.value || needsReview(entry))) return;
    const key = revoke ? 'revokeConfirm' : currentStatus === 'Enable' ? 'disableConfirm' : 'enableConfirm';
    try {
        await ElMessageBox.confirm(
            i18n.global.t(`panelMCP.${key}`, [entry.name]),
            i18n.global.t('panelMCP.credentials'),
        );
    } catch {
        return;
    }
    if (!isCurrentPage(version) || mutating.value) return;
    mutating.value = true;
    try {
        const reference = { id: entry.id, revision: entry.revision };
        if (revoke) await revokePanelMCPCredential(reference);
        else
            await setPanelMCPCredentialStatus({
                ...reference,
                status: currentStatus === 'Enable' ? 'Disable' : 'Enable',
            });
        if (!isCurrentPage(version)) return;
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    } finally {
        mutating.value = false;
        if (isCurrentPage(version)) await search();
    }
};
const buttons = computed<FuTableOperationButton<PanelMCP.Credential>[]>(() => [
    { label: i18n.global.t('panelMCP.connection'), click: showConnection },
    {
        label: i18n.global.t('commons.button.edit'),
        permission: 'mcp_access_manage',
        show: (row) => row.status !== 'Revoked',
        disabled: () => mutating.value || !catalogAvailable.value,
        click: (row) => editor.value?.open(row),
    },
    { label: i18n.global.t('panelMCP.invocations'), click: (row) => invocations.value?.open(row) },
    {
        label: i18n.global.t('commons.button.disable'),
        permission: 'mcp_access_manage',
        show: (row) => row.status === 'Enable',
        disabled: () => mutating.value,
        click: (row) => operate(row),
    },
    {
        label: i18n.global.t('commons.button.enable'),
        permission: 'mcp_access_manage',
        show: (row) => row.status === 'Disable',
        disabled: (row) => mutating.value || !catalogAvailable.value || needsReview(row),
        click: (row) => operate(row),
    },
    {
        label: i18n.global.t('apiKeyManagement.revoke'),
        permission: 'mcp_access_manage',
        show: (row) => row.status !== 'Revoked',
        disabled: () => mutating.value,
        click: (row) => operate(row, true),
    },
]);
const activate = () => {
    if (active) return;
    active = true;
    refresh();
    timer = setInterval(() => {
        now.value = Date.now();
    }, 1000);
};
const deactivate = () => {
    clearConnectionSecret();
    active = false;
    pageVersion++;
    requestVersion++;
    catalogVersion++;
    catalogAvailable.value = false;
    catalogLoading.value = false;
    serviceStatus.value = undefined;
    serviceChecked.value = false;
    credentialsLoaded.value = false;
    if (timer) clearInterval(timer);
    timer = undefined;
    connectionVisible.value = false;
    detail.value = undefined;
};
watch(
    connectionVisible,
    (value) => {
        if (!value) clearConnectionSecret();
    },
    { flush: 'sync' },
);
watch(() => `${detail.value?.id}:${detail.value?.revision}`, clearConnectionSecret, { flush: 'sync' });
watch(
    () => globalStore.isLogin,
    (value) => {
        if (!value) {
            clearConnectionSecret();
            connectionVisible.value = false;
        }
    },
    { flush: 'sync' },
);
watch(
    canManageCredential,
    (value) => {
        if (!value) clearConnectionSecret();
    },
    { flush: 'sync' },
);
watch(
    detailStatus,
    (value) => {
        if (value !== 'Enable') clearConnectionSecret();
    },
    { flush: 'sync' },
);
onBeforeRouteLeave(() => {
    clearConnectionSecret();
    connectionVisible.value = false;
});
onMounted(activate);
onActivated(activate);
onDeactivated(deactivate);
onBeforeUnmount(deactivate);
</script>

<style scoped lang="scss">
.connection-details {
    :deep(.el-descriptions__table) {
        table-layout: fixed;
    }
    :deep(.el-descriptions__label) {
        white-space: normal;
        overflow-wrap: anywhere;
    }
    :deep(.el-descriptions__content) {
        overflow-wrap: anywhere;
    }
}
</style>
