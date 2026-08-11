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
                            v-if="provider === 'ufw'"
                            class="mb-2"
                            type="info"
                            :closable="false"
                            :title="$t('firewall.ufwReorderUnsupported')"
                        />
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
                        <div class="firewall-filter-bar">
                            <el-select v-model="selectedFamily" class="p-w-200" @change="resetPagination">
                                <template #prefix>IP</template>
                                <el-option :label="$t('commons.table.all')" value="all" />
                                <el-option label="IPv4" value="ipv4" />
                                <el-option label="IPv6" value="ipv6" />
                            </el-select>
                        </div>
                        <TableSearch v-model:searchName="searchName" @search="resetPagination" />
                        <TableRefresh @search="search" />
                        <TableSetting title="firewall-rule-refresh" @search="search" />
                    </template>
                    <template #main>
                        <div ref="tableContainerRef">
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
                                        <el-tag :type="actionTagType(row.rule.action)">
                                            {{ actionLabel(row.rule.action) }}
                                        </el-tag>
                                    </template>
                                </el-table-column>
                                <el-table-column label="排序" align="center" width="80">
                                    <template #default="{ row }">
                                        <el-tooltip
                                            v-if="provider === 'iptables'"
                                            :content="
                                                $t(canDragRule(row) ? 'firewall.reorderTip' : 'firewall.reorderBlocked')
                                            "
                                            placement="top"
                                        >
                                            <el-icon
                                                class="firewall-sort-handle"
                                                :class="{ 'is-enabled': canDragRule(row) }"
                                            >
                                                <Rank />
                                            </el-icon>
                                        </el-tooltip>
                                    </template>
                                </el-table-column>
                                <el-table-column :label="$t('commons.table.status')" width="95" align="center">
                                    <template #default="{ row }">
                                        <el-tooltip v-if="stateIcon(row.state)" placement="right">
                                            <template #content>
                                                <div>
                                                    <div>
                                                        {{ $t(`firewall.${row.state}`) }}
                                                    </div>
                                                    <div>{{ $t(`firewall.${row.state}Helper`) }}</div>
                                                </div>
                                            </template>
                                            <el-icon>
                                                <component :is="stateIcon(row.state)" />
                                            </el-icon>
                                        </el-tooltip>
                                        <el-tag v-else :type="stateTagType(row.state)" effect="plain">
                                            {{ $t(`firewall.${row.state}`) }}
                                        </el-tag>
                                    </template>
                                </el-table-column>
                                <el-table-column :label="$t('commons.table.protocol')" width="85">
                                    <template #default="{ row }">
                                        {{ row.rule.protocol?.toUpperCase() || '-' }}
                                    </template>
                                </el-table-column>
                                <el-table-column :label="accessSourceLabel" min-width="240" show-overflow-tooltip>
                                    <template #default="{ row }">
                                        <span>
                                            {{ displayAddress(row.rule, row.rule.sourceAddress) }}
                                        </span>
                                    </template>
                                </el-table-column>
                                <el-table-column :label="destinationPortLabel" min-width="140" show-overflow-tooltip>
                                    <template #default="{ row }">
                                        <span>
                                            {{ displayPort(row.rule.destinationPort) }}
                                        </span>
                                    </template>
                                </el-table-column>
                                <el-table-column
                                    v-if="provider === 'firewalld'"
                                    :label="$t('firewall.priority')"
                                    width="80"
                                >
                                    <template #default="{ row }">{{ row.rule.priority ?? '-' }}</template>
                                </el-table-column>
                                <el-table-column :label="$t('firewall.used')" min-width="220">
                                    <template #default="{ row }">
                                        <el-tag v-if="ruleUsageEntries(row).length === 0" type="info" size="small">
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
    </div>
    <RuleOperate ref="ruleOperateRef" @search="search" />
    <RuleImport ref="ruleImportRef" @search="search" />
    <ProcessDetail ref="processDetailRef" />
</template>

<script lang="ts" setup>
import { Firewall } from '@/api/interface/firewall';
import { Process } from '@/api/interface/process';
import {
    checkFirewallRule,
    createFirewallRule,
    deleteFirewallRule,
    reorderFirewallRule,
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
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue';
import { ElMessageBox } from 'element-plus';
import { CirclePlus, Expand, Link, Lock, Rank } from '@element-plus/icons-vue';
import Sortable from 'sortablejs';

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

const fireStatusRef = ref<InstanceType<typeof FireStatus>>();
const ruleOperateRef = ref<InstanceType<typeof RuleOperate>>();
const ruleImportRef = ref<InstanceType<typeof RuleImport>>();
const processDetailRef = ref<InstanceType<typeof ProcessDetail>>();
const loading = ref(false);
const maskShow = ref(true);
const isActive = ref(false);
const isBind = ref(false);
const provider = ref('');
const selectedFamily = ref<'all' | 'ipv4' | 'ipv6'>('all');
const searchName = ref('');
const inventoryItems = ref<Firewall.InventoryItem[]>([]);
const listeningProcesses = ref<Process.ListeningProcess[]>([]);
const selects = ref<RuleRow[]>([]);
const scopeNotices = ref<Firewall.ScopeNotice[]>([]);
const tableContainerRef = ref<HTMLElement>();
let tableSortable: Sortable | undefined;

const accessSourceLabel = computed(() => i18n.global.t('firewall.accessSource'));
const destinationPortLabel = computed(
    () => `${i18n.global.t('firewall.destPort')} (${i18n.global.t('commons.table.local')})`,
);

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

const displayAddress = (rule: Firewall.Rule, address?: string) => address || wildcardAddress(rule.scope.family);
const displayPort = (port?: string) => port || '*';
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
    if (row.rule.scope.direction !== 'input') return [];
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

const matchesFamilyFilter = (rule: Firewall.Rule) => {
    if (selectedFamily.value === 'all') return true;
    if (rule.scope.family !== 'inet') return rule.scope.family === selectedFamily.value;

    const address = rule.sourceAddress;
    if (!address) return true;
    return selectedFamily.value === 'ipv6' ? address.includes(':') : !address.includes(':');
};

const filteredItems = computed<RuleRow[]>(() => {
    const keyword = searchName.value.trim().toLowerCase();
    return allRows.value.filter((item) => {
        if (!matchesFamilyFilter(item.rule)) {
            return false;
        }
        if (!keyword) return true;
        return [
            item.rule.protocol,
            displayAddress(item.rule, item.rule.sourceAddress),
            displayPort(item.rule.destinationPort),
            item.rule.sourceAddress,
            item.rule.sourcePort,
            item.rule.destinationAddress,
            item.rule.description,
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

const actionTagType = (action: Firewall.Action) => (action === 'accept' ? 'success' : 'danger');

const stateTagType = (state: Firewall.InventoryState) => {
    if (state === 'managed' || state === 'adopted') {
        return 'success';
    }
    if (state === 'drifted' || state === 'protected') {
        return 'danger';
    }
    return 'info';
};

const stateIcon = (state: Firewall.InventoryState) => {
    if (state === 'managed') return CirclePlus;
    if (state === 'adopted') return Link;
    if (state === 'protected') return Lock;
    return undefined;
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
    ruleOperateRef.value?.acceptParams(provider.value as Firewall.Provider);
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
    const raw = row.observed?.raw?.trim();
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

const canDragRule = (row: Firewall.InventoryItem) =>
    provider.value === 'iptables' &&
    row.rule.scope.provider === 'iptables' &&
    Boolean(row.observed?.locator.position) &&
    isEditableManagedRule(row);

const openEdit = (row: RuleRow) => {
    if (!isEditableManagedRule(row)) return;
    ruleOperateRef.value?.acceptParams(provider.value as Firewall.Provider, row);
};

interface FirewallSortEvent {
    oldIndex?: number;
    newIndex?: number;
}

const handleTableDrag = async (event: FirewallSortEvent) => {
    const oldIndex = event.oldIndex;
    const newIndex = event.newIndex;
    if (oldIndex === undefined || newIndex === undefined || oldIndex === newIndex) return;
    const moving = pagedItems.value[oldIndex];
    const target = pagedItems.value[newIndex];
    const movingPosition = moving?.observed?.locator.position;
    const targetPosition = target?.observed?.locator.position;
    const scopeKey = moving ? scopeIdentity(moving.rule) : '';
    const firstPosition = Math.min(movingPosition || 0, targetPosition || 0);
    const lastPosition = Math.max(movingPosition || 0, targetPosition || 0);
    const crossedRules = inventoryItems.value.filter((row) => {
        const position = row.observed?.locator.position;
        return (
            position !== undefined &&
            scopeIdentity(row.rule) === scopeKey &&
            position >= firstPosition &&
            position <= lastPosition
        );
    });
    if (
        !moving?.desired?.uuid ||
        !movingPosition ||
        !targetPosition ||
        scopeIdentity(moving.rule) !== scopeIdentity(target.rule) ||
        crossedRules.some((row) => !canDragRule(row))
    ) {
        MsgError(i18n.global.t('firewall.reorderBlocked'));
        await search();
        return;
    }
    loading.value = true;
    try {
        await reorderFirewallRule(moving.desired.uuid, {
            targetPosition,
        });
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    } catch {
        // The shared HTTP interceptor displays the provider error.
    } finally {
        await search();
        loading.value = false;
    }
};

const destroyTableSortable = () => {
    tableSortable?.destroy();
    tableSortable = undefined;
};

const rebuildTableSortable = async () => {
    destroyTableSortable();
    if (provider.value !== 'iptables') return;
    await nextTick();
    const body = tableContainerRef.value?.querySelector<HTMLElement>('.el-table__body-wrapper tbody');
    if (!body) return;
    tableSortable = Sortable.create(body, {
        handle: '.firewall-sort-handle.is-enabled',
        draggable: 'tr',
        animation: 150,
        ghostClass: 'firewall-sort-ghost',
        onEnd: (event: FirewallSortEvent) => void handleTableDrag(event),
    });
};

const setRulePriority = async (row: RuleRow) => {
    if (!row.desired?.uuid) return;
    let value: string;
    try {
        const result = await ElMessageBox.prompt('', i18n.global.t('firewall.priority'), {
            inputValue: String(row.rule.priority ?? 0),
            inputPattern: /^-?\d+$/,
            inputErrorMessage: i18n.global.t('commons.msg.inputOrSelect'),
        });
        value = result.value;
    } catch {
        return;
    }
    loading.value = true;
    try {
        await reorderFirewallRule(row.desired.uuid, { priority: Number(value) });
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        await search();
    } finally {
        loading.value = false;
    }
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
        label: i18n.global.t('firewall.priority'),
        permission: true,
        nodeAdmin: true,
        show: (row: RuleRow) =>
            row.rule.scope.provider === 'firewalld' &&
            row.rule.nativeKind === 'rich_rule' &&
            row.rule.priority !== undefined &&
            isEditableManagedRule(row),
        click: setRulePriority,
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

watch([pagedItems, searchName, selectedFamily, provider], () => void rebuildTableSortable(), {
    flush: 'post',
});
onBeforeUnmount(destroyTableSortable);
</script>

<style lang="scss" scoped>
.firewall-filter-bar {
    display: inline-flex;
    flex: none;
    flex-wrap: nowrap;
    gap: 8px;
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

.firewall-sort-handle {
    opacity: 0.3;
    cursor: not-allowed;
}

.firewall-sort-handle.is-enabled {
    opacity: 1;
    cursor: grab;
}

.firewall-sort-handle.is-enabled:active {
    cursor: grabbing;
}

:deep(.firewall-sort-ghost) {
    opacity: 0.45;
}

:global(.firewall-raw-rule-message .el-message-box__message) {
    font-family: monospace;
    white-space: pre-wrap;
    word-break: break-all;
}
</style>
