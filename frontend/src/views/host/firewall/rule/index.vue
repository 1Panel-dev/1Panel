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
                <template v-if="canReset" #actions>
                    <el-divider direction="vertical" />
                    <el-button
                        v-permission
                        v-node-admin
                        type="primary"
                        link
                        :disabled="loading"
                        :loading="resetting"
                        @click="resetRules"
                    >
                        {{ $t('firewall.cleanupAction') }}
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
                        <el-alert
                            v-for="notice in notices"
                            :key="notice.key"
                            class="mb-2"
                            type="warning"
                            :closable="false"
                            :title="notice.text"
                        />
                    </template>
                    <template #leftToolBar>
                        <el-button v-permission v-node-admin type="primary" :disabled="loading" @click="openCreate">
                            {{ $t('commons.button.create') }}
                        </el-button>
                        <el-button
                            v-permission
                            v-node-admin
                            :disabled="loading || !canSyncRules"
                            :loading="syncOpening"
                            @click="openRuleSync"
                        >
                            {{ $t('commons.button.sync') }}
                        </el-button>
                        <el-button
                            v-permission
                            v-node-admin
                            :disabled="loading || selects.length === 0"
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
                                :disabled="loading || allManagedRules.length === 0"
                                @click="exportRulesBySelection"
                            >
                                {{ $t('commons.button.export') }}
                            </el-button>
                        </el-button-group>
                    </template>
                    <template #rightToolBar>
                        <el-popover
                            v-if="provider === 'iptables' || provider === 'nftables'"
                            placement="bottom"
                            trigger="click"
                            :width="230"
                        >
                            <template #reference>
                                <el-button
                                    :type="iptablesChainFilterActive ? 'primary' : 'default'"
                                    :icon="Filter"
                                    :title="$t('menu.filter')"
                                    plain
                                />
                            </template>
                            <div class="firewall-chain-filter-title">{{ $t('firewall.chain') }}</div>
                            <el-checkbox-group
                                v-model="visibleIptablesChains"
                                class="firewall-chain-filter-options"
                                @change="changeIptablesChainFilter"
                            >
                                <el-checkbox v-for="chain in iptablesChains" :key="chain" :value="chain">
                                    {{ chain }}
                                </el-checkbox>
                            </el-checkbox-group>
                        </el-popover>
                        <div class="firewall-filter-bar">
                            <el-select
                                v-model="selectedRuleFilters"
                                class="firewall-rule-filter"
                                :placeholder="$t('menu.filter')"
                                multiple
                                clearable
                                collapse-tags
                                collapse-tags-tooltip
                                :max-collapse-tags="4"
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
                                <el-option-group :label="$t('commons.table.status')">
                                    <el-option :label="$t('firewall.stateShort.managed')" value="state:managed">
                                        <div class="firewall-state-filter-option">
                                            <span>{{ $t('firewall.stateShort.managed') }}</span>
                                            <span class="firewall-state-filter-description">
                                                {{ $t('firewall.managedHelper') }}
                                            </span>
                                        </div>
                                    </el-option>
                                    <el-option :label="$t('firewall.stateShort.adopted')" value="state:adopted">
                                        <div class="firewall-state-filter-option">
                                            <span>{{ $t('firewall.stateShort.adopted') }}</span>
                                            <span class="firewall-state-filter-description">
                                                {{ $t('firewall.adoptedHelper') }}
                                            </span>
                                        </div>
                                    </el-option>
                                    <el-option :label="$t('firewall.stateShort.external')" value="state:external">
                                        <div class="firewall-state-filter-option">
                                            <span>{{ $t('firewall.stateShort.external') }}</span>
                                            <span class="firewall-state-filter-description">
                                                {{ $t('firewall.externalHelper') }}
                                            </span>
                                        </div>
                                    </el-option>
                                    <el-option :label="$t('firewall.stateShort.protected')" value="state:protected">
                                        <div class="firewall-state-filter-option">
                                            <span>{{ $t('firewall.stateShort.protected') }}</span>
                                            <span class="firewall-state-filter-description">
                                                {{ $t('firewall.protectedHelper') }}
                                            </span>
                                        </div>
                                    </el-option>
                                    <el-option :label="$t('firewall.stateShort.drifted')" value="state:drifted">
                                        <div class="firewall-state-filter-option">
                                            <span>{{ $t('firewall.stateShort.drifted') }}</span>
                                            <span class="firewall-state-filter-description">
                                                {{ $t('firewall.plan_managed_rule_drifted') }}
                                            </span>
                                        </div>
                                    </el-option>
                                </el-option-group>
                            </el-select>
                        </div>
                        <TableSearch v-model:searchName="searchName" @search="resetPagination" />
                        <TableRefresh @search="search" />
                        <TableSetting title="firewall-rule-refresh" @search="search" />
                    </template>
                    <template #main>
                        <div v-loading="loading">
                            <ComplexTable
                                v-model:selects="selects"
                                :pagination-config="paginationConfig"
                                :data="pagedItems"
                                :heightDiff="370"
                                row-key="rowKey"
                            >
                                <el-table-column type="selection" :selectable="isDeletableManagedRule" width="48" fix />
                                <el-table-column :label="$t('firewall.action')" width="76">
                                    <template #default="{ row }">
                                        <span
                                            class="firewall-action"
                                            :class="row.rule.action === 'accept' ? 'is-accept' : 'is-drop'"
                                        >
                                            <i
                                                class="iconfont firewall-action-icon"
                                                :class="row.rule.action === 'accept' ? 'p-yunxu1' : 'p-a-44tubiao-226'"
                                                aria-hidden="true"
                                            />
                                            {{ actionLabel(row.rule.action) }}
                                        </span>
                                    </template>
                                </el-table-column>
                                <el-table-column :label="$t('firewall.priority')" width="90">
                                    <template #default="{ row }">
                                        {{ displayRulePriority(row) }}
                                    </template>
                                </el-table-column>
                                <el-table-column :label="$t('commons.table.status')" width="64">
                                    <template #default="{ row }">
                                        <el-tooltip :content="ruleStateTooltip(row)" placement="top" :show-after="200">
                                            <span
                                                class="firewall-rule-state"
                                                :aria-label="`${ruleStateTitle(row)}：${ruleStateDetail(row)}`"
                                                tabindex="0"
                                            >
                                                <el-icon
                                                    v-if="row.state === 'drifted'"
                                                    class="firewall-rule-state-warning"
                                                >
                                                    <WarningFilled />
                                                </el-icon>
                                                <i
                                                    v-if="row.state === 'adopted'"
                                                    class="iconfont firewall-rule-source-icon"
                                                    :class="'p-yinaguan'"
                                                    aria-hidden="true"
                                                />
                                                <i
                                                    v-if="row.state === 'managed'"
                                                    class="iconfont firewall-rule-source-icon"
                                                    :class="'p-shuju'"
                                                    aria-hidden="true"
                                                />
                                                <el-icon
                                                    v-if="row.state === 'external'"
                                                    class="firewall-rule-source-icon"
                                                >
                                                    <Link />
                                                </el-icon>
                                                <el-icon
                                                    v-if="row.state === 'protected'"
                                                    class="firewall-rule-source-icon"
                                                >
                                                    <Lock />
                                                </el-icon>
                                            </span>
                                        </el-tooltip>
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
                                    <template #default="{ row }">
                                        <span v-if="isReadOnlyNativeRule(row)">-</span>
                                        <el-tag v-else-if="ruleUsageEntries(row).length === 0" type="info" size="small">
                                            {{ $t('firewall.unUsed') }}
                                        </el-tag>
                                        <div v-else class="firewall-used-cell">
                                            <el-tooltip
                                                :content="usageEntryLabel(ruleUsageEntries(row)[0])"
                                                placement="top"
                                            >
                                                <el-tag
                                                    class="cursor-pointer firewall-used-entry"
                                                    type="info"
                                                    effect="plain"
                                                    size="small"
                                                    @click.stop="openUsageDetail(ruleUsageEntries(row)[0])"
                                                >
                                                    <span class="firewall-used-entry-owner">
                                                        {{ ruleUsageEntries(row)[0].owner }}
                                                    </span>
                                                    <span class="firewall-used-entry-port">
                                                        ({{ usageEntryPortText(ruleUsageEntries(row)[0]) }})
                                                    </span>
                                                    <el-icon class="firewall-used-entry-icon"><Expand /></el-icon>
                                                </el-tag>
                                            </el-tooltip>
                                            <el-popover
                                                v-if="ruleUsageEntries(row).length > 1"
                                                placement="right"
                                                trigger="click"
                                                :width="340"
                                            >
                                                <template #reference>
                                                    <el-tag
                                                        class="cursor-pointer firewall-used-more"
                                                        type="info"
                                                        effect="plain"
                                                        size="small"
                                                    >
                                                        +{{ ruleUsageEntries(row).length - 1 }}
                                                    </el-tag>
                                                </template>
                                                <div class="firewall-used-popover-list">
                                                    <el-tooltip
                                                        v-for="entry in ruleUsageEntries(row)"
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
        <RuleOperate ref="ruleOperateRef" @search="search" />
        <RuleImport ref="ruleImportRef" @search="search" />
        <RuleSync ref="ruleSyncRef" @search="search" />
        <ProcessDetail ref="processDetailRef" />
        <ConfirmDialog ref="resetConfirmRef" @confirm="submitResetRules" />
    </div>
</template>

<script lang="ts" setup>
import { Firewall } from '@/api/interface/firewall';
import { Process } from '@/api/interface/process';
import {
    checkFirewallRules,
    createFirewallRules,
    deleteFirewallRules,
    loadDockerPublishedPorts,
    loadFirewallNativeDetail,
    loadFirewallRuleSyncTask,
    resetFirewallRules,
    searchFirewallRules,
} from '@/api/modules/firewall';
import { getListeningProcess } from '@/api/modules/process';
import i18n from '@/lang';
import { getCurrentDateFormatted } from '@/utils/date';
import { downloadWithContent } from '@/utils/file';
import { MsgError, MsgSuccess } from '@/utils/message';
import { formatHostAddress } from '@/views/host/firewall/utils/validation';
import RuleImport from '@/views/host/firewall/rule/import/index.vue';
import RuleOperate from '@/views/host/firewall/rule/operate/index.vue';
import RuleSync from '@/views/host/firewall/sync/index.vue';
import FireRouter from '@/views/host/firewall/index.vue';
import FireStatus from '@/views/host/firewall/status/index.vue';
import ProcessDetail from '@/views/host/process/process/detail/index.vue';
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { ElMessageBox } from 'element-plus';
import { Expand, Filter, Lock, WarningFilled } from '@element-plus/icons-vue';

interface RuleRow extends Firewall.InventoryItem {
    rowKey: string;
}

interface UsageEntry {
    key: string;
    ports: number[];
    owner: string;
    pid?: number;
    docker?: boolean;
}

interface DisplayNotice {
    key: string;
    text: string;
}

interface PriorityPositionRange {
    min: number;
    max: number;
}

type RuleFilter = 'family:ipv4' | 'family:ipv6' | 'action:accept' | 'action:deny' | `state:${Firewall.InventoryState}`;

const fireStatusRef = ref<InstanceType<typeof FireStatus>>();
const ruleOperateRef = ref<InstanceType<typeof RuleOperate>>();
const ruleImportRef = ref<InstanceType<typeof RuleImport>>();
const ruleSyncRef = ref<InstanceType<typeof RuleSync>>();
const processDetailRef = ref<InstanceType<typeof ProcessDetail>>();
const resetConfirmRef = ref<InstanceType<typeof ConfirmDialog>>();
const loading = ref(false);
const resetting = ref(false);
const syncOpening = ref(false);
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
const selectedRuleFilters = ref<RuleFilter[]>([]);
const iptablesChains = ['1PANEL_BASIC_BEFORE', '1PANEL_BASIC', '1PANEL_BASIC_AFTER'] as const;
const visibleIptablesChains = ref<string[]>([...iptablesChains]);
const searchName = ref('');
const inventoryItems = ref<Firewall.InventoryItem[]>([]);
const listeningProcesses = ref<Process.ListeningProcess[]>([]);
const dockerEndpoints = ref<Firewall.DockerGuardEndpoint[]>([]);
const selects = ref<RuleRow[]>([]);
const scopeNotices = ref<Firewall.ScopeNotice[]>([]);

const iptablesChainFilterActive = computed(() => visibleIptablesChains.value.length < iptablesChains.length);
const supportsFirewalldPriority = computed(() => {
    if (provider.value !== 'firewalld') return true;
    const match = firewallVersion.value.trim().match(/^(\d+)\.(\d+)/);
    if (!match) return false;
    const major = Number(match[1]);
    const minor = Number(match[2]);
    return major > 0 || minor >= 7;
});
const canSyncRules = computed(
    () =>
        ['iptables', 'nftables', 'firewalld', 'ufw'].includes(provider.value) &&
        isActive.value &&
        ((provider.value !== 'iptables' && provider.value !== 'nftables') || isBind.value),
);

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

const search = async () => {
    if (!isFirewallReady.value) {
        loading.value = false;
        inventoryItems.value = [];
        scopeNotices.value = [];
        return;
    }

    const scopes = providerScopes();
    if (scopes.length === 0) {
        loading.value = false;
        inventoryItems.value = [];
        scopeNotices.value = [];
        return;
    }

    loading.value = true;
    try {
        const [responses] = await Promise.all([
            Promise.all(scopes.map((scope) => searchFirewallRules({ scope }))),
            loadListeningProcesses(),
            loadDockerEndpoints(),
        ]);
        inventoryItems.value = responses.flatMap((response) => response.data.items || []);
        scopeNotices.value = responses.flatMap((response) => response.data.notices || []);
        selects.value = [];
    } finally {
        loading.value = false;
    }
};

const isIptablesSystemPresetScope = (scope: Firewall.Scope) =>
    (scope.provider === 'iptables' || scope.provider === 'nftables') &&
    (scope.chain === '1PANEL_BASIC_BEFORE' || scope.chain === '1PANEL_BASIC_AFTER');

const wildcardAddress = (family: Firewall.Family) => {
    if (family === 'ipv6') return '::/0';
    if (family === 'inet') return '0.0.0.0/0, ::/0';
    return '0.0.0.0/0';
};

const isOpaqueRule = (row: Firewall.InventoryItem) => row.observed?.parseStatus === 'opaque';
const isReadOnlyNativeRule = (row: Firewall.InventoryItem) =>
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
    const description = row.rule.description?.trim() || row.observed?.rule.description?.trim();
    if (description) return description;
    const canonical = row.observed?.locator.canonical || row.observed?.locator.nativeId || '';
    if (canonical.startsWith('service:')) return canonical.slice('service:'.length).trim();
    return row.observed?.raw?.trim().split(/\r?\n/, 1)[0]?.trim() || '';
};

const ufwApplicationName = (row: Firewall.InventoryItem) => {
    const description = row.rule.description?.trim() || row.observed?.rule.description?.trim();
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
const loadListeningProcesses = async () => {
    try {
        const response = await getListeningProcess();
        listeningProcesses.value = response.data || [];
    } catch {
        listeningProcesses.value = [];
    }
};
const loadDockerEndpoints = async () => {
    try {
        dockerEndpoints.value = (await loadDockerPublishedPorts()).data.flatMap(
            (container) => container.endpoints || [],
        );
    } catch {
        dockerEndpoints.value = [];
    }
};
const ruleUsageEntries = (row: RuleRow): UsageEntry[] => {
    if (row.rule.scope.direction !== 'input' || isReadOnlyNativeRule(row)) return [];
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
        .filter((endpoint) => protocols.includes(endpoint.protocol === 'tcp' ? 1 : 2))
        .filter((endpoint) => isPortInRule(row.rule.destinationPort, endpoint.hostPort))
        .map((endpoint) => ({
            key: `docker:${endpoint.family}:${endpoint.hostIP}:${endpoint.hostPort}:${endpoint.protocol}`,
            ports: [endpoint.hostPort],
            owner: `Docker: ${endpoint.containerName || endpoint.containerID?.slice(0, 12) || '-'}`,
            docker: true,
        }));
    return [...processes, ...docker];
};
const usageEntryPortText = (entry: UsageEntry) => entry.ports.join(', ') || '-';
const usageEntryLabel = (entry: UsageEntry) =>
    entry.docker
        ? `${entry.owner} (${usageEntryPortText(entry)}) — ${i18n.global.t('firewall.dockerInputNotProtected')}`
        : `${entry.owner} (${usageEntryPortText(entry)})`;
const openUsageDetail = (entry: UsageEntry) => {
    if (entry.docker) {
        window.location.href = '/hosts/firewall/docker';
        return;
    }
    if (entry.pid !== undefined) processDetailRef.value?.acceptParams(entry.pid);
};
const scopeIdentity = (rule: Firewall.Rule) => JSON.stringify(rule.scope);
const sameIptablesPositionScope = (rule: Firewall.Rule, family: Firewall.Family, chain: string) =>
    (rule.scope.provider === 'iptables' || rule.scope.provider === 'nftables') &&
    rule.scope.family === family &&
    rule.scope.table === 'filter' &&
    rule.scope.chain === chain &&
    rule.scope.direction === 'input';

const priorityPositionRanges = (
    item?: Firewall.InventoryItem,
): Partial<Record<Firewall.Family, PriorityPositionRange>> => {
    if (provider.value === 'firewalld') return {};
    const extraPosition = item ? 0 : 1;
    if (provider.value === 'ufw') {
        if (!item) {
            const maxPosition = inventoryItems.value.reduce(
                (max, row) => Math.max(max, row.observed?.locator.position || 0),
                0,
            );
            const range = { min: 1, max: Math.max(1, maxPosition + 1) };
            return { ipv4: range, ipv6: range };
        }
        const family = item.rule.scope.family;
        const positions = inventoryItems.value
            .filter((row) => row.rule.scope.family === family && row.observed?.locator.position)
            .map((row) => row.observed!.locator.position!);
        const currentPosition = item.observed?.locator.position || 1;
        return {
            [family]: {
                min: positions.length > 0 ? Math.min(...positions) : currentPosition,
                max: positions.length > 0 ? Math.max(...positions) : currentPosition,
            },
        };
    }
    const chain = item?.rule.scope.chain || '1PANEL_BASIC';
    return Object.fromEntries(
        (['ipv4', 'ipv6'] as Firewall.Family[]).map((family) => {
            const scopeRows = inventoryItems.value
                .filter((row) => sameIptablesPositionScope(row.rule, family, chain))
                .sort(
                    (left, right) => (left.observed?.locator.position || 0) - (right.observed?.locator.position || 0),
                );
            const maxPosition = scopeRows.reduce((max, row) => Math.max(max, row.observed?.locator.position || 0), 0);
            if (!item || item.rule.scope.family !== family) {
                return [family, { min: 1, max: Math.max(1, maxPosition + extraPosition) }];
            }
            const currentPosition = item.observed?.locator.position || 1;
            const currentIndex = scopeRows.findIndex(
                (row) => row.observed?.locator.position === item.observed?.locator.position,
            );
            let min = currentPosition;
            let max = currentPosition;
            for (let index = currentIndex - 1; index >= 0 && isEditableManagedRule(scopeRows[index]); index--) {
                min = scopeRows[index].observed?.locator.position || min;
            }
            for (
                let index = currentIndex + 1;
                index < scopeRows.length && isEditableManagedRule(scopeRows[index]);
                index++
            ) {
                max = scopeRows[index].observed?.locator.position || max;
            }
            return [family, { min, max }];
        }),
    );
};

const allRows = computed<RuleRow[]>(() =>
    inventoryItems.value.map((item, index) => {
        const nativeGroup = item.rule.orderBucket || item.rule.nativeKind || 'default';
        return {
            ...item,
            rowKey:
                item.desired?.uuid ||
                item.observed?.instanceKey ||
                item.observed?.marker ||
                `${scopeIdentity(item.rule)}:${nativeGroup}:${item.observed?.locator.position ?? index}`,
        };
    }),
);

const allManagedRules = computed(() => allRows.value.filter((row) => isDeletableManagedRule(row)));

const matchesRuleFamily = (rule: Firewall.Rule, family: Firewall.Family) => {
    if (rule.scope.family !== 'inet') return rule.scope.family === family;

    const address = rule.sourceAddress;
    if (!address) return true;
    return family === 'ipv6' ? address.includes(':') : !address.includes(':');
};

const matchesRuleFilters = (item: Firewall.InventoryItem) => {
    const familyFilters = selectedRuleFilters.value.filter((filter) => filter.startsWith('family:'));
    if (
        familyFilters.length > 0 &&
        !familyFilters.some((filter) => matchesRuleFamily(item.rule, filter.slice('family:'.length) as Firewall.Family))
    ) {
        return false;
    }

    const actionFilters = selectedRuleFilters.value.filter((filter) => filter.startsWith('action:'));
    if (
        actionFilters.length > 0 &&
        !actionFilters.some((filter) =>
            filter === 'action:accept' ? item.rule.action === 'accept' : item.rule.action !== 'accept',
        )
    ) {
        return false;
    }

    const stateFilters = selectedRuleFilters.value.filter((filter) => filter.startsWith('state:'));
    return stateFilters.length === 0 || stateFilters.some((filter) => item.state === filter.slice('state:'.length));
};

const filteredItems = computed<RuleRow[]>(() => {
    const keyword = searchName.value.trim().toLowerCase();
    return allRows.value.filter((item) => {
        if (
            (item.rule.scope.provider === 'iptables' || item.rule.scope.provider === 'nftables') &&
            !visibleIptablesChains.value.includes(item.rule.scope.chain || '')
        ) {
            return false;
        }
        if (!matchesRuleFilters(item)) {
            return false;
        }
        if (!keyword) return true;
        return [
            displayProtocol(item),
            displayAddress(item),
            displayPort(item),
            item.rule.sourceAddress,
            item.rule.sourcePort,
            item.rule.destinationAddress,
            item.rule.description,
            item.observed?.rule.description,
            item.desired?.rule.description,
            item.observed?.raw,
            item.rule.action,
            item.state,
        ].some((value) => value?.toLowerCase().includes(keyword));
    });
});

const pagedItems = computed(() => {
    const start = (paginationConfig.currentPage - 1) * paginationConfig.pageSize;
    return filteredItems.value.slice(start, start + paginationConfig.pageSize);
});

const resetPagination = () => {
    paginationConfig.currentPage = 1;
};

const changeIptablesChainFilter = () => {
    selects.value = [];
    resetPagination();
};

const changeRuleFilter = () => {
    selects.value = [];
    resetPagination();
};

watch(
    filteredItems,
    (items) => {
        paginationConfig.total = items.length;
        const lastPage = Math.max(1, Math.ceil(items.length / paginationConfig.pageSize));
        if (paginationConfig.currentPage > lastPage) {
            paginationConfig.currentPage = lastPage;
        }
    },
    { immediate: true },
);

const notices = computed<DisplayNotice[]>(() => {
    const unique = new Map<string, DisplayNotice>();
    scopeNotices.value.forEach((notice) => {
        if (notice.code === 'managed_scope_missing') return;
        const key = `${notice.code}:${(notice.values || []).join(',')}`;
        if (!unique.has(key)) {
            unique.set(key, { key, text: scopeNoticeText(notice) });
        }
    });
    return [...unique.values()];
});

const scopeNoticeText = (notice: Firewall.ScopeNotice) => {
    const value = (notice.values || []).join(', ') || '-';
    switch (notice.code) {
        case 'default_scope_mismatch':
            return i18n.global.t('firewall.scopeDefaultMismatch', [value]);
        case 'managed_scope_inactive':
            return i18n.global.t('firewall.scopeInactive');
        case 'managed_scope_missing':
            return i18n.global.t('firewall.scopeMissing', [value]);
        case 'unmanaged_active_scopes':
            return i18n.global.t('firewall.scopeUnmanagedActive', [value]);
        case 'runtime_permanent_mismatch':
            return i18n.global.t('firewall.scopeRuntimeMismatch', [value]);
    }
};

const actionLabel = (action: Firewall.Action) => {
    if (action === 'accept') {
        return i18n.global.t('firewall.accept');
    }
    if (action === 'reject') {
        return i18n.global.t('firewall.reject');
    }
    return i18n.global.t('firewall.drop');
};

const ruleSourceLabel = (row: Firewall.InventoryItem) => {
    if (row.state === 'protected') return i18n.global.t('firewall.protected');
    if (row.state === 'external') return i18n.global.t('firewall.external');
    if (row.state === 'adopted' || row.desired?.origin === 'adopted') {
        return i18n.global.t('firewall.adopted');
    }
    return i18n.global.t('firewall.managed');
};

const ruleSourceDetail = (row: Firewall.InventoryItem) => {
    if (row.state === 'protected') return i18n.global.t('firewall.protectedHelper');
    if (row.state === 'external') return i18n.global.t('firewall.externalHelper');
    if (row.state === 'adopted' || row.desired?.origin === 'adopted') {
        return i18n.global.t('firewall.adoptedHelper');
    }
    return i18n.global.t('firewall.managedHelper');
};

const ruleStateTitle = (row: Firewall.InventoryItem) =>
    row.state === 'drifted' ? i18n.global.t('firewall.stateShort.drifted') : ruleSourceLabel(row);

const ruleStateDetail = (row: Firewall.InventoryItem) =>
    row.state === 'drifted' ? ruleIssueText(row) : ruleSourceDetail(row);

const ruleStateTooltip = (row: Firewall.InventoryItem) => `${ruleStateTitle(row)}：${ruleStateDetail(row)}`;

const ruleIssueText = (row: Firewall.InventoryItem) => {
    if (row.observed?.persistence && row.observed.persistence !== 'converged') {
        return i18n.global.t('firewall.plan_runtime_permanent_mismatch');
    }
    if (row.match === 'opaque') {
        return i18n.global.t('firewall.plan_opaque_rule_in_target_scope');
    }
    return i18n.global.t('firewall.plan_managed_rule_drifted');
};

const openCreate = async () => {
    const unavailableScope = scopeNotices.value.find((notice) =>
        ['managed_scope_inactive', 'managed_scope_missing'].includes(notice.code),
    );
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
    ruleOperateRef.value?.acceptParams(
        provider.value as Firewall.Provider,
        undefined,
        priorityPositionRanges(),
        supportsFirewalldPriority.value,
    );
};

const openImport = () => {
    ruleImportRef.value?.acceptParams(provider.value as Firewall.Provider);
};

const openRuleSync = async () => {
    if (!canSyncRules.value) return;
    syncOpening.value = true;
    try {
        const running = (await loadFirewallRuleSyncTask()).data;
        if (running.executing && running.taskID) {
            ruleSyncRef.value?.openTask(running.taskID);
            return;
        }
        await ruleSyncRef.value?.acceptParams(provider.value as Firewall.Provider);
    } finally {
        syncOpening.value = false;
    }
};

const exportRules = async (rows: RuleRow[]) => {
    if (rows.length === 0) return;
    try {
        await ElMessageBox.confirm(
            i18n.global.t('firewall.exportHelper', [rows.length]),
            i18n.global.t('commons.button.export'),
            {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
            },
        );
    } catch {
        return;
    }
    const exported = rows.map(({ rule }) => ({
        ...rule,
        uuid: undefined,
        scope: { ...rule.scope },
    }));
    downloadWithContent(JSON.stringify(exported, null, 2), `1panel-firewall-rules-${getCurrentDateFormatted()}.json`);
};

const exportRulesBySelection = () => {
    const selected = selects.value.filter((row) => isDeletableManagedRule(row));
    return exportRules(selected.length > 0 ? selected : allManagedRules.value);
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
    const accepted = selected.filter((row) => row.rule.action === 'accept' && Boolean(row.observed));
    const risky = accepted.filter((row) => isWildcardDestinationPort(row.rule) || ruleUsageEntries(row).length > 0);
    if (selected.length > 1 && risky.length > 0) {
        return i18n.global.t('firewall.deleteRiskRulesConfirm', [selected.length, risky.length]);
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
    return i18n.global.t('firewall.deleteRuleConfirm', [selected.length]);
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
    const uuids = selected.flatMap((row) => (row.desired?.uuid ? [row.desired.uuid] : []));
    try {
        let succeeded = 0;
        let failed = 0;
        for (let offset = 0; offset < uuids.length; offset += 256) {
            const batch = uuids.slice(offset, offset + 256);
            const result = (await deleteFirewallRules({ uuids: batch })).data;
            succeeded += result.succeeded;
            failed += result.failed;
        }
        if (succeeded > 0) {
            MsgSuccess(`${i18n.global.t('commons.msg.operationSuccess')} (${succeeded}/${uuids.length})`);
        }
        if (failed > 0) {
            MsgError(`${i18n.global.t('commons.msg.operationFailed')} (${failed}/${uuids.length})`);
        }
    } finally {
        try {
            await search();
        } finally {
            loading.value = false;
        }
    }
};

const removeSelectedRules = () => removeRules(selects.value.filter((row) => isDeletableManagedRule(row)));

const resetRules = () => {
    const message = i18n.global.t(
        isDirectBackend.value ? 'firewall.resetDirectRulesHelper' : 'firewall.resetWhitelistRulesHelper',
        [provider.value],
    );
    resetConfirmRef.value?.acceptParams({
        header: i18n.global.t('firewall.cleanupAction'),
        operationInfo: message,
        submitInputInfo: provider.value,
    });
};

const submitResetRules = async () => {
    resetting.value = true;
    loading.value = true;
    try {
        await resetFirewallRules({ provider: provider.value as Firewall.Provider });
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

const adoptRule = async (row: RuleRow) => {
    if (!row.observed) return;
    try {
        await ElMessageBox.confirm(
            i18n.global.t('firewall.adoptRuleConfirm'),
            i18n.global.t('firewall.resolution_adopt'),
            {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
            },
        );
    } catch {
        return;
    }
    loading.value = true;
    try {
        const plan = (await checkFirewallRules({ items: [{ rule: row.rule }] })).data.items[0];
        if (plan.decision !== 'confirmation_required' || plan.classification !== 'exact_external') {
            MsgError(i18n.global.t('firewall.plan_blocked'));
            return;
        }
        const candidate = plan.candidates?.find(
            (item) =>
                item.locator.position === row.observed?.locator.position &&
                item.locator.nativeId === row.observed?.locator.nativeId &&
                item.locator.canonical === row.observed?.locator.canonical,
        );
        const resolution: Firewall.ApplicableCheckAction = plan.candidates?.length === 1 ? 'adopt' : 'select_adopt';
        if (resolution === 'select_adopt' && !candidate?.instanceKey) {
            MsgError(i18n.global.t('firewall.plan_blocked'));
            return;
        }
        const result = (
            await createFirewallRules({
                items: [
                    {
                        checkFlag: plan.checkFlag,
                        action: resolution,
                        adoptInstanceKey:
                            resolution === 'select_adopt' ? candidate?.instanceKey : plan.candidates?.[0]?.instanceKey,
                        rule: plan.requestedRule,
                        sourceKind: 'user',
                    },
                ],
            })
        ).data;
        if (result.failed > 0 || result.skipped > 0) {
            MsgError(result.errors?.[0]?.error || i18n.global.t('commons.msg.operationFailed'));
            return;
        }
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        await search();
    } finally {
        loading.value = false;
    }
};

const removeRule = (row: RuleRow) => removeRules([row]);

const isEditableManagedRule = (row: Firewall.InventoryItem) =>
    Boolean(row.desired?.uuid) &&
    (row.desired?.origin === 'created' || row.desired?.origin === 'adopted') &&
    !row.desired?.protected &&
    !isIptablesSystemPresetScope(row.rule.scope) &&
    row.state !== 'drifted' &&
    row.state !== 'protected';

const isMissingManagedRule = (row: Firewall.InventoryItem) =>
    row.state === 'drifted' && row.match === 'missing' && !row.observed;

const isDeletableManagedRule = (row: Firewall.InventoryItem) =>
    Boolean(row.desired?.uuid) &&
    (row.desired?.origin === 'created' || row.desired?.origin === 'adopted') &&
    !row.desired?.protected &&
    !isIptablesSystemPresetScope(row.rule.scope) &&
    row.state !== 'protected' &&
    (row.state !== 'drifted' || isMissingManagedRule(row));

const displayRulePriority = (row: Firewall.InventoryItem) => {
    if (row.rule.scope.provider === 'firewalld') return row.rule.priority ?? '-';
    if (
        (row.rule.scope.provider === 'iptables' || row.rule.scope.provider === 'nftables') &&
        row.rule.scope.chain !== '1PANEL_BASIC'
    )
        return '-';
    return row.observed?.locator.position ?? '-';
};

const openEdit = (row: RuleRow) => {
    if (!isEditableManagedRule(row)) return;
    ruleOperateRef.value?.acceptParams(
        provider.value as Firewall.Provider,
        row,
        priorityPositionRanges(row),
        supportsFirewalldPriority.value,
    );
};

const operationButtons = [
    {
        label: i18n.global.t('commons.button.view'),
        permission: true,
        nodeAdmin: true,
        show: (row: RuleRow) => isReadOnlyNativeRule(row) && Boolean(row.observed?.raw),
        click: viewRawRule,
    },
    {
        label: i18n.global.t('firewall.resolution_adopt'),
        permission: true,
        nodeAdmin: true,
        show: (row: RuleRow) => row.state === 'external' && row.observed?.parseStatus === 'supported',
        click: adoptRule,
    },
    {
        label: i18n.global.t('commons.button.edit'),
        permission: true,
        nodeAdmin: true,
        show: (row: RuleRow) => isEditableManagedRule(row),
        click: openEdit,
    },
    {
        label: i18n.global.t('commons.button.delete'),
        permission: true,
        nodeAdmin: true,
        show: (row: RuleRow) => isDeletableManagedRule(row),
        click: removeRule,
    },
];

onMounted(() => {
    loading.value = true;
    fireStatusRef.value?.acceptParams();
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
    width: 400px;
}

.firewall-chain-filter-title {
    margin-bottom: 8px;
    font-weight: 500;
}

.firewall-chain-filter-options {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
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
}

.firewall-action-icon {
    flex: none;
    font-size: 14px;
    line-height: 1;
}

.firewall-rule-state {
    display: inline-flex;
    align-items: center;
    color: var(--el-text-color-regular);
}

.firewall-rule-state-warning {
    flex: none;
    color: var(--el-color-warning);
    cursor: help;
    font-size: 15px;
}

.firewall-rule-source-icon {
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
