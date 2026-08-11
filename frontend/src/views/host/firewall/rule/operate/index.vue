<template>
    <DrawerPro
        v-model="drawerVisible"
        :header="$t(mode === 'edit' ? 'firewall.edit' : 'firewall.create')"
        size="large"
        :auto-close="!loading"
        @close="handleClose"
    >
        <el-form
            v-if="!showingPreview"
            ref="formRef"
            v-loading="loading"
            label-position="top"
            :model="form"
            :rules="rules"
            :disabled="Boolean(plan)"
        >
            <el-form-item :label="$t('firewall.action')" prop="action">
                <el-radio-group v-model="form.action">
                    <el-radio-button value="accept">
                        {{ $t('firewall.accept') }}
                    </el-radio-button>
                    <el-radio-button value="drop">
                        {{ $t('firewall.drop') }}
                    </el-radio-button>
                </el-radio-group>
            </el-form-item>
            <el-form-item :label="$t('commons.table.protocol')" prop="protocol">
                <el-select v-model="form.protocol" class="w-full" @change="changeProtocol">
                    <el-option label="ALL" value="all" />
                    <el-option label="TCP" value="tcp" />
                    <el-option label="UDP" value="udp" />
                    <el-option v-if="provider !== 'ufw'" label="ICMP" value="icmp" />
                    <el-option v-if="provider !== 'ufw'" label="ICMPV6" value="icmpv6" />
                </el-select>
            </el-form-item>
            <el-form-item :label="accessSourceLabel" prop="sourceAddresses">
                <div v-for="(item, index) of form.sourceAddresses" :key="index" class="w-full">
                    <el-input
                        v-model.trim="item.address"
                        class="mt-2"
                        clearable
                        :placeholder="sourceAddressPlaceholder(item.family)"
                    >
                        <template #prepend>
                            <el-select v-model="item.family" class="ip-family-select" :disabled="mode === 'edit'">
                                <el-option label="IPv4" value="ipv4" />
                                <el-option label="IPv6" value="ipv6" />
                                <el-option v-if="provider === 'firewalld'" label="INET" value="inet" />
                            </el-select>
                        </template>
                        <template #append v-if="mode === 'create'">
                            <el-button link icon="Delete" @click="removeSourceAddress(index)" />
                        </template>
                    </el-input>
                </div>
                <el-button v-if="mode === 'create'" class="mt-2" @click="addSourceAddress">
                    {{ $t('commons.button.add') }}
                </el-button>
            </el-form-item>
            <el-form-item :label="destinationPortLabel" prop="destinationPorts">
                <div v-for="(_, index) of form.destinationPorts" :key="index" class="w-full">
                    <el-input
                        v-model.trim="form.destinationPorts[index]"
                        class="mt-2"
                        clearable
                        :disabled="!portProtocol"
                        placeholder="80 或 8080-8089"
                    >
                        <template #append v-if="mode === 'create'">
                            <el-button link icon="Delete" :disabled="!portProtocol" @click="removeRuleRow(index)" />
                        </template>
                    </el-input>
                </div>
                <el-button v-if="mode === 'create'" class="mt-2" :disabled="!portProtocol" @click="addRuleRow">
                    {{ $t('commons.button.add') }}
                </el-button>
            </el-form-item>
            <el-form-item
                v-if="provider === 'firewalld' && (mode === 'create' || editingRule?.nativeKind === 'rich_rule')"
                :label="$t('firewall.priority')"
            >
                <el-input-number v-model="form.priority" :min="-32768" :max="32767" controls-position="right" />
            </el-form-item>
            <el-form-item :label="$t('commons.table.description')">
                <el-input v-model.trim="form.description" clearable />
            </el-form-item>
        </el-form>

        <el-card v-if="showingPreview" class="rule-preview" shadow="never">
            <div class="rule-preview-title">
                <span>{{ $t('commons.button.preview') }}</span>
                <el-tag type="info">{{ $t('commons.table.total', [previewRules.length]) }}</el-tag>
            </div>
            <el-table :data="previewRules" max-height="420">
                <el-table-column type="index" :label="$t('commons.table.serialNumber')" width="70" />
                <el-table-column :label="$t('commons.table.protocol')" width="85">
                    <template #default="{ row }">{{ row.protocol.toUpperCase() }}</template>
                </el-table-column>
                <el-table-column :label="accessSourceLabel" min-width="240" show-overflow-tooltip>
                    <template #default="{ row }">{{ previewAddress(row) }}</template>
                </el-table-column>
                <el-table-column :label="destinationPortLabel" min-width="140" show-overflow-tooltip>
                    <template #default="{ row }">{{ row.destinationPort || '*' }}</template>
                </el-table-column>
                <el-table-column :label="$t('firewall.action')" width="80">
                    <template #default="{ row }">
                        <el-tag :type="row.action === 'accept' ? 'success' : 'danger'">
                            {{ $t(`firewall.${row.action === 'accept' ? 'accept' : 'drop'}`) }}
                        </el-tag>
                    </template>
                </el-table-column>
                <el-table-column
                    :label="$t('commons.table.description')"
                    prop="description"
                    min-width="180"
                    show-overflow-tooltip
                />
            </el-table>
        </el-card>

        <div v-if="plan" class="plan-confirmation">
            <el-alert :title="planMessage" type="warning" :closable="false" show-icon />
            <el-radio-group v-model="resolution" class="resolution-list">
                <el-radio v-for="item in resolutionOptions" :key="item.value" :value="item.value">
                    {{ item.label }}
                </el-radio>
            </el-radio-group>
            <el-radio-group v-if="resolution === 'select_adopt'" v-model="selectedInstanceKey" class="candidate-list">
                <el-radio
                    v-for="candidate in plan.candidates || []"
                    :key="candidate.instanceKey"
                    :value="candidate.instanceKey"
                >
                    {{ candidateLabel(candidate) }}
                </el-radio>
            </el-radio-group>
        </div>

        <template #footer>
            <el-button :disabled="loading" @click="drawerVisible = false">
                {{ $t('commons.button.cancel') }}
            </el-button>
            <el-button v-if="previewRules.length > 0" :disabled="loading" @click="backToForm">
                {{ $t('commons.button.back') }}
            </el-button>
            <el-button type="primary" :loading="loading" @click="onSubmit">
                {{ submitButtonLabel }}
            </el-button>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { Firewall } from '@/api/interface/firewall';
import { checkFirewallRulesBatch, createFirewallRulesBatch, updateFirewallRule } from '@/api/modules/firewall';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';
import { computed, reactive, ref } from 'vue';
import { ElMessageBox, type FormInstance, type FormRules } from 'element-plus';

const provider = ref<Firewall.Provider>('iptables');
const mode = ref<'create' | 'edit'>('create');
const editingUUID = ref('');
const editingRule = ref<Firewall.Rule>();
const drawerVisible = ref(false);
const loading = ref(false);
const formRef = ref<FormInstance>();
const plan = ref<Firewall.RuleCheckResult>();
const resolution = ref<Firewall.ApplicableCheckAction>();
const selectedInstanceKey = ref('');
const previewRules = ref<Firewall.Rule[]>([]);

interface BatchPlanItem {
    rule: Firewall.Rule;
    plan: Firewall.RuleCheckResult;
    resolution?: Firewall.ApplicableCheckAction;
    selectedInstanceKey?: string;
}

interface SourceAddressItem {
    family: Firewall.Family;
    address: string;
}

const batchPlans = ref<BatchPlanItem[]>([]);
const confirmationIndexes = ref<number[]>([]);
const activeConfirmationIndex = ref(-1);

const form = reactive({
    protocol: 'tcp',
    sourceAddresses: [{ family: 'ipv4', address: '' }] as SourceAddressItem[],
    sourcePort: '',
    destinationAddress: '',
    destinationPorts: [''] as string[],
    action: 'accept' as Firewall.Action,
    priority: undefined as number | undefined,
    description: '',
});

const defaultFamily = (): Firewall.Family => (provider.value === 'firewalld' ? 'inet' : 'ipv4');

const rules = reactive<FormRules>({
    protocol: [Rules.requiredSelect],
    action: [Rules.requiredSelect],
});

const portProtocol = computed(() => form.protocol === 'tcp' || form.protocol === 'udp');
const sourceAddressPlaceholder = (family: Firewall.Family) => {
    if (family === 'ipv6') return '2001:db8::1 或 2001:db8::/64';
    if (family === 'inet') return '0.0.0.0/0 或 ::/0';
    return '172.16.10.11 或 172.16.0.0/24';
};

const accessSourceLabel = computed(() => i18n.global.t('firewall.accessSource'));
const destinationPortLabel = computed(
    () => `${i18n.global.t('firewall.destPort')} (${i18n.global.t('commons.table.local')})`,
);
const showingPreview = computed(() => previewRules.value.length > 0 && !plan.value);
const submitButtonLabel = computed(() => {
    if (plan.value || showingPreview.value || mode.value === 'edit') {
        return i18n.global.t('commons.button.confirm');
    }
    return i18n.global.t('commons.button.preview');
});

const resolutionOptions = computed(() =>
    (plan.value?.allowedActions || [])
        .filter((item): item is Firewall.ApplicableCheckAction => item !== 'cancel')
        .map((item) => ({ value: item, label: i18n.global.t(`firewall.resolution_${item}`) })),
);

const supportedPlanReasons = new Set([
    'equivalent_external_rule',
    'multiple_equivalent_external_rules',
    'requested_rule_is_covered',
    'equivalent_managed_rule',
    'managed_rule_drifted',
    'opaque_rule_in_target_scope',
    'runtime_permanent_mismatch',
    'protected_rule',
    'overlapping_rule_with_different_action',
]);

const planReasonMessage = (reason: string) =>
    i18n.global.t(`firewall.plan_${supportedPlanReasons.has(reason) ? reason : 'blocked'}`);

const planMessage = computed(() => planReasonMessage(plan.value?.reason || 'blocked'));
const splitTagValues = (values: string[]) => [
    ...new Set(
        values
            .flatMap((value) => value.split(/[,，;；\s]+/))
            .map((value) => value.trim())
            .filter(Boolean),
    ),
];
const normalizeSourceAddresses = () => {
    const seen = new Set<string>();
    const normalized = form.sourceAddresses.flatMap((item) =>
        splitTagValues([item.address]).flatMap((address) => {
            const key = `${item.family}:${address}`;
            if (seen.has(key)) return [];
            seen.add(key);
            return [{ family: item.family, address }];
        }),
    );
    const fallback = form.sourceAddresses[0] || { family: 'ipv4', address: '' };
    form.sourceAddresses =
        mode.value === 'edit'
            ? [normalized[0] || { ...fallback, address: '' }]
            : normalized.length > 0
              ? normalized
              : [{ ...fallback, address: '' }];
};
const normalizeDestinationPorts = (values = form.destinationPorts) => {
    const normalized = splitTagValues(values);
    form.destinationPorts = mode.value === 'edit' ? [normalized[0] || ''] : normalized.length > 0 ? normalized : [''];
};
const addSourceAddress = () => {
    const family = form.sourceAddresses.at(-1)?.family || 'ipv4';
    form.sourceAddresses.push({ family, address: '' });
};
const removeSourceAddress = (index: number) => {
    form.sourceAddresses.splice(index, 1);
};
const addRuleRow = () => {
    form.destinationPorts.push('');
};
const removeRuleRow = (index: number) => {
    form.destinationPorts.splice(index, 1);
};

const resetForm = () => {
    form.protocol = 'tcp';
    form.sourceAddresses = [{ family: defaultFamily(), address: '' }];
    form.sourcePort = '';
    form.destinationAddress = '';
    form.destinationPorts = [''];
    form.action = 'accept';
    form.priority = undefined;
    form.description = '';
    editingUUID.value = '';
    editingRule.value = undefined;
    resetBatch();
    formRef.value?.clearValidate();
};

const acceptParams = (value: Firewall.Provider, item?: Firewall.InventoryItem) => {
    provider.value = value;
    mode.value = item?.desired?.uuid ? 'edit' : 'create';
    resetForm();
    if (mode.value === 'edit' && item?.desired?.uuid) {
        const rule = item.rule;
        editingUUID.value = item.desired.uuid;
        editingRule.value = { ...rule, scope: { ...rule.scope } };
        form.protocol = rule.protocol;
        form.sourceAddresses = [{ family: rule.scope.family, address: rule.sourceAddress || '' }];
        form.sourcePort = rule.sourcePort || '';
        form.destinationAddress = rule.destinationAddress || '';
        form.destinationPorts = splitTagValues([rule.destinationPort || '']);
        if (form.destinationPorts.length === 0) form.destinationPorts = [''];
        form.action = rule.action === 'reject' ? 'drop' : rule.action;
        form.priority = rule.priority;
        form.description = rule.description || '';
    }
    drawerVisible.value = true;
};

const handleClose = () => {
    drawerVisible.value = false;
    mode.value = 'create';
    resetBatch();
};

const resetPlan = () => {
    plan.value = undefined;
    resolution.value = undefined;
    selectedInstanceKey.value = '';
};

const resetBatch = () => {
    previewRules.value = [];
    batchPlans.value = [];
    confirmationIndexes.value = [];
    activeConfirmationIndex.value = -1;
    resetPlan();
};

const backToForm = () => resetBatch();

const changeProtocol = () => {
    if (!portProtocol.value) {
        form.sourcePort = '';
        form.destinationPorts = [''];
    }
};

const buildRule = (
    source: SourceAddressItem = form.sourceAddresses[0] || { family: 'ipv4', address: '' },
    destinationPort = form.destinationPorts[0] || '',
): Firewall.Rule => {
    const action =
        mode.value === 'edit' && editingRule.value?.action === 'reject' && form.action === 'drop'
            ? 'reject'
            : form.action;
    return {
        ...(editingRule.value || {}),
        scope:
            provider.value === 'iptables'
                ? {
                      provider: provider.value,
                      family: source.family,
                      table: 'filter',
                      chain: mode.value === 'edit' ? editingRule.value?.scope.chain || '1PANEL_BASIC' : '1PANEL_BASIC',
                      direction: 'input',
                  }
                : provider.value === 'firewalld'
                  ? {
                        provider: provider.value,
                        family: source.family,
                        zone: 'public',
                        direction: 'input',
                    }
                  : {
                        provider: provider.value,
                        family: source.family,
                        chain: 'incoming',
                        direction: 'input',
                    },
        protocol: form.protocol,
        sourceAddress: source.address,
        sourcePort: form.sourcePort,
        destinationAddress: form.destinationAddress,
        destinationPort,
        action,
        priority: form.priority,
        description: form.description,
    };
};

const editableFieldLabels: Array<[keyof Firewall.Rule, string]> = [
    ['protocol', 'commons.table.protocol'],
    ['sourceAddress', 'firewall.sourceIP'],
    ['sourcePort', 'firewall.sourcePort'],
    ['destinationAddress', 'firewall.destIP'],
    ['destinationPort', 'firewall.destPort'],
    ['action', 'firewall.action'],
    ['priority', 'firewall.priority'],
    ['description', 'commons.table.description'],
];

const changedFieldLabels = (before: Firewall.Rule, after: Firewall.Rule) =>
    editableFieldLabels
        .filter(([field]) => JSON.stringify(before[field] ?? '') !== JSON.stringify(after[field] ?? ''))
        .map(([, label]) => i18n.global.t(label));

const availableResolutions = (result: Firewall.RuleCheckResult) =>
    (result.allowedActions || []).filter((item): item is Firewall.ApplicableCheckAction => item !== 'cancel');

const previewAddress = (rule: Firewall.Rule) => {
    if (rule.sourceAddress) return rule.sourceAddress;
    if (rule.scope.family === 'ipv6') return '::/0';
    if (rule.scope.family === 'inet') return '0.0.0.0/0, ::/0';
    return '0.0.0.0/0';
};

const buildPreviewRules = () => {
    normalizeSourceAddresses();
    normalizeDestinationPorts();
    const addresses =
        form.sourceAddresses.length > 0 ? form.sourceAddresses : [{ family: 'ipv4' as const, address: '' }];
    const ports = form.destinationPorts.length > 0 ? form.destinationPorts : [''];
    const rules = addresses.flatMap((address) => ports.map((port) => buildRule(address, port)));
    if (rules.length > 256) {
        MsgError(i18n.global.t('firewall.batchRuleLimit', [256]));
        return;
    }
    previewRules.value = rules;
};

const showBatchConfirmation = (index: number) => {
    const item = batchPlans.value[index];
    activeConfirmationIndex.value = index;
    plan.value = item.plan;
    resolution.value = item.resolution || availableResolutions(item.plan)[0];
    selectedInstanceKey.value = item.selectedInstanceKey || '';
    if (resolution.value === 'select_adopt') {
        selectedInstanceKey.value = selectedInstanceKey.value || item.plan.candidates?.[0]?.instanceKey || '';
    }
};

const prepareBatchPlans = async () => {
    batchPlans.value = [];
    confirmationIndexes.value = [];
    const results = (await checkFirewallRulesBatch({ rules: previewRules.value })).data.items || [];
    if (results.length !== previewRules.value.length) {
        MsgError(i18n.global.t('commons.msg.operationFailed'));
        return;
    }
    for (let ruleIndex = 0; ruleIndex < previewRules.value.length; ruleIndex++) {
        const rule = previewRules.value[ruleIndex];
        const result = results[ruleIndex];
        if (result.decision === 'blocked') {
            batchPlans.value = [];
            MsgError(planReasonMessage(result.reason));
            return;
        }
        const available = availableResolutions(result);
        const item: BatchPlanItem = {
            rule,
            plan: result,
            resolution: result.decision === 'ready' && available.length === 1 ? available[0] : undefined,
        };
        const index = batchPlans.value.push(item) - 1;
        if (result.decision === 'confirmation_required') {
            confirmationIndexes.value.push(index);
        }
    }
    if (confirmationIndexes.value.length > 0) {
        showBatchConfirmation(confirmationIndexes.value[0]);
        return;
    }
    await executeBatchPlans();
};

const confirmBatchPlan = async () => {
    if (!resolution.value || (resolution.value === 'select_adopt' && !selectedInstanceKey.value)) {
        MsgError(i18n.global.t('commons.msg.selectOne', [i18n.global.t('firewall.external')]));
        return;
    }
    const item = batchPlans.value[activeConfirmationIndex.value];
    item.resolution = resolution.value;
    item.selectedInstanceKey = resolution.value === 'select_adopt' ? selectedInstanceKey.value : undefined;
    confirmationIndexes.value.shift();
    resetPlan();
    if (confirmationIndexes.value.length > 0) {
        showBatchConfirmation(confirmationIndexes.value[0]);
        return;
    }
    await executeBatchPlans();
};

const executeBatchPlans = async () => {
    const items: Firewall.CreateRequest[] = [];
    for (const item of batchPlans.value) {
        if (item.plan.decision === 'no_change') continue;
        const available = availableResolutions(item.plan);
        const selectedResolution =
            item.plan.decision === 'ready' && available.length === 1
                ? available[0]
                : available.includes(item.resolution as Firewall.ApplicableCheckAction)
                  ? item.resolution
                  : undefined;
        if (!selectedResolution) {
            MsgError(i18n.global.t('commons.msg.operationFailed'));
            return;
        }
        items.push({
            checkFlag: item.plan.checkFlag,
            action: selectedResolution,
            adoptInstanceKey:
                selectedResolution === 'adopt'
                    ? item.plan.candidates?.[0]?.instanceKey
                    : selectedResolution === 'select_adopt'
                      ? item.selectedInstanceKey
                      : undefined,
            rule: item.plan.requestedRule,
            sourceKind: 'user',
        });
    }

    if (items.length === 0) {
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        drawerVisible.value = false;
        return;
    }
    const result = (await createFirewallRulesBatch({ items })).data;
    if (result.succeeded > 0) emit('search');
    if (result.failed > 0) {
        MsgError(`${i18n.global.t('commons.msg.operationFailed')} (${result.failed}/${items.length})`);
        batchPlans.value = [];
        return;
    }
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    drawerVisible.value = false;
};

const onSubmit = async () => {
    if (loading.value) return;
    if (plan.value) {
        loading.value = true;
        try {
            await confirmBatchPlan();
        } finally {
            loading.value = false;
        }
        return;
    }
    if (showingPreview.value) {
        loading.value = true;
        try {
            await prepareBatchPlans();
        } finally {
            loading.value = false;
        }
        return;
    }
    if (!formRef.value) return;
    normalizeSourceAddresses();
    normalizeDestinationPorts();
    if (
        splitTagValues(form.sourceAddresses.map((item) => item.address)).length === 0 &&
        !form.destinationAddress &&
        !form.sourcePort &&
        splitTagValues(form.destinationPorts).length === 0
    ) {
        MsgError(i18n.global.t('firewall.ruleTargetRequired'));
        return;
    }
    const valid = await formRef.value.validate().catch(() => false);
    if (!valid) return;
    loading.value = true;
    try {
        if (mode.value === 'edit' && editingUUID.value) {
            const updatedRule = buildRule();
            const changed = editingRule.value ? changedFieldLabels(editingRule.value, updatedRule) : [];
            if (changed.length === 0) {
                drawerVisible.value = false;
                return;
            }
            try {
                await ElMessageBox.confirm(
                    i18n.global.t('firewall.editRuleConfirm', [changed.join(', ')]),
                    i18n.global.t('firewall.edit'),
                    {
                        confirmButtonText: i18n.global.t('commons.button.confirm'),
                        cancelButtonText: i18n.global.t('commons.button.cancel'),
                        type: 'warning',
                    },
                );
            } catch {
                return;
            }
            await updateFirewallRule(editingUUID.value, { rule: updatedRule });
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            emit('search');
            drawerVisible.value = false;
        } else {
            buildPreviewRules();
        }
    } finally {
        loading.value = false;
    }
};

const candidateLabel = (candidate: Firewall.ObservedRule) => {
    const rule = candidate.rule;
    const target = [rule.sourceAddress, rule.sourcePort, rule.destinationAddress, rule.destinationPort]
        .filter(Boolean)
        .join(' → ');
    return `#${candidate.locator.position || '-'} · ${rule.protocol} · ${target || i18n.global.t('commons.table.all')}`;
};

const emit = defineEmits<{ (event: 'search'): void }>();

defineExpose({ acceptParams });
</script>

<style lang="scss" scoped>
.ip-family-select {
    width: 100px;
}

.plan-confirmation {
    margin-top: 16px;
}

.rule-preview {
    margin-top: 4px;
}

.rule-preview-title {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 12px;
    font-weight: 500;
}

.resolution-list,
.candidate-list {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
    margin-top: 16px;
}

.candidate-list {
    margin-left: 24px;
}
</style>
