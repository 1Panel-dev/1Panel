<template>
    <DialogPro v-model="visible" :title="$t('firewall.ruleSyncTitle')" size="large">
        <el-form v-loading="loading" @submit.prevent>
            <el-alert type="info" :closable="false" show-icon :title="syncHelper" class="sync-helper" />

            <div class="sync-route">
                <div class="sync-provider">
                    <span class="sync-provider-label">{{ $t('firewall.ruleSyncSource') }}</span>
                    <el-select
                        v-if="subsystem === 'system'"
                        v-model="sourceProvider"
                        class="sync-provider-select"
                        @change="changeSource"
                    >
                        <el-option v-for="item in sourceOptions" :key="item" :label="item" :value="item" />
                    </el-select>
                    <div v-else class="sync-database-source">
                        <el-icon><Coin /></el-icon>
                        <span>{{ $t('firewall.ruleSyncDatabase') }}</span>
                    </div>
                </div>
                <div class="sync-direction" aria-hidden="true">
                    <span class="sync-direction-line" />
                    <span class="sync-direction-icon">
                        <el-icon><Right /></el-icon>
                    </span>
                    <span class="sync-direction-line" />
                </div>
                <div class="sync-provider is-target">
                    <span class="sync-provider-label">{{ $t('firewall.ruleSyncTarget') }}</span>
                    <div class="sync-target-provider">
                        <el-icon><Lock /></el-icon>
                        <span>{{ targetProvider }}</span>
                    </div>
                </div>
            </div>

            <div v-if="subsystem === 'system'" class="sync-reset-source">
                <el-checkbox v-model="resetSource" :disabled="Boolean(preview?.blocked)">
                    {{ $t('firewall.ruleSyncResetSource', [sourceProvider]) }}
                </el-checkbox>
                <div class="input-help">
                    {{
                        preview?.blocked
                            ? $t('firewall.ruleSyncResetSourceBlocked')
                            : $t('firewall.ruleSyncResetSourceHelper')
                    }}
                </div>
            </div>

            <div v-if="preview" class="sync-summary">
                <button
                    type="button"
                    class="sync-summary-item"
                    :class="{ 'is-active': detailFilter === 'total' }"
                    :disabled="preview.total === 0"
                    :aria-pressed="detailFilter === 'total'"
                    @click="toggleDetail('total', preview.total)"
                >
                    <span class="sync-summary-label">{{ totalLabel }}</span>
                    <strong class="sync-summary-value">{{ preview.total }}</strong>
                </button>
                <button
                    type="button"
                    class="sync-summary-item is-ready"
                    :class="{ 'is-active': detailFilter === 'ready' }"
                    :disabled="preview.ready === 0"
                    :aria-pressed="detailFilter === 'ready'"
                    @click="toggleDetail('ready', preview.ready)"
                >
                    <span class="sync-summary-label">{{ $t('firewall.ruleSyncReady') }}</span>
                    <strong class="sync-summary-value">{{ preview.ready }}</strong>
                </button>
                <button
                    type="button"
                    class="sync-summary-item is-existing"
                    :class="{ 'is-active': detailFilter === 'existing' }"
                    :disabled="preview.existing === 0"
                    :aria-pressed="detailFilter === 'existing'"
                    @click="toggleDetail('existing', preview.existing)"
                >
                    <span class="sync-summary-label">{{ $t('firewall.ruleSyncExisting') }}</span>
                    <strong class="sync-summary-value">{{ preview.existing }}</strong>
                </button>
                <button
                    type="button"
                    class="sync-summary-item is-remove"
                    :class="{ 'is-active': detailFilter === 'remove' }"
                    :disabled="preview.removed === 0"
                    :aria-pressed="detailFilter === 'remove'"
                    @click="toggleDetail('remove', preview.removed)"
                >
                    <span class="sync-summary-label">{{ $t('firewall.ruleSyncRemove') }}</span>
                    <strong class="sync-summary-value">{{ preview.removed }}</strong>
                </button>
                <button
                    type="button"
                    class="sync-summary-item"
                    :class="{ 'is-blocked': preview.blocked > 0, 'is-active': detailFilter === 'blocked' }"
                    :disabled="preview.blocked === 0"
                    :aria-pressed="detailFilter === 'blocked'"
                    @click="toggleDetail('blocked', preview.blocked)"
                >
                    <span class="sync-summary-label">{{ $t('firewall.ruleSyncBlocked') }}</span>
                    <strong class="sync-summary-value">{{ preview.blocked }}</strong>
                </button>
            </div>

            <ComplexTable v-if="detailFilter" class="sync-rule-table" :data="detailItems" :height="360">
                <el-table-column :label="$t('commons.table.status')" width="105">
                    <template #default="{ row }">
                        <el-tag :type="statusType(row.status)" effect="plain">
                            {{ statusText(row.status) }}
                        </el-tag>
                    </template>
                </el-table-column>
                <template v-if="subsystem === 'system'">
                    <el-table-column :label="$t('commons.table.protocol')" prop="rule.protocol" width="95" />
                    <el-table-column label="IP" min-width="190" show-overflow-tooltip>
                        <template #default="{ row }">
                            {{ row.rule ? displayAddress(row.rule) : '-' }}
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.port')" min-width="120">
                        <template #default="{ row }">
                            {{ row.rule?.destinationPort || $t('firewall.allPorts') }}
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('firewall.action')" prop="rule.action" width="90" />
                </template>
                <template v-else-if="subsystem === 'forwarding'">
                    <el-table-column label="IP" prop="forwardRule.family" width="80" />
                    <el-table-column :label="$t('commons.table.protocol')" prop="forwardRule.protocol" width="90" />
                    <el-table-column :label="$t('firewall.sourcePort')" prop="forwardRule.port" min-width="105" />
                    <el-table-column :label="$t('firewall.targetIP')" prop="forwardRule.targetIP" min-width="150" />
                    <el-table-column :label="$t('firewall.targetPort')" prop="forwardRule.targetPort" min-width="105" />
                    <el-table-column
                        :label="$t('firewall.forwardInboundInterface')"
                        prop="forwardRule.interface"
                        min-width="120"
                    >
                        <template #default="{ row }">
                            {{ row.forwardRule?.interface || $t('commons.table.all') }}
                        </template>
                    </el-table-column>
                </template>
                <template v-else>
                    <el-table-column label="IP" min-width="180">
                        <template #default="{ row }">
                            {{ dockerAddress(row.dockerRule) }}
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.protocol')" prop="dockerRule.protocol" width="90" />
                    <el-table-column :label="$t('firewall.protectionMode')" min-width="145">
                        <template #default="{ row }">
                            {{ dockerMode(row.dockerRule?.mode) }}
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('firewall.sourceIP')" min-width="170" show-overflow-tooltip>
                        <template #default="{ row }">
                            {{ row.dockerRule?.sources?.join(', ') || '-' }}
                        </template>
                    </el-table-column>
                </template>
                <el-table-column :label="$t('firewall.ruleSyncReason')" min-width="220" show-overflow-tooltip>
                    <template #default="{ row }">
                        {{ reasonText(row.reason) }}
                    </template>
                </el-table-column>
            </ComplexTable>
        </el-form>

        <template #footer>
            <el-button @click="visible = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button :disabled="loading" @click="loadPreview">
                {{ $t('commons.button.refresh') }}
            </el-button>
            <el-button type="primary" :disabled="loading || syncDisabled" @click="onSync">
                {{ $t('firewall.ruleSyncAction') }}
            </el-button>
        </template>
    </DialogPro>
    <TaskLog ref="taskLogRef" width="60%" @close="handleTaskClose" />
</template>

<script lang="ts" setup>
import { Firewall } from '@/api/interface/firewall';
import { previewFirewallRuleSync, syncFirewallRules } from '@/api/modules/firewall';
import TaskLog from '@/components/log/task/index.vue';
import { useGlobalStore } from '@/composables/useGlobalStore';
import i18n from '@/lang';
import { newUUID } from '@/utils/id';
import { MsgSuccess, MsgWarning } from '@/utils/message';
import { Coin, Lock, Right } from '@element-plus/icons-vue';
import { ElMessageBox } from 'element-plus';
import { computed, ref } from 'vue';

const emit = defineEmits<{ (event: 'search'): void }>();
const { currentNode } = useGlobalStore();
const providers: Firewall.Provider[] = ['iptables', 'nftables', 'firewalld', 'ufw'];
const visible = ref(false);
const loading = ref(false);
const subsystem = ref<Firewall.BackendSubsystem>('system');
const targetProvider = ref<Firewall.Provider>('iptables');
const sourceProvider = ref<Firewall.Provider>();
const preview = ref<Firewall.RuleSyncPreview>();
const resetSource = ref(false);
const taskLogRef = ref<InstanceType<typeof TaskLog>>();
type RuleDetailFilter = Firewall.RuleSyncStatus | 'total';
const detailFilter = ref<RuleDetailFilter>();

const sourceOptions = computed(() => providers.filter((item) => item !== targetProvider.value));
const isDatabaseSync = computed(() => subsystem.value !== 'system');
const syncHelper = computed(() =>
    i18n.global.t(isDatabaseSync.value ? 'firewall.ruleSyncDatabaseHelper' : 'firewall.ruleSyncHelper'),
);
const totalLabel = computed(() =>
    i18n.global.t(isDatabaseSync.value ? 'firewall.ruleSyncDatabaseTotal' : 'firewall.ruleSyncTotal'),
);
const syncDisabled = computed(() => {
    if (!preview.value) return true;
    if (isDatabaseSync.value) return preview.value.blocked > 0;
    return Boolean(preview.value.blocked && resetSource.value) || (!resetSource.value && preview.value.ready === 0);
});
const detailItems = computed(() => {
    if (!preview.value || !detailFilter.value) return [];
    if (detailFilter.value === 'total') return preview.value.items.filter((item) => item.status !== 'remove');
    return preview.value.items.filter((item) => item.status === detailFilter.value);
});

const toggleDetail = (filter: RuleDetailFilter, count: number) => {
    if (count === 0) return;
    detailFilter.value = detailFilter.value === filter ? undefined : filter;
};

const syncRequest = (taskID?: string): Firewall.RuleSyncRequest => ({
    subsystem: subsystem.value,
    targetProvider: targetProvider.value,
    ...(subsystem.value === 'system' && sourceProvider.value ? { sourceProvider: sourceProvider.value } : {}),
    ...(subsystem.value === 'system'
        ? { resetSource: resetSource.value, ...(taskID ? { taskID } : {}) }
        : { resetSource: false }),
});

const loadPreview = async () => {
    if (subsystem.value === 'system' && (!sourceProvider.value || sourceProvider.value === targetProvider.value))
        return;
    detailFilter.value = undefined;
    loading.value = true;
    try {
        preview.value = (await previewFirewallRuleSync(syncRequest())).data;
        if (preview.value.blocked > 0) resetSource.value = false;
    } finally {
        loading.value = false;
    }
};

const onSync = async () => {
    if (syncDisabled.value || !preview.value) return;
    try {
        const message = resetSource.value
            ? i18n.global.t('firewall.ruleSyncResetSourceConfirm', [
                  preview.value.ready,
                  sourceProvider.value,
                  targetProvider.value,
              ])
            : isDatabaseSync.value
              ? i18n.global.t('firewall.ruleSyncDatabaseConfirm', [
                    preview.value.total,
                    targetProvider.value,
                    preview.value.removed,
                ])
              : i18n.global.t('firewall.ruleSyncConfirm', [
                    preview.value.ready,
                    sourceProvider.value,
                    targetProvider.value,
                ]);
        await ElMessageBox.confirm(message, i18n.global.t('firewall.ruleSyncTitle'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'warning',
        });
    } catch {
        return;
    }

    loading.value = true;
    try {
        const result = (await syncFirewallRules(syncRequest(subsystem.value === 'system' ? newUUID() : undefined)))
            .data;
        if (result.queued && result.taskID) {
            visible.value = false;
            taskLogRef.value?.openWithTaskID(result.taskID, true, currentNode.value);
            return;
        }
        if (result.failed > 0) {
            MsgWarning(i18n.global.t('firewall.ruleSyncPartial', [result.succeeded, result.skipped, result.failed]));
            await loadPreview();
        } else {
            MsgSuccess(i18n.global.t('firewall.ruleSyncSuccess', [result.succeeded, result.skipped, result.removed]));
            visible.value = false;
        }
        emit('search');
    } finally {
        loading.value = false;
    }
};

const changeSource = async () => {
    resetSource.value = false;
    await loadPreview();
};

const handleTaskClose = () => {
    emit('search');
};

const openTask = (taskID: string) => {
    visible.value = false;
    taskLogRef.value?.openWithTaskID(taskID, true, currentNode.value);
};

const statusType = (status: Firewall.RuleSyncStatus) => {
    if (status === 'ready') return 'success';
    if (status === 'remove') return 'warning';
    if (status === 'blocked') return 'danger';
    return 'info';
};

const statusText = (status: Firewall.RuleSyncStatus) => i18n.global.t(`firewall.ruleSyncStatus.${status}`);

const reasonText = (reason?: string) => {
    if (!reason) return '-';
    const key = `firewall.plan_${reason}`;
    return i18n.global.te(key) ? i18n.global.t(key) : reason;
};

const displayAddress = (rule: Firewall.Rule) => {
    const values = [rule.sourceAddress, rule.destinationAddress].filter(Boolean);
    return values.length > 0 ? values.join(' → ') : i18n.global.t('firewall.anyWhere');
};

const dockerAddress = (rule?: Firewall.DockerGuardEndpoint) => {
    if (!rule) return '-';
    const address = rule.hostIP.includes(':') ? `[${rule.hostIP}]` : rule.hostIP;
    return `${address}:${rule.hostPort}`;
};

const dockerMode = (mode?: Firewall.DockerGuardPolicy['mode']) => {
    if (!mode) return '-';
    if (mode === 'deny_all') return i18n.global.t('firewall.denyAll');
    return i18n.global.t(mode === 'allow_sources' ? 'firewall.allowSources' : 'firewall.denySources');
};

const acceptParams = async (
    target: Firewall.Provider,
    syncSubsystem: Firewall.BackendSubsystem = 'system',
    source?: Firewall.Provider,
) => {
    subsystem.value = syncSubsystem;
    targetProvider.value = target;
    sourceProvider.value =
        syncSubsystem === 'system'
            ? source && sourceOptions.value.includes(source)
                ? source
                : sourceOptions.value[0]
            : undefined;
    preview.value = undefined;
    detailFilter.value = undefined;
    resetSource.value = false;
    visible.value = true;
    await loadPreview();
};

defineExpose({ acceptParams, openTask });
</script>

<style lang="scss" scoped>
.sync-helper {
    margin-bottom: 16px;

    :deep(.el-alert__title) {
        line-height: 20px;
    }
}

.sync-reset-source {
    margin-bottom: 16px;
    padding: 12px 14px;
    border: 1px solid var(--el-border-color-light);
    border-radius: 6px;
    background: var(--el-fill-color-lighter);
}

.sync-route {
    display: grid;
    grid-template-columns: minmax(0, 1fr) 96px minmax(0, 1fr);
    align-items: stretch;
    margin-bottom: 16px;
}

.sync-provider {
    min-width: 0;
    padding: 14px 16px;
    border: 1px solid var(--el-border-color-light);
    border-radius: 8px;
    background: var(--el-fill-color-light);

    &.is-target {
        border-color: var(--el-color-primary-light-7);
        background: var(--el-color-primary-light-9);
    }
}

.sync-provider-label {
    display: block;
    margin-bottom: 8px;
    color: var(--el-text-color-secondary);
    font-size: 12px;
    line-height: 18px;
}

.sync-provider-select {
    width: 100%;
}

.sync-target-provider {
    display: flex;
    height: 32px;
    align-items: center;
    gap: 8px;
    color: var(--el-color-primary);
    font-size: 14px;
    font-weight: 600;

    .el-icon {
        font-size: 15px;
    }
}

.sync-database-source {
    display: flex;
    height: 32px;
    align-items: center;
    gap: 8px;
    color: var(--el-text-color-primary);
    font-size: 14px;
    font-weight: 600;

    .el-icon {
        color: var(--el-color-success);
        font-size: 16px;
    }
}

.sync-direction {
    display: flex;
    align-items: center;
}

.sync-direction-line {
    height: 1px;
    flex: 1;
    background: var(--el-border-color);
}

.sync-direction-icon {
    display: flex;
    width: 28px;
    height: 28px;
    flex: none;
    align-items: center;
    justify-content: center;
    border: 1px solid var(--el-border-color);
    border-radius: 50%;
    background: var(--el-bg-color);
    color: var(--el-text-color-secondary);
}

.sync-summary {
    display: grid;
    grid-template-columns: repeat(5, minmax(0, 1fr));
    gap: 10px;
    margin-bottom: 16px;
}

.sync-summary-item {
    appearance: none;
    display: flex;
    min-width: 0;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 12px 14px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 6px;
    background: var(--el-fill-color-blank);
    color: inherit;
    font-family: inherit;
    text-align: left;

    &:not(:disabled) {
        cursor: pointer;

        &:hover {
            border-color: var(--el-color-primary-light-5);
            background: var(--el-color-primary-light-9);
        }
    }

    &.is-active {
        box-shadow: 0 0 0 1px var(--el-color-primary) inset;
    }

    &:disabled {
        cursor: default;
    }

    &.is-ready .sync-summary-value {
        color: var(--el-color-success);
    }

    &.is-remove .sync-summary-value {
        color: var(--el-color-warning);
    }

    &.is-existing .sync-summary-value {
        color: var(--el-text-color-secondary);
    }

    &.is-blocked {
        border-color: var(--el-color-danger-light-7);
        background: var(--el-color-danger-light-9);

        .sync-summary-value {
            color: var(--el-color-danger);
        }
    }
}

.sync-summary-label {
    overflow: hidden;
    color: var(--el-text-color-secondary);
    font-size: 13px;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.sync-summary-value {
    flex: none;
    color: var(--el-text-color-primary);
    font-size: 20px;
    font-weight: 600;
    line-height: 24px;
}

.sync-rule-table {
    border-top: 1px solid var(--el-border-color-lighter);
}

@media (max-width: 768px) {
    .sync-route {
        grid-template-columns: 1fr;
    }

    .sync-direction {
        height: 40px;
        flex-direction: column;
    }

    .sync-direction-line {
        width: 1px;
        height: auto;
    }

    .sync-direction-icon {
        transform: rotate(90deg);
    }

    .sync-summary {
        grid-template-columns: repeat(2, minmax(0, 1fr));
    }
}
</style>
