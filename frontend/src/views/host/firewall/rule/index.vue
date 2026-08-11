<template>
    <div>
        <FireRouter />

        <div v-loading="loading">
            <FireStatus
                ref="fireStatusRef"
                v-model:loading="loading"
                v-model:mask-show="maskShow"
                v-model:is-active="isActive"
                v-model:is-bind="isBind"
                v-model:name="provider"
                current-tab="base"
                @search="search"
            />

            <div v-if="provider !== '-'">
                <el-card v-if="!isActive && maskShow" class="mask-prompt">
                    <span>{{ $t('firewall.firewallNotStart') }}</span>
                </el-card>
                <el-card v-if="provider === 'iptables' && !isBind && maskShow" class="mask-prompt">
                    <span>{{ $t('firewall.basicStatus', ['1PANEL_BASIC']) }}</span>
                </el-card>

                <LayoutContent
                    :title="$t('menu.firewall')"
                    :class="{ mask: !isActive || (provider === 'iptables' && !isBind) }"
                >
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
                        <el-button v-permission v-node-admin type="primary" @click="openCreate">
                            {{ $t('commons.button.create') }}
                        </el-button>
                        <el-button
                            v-permission
                            v-node-admin
                            type="primary"
                            plain
                            :disabled="selects.length === 0"
                            @click="removeSelectedRules"
                        >
                            {{ $t('commons.button.delete') }}
                        </el-button>
                        <el-button-group>
                            <el-button v-permission v-node-admin @click="openImport">
                                {{ $t('commons.button.import') }}
                            </el-button>
                            <el-button v-permission :disabled="selects.length === 0" @click="exportSelectedRules">
                                {{ $t('commons.button.export') }}
                            </el-button>
                        </el-button-group>
                    </template>
                    <template #rightToolBar>
                        <el-popover v-if="provider === 'iptables'" placement="bottom" trigger="click" :width="230">
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
                                :min="1"
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
                                    <el-option :label="$t('firewall.managed')" value="state:managed" />
                                    <el-option :label="$t('firewall.adopted')" value="state:adopted" />
                                    <el-option :label="$t('firewall.external')" value="state:external" />
                                    <el-option :label="$t('firewall.protected')" value="state:protected" />
                                    <el-option :label="$t('firewall.drifted')" value="state:drifted" />
                                </el-option-group>
                            </el-select>
                        </div>
                        <TableSearch v-model:searchName="searchName" @search="resetPagination" />
                        <TableRefresh @search="search" />
                        <TableSetting title="firewall-rule-refresh" @search="search" />
                    </template>
                    <template #main>
                        <div>
                            <ComplexTable
                                v-model:selects="selects"
                                :pagination-config="paginationConfig"
                                :data="pagedItems"
                                :heightDiff="420"
                                row-key="rowKey"
                            >
                                <el-table-column type="selection" :selectable="isEditableManagedRule" width="48" fix />
                                <el-table-column :label="$t('firewall.action')" width="80">
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
                                <el-table-column :label="$t('firewall.priority')" align="center" width="100">
                                    <template #default="{ row }">
                                        {{ displayRulePriority(row) }}
                                    </template>
                                </el-table-column>
                                <el-table-column :label="$t('commons.table.status')" width="100" align="center">
                                    <template #default="{ row }">
                                        <el-tag
                                            class="firewall-state-tag"
                                            :type="stateTagType(row.state)"
                                            effect="plain"
                                            size="small"
                                        >
                                            {{ $t(`firewall.stateShort.${row.state}`) }}
                                        </el-tag>
                                    </template>
                                </el-table-column>
                                <el-table-column :label="$t('commons.table.protocol')" width="85">
                                    <template #default="{ row }">
                                        {{ displayProtocol(row) }}
                                    </template>
                                </el-table-column>
                                <el-table-column label="IP" min-width="200" show-overflow-tooltip>
                                    <template #default="{ row }">
                                        <span>
                                            {{ displayAddress(row) }}
                                        </span>
                                    </template>
                                </el-table-column>
                                <el-table-column
                                    :label="$t('commons.table.port')"
                                    min-width="140"
                                    show-overflow-tooltip
                                >
                                    <template #default="{ row }">
                                        <span>
                                            {{ displayPort(row) }}
                                        </span>
                                    </template>
                                </el-table-column>
                                <el-table-column :label="$t('firewall.used')" min-width="220">
                                    <template #default="{ row }">
                                        <span v-if="isOpaqueRule(row)">-</span>
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
                                                popper-class="firewall-used-popper"
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
                                    min-width="180"
                                    prop="rule.description"
                                    show-overflow-tooltip
                                />
                                <fu-table-operations
                                    width="220px"
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
        <ProcessDetail ref="processDetailRef" />
    </div>
</template>

<script lang="ts" setup>
import { Firewall } from '@/api/interface/firewall';
import { Process } from '@/api/interface/process';
import {
    checkFirewallRule,
    createFirewallRule,
    deleteFirewallRule,
    loadFirewallNativeDetail,
    searchFirewallRules,
} from '@/api/modules/firewall';
import { getListeningProcess } from '@/api/modules/process';
import i18n from '@/lang';
import { getCurrentDateFormatted } from '@/utils/date';
import { downloadWithContent } from '@/utils/file';
import { MsgError, MsgSuccess } from '@/utils/message';
import RuleImport from '@/views/host/firewall/rule/import/index.vue';
import RuleOperate from '@/views/host/firewall/rule/operate/index.vue';
import FireRouter from '@/views/host/firewall/index.vue';
import FireStatus from '@/views/host/firewall/status/index.vue';
import ProcessDetail from '@/views/host/process/process/detail/index.vue';
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { ElMessageBox } from 'element-plus';
import { Expand, Filter } from '@element-plus/icons-vue';

interface RuleRow extends Firewall.InventoryItem {
    rowKey: string;
}

interface UsageEntry {
    key: string;
    ports: number[];
    owner: string;
    pid: number;
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
const processDetailRef = ref<InstanceType<typeof ProcessDetail>>();
const loading = ref(false);
const maskShow = ref(true);
const isActive = ref(false);
const isBind = ref(false);
const provider = ref('');
const selectedRuleFilters = ref<RuleFilter[]>([]);
const iptablesChains = ['1PANEL_BASIC_BEFORE', '1PANEL_BASIC', '1PANEL_BASIC_AFTER'] as const;
const visibleIptablesChains = ref<string[]>(['1PANEL_BASIC']);
const searchName = ref('');
const inventoryItems = ref<Firewall.InventoryItem[]>([]);
const listeningProcesses = ref<Process.ListeningProcess[]>([]);
const selects = ref<RuleRow[]>([]);
const scopeNotices = ref<Firewall.ScopeNotice[]>([]);

const iptablesChainFilterActive = computed(() => visibleIptablesChains.value.length < iptablesChains.length);

const paginationConfig = reactive({
    cacheSizeKey: 'firewall-rule-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('firewall-rule-page-size')) || 20,
    total: 0,
});

const providerScopes = (): Firewall.Scope[] => {
    if (provider.value === 'iptables') {
        return (['ipv4', 'ipv6'] as Firewall.Family[]).flatMap((family) =>
            ['1PANEL_BASIC_BEFORE', '1PANEL_BASIC', '1PANEL_BASIC_AFTER'].map((chain) => ({
                provider: 'iptables' as const,
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
        return (['ipv4', 'ipv6'] as Firewall.Family[]).map((family) => ({
            provider: 'ufw' as const,
            family,
            chain: 'incoming',
            direction: 'input' as const,
        }));
    }
    return [];
};

const search = async () => {
    if (!isActive.value) {
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
        ]);
        inventoryItems.value = responses.flatMap((response) => response.data.items || []);
        scopeNotices.value = responses.flatMap((response) => response.data.notices || []);
        selects.value = [];
    } finally {
        loading.value = false;
    }
};

const isIptablesSystemPresetScope = (scope: Firewall.Scope) =>
    scope.provider === 'iptables' && (scope.chain === '1PANEL_BASIC_BEFORE' || scope.chain === '1PANEL_BASIC_AFTER');

const wildcardAddress = (family: Firewall.Family) => {
    if (family === 'ipv6') return '::/0';
    if (family === 'inet') return '0.0.0.0/0, ::/0';
    return '0.0.0.0/0';
};

const isOpaqueRule = (row: Firewall.InventoryItem) => row.observed?.parseStatus === 'opaque';
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
    if (isUFWApplication(row)) return 'APP';
    if (isOpaqueRule(row)) return '-';
    return row.rule.protocol?.toUpperCase() || '-';
};
const displayAddress = (row: Firewall.InventoryItem) => {
    if (isOpaqueRule(row)) return '-';
    const wildcard = wildcardAddress(row.rule.scope.family);
    const address = row.rule.sourceAddress;
    if (address && address !== wildcard) return address;
    return `${wildcard}（${i18n.global.t('firewall.anyWhere')}）`;
};
const displayPort = (row: Firewall.InventoryItem) => {
    if (isOpaqueRule(row)) return '-';
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
const ruleUsageEntries = (row: RuleRow): UsageEntry[] => {
    if (row.rule.scope.direction !== 'input' || isOpaqueRule(row)) return [];
    const protocols = listeningProtocolNumbers(row.rule.protocol);
    return listeningProcesses.value.flatMap((process) => {
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
};
const usageEntryPortText = (entry: UsageEntry) => entry.ports.join(', ') || '-';
const usageEntryLabel = (entry: UsageEntry) => `${entry.owner} (${usageEntryPortText(entry)})`;
const openUsageDetail = (entry: UsageEntry) => processDetailRef.value?.acceptParams(entry.pid);
const scopeIdentity = (rule: Firewall.Rule) => JSON.stringify(rule.scope);

const priorityPositionRanges = (
    item?: Firewall.InventoryItem,
): Partial<Record<Firewall.Family, PriorityPositionRange>> => {
    if (provider.value === 'firewalld') return {};
    const extraPosition = item ? 0 : 1;
    if (provider.value === 'ufw') {
        const maxPosition = inventoryItems.value.reduce(
            (max, row) => Math.max(max, row.observed?.locator.position || 0),
            0,
        );
        const limit = Math.max(1, maxPosition + extraPosition);
        return { ipv4: { min: 1, max: limit }, ipv6: { min: 1, max: limit } };
    }
    const chain = item?.rule.scope.chain || '1PANEL_BASIC';
    return Object.fromEntries(
        (['ipv4', 'ipv6'] as Firewall.Family[]).map((family) => {
            const scopeRows = inventoryItems.value
                .filter((row) => row.rule.scope.family === family && row.rule.scope.chain === chain)
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
            item.rule.scope.provider === 'iptables' &&
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
        case 'default_policy':
            return i18n.global.t('firewall.scopeDefaultPolicy', [value]);
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

const stateTagType = (state: Firewall.InventoryState) => {
    if (state === 'managed' || state === 'adopted') return 'primary';
    if (state === 'protected') return 'warning';
    if (state === 'drifted') return 'danger';
    return 'info';
};

const openCreate = async () => {
    if (scopeNotices.value.some((notice) => notice.code === 'managed_scope_inactive')) {
        try {
            await ElMessageBox.confirm(
                i18n.global.t('firewall.scopeInactive'),
                i18n.global.t('commons.msg.infoTitle'),
                {
                    confirmButtonText: i18n.global.t('commons.button.confirm'),
                    cancelButtonText: i18n.global.t('commons.button.cancel'),
                },
            );
        } catch {
            return;
        }
    }
    ruleOperateRef.value?.acceptParams(provider.value as Firewall.Provider, undefined, priorityPositionRanges());
};

const openImport = () => {
    ruleImportRef.value?.acceptParams(provider.value as Firewall.Provider);
};

const exportSelectedRules = async () => {
    if (selects.value.length === 0) return;
    try {
        await ElMessageBox.confirm(
            i18n.global.t('firewall.exportHelper', [selects.value.length]),
            i18n.global.t('commons.button.export'),
            {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
            },
        );
    } catch {
        return;
    }
    const exported = selects.value.map(({ rule }) => ({
        ...rule,
        uuid: undefined,
        scope: { ...rule.scope },
    }));
    downloadWithContent(JSON.stringify(exported, null, 2), `1panel-firewall-rules-${getCurrentDateFormatted()}.json`);
};

const removeSelectedRules = async () => {
    const selected = selects.value.filter((row) => isEditableManagedRule(row));
    if (selected.length === 0) return;
    const usedBy = selected.flatMap((row) =>
        row.rule.action === 'accept' && row.usage?.used ? row.usage.usedBy || [] : [],
    );
    const message = usedBy.length
        ? `${i18n.global.t('firewall.deleteRuleConfirm', [selected.length])}\n${i18n.global.t(
              'firewall.deleteUsedRuleConfirm',
              [usedBy.join(', ')],
          )}`
        : i18n.global.t('firewall.deleteRuleConfirm', [selected.length]);
    try {
        await ElMessageBox.confirm(message, i18n.global.t('commons.button.delete'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
        });
    } catch {
        return;
    }
    loading.value = true;
    let success = 0;
    for (const row of selected) {
        if (!row.desired?.uuid) continue;
        try {
            await deleteFirewallRule(row.desired.uuid);
            success++;
        } catch {
            // The shared interceptor reports the failed rule and the remaining rules continue.
        }
    }
    if (success > 0) {
        MsgSuccess(`${i18n.global.t('commons.msg.operationSuccess')} (${success}/${selected.length})`);
    }
    await search();
    loading.value = false;
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
        // Closing the read-only dialog does not require follow-up.
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
        const plan = (await checkFirewallRule({ rule: row.rule })).data;
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
        await createFirewallRule({
            checkFlag: plan.checkFlag,
            action: resolution,
            adoptInstanceKey:
                resolution === 'select_adopt' ? candidate?.instanceKey : plan.candidates?.[0]?.instanceKey,
            rule: plan.requestedRule,
            sourceKind: 'user',
        });
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        await search();
    } finally {
        loading.value = false;
    }
};

const removeRule = async (row: RuleRow) => {
    if (!row.desired?.uuid) return;
    const usedBy =
        row.rule.action === 'accept' && row.usage?.used
            ? row.usage.usedBy?.join(', ') || i18n.global.t('firewall.used')
            : '';
    const message = usedBy
        ? `${i18n.global.t('firewall.deleteRuleConfirm', [1])}\n${i18n.global.t('firewall.deleteUsedRuleConfirm', [usedBy])}`
        : i18n.global.t('firewall.deleteRuleConfirm', [1]);
    try {
        await ElMessageBox.confirm(message, i18n.global.t('commons.button.delete'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
        });
    } catch {
        return;
    }
    loading.value = true;
    try {
        await deleteFirewallRule(row.desired.uuid);
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        await search();
    } finally {
        loading.value = false;
    }
};

const isEditableManagedRule = (row: Firewall.InventoryItem) =>
    Boolean(row.desired?.uuid) &&
    (row.desired?.origin === 'created' || row.desired?.origin === 'adopted') &&
    !isIptablesSystemPresetScope(row.rule.scope) &&
    row.state !== 'drifted' &&
    row.state !== 'protected';

const displayRulePriority = (row: Firewall.InventoryItem) => {
    if (row.rule.scope.provider === 'firewalld') return row.rule.priority ?? '-';
    if (row.rule.scope.provider === 'iptables' && row.rule.scope.chain !== '1PANEL_BASIC') return '-';
    return row.observed?.locator.position ?? '-';
};

const openEdit = (row: RuleRow) => {
    if (!isEditableManagedRule(row)) return;
    ruleOperateRef.value?.acceptParams(provider.value as Firewall.Provider, row, priorityPositionRanges(row));
};

const operationButtons = [
    {
        label: i18n.global.t('commons.button.view'),
        permission: true,
        nodeAdmin: true,
        show: (row: RuleRow) => row.observed?.parseStatus === 'opaque' && Boolean(row.observed.raw),
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
        show: (row: RuleRow) => isEditableManagedRule(row),
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
    width: 480px;
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

.firewall-state-tag {
    min-width: 58px;
    justify-content: center;
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
</style>
