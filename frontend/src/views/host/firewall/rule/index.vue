<template>
    <div>
        <FireRouter />

        <div>
            <FireStatus
                ref="fireStatusRef"
                v-model:loading="loading"
                v-model:is-active="isActive"
                v-model:is-init="isInit"
                v-model:is-bind="isBind"
                v-model:name="provider"
                v-model:version="firewallVersion"
                current-tab="base"
                @search="search"
            >
                <template #actions>
                    <el-divider v-if="canReset" direction="vertical" />
                    <el-button
                        v-permission
                        v-node-admin
                        type="primary"
                        link
                        :disabled="loading"
                        v-if="canReset"
                        :loading="resetting"
                        @click="resetRules"
                    >
                        {{ $t('commons.button.reset') }}
                    </el-button>
                </template>
            </FireStatus>

            <div v-if="provider !== '-'">
                <el-card v-if="showFirewallUnavailablePrompt" class="mask-prompt">
                    <span v-if="isServiceBackend">{{ $t('firewall.firewallNotStart') }}</span>
                    <span v-else-if="!isInit">{{ $t('firewall.initHelper', [provider]) }}</span>
                    <span v-else>{{ $t('firewall.basicStatus') }}</span>
                </el-card>

                <LayoutContent :title="$t('menu.firewall')" :class="{ mask: !isFirewallReady }">
                    <template #prompt>
                        <div ref="noticeRef" class="flow-root">
                            <el-alert
                                v-for="notice in notices"
                                :key="notice.key"
                                class="mb-2"
                                type="warning"
                                :closable="false"
                                :title="notice.text"
                            />
                        </div>
                    </template>
                    <template #leftToolBar>
                        <el-button v-permission v-node-admin type="primary" :disabled="loading" @click="openCreate">
                            {{ $t('commons.button.create') }}
                        </el-button>
                        <el-button
                            v-permission
                            v-node-admin
                            :disabled="loading || !selects.some(isDeletableRule)"
                            @click="removeSelectedRules"
                        >
                            {{ $t('commons.button.delete') }}
                        </el-button>
                        <el-button-group>
                            <el-button v-permission v-node-admin :disabled="loading" @click="openImport">
                                {{ $t('commons.button.import') }}
                            </el-button>
                            <el-button
                                v-permission
                                :disabled="loading || inventoryTotal === 0"
                                @click="exportRulesBySelection"
                            >
                                {{ $t('commons.button.export') }}
                            </el-button>
                        </el-button-group>
                    </template>
                    <template #rightToolBar>
                        <div class="firewall-filter-bar">
                            <el-select
                                v-model="selectedRuleFilters"
                                class="firewall-rule-filter"
                                :placeholder="$t('menu.filter')"
                                multiple
                                clearable
                                collapse-tags
                                collapse-tags-tooltip
                                :max-collapse-tags="2"
                                popper-class="firewall-rule-filter-popper"
                                @change="changeRuleFilter"
                            >
                                <el-option-group label="IP">
                                    <el-option label="IPv4" value="family:ipv4" />
                                    <el-option label="IPv6" value="family:ipv6" />
                                </el-option-group>
                                <el-option-group :label="$t('firewall.action')">
                                    <el-option :label="$t('firewall.accept')" value="action:accept" />
                                    <el-option :label="$t('firewall.drop')" value="action:deny" />
                                </el-option-group>
                            </el-select>
                        </div>
                        <TableSearch v-model:searchName="searchName" @search="searchWithReset" />
                        <TableRefresh @search="search" />
                        <TableSetting title="firewall-rule-refresh" @search="!loading && !usageLoading && search()" />
                    </template>
                    <template #main>
                        <div v-loading="loading">
                            <ComplexTable
                                v-model:selects="selects"
                                :pagination-config="paginationConfig"
                                :data="allRows"
                                :heightDiff="320 + noticeHeight"
                                row-key="rowKey"
                                @search="search"
                            >
                                <el-table-column type="selection" width="48" fix />
                                <el-table-column :label="$t('firewall.action')" width="110">
                                    <template #default="{ row }">
                                        <span
                                            class="firewall-action"
                                            :class="{
                                                'is-accept': row.rule.action === 'accept',
                                                'is-drop': isDenyAction(row.rule.action),
                                                'is-unknown': !isKnownAction(row.rule.action),
                                            }"
                                        >
                                            <i
                                                v-if="isKnownAction(row.rule.action)"
                                                class="iconfont firewall-action-icon"
                                                :class="row.rule.action === 'accept' ? 'p-yunxu1' : 'p-a-44tubiao-226'"
                                                aria-hidden="true"
                                            />
                                            {{ actionLabel(row.rule.action) }}
                                            <el-tooltip
                                                v-if="row.observed.protected && !row.isWhitelist"
                                                :content="$t('firewall.builtinRuleProtected')"
                                                placement="top"
                                                :show-after="200"
                                                :popper-style="{ maxWidth: '320px' }"
                                            >
                                                <el-icon
                                                    class="firewall-action-lock"
                                                    :aria-label="$t('firewall.builtinRuleProtected')"
                                                    tabindex="0"
                                                >
                                                    <Lock />
                                                </el-icon>
                                            </el-tooltip>
                                        </span>
                                    </template>
                                </el-table-column>
                                <el-table-column :label="$t('firewall.priority')" width="90">
                                    <template #default="{ row }">
                                        {{ displayRulePriority(row) }}
                                    </template>
                                </el-table-column>
                                <el-table-column :label="$t('commons.table.protocol')" width="110">
                                    <template #default="{ row }">
                                        {{ displayProtocol(row) }}
                                    </template>
                                </el-table-column>
                                <el-table-column label="IP" min-width="240" show-overflow-tooltip>
                                    <template #default="{ row }">
                                        <span>
                                            {{ displayAddress(row) }}
                                        </span>
                                    </template>
                                </el-table-column>
                                <el-table-column
                                    :label="$t('commons.table.port')"
                                    min-width="180"
                                    show-overflow-tooltip
                                >
                                    <template #default="{ row }">
                                        <span>
                                            {{ displayPort(row) }}
                                        </span>
                                    </template>
                                </el-table-column>
                                <el-table-column :label="$t('firewall.used')" min-width="200">
                                    <template #header>
                                        <span>{{ $t('firewall.used') }}</span>
                                        <el-button
                                            link
                                            :icon="Refresh"
                                            :disabled="usageLoading"
                                            :aria-label="$t('commons.button.refresh')"
                                            @click="refreshUsage"
                                        />
                                    </template>
                                    <template #default="{ row }">
                                        <span v-if="hasIncompleteParsing(row)">-</span>
                                        <el-icon v-else-if="usageLoading" class="is-loading"><Loading /></el-icon>
                                        <span v-else-if="usageFailed">-</span>
                                        <el-tag
                                            v-else-if="usageEntriesByRow[row.rowKey].length === 0"
                                            type="info"
                                            size="small"
                                        >
                                            {{ $t('firewall.unUsed') }}
                                        </el-tag>
                                        <div v-else class="firewall-used-cell">
                                            <el-tooltip
                                                :content="usageEntryLabel(usageEntriesByRow[row.rowKey][0])"
                                                placement="top"
                                            >
                                                <el-tag
                                                    class="cursor-pointer firewall-used-entry"
                                                    type="info"
                                                    effect="plain"
                                                    size="small"
                                                    @click.stop="openUsageDetail(usageEntriesByRow[row.rowKey][0])"
                                                >
                                                    <span class="firewall-used-entry-owner">
                                                        {{ usageEntriesByRow[row.rowKey][0].owner }}
                                                    </span>
                                                    <span class="firewall-used-entry-port">
                                                        ({{ usageEntryPortText(usageEntriesByRow[row.rowKey][0]) }})
                                                    </span>
                                                    <el-icon class="firewall-used-entry-icon"><Expand /></el-icon>
                                                </el-tag>
                                            </el-tooltip>
                                            <el-popover
                                                v-if="usageEntriesByRow[row.rowKey].length > 1"
                                                placement="right"
                                                trigger="click"
                                                :width="340"
                                                :persistent="false"
                                            >
                                                <template #reference>
                                                    <el-tag
                                                        class="cursor-pointer firewall-used-more"
                                                        type="info"
                                                        effect="plain"
                                                        size="small"
                                                    >
                                                        +{{ usageEntriesByRow[row.rowKey].length - 1 }}
                                                    </el-tag>
                                                </template>
                                                <div class="firewall-used-popover-list">
                                                    <el-tooltip
                                                        v-for="entry in usageEntriesByRow[row.rowKey]"
                                                        :key="`${row.rowKey}:${entry.key}`"
                                                        :content="usageEntryLabel(entry)"
                                                        placement="top"
                                                    >
                                                        <el-tag
                                                            class="cursor-pointer firewall-used-popover-entry"
                                                            type="info"
                                                            effect="plain"
                                                            size="small"
                                                            @click.stop="openUsageDetail(entry)"
                                                        >
                                                            <span class="firewall-used-entry-owner">
                                                                {{ entry.owner }}
                                                            </span>
                                                            <span class="firewall-used-entry-port">
                                                                ({{ usageEntryPortText(entry) }})
                                                            </span>
                                                            <el-icon class="firewall-used-entry-icon">
                                                                <Expand />
                                                            </el-icon>
                                                        </el-tag>
                                                    </el-tooltip>
                                                </div>
                                            </el-popover>
                                        </div>
                                    </template>
                                </el-table-column>
                                <el-table-column
                                    :label="$t('commons.table.description')"
                                    min-width="160"
                                    prop="rule.description"
                                    show-overflow-tooltip
                                />
                                <fu-table-operations
                                    width="160px"
                                    :buttons="operationButtons"
                                    :label="$t('commons.table.operate')"
                                    fix
                                />
                            </ComplexTable>
                        </div>
                    </template>
                </LayoutContent>
            </div>
        </div>
        <RuleOperate ref="ruleOperateRef" @search="search" @created="openRuleTask" />
        <RuleImport ref="ruleImportRef" @created="openRuleTask" />
        <TaskLog ref="ruleTaskLogRef" @close="fireStatusRef?.acceptParams()" />
        <ProcessDetail ref="processDetailRef" />
        <RuleReset ref="resetConfirmRef" @confirm="prepareResetRules" />
        <DockerRestart
            ref="dockerRestartRef"
            v-model:withDockerRestart="withDockerRestart"
            :title="$t('commons.button.reset')"
            @submit="submitResetRules"
        />
    </div>
</template>

<script lang="ts" setup>
import { Firewall } from '@/api/interface/firewall';
import type { FuTableOperationButton } from '@/components/table/shared';
import { buildHostRuleExport } from './transfer';
import { Process } from '@/api/interface/process';
import {
    deleteFirewallRules,
    loadDockerPublishedPorts,
    loadFirewallNativeDetail,
    resetFirewallRules,
    searchFirewallRules,
} from '@/api/modules/firewall';
import { getListeningProcess } from '@/api/modules/process';
import i18n from '@/lang';
import { getCurrentDateFormatted } from '@/utils/date';
import { downloadWithContent } from '@/utils/file';
import { MsgError, MsgInfo, MsgSuccess } from '@/utils/message';
import { dockerGuardEndpointManagementMessage, dockerGuardManagementTarget } from '@/views/host/firewall/docker/model';
import { formatHostAddress } from '@/views/host/firewall/utils/validation';
import RuleImport from '@/views/host/firewall/rule/import/index.vue';
import RuleOperate from '@/views/host/firewall/rule/operate/index.vue';
import FireRouter from '@/views/host/firewall/index.vue';
import FireStatus from '@/views/host/firewall/status/index.vue';
import ProcessDetail from '@/views/host/process/process/detail/index.vue';
import RuleReset from '@/views/host/firewall/components/rule-reset.vue';
import TaskLog from '@/components/log/task/index.vue';
import DockerRestart from '@/components/docker-proxy/docker-restart.vue';
import { loadDockerStatus } from '@/api/modules/container';
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue';
import { useElementSize } from '@vueuse/core';
import { ElMessageBox } from 'element-plus';
import { Expand, Loading, Lock, Refresh } from '@element-plus/icons-vue';

interface RuleRow extends Firewall.InventoryItem {
    rowKey: string;
}

interface UsageEntry {
    key: string;
    ports: number[];
    owner: string;
    pid?: number;
    docker?: boolean;
    dockerManagementTarget?: Firewall.DockerGuardEndpoint['managementTarget'];
    dockerEndpoint?: Firewall.DockerGuardEndpoint;
}

interface DisplayNotice {
    key: string;
    text: string;
}

type RuleFilter = 'family:ipv4' | 'family:ipv6' | 'action:accept' | 'action:deny';

const ruleFilterStorageKey = 'firewall-rule-filters';
const ruleFilterOptions: RuleFilter[] = ['family:ipv4', 'family:ipv6', 'action:accept', 'action:deny'];

const loadCachedFilterValues = <T extends string>(key: string, allowed: readonly T[], defaults: readonly T[]): T[] => {
    try {
        const cached = JSON.parse(localStorage.getItem(key) || 'null');
        if (!Array.isArray(cached)) return [...defaults];
        return cached.filter((item): item is T => typeof item === 'string' && allowed.includes(item as T));
    } catch {
        return [...defaults];
    }
};

const cacheFilterValues = (key: string, values: readonly string[]) => {
    localStorage.setItem(key, JSON.stringify(values));
};

const fireStatusRef = ref<InstanceType<typeof FireStatus>>();
const noticeRef = ref<HTMLElement>();
const { height: noticeHeight } = useElementSize(noticeRef);
const ruleOperateRef = ref<InstanceType<typeof RuleOperate>>();
const ruleImportRef = ref<InstanceType<typeof RuleImport>>();
const ruleTaskLogRef = ref<InstanceType<typeof TaskLog>>();
const openRuleTask = (taskID: string) => ruleTaskLogRef.value?.openWithTaskID(taskID, true);
const processDetailRef = ref<InstanceType<typeof ProcessDetail>>();
const resetConfirmRef = ref<InstanceType<typeof RuleReset>>();
const dockerRestartRef = ref<InstanceType<typeof DockerRestart>>();
const withDockerRestart = ref(false);
const loading = ref(false);
const resetting = ref(false);
const backupBeforeReset = ref(true);
const isActive = ref(false);
const isInit = ref(false);
const isBind = ref(false);
const provider = ref('');
const isDirectBackend = computed(() => provider.value === 'iptables' || provider.value === 'nftables');
const isServiceBackend = computed(() => provider.value === 'firewalld' || provider.value === 'ufw');
const canReset = computed(
    () =>
        ['iptables', 'nftables', 'firewalld', 'ufw'].includes(provider.value) &&
        (isDirectBackend.value ? isInit.value : true),
);
const isFirewallReady = computed(() => (isDirectBackend.value ? isInit.value && isBind.value : isActive.value));
const showFirewallUnavailablePrompt = computed(
    () => (isDirectBackend.value && (!isInit.value || !isBind.value)) || (isServiceBackend.value && !isActive.value),
);
const firewallVersion = ref('');
const selectedRuleFilters = ref<RuleFilter[]>(loadCachedFilterValues(ruleFilterStorageKey, ruleFilterOptions, []));
const searchName = ref('');
const inventoryItems = ref<Firewall.InventoryItem[]>([]);
const positionRanges = ref<Partial<Record<Firewall.Family, Firewall.PositionRange>>>({});
const inventoryTotal = ref(0);
const listeningProcesses = ref<Process.ListeningProcess[]>([]);
const dockerEndpoints = ref<Firewall.DockerGuardEndpoint[]>([]);
const usageLoading = ref(false);
const usageFailed = ref(false);
let searchRequestID = 0;
let usageRequestID = 0;
let usageLoadedAt = 0;
const selects = ref<RuleRow[]>([]);
const scopeNotices = ref<Firewall.ScopeNotice[]>([]);

const supportsFirewalldPriority = computed(() => {
    if (provider.value !== 'firewalld') return true;
    const match = firewallVersion.value.trim().match(/^(\d+)\.(\d+)/);
    if (!match) return false;
    const major = Number(match[1]);
    const minor = Number(match[2]);
    return major > 0 || minor >= 7;
});
const paginationConfig = reactive({
    cacheSizeKey: 'firewall-rule-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('firewall-rule-page-size')) || 20,
    total: 0,
});

const providerScopes = (): Firewall.Scope[] => {
    if (provider.value === 'iptables' || provider.value === 'nftables') {
        return (['ipv4', 'ipv6'] as Firewall.Family[]).flatMap((family) =>
            ['1PANEL_BASIC_BEFORE', '1PANEL_BASIC', '1PANEL_BASIC_AFTER'].map((chain) => ({
                provider: provider.value as 'iptables' | 'nftables',
                family,
                table: 'filter',
                chain,
                direction: 'input' as const,
            })),
        );
    }
    if (provider.value === 'firewalld') {
        return [{ provider: 'firewalld', family: 'inet', zone: 'public', direction: 'input' }];
    }
    if (provider.value === 'ufw') {
        return [
            {
                provider: 'ufw' as const,
                family: 'inet',
                chain: 'incoming',
                direction: 'input' as const,
            },
        ];
    }
    return [];
};

const inventoryFilters = () => ({
    families: selectedRuleFilters.value
        .filter((filter) => filter.startsWith('family:'))
        .map((filter) => filter.slice('family:'.length) as 'ipv4' | 'ipv6'),
    actions: selectedRuleFilters.value
        .filter((filter) => filter.startsWith('action:'))
        .map((filter) => filter.slice('action:'.length) as 'accept' | 'deny'),
});

const inventoryRequest = (page = paginationConfig.currentPage, pageSize = paginationConfig.pageSize) => ({
    scopes: providerScopes(),
    info: searchName.value,
    page,
    pageSize,
    ...inventoryFilters(),
});

const search = async () => {
    const requestID = ++searchRequestID;
    if (!isFirewallReady.value) {
        clearUsage();
        loading.value = false;
        inventoryItems.value = [];
        positionRanges.value = {};
        scopeNotices.value = [];
        paginationConfig.total = 0;
        inventoryTotal.value = 0;
        return;
    }

    const scopes = providerScopes();
    if (scopes.length === 0) {
        clearUsage();
        loading.value = false;
        inventoryItems.value = [];
        positionRanges.value = {};
        scopeNotices.value = [];
        paginationConfig.total = 0;
        inventoryTotal.value = 0;
        return;
    }

    loading.value = true;
    try {
        const response = await searchFirewallRules(inventoryRequest());
        if (requestID !== searchRequestID) return;
        const total = response.data.total || 0;
        const lastPage = Math.max(1, Math.ceil(total / paginationConfig.pageSize));
        if (paginationConfig.currentPage > lastPage) {
            paginationConfig.currentPage = lastPage;
            await search();
            return;
        }
        inventoryItems.value = response.data.items || [];
        positionRanges.value = {
            ipv4: response.data.ipv4Range,
            ipv6: response.data.ipv6Range,
            inet: response.data.ipv4Range,
        };
        scopeNotices.value = response.data.notices || [];
        paginationConfig.total = total;
        inventoryTotal.value = response.data.allTotal || 0;
        selects.value = [];
        if (
            inventoryItems.value.some(
                (row) => !hasIncompleteParsing(row) && listeningProtocolNumbers(row.rule.protocol).length > 0,
            ) &&
            Date.now() - usageLoadedAt >= 30_000
        ) {
            refreshUsage();
        }
    } finally {
        if (requestID === searchRequestID) loading.value = false;
    }
};

const wildcardAddress = (family: Firewall.Family) => {
    if (family === 'ipv6') return '::/0';
    if (family === 'inet') return '0.0.0.0/0, ::/0';
    return '0.0.0.0/0';
};

const isOpaqueRule = (row: Firewall.InventoryItem) => row.observed?.parseStatus === 'opaque';
const hasIncompleteParsing = (row: Firewall.InventoryItem) =>
    Boolean(row.observed) && row.observed?.parseStatus !== 'supported';
const rowProvider = (row: Firewall.InventoryItem) => row.rule.scope.provider || row.observed?.locator.provider;
const rowNativeKind = (row: Firewall.InventoryItem) => row.rule.nativeKind || row.observed?.rule.nativeKind;
const isFirewalldService = (row: Firewall.InventoryItem) => {
    const canonical = row.observed?.locator.canonical || row.observed?.locator.nativeId || '';
    return (
        rowProvider(row) === 'firewalld' && (rowNativeKind(row) === 'zone_service' || canonical.startsWith('service:'))
    );
};
const isUFWApplication = (row: Firewall.InventoryItem) =>
    rowProvider(row) === 'ufw' && rowNativeKind(row) === 'ufw_application';
const hasParsedUFWFields = (row: Firewall.InventoryItem) => rowProvider(row) === 'ufw' && Boolean(row.rule.protocol);

const firewalldServiceName = (row: Firewall.InventoryItem) => {
    const description = row.observed?.rule.description?.trim();
    if (description) return description;
    const canonical = row.observed?.locator.canonical || row.observed?.locator.nativeId || '';
    if (canonical.startsWith('service:')) return canonical.slice('service:'.length).trim();
    return row.observed?.raw?.trim().split(/\r?\n/, 1)[0]?.trim() || '';
};

const ufwApplicationName = (row: Firewall.InventoryItem) => {
    const description = row.observed?.rule.description?.trim();
    if (description) return description;
    const raw = row.observed?.raw || '';
    return raw.match(/^\s*\[\s*\d+\]\s+(.+?)\s+(?:ALLOW|DENY|REJECT)(?:\s+(?:IN|OUT|FWD))?\s+/)?.[1]?.trim() || '';
};

const nativeDetailTarget = (row: Firewall.InventoryItem): Firewall.NativeDetailRequest | undefined => {
    if (isFirewalldService(row)) {
        const name = firewalldServiceName(row);
        if (!name) return undefined;
        return {
            provider: 'firewalld',
            nativeKind: 'zone_service',
            name,
            permanent: row.observed?.persistence === 'permanent_only',
        };
    }
    if (isUFWApplication(row)) {
        const name = ufwApplicationName(row);
        if (!name) return undefined;
        return { provider: 'ufw', nativeKind: 'ufw_application', name, permanent: false };
    }
};
const displayProtocol = (row: Firewall.InventoryItem) => {
    if (isFirewalldService(row)) return 'SERVICE';
    if (isUFWApplication(row) && !row.rule.protocol) return 'APP';
    if (isOpaqueRule(row) && !hasParsedUFWFields(row)) return '-';
    if (rowProvider(row) === 'ufw' && row.rule.protocol === 'all' && row.rule.destinationPort) return 'TCP/UDP';
    return row.rule.protocol?.toUpperCase() || '-';
};
const displayAddress = (row: Firewall.InventoryItem) => {
    if (isOpaqueRule(row) && !hasParsedUFWFields(row)) return '-';
    const wildcard = wildcardAddress(row.rule.scope.family);
    const address = row.rule.sourceAddress;
    if (address && address !== wildcard) return formatHostAddress(address, row.rule.scope.family);
    return `${wildcard}（${i18n.global.t('firewall.anyWhere')}）`;
};
const displayPort = (row: Firewall.InventoryItem) => {
    if (isOpaqueRule(row) && !hasParsedUFWFields(row)) return '-';
    return row.rule.destinationPort || '*';
};
const extractListeningPorts = (portMap: Process.ListeningProcess['Port']) =>
    Object.keys(portMap || {})
        .map(Number)
        .filter((port) => Number.isInteger(port) && port > 0 && port <= 65535);
const isPortInRule = (rulePort: string | undefined, port: number) => {
    const value = rulePort?.trim();
    if (!value || value === '*') return true;
    return value.split(',').some((rawSegment) => {
        const segment = rawSegment.trim();
        const delimiter = segment.includes('-') ? '-' : segment.includes(':') ? ':' : '';
        if (!delimiter) return Number(segment) === port;
        const [rawStart, rawEnd] = segment.split(delimiter, 2);
        const start = Number(rawStart);
        const end = Number(rawEnd);
        return Number.isInteger(start) && Number.isInteger(end) && port >= start && port <= end;
    });
};
const listeningProtocolNumbers = (protocol: string) => {
    switch (protocol.toLowerCase()) {
        case 'tcp':
            return [1];
        case 'udp':
            return [2];
        case 'all':
            return [1, 2];
        default:
            return [];
    }
};
const clearUsage = () => {
    usageRequestID++;
    usageLoadedAt = 0;
    usageLoading.value = false;
    usageFailed.value = false;
    listeningProcesses.value = [];
    dockerEndpoints.value = [];
};
const loadUsage = async () => {
    const requestID = ++usageRequestID;
    usageLoading.value = true;
    usageFailed.value = false;
    const [processes, containers] = await Promise.allSettled([getListeningProcess(), loadDockerPublishedPorts()]);
    if (requestID !== usageRequestID) return;
    listeningProcesses.value = processes.status === 'fulfilled' ? processes.value.data || [] : [];
    dockerEndpoints.value =
        containers.status === 'fulfilled'
            ? (containers.value.data || []).flatMap((container) => container.endpoints || [])
            : [];
    usageFailed.value = processes.status === 'rejected' || containers.status === 'rejected';
    usageLoadedAt = Date.now();
    usageLoading.value = false;
};
const refreshUsage = () => {
    if (!usageLoading.value) loadUsage();
};
const ruleUsageEntries = (row: RuleRow): UsageEntry[] => {
    if (row.rule.scope.direction !== 'input' || hasIncompleteParsing(row)) return [];
    const protocols = listeningProtocolNumbers(row.rule.protocol);
    const processes = listeningProcesses.value.flatMap((process) => {
        if (!protocols.includes(process.Protocol)) return [];
        const ports = extractListeningPorts(process.Port)
            .filter((port) => isPortInRule(row.rule.destinationPort, port))
            .sort((left, right) => left - right);
        if (ports.length === 0) return [];
        return [
            {
                key: `${process.PID}:${process.Protocol}:${ports.join(',')}`,
                ports,
                owner: process.Name?.trim() || `PID ${process.PID}`,
                pid: process.PID,
            },
        ];
    });
    const docker = dockerEndpoints.value
        .filter((endpoint) => row.rule.scope.family === 'inet' || endpoint.family === row.rule.scope.family)
        .filter((endpoint) => protocols.includes(endpoint.protocol === 'tcp' ? 1 : 2))
        .filter((endpoint) => isPortInRule(row.rule.destinationPort, endpoint.hostPort))
        .map((endpoint) => ({
            key: `docker:${endpoint.family}:${endpoint.hostIP}:${endpoint.hostPort}:${endpoint.protocol}`,
            ports: [endpoint.hostPort],
            owner: `Docker: ${endpoint.containerName || endpoint.containerID?.slice(0, 12) || '-'}`,
            docker: true,
            dockerManagementTarget: dockerGuardManagementTarget(endpoint),
            dockerEndpoint: endpoint,
        }));
    return [...processes, ...docker];
};
const usageEntryPortText = (entry: UsageEntry) => entry.ports.join(', ') || '-';
const dockerUsageMessage = (entry: UsageEntry) => {
    if (entry.dockerManagementTarget === 'host_firewall') {
        return i18n.global.t('firewall.dockerInputUseHostFirewall');
    }
    if (entry.dockerManagementTarget === 'container_guard') {
        return i18n.global.t('firewall.dockerInputNotProtected');
    }
    return entry.dockerEndpoint
        ? dockerGuardEndpointManagementMessage(entry.dockerEndpoint)
        : i18n.global.t('firewall.dockerTrafficPathUnknown');
};
const usageEntryLabel = (entry: UsageEntry) =>
    entry.docker
        ? `${entry.owner} (${usageEntryPortText(entry)}) — ${dockerUsageMessage(entry)}`
        : `${entry.owner} (${usageEntryPortText(entry)})`;
const openUsageDetail = (entry: UsageEntry) => {
    if (entry.docker) {
        window.location.href = '/hosts/firewall/docker';
        return;
    }
    if (entry.pid !== undefined) processDetailRef.value?.acceptParams(entry.pid);
};
const scopeIdentity = (rule: Firewall.Rule) => JSON.stringify(rule.scope);
const toRuleRows = (items: Firewall.InventoryItem[]): RuleRow[] =>
    items.map((item, index) => {
        const nativeGroup = item.rule.orderBucket || item.rule.nativeKind || 'default';
        return {
            ...item,
            rowKey:
                item.observed?.instanceKey ||
                item.observed?.marker ||
                `${scopeIdentity(item.rule)}:${nativeGroup}:${item.observed?.locator.position ?? index}`,
        };
    });

const allRows = computed<RuleRow[]>(() => toRuleRows(inventoryItems.value));
const usageEntriesByRow = computed<Record<string, UsageEntry[]>>(() =>
    Object.fromEntries(allRows.value.map((row) => [row.rowKey, ruleUsageEntries(row)])),
);

const loadAllInventoryItems = async () => {
    if (inventoryTotal.value === 0) return [];
    const response = await searchFirewallRules({
        ...inventoryRequest(1, paginationConfig.pageSize),
        all: true,
        info: '',
        families: [],
        actions: [],
    });
    return response.data.items || [];
};

const searchWithReset = () => {
    paginationConfig.currentPage = 1;
    return search();
};

const changeRuleFilter = () => {
    cacheFilterValues(ruleFilterStorageKey, selectedRuleFilters.value);
    selects.value = [];
    return searchWithReset();
};

const notices = computed<DisplayNotice[]>(() => {
    const unique = new Map<string, DisplayNotice>();
    if (inventoryTotal.value > 1000 && isServiceBackend.value) {
        unique.set('largeRuleSet', {
            key: 'largeRuleSet',
            text: i18n.global.t('firewall.largeRuleSet'),
        });
    }
    scopeNotices.value.forEach((notice) => {
        if (notice.code === 'managed_scope_missing') return;
        const text = scopeNoticeText(notice);
        if (!text) return;
        const key = `${notice.code}:${(notice.values || []).join(',')}`;
        if (!unique.has(key)) {
            unique.set(key, { key, text });
        }
    });
    return [...unique.values()];
});

const scopeNoticeText = (notice: Firewall.ScopeNotice) => {
    const value = (notice.values || []).join(', ') || '-';
    switch (notice.code) {
        case 'family_unavailable':
            return value;
        case 'managed_scope_missing':
            return i18n.global.t('firewall.scopeMissing', [value]);
        default:
            return '';
    }
};

const isDenyAction = (action: string) => action === 'drop' || action === 'reject';
const isKnownAction = (action: string) => action === 'accept' || isDenyAction(action);

const actionLabel = (action: string) => {
    if (action === 'accept') {
        return i18n.global.t('firewall.accept');
    }
    if (action === 'reject') {
        return i18n.global.t('firewall.reject');
    }
    if (action === 'drop') {
        return i18n.global.t('firewall.drop');
    }
    return i18n.global.t('commons.status.unknown');
};

const openCreate = async () => {
    const unavailableScope = scopeNotices.value.find((notice) => notice.code === 'managed_scope_missing');
    if (unavailableScope) {
        try {
            await ElMessageBox.confirm(scopeNoticeText(unavailableScope), i18n.global.t('commons.msg.infoTitle'), {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
            });
        } catch {
            return;
        }
    }
    const ranges = { ...positionRanges.value };
    if (isDirectBackend.value) {
        for (const family of ['ipv4', 'ipv6'] as const) {
            ranges[family] = { min: 1, max: (ranges[family]?.max || 0) + 1 };
        }
    }
    ruleOperateRef.value?.acceptParams(
        provider.value as Firewall.Provider,
        undefined,
        ranges,
        supportsFirewalldPriority.value,
    );
};

const openImport = () => {
    ruleImportRef.value?.acceptParams(provider.value as Firewall.Provider);
};

const exportRules = async (rows: RuleRow[]) => {
    const exported = buildHostRuleExport(rows);
    if (exported.length === 0) {
        MsgInfo(i18n.global.t('commons.msg.noneData'));
        return;
    }
    try {
        await ElMessageBox.confirm(
            i18n.global.t('firewall.exportHelper', [exported.length]),
            i18n.global.t('commons.button.export'),
            {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
            },
        );
    } catch {
        return;
    }
    downloadWithContent(JSON.stringify(exported, null, 2), `1panel-firewall-rules-${getCurrentDateFormatted()}.json`);
};

const exportRulesBySelection = async () => {
    if (selects.value.length > 0) {
        const selected = new Set(selects.value.map((row) => row.rowKey));
        return exportRules(allRows.value.filter((row) => selected.has(row.rowKey)));
    }
    return exportRules(toRuleRows(await loadAllInventoryItems()));
};

const isWildcardDestinationPort = (rule: Firewall.Rule) => {
    const port = rule.destinationPort?.trim();
    return !port || port === '*';
};

const ruleUsageOwners = (row: RuleRow) =>
    [...new Set(ruleUsageEntries(row).map((entry) => entry.owner))].sort((left, right) => left.localeCompare(right));

const usageOwnersSummary = (owners: string[]) => {
    const visibleOwners = owners.slice(0, 5);
    const remaining = owners.length - visibleOwners.length;
    return remaining > 0 ? `${visibleOwners.join(', ')} (+${remaining})` : visibleOwners.join(', ');
};

const deleteRulesConfirmMessage = (selected: RuleRow[]) => {
    const count = new Set(selected.map((row) => row.observed?.instanceKey)).size;
    const accepted = selected.filter((row) => row.rule.action === 'accept' && Boolean(row.observed));
    const risky = accepted.filter((row) => isWildcardDestinationPort(row.rule) || ruleUsageEntries(row).length > 0);
    if (selected.length > 1 && risky.length > 0) {
        return i18n.global.t('firewall.deleteRiskRulesConfirm', [
            count,
            new Set(risky.map((row) => row.observed?.instanceKey)).size,
        ]);
    }
    if (selected.length === 1 && accepted.length === 1) {
        const [row] = accepted;
        if (isWildcardDestinationPort(row.rule)) {
            return i18n.global.t('firewall.deleteWildcardRuleConfirm', [
                formatHostAddress(row.rule.sourceAddress, row.rule.scope.family) || i18n.global.t('firewall.anyWhere'),
                `${row.rule.protocol.toUpperCase()}/*`,
            ]);
        }
        const owners = ruleUsageOwners(row);
        if (owners.length > 0) {
            return i18n.global.t('firewall.deleteUsedRuleConfirm', [usageOwnersSummary(owners)]);
        }
    }
    return i18n.global.t('firewall.deleteRuleConfirm', [count]);
};

const removeRules = async (selected: RuleRow[]) => {
    if (selected.length === 0) return;
    try {
        await ElMessageBox.confirm(deleteRulesConfirmMessage(selected), i18n.global.t('commons.button.delete'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
        });
    } catch {
        return;
    }
    loading.value = true;
    const targets = selected.filter(isDeletableRule).map((row) => ({
        scope: row.rule.scope,
        instanceKey: row.observed.instanceKey!,
        observed: row.observed,
    }));
    const count = targets.length;
    try {
        if (count === 0) return;
        const { taskID, queued, succeeded, failed } = (await deleteFirewallRules({ targets })).data;
        if (queued && taskID) {
            selects.value = [];
            openRuleTask(taskID);
            return;
        }
        if (succeeded > 0) {
            MsgSuccess(`${i18n.global.t('commons.msg.operationSuccess')} (${succeeded}/${count})`);
        }
        if (failed > 0) {
            MsgError(`${i18n.global.t('commons.msg.operationFailed')} (${failed}/${count})`);
        }
        await search();
    } finally {
        loading.value = false;
    }
};

const removeSelectedRules = () => removeRules(selects.value.filter((row) => isDeletableRule(row)));

const resetRules = () => {
    withDockerRestart.value = false;
    const message = i18n.global.t(
        isDirectBackend.value ? 'firewall.resetDirectRulesHelper' : 'firewall.resetWhitelistRulesHelper',
        [provider.value],
    );
    resetConfirmRef.value?.acceptParams({
        message,
        provider: provider.value,
    });
};

const prepareResetRules = async (backup: boolean) => {
    backupBeforeReset.value = backup;
    if (provider.value === 'firewalld') {
        const status = await loadDockerStatus();
        if (status.data.isActive) {
            dockerRestartRef.value?.acceptParams({ title: i18n.global.t('firewall.dockerRestart') });
            return;
        }
    }
    await submitResetRules();
};

const submitResetRules = async () => {
    resetting.value = true;
    loading.value = true;
    try {
        const response = await resetFirewallRules({
            backup: backupBeforeReset.value,
            provider: provider.value as Firewall.Provider,
            withDockerRestart: provider.value === 'firewalld' && withDockerRestart.value,
        });
        if (response.data.backupPath) {
            await ElMessageBox.alert(response.data.backupPath, i18n.global.t('commons.button.export'), {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
            });
        }
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    } finally {
        resetting.value = false;
        fireStatusRef.value?.acceptParams();
    }
};

const viewRawRule = async (row: RuleRow) => {
    let raw = row.observed?.raw?.trim();
    const target = nativeDetailTarget(row);
    if (target) {
        const response = await loadFirewallNativeDetail(target);
        raw = response.data.trim();
    }
    if (!raw) return;
    try {
        await ElMessageBox.alert(raw, i18n.global.t('commons.button.view'), {
            confirmButtonText: i18n.global.t('commons.button.close'),
            customClass: 'firewall-raw-rule-message',
        });
    } catch {
        return;
    }
};

const removeRule = async (row: RuleRow) => {
    if (row.isWhitelist) {
        await ElMessageBox.alert(
            i18n.global.t('firewall.whitelistRuleProtected'),
            i18n.global.t('commons.msg.infoTitle'),
            { type: 'info', confirmButtonText: i18n.global.t('commons.button.close') },
        ).catch(() => {});
        return;
    }
    return removeRules([row]);
};

const canEditDescription = (row: Firewall.InventoryItem) =>
    (row.isWhitelist || !row.observed.protected) && Boolean(row.observed.instanceKey);
const isDeletableRule = (row: Firewall.InventoryItem) => !row.observed.protected;

const displayRulePriority = (row: Firewall.InventoryItem) => {
    if (row.rule.scope.provider === 'firewalld') {
        if (!supportsFirewalldPriority.value || row.rule.nativeKind !== 'rich_rule') return '-';
        return row.rule.priority ?? '-';
    }
    return row.observed?.locator.position ?? '-';
};

const openEdit = async (row: RuleRow) => {
    if (!canEditDescription(row)) return;
    let ranges = positionRanges.value;
    if (isDirectBackend.value && row.rule.scope.chain !== '1PANEL_BASIC') {
        const { data } = await searchFirewallRules({ scopes: [row.rule.scope], info: '', page: 1, pageSize: 1 });
        ranges = { ipv4: data.ipv4Range, ipv6: data.ipv6Range };
    }
    ruleOperateRef.value?.acceptParams(
        provider.value as Firewall.Provider,
        row,
        ranges,
        supportsFirewalldPriority.value,
    );
};

const operationButtons: FuTableOperationButton<RuleRow>[] = [
    {
        label: i18n.global.t('commons.button.view'),
        permission: true,
        nodeAdmin: true,
        show: (row: RuleRow) => hasIncompleteParsing(row) && Boolean(row.observed?.raw),
        click: viewRawRule,
    },
    {
        label: i18n.global.t('commons.button.edit'),
        permission: true,
        nodeAdmin: true,
        disabled: (row: RuleRow) => !canEditDescription(row),
        click: openEdit,
    },
    {
        label: i18n.global.t('commons.button.delete'),
        permission: true,
        nodeAdmin: true,
        show: (row: RuleRow) => row.isWhitelist || isDeletableRule(row),
        click: removeRule,
    },
];

onMounted(() => {
    loading.value = true;
    fireStatusRef.value?.acceptParams();
});

onBeforeUnmount(() => {
    searchRequestID++;
    usageRequestID++;
});
</script>

<style lang="scss" scoped>
.firewall-filter-bar {
    display: inline-flex;
    flex: none;
    flex-wrap: nowrap;
    gap: 8px;
}

.firewall-rule-filter {
    width: 240px;
}

.firewall-action {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    color: var(--el-text-color-regular);
    font-size: 12px;
    line-height: 18px;

    &.is-accept .firewall-action-icon {
        color: var(--el-color-primary);
    }

    &.is-drop .firewall-action-icon {
        color: var(--el-color-info);
    }

    &.is-unknown {
        color: var(--el-text-color-secondary);
    }
}

.firewall-action-icon {
    flex: none;
    font-size: 14px;
    line-height: 1;
}

.firewall-action-lock {
    flex: none;
    color: var(--el-text-color-secondary);
    font-size: 15px;
    line-height: 1;
}

.firewall-used-cell {
    display: flex;
    align-items: center;
    flex-wrap: nowrap;
    gap: 8px;
    width: 100%;
    min-width: 0;
    white-space: nowrap;
}

.firewall-used-more {
    flex: none;
}

.firewall-used-entry {
    flex: 0 1 auto;
    min-width: 0;
    max-width: 100%;
    overflow: hidden;

    :deep(.el-tag__content) {
        display: flex;
        align-items: center;
        gap: 4px;
        min-width: 0;
        max-width: 100%;
    }
}

.firewall-used-entry-owner {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.firewall-used-entry-port,
.firewall-used-entry-icon {
    flex: none;
}

.firewall-used-entry-icon {
    margin-left: 2px;
}

.firewall-used-popover-list {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
    max-height: 280px;
    overflow-x: hidden;
    overflow-y: auto;
}

.firewall-used-popover-entry {
    display: inline-flex;
    flex: none;
    min-width: 0;
    max-width: 100%;
    overflow: hidden;

    :deep(.el-tag__content) {
        display: flex;
        align-items: center;
        gap: 4px;
        min-width: 0;
        max-width: 100%;
    }
}

:global(.firewall-raw-rule-message .el-message-box__message) {
    font-family: monospace;
    white-space: pre-wrap;
    word-break: break-all;
}

:global(.firewall-rule-filter-popper .el-select-group__wrap) {
    padding: 6px 8px 8px;
}

:global(.firewall-rule-filter-popper .el-select-group__wrap:not(:last-of-type)) {
    padding-bottom: 10px;
}

:global(.firewall-rule-filter-popper .el-select-group__wrap:not(:last-of-type)::after) {
    display: none;
}

:global(.firewall-rule-filter-popper .el-select-group__title) {
    height: 30px;
    margin-bottom: 4px;
    padding: 0 10px;
    border-left: 3px solid var(--el-color-primary);
    border-radius: 4px;
    background: var(--el-fill-color-light);
    color: var(--el-text-color-primary);
    font-size: 13px;
    font-weight: 600;
    line-height: 30px;
}

:global(.firewall-rule-filter-popper .firewall-state-filter-option) {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 20px;
    width: 100%;
}

:global(.firewall-rule-filter-popper .firewall-state-filter-description) {
    overflow: hidden;
    color: var(--el-text-color-secondary);
    font-size: 12px;
    text-overflow: ellipsis;
    white-space: nowrap;
}
</style>
