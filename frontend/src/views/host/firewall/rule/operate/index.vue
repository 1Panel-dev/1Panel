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
                    <el-option v-if="mode === 'create' || provider === 'ufw'" label="TCP/UDP" value="tcp/udp" />
                    <el-option label="TCP" value="tcp" />
                    <el-option label="UDP" value="udp" />
                    <el-option v-if="provider !== 'ufw'" label="ICMP" value="icmp" />
                    <el-option v-if="provider !== 'ufw'" label="ICMPV6" value="icmpv6" />
                </el-select>
            </el-form-item>
            <el-form-item label="IP" prop="sourceAddresses">
                <div v-for="(item, index) of form.sourceAddresses" :key="index" class="source-address-row mt-2">
                    <el-select v-model="item.family" class="ip-family-select" :disabled="mode === 'edit'">
                        <el-option label="IPv4" value="ipv4" />
                        <el-option label="IPv6" value="ipv6" />
                        <el-option v-if="provider === 'firewalld'" label="INET" value="inet" />
                    </el-select>
                    <el-select
                        ref="sourceAddressRefs"
                        v-model="item.address"
                        class="source-address-select"
                        clearable
                        filterable
                        allow-create
                        default-first-option
                        :placeholder="sourceAddressPlaceholder(item.family)"
                        @keyup.enter.prevent="addSourceAddressOnEnter(index)"
                    >
                        <el-option :label="wildcardAddressLabel(item.family)" :value="anywhereSourceValue" />
                    </el-select>
                    <el-button v-if="mode === 'create'" link icon="Delete" @click="removeSourceAddress(index)" />
                </div>
                <el-button v-if="mode === 'create'" class="mt-2" @click="addSourceAddress">
                    {{ $t('commons.button.add') }}
                </el-button>
            </el-form-item>
            <el-form-item :label="$t('commons.table.port')" prop="destinationPorts">
                <div v-for="(_, index) of form.destinationPorts" :key="index" class="destination-port-row mt-2">
                    <el-input
                        ref="destinationPortRefs"
                        v-model.trim="form.destinationPorts[index]"
                        class="destination-port-input"
                        clearable
                        :disabled="!portProtocol"
                        placeholder="80、80,443 或 8080-8089"
                        @keyup.enter.prevent="addDestinationPortOnEnter(index)"
                    />
                    <el-button
                        v-if="mode === 'create'"
                        link
                        icon="Delete"
                        :disabled="!portProtocol"
                        @click="removeRuleRow(index)"
                    />
                </div>
                <el-button v-if="mode === 'create'" class="mt-2" :disabled="!portProtocol" @click="addRuleRow">
                    {{ $t('commons.button.add') }}
                </el-button>
            </el-form-item>
            <el-form-item v-if="showPriorityField" :label="priorityFieldLabel">
                <el-input-number
                    v-model="form.priority"
                    :min="priorityMin"
                    :max="priorityMax"
                    controls-position="right"
                />
                <span class="priority-range">{{ priorityMin }} ~ {{ priorityMax }}</span>
            </el-form-item>
            <el-form-item :label="$t('commons.table.description')">
                <el-input v-model.trim="form.description" clearable />
            </el-form-item>
        </el-form>

        <div v-if="showingPreview" class="rule-preview">
            <div class="rule-preview-title">
                <span>{{ $t('firewall.ruleCheckResult') }}</span>
            </div>
            <div class="rule-check-groups">
                <section v-for="group in ruleCheckGroups" :key="group.status" class="rule-check-group">
                    <div class="rule-check-group-header">
                        <div class="rule-check-group-title">
                            <div class="rule-check-group-status">
                                <span class="rule-check-group-label">{{ group.label }} · {{ group.items.length }}</span>
                                <span class="rule-check-group-description">
                                    {{ ruleCheckGroupDescription(group.status) }}
                                </span>
                            </div>
                        </div>
                    </div>

                    <div class="rule-check-items">
                        <div v-for="item in group.items" :key="ruleCheckItemKey(item)" class="rule-check-item">
                            <div class="rule-check-item-main">
                                <div class="rule-check-rule-summary">
                                    <span class="rule-check-protocol">
                                        {{ previewProtocol(previewRule(item)) }}
                                    </span>
                                    <span class="rule-check-separator">·</span>
                                    <span class="rule-check-address">{{ previewAddress(previewRule(item)) }}</span>
                                    <span class="rule-check-arrow">→</span>
                                    <span>
                                        {{ $t('commons.table.port') }}
                                        {{ previewRule(item).destinationPort || '*' }}
                                    </span>
                                </div>
                            </div>
                            <div class="rule-check-item-meta">
                                <span
                                    class="rule-check-action"
                                    :class="previewRule(item).action === 'accept' ? 'is-accept' : 'is-drop'"
                                >
                                    <i
                                        class="iconfont rule-check-action-icon"
                                        :class="previewRule(item).action === 'accept' ? 'p-yunxu' : 'p-a-44tubiao-139'"
                                        aria-hidden="true"
                                    />
                                    {{ $t(`firewall.${previewRule(item).action === 'accept' ? 'accept' : 'drop'}`) }}
                                </span>
                                <span v-if="previewRulePriority(previewRule(item)) !== undefined">
                                    {{ priorityFieldLabel }}：{{ previewRulePriority(previewRule(item)) }}
                                </span>
                                <span v-if="previewRule(item).description" class="rule-check-description">
                                    {{ previewRule(item).description }}
                                </span>
                                <span
                                    v-if="['warning', 'error'].includes(ruleCheckStatus(item.plan))"
                                    class="rule-check-item-reason"
                                    :class="`is-${ruleCheckStatus(item.plan)}`"
                                >
                                    {{ ruleCheckDescription(item.plan) }}
                                </span>
                            </div>
                        </div>
                    </div>
                </section>
            </div>
        </div>

        <template #footer>
            <el-button :disabled="loading" @click="drawerVisible = false">
                {{ $t('commons.button.cancel') }}
            </el-button>
            <el-button v-if="showingPreview" :disabled="loading" @click="backToForm">
                {{ $t('commons.button.back') }}
            </el-button>
            <el-button
                type="primary"
                :loading="loading"
                :disabled="showingPreview && hasBlockingRules"
                @click="onCheckOrSubmit"
            >
                {{ $t(showingPreview ? 'commons.button.submit' : 'commons.button.check') }}
            </el-button>
        </template>
    </DrawerPro>
    <ErrDialog ref="errDialogRef" @close="backToForm" />
</template>

<script lang="ts" setup>
import { Firewall } from '@/api/interface/firewall';
import {
    checkFirewallRule,
    checkFirewallRulesBatch,
    createFirewallRulesBatch,
    updateFirewallRule,
} from '@/api/modules/firewall';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgError, MsgSuccess, MsgWarning } from '@/utils/message';
import { computed, nextTick, reactive, ref, watch } from 'vue';
import { ElMessageBox, type FormInstance, type FormRules } from 'element-plus';
import ErrDialog from './err-message.vue';

const provider = ref<Firewall.Provider>('iptables');
const mode = ref<'create' | 'edit'>('create');
const editingUUID = ref('');
const editingRule = ref<Firewall.Rule>();
const drawerVisible = ref(false);
const loading = ref(false);
const formRef = ref<FormInstance>();
const sourceAddressRefs = ref<Array<{ focus: () => void }>>([]);
const destinationPortRefs = ref<Array<{ focus: () => void }>>([]);
const errDialogRef = ref<InstanceType<typeof ErrDialog>>();
const previewRules = ref<Firewall.Rule[]>([]);
const previewVisible = ref(false);
const checkCompleted = ref(false);

interface PriorityPositionRange {
    min: number;
    max: number;
}

const positionRanges = ref<Partial<Record<Firewall.Family, PriorityPositionRange>>>({});
const firewalldPrioritySupported = ref(true);

interface BatchPlanItem {
    rule: Firewall.Rule;
    plan: Firewall.RuleCheckResult;
    resolution?: Firewall.ApplicableCheckAction;
}

interface SourceAddressItem {
    family: Firewall.Family;
    address: string;
}

const batchPlans = ref<BatchPlanItem[]>([]);
const anywhereSourceValue = '__1panel_anywhere__';

const form = reactive({
    protocol: 'tcp',
    sourceAddresses: [{ family: 'ipv4', address: anywhereSourceValue }] as SourceAddressItem[],
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

const portProtocol = computed(() => ['tcp', 'udp', 'tcp/udp'].includes(form.protocol));
const sourceAddressPlaceholder = (family: Firewall.Family) => {
    if (family === 'ipv6') return '2001:db8::1 或 2001:db8::/64';
    if (family === 'inet') return '0.0.0.0/0 或 ::/0';
    return '172.16.10.11 或 172.16.0.0/24';
};
const wildcardAddress = (family: Firewall.Family) => {
    if (family === 'ipv6') return '::/0';
    if (family === 'inet') return '0.0.0.0/0, ::/0';
    return '0.0.0.0/0';
};
const wildcardAddressLabel = (family: Firewall.Family) =>
    `${wildcardAddress(family)}（${i18n.global.t('firewall.anyWhere')}）`;
const isWildcardAddress = (family: Firewall.Family, address?: string) => {
    const value = address?.trim();
    return !value || value === anywhereSourceValue || value === wildcardAddress(family);
};

const priorityFieldLabel = computed(() => i18n.global.t('firewall.priority'));
const showPriorityField = computed(() => {
    if (provider.value !== 'firewalld') return mode.value === 'edit';
    return (
        firewalldPrioritySupported.value && (mode.value === 'create' || editingRule.value?.nativeKind === 'rich_rule')
    );
});
const selectedPositionRanges = computed(() => {
    const families = [...new Set(form.sourceAddresses.map((item) => item.family))];
    if (families.length === 0) return [{ min: 1, max: 1 }];
    return families.map((family) => positionRanges.value[family] || { min: 1, max: 1 });
});
const positionalPriorityMin = computed(() => Math.max(1, ...selectedPositionRanges.value.map((range) => range.min)));
const positionalPriorityMax = computed(() => Math.min(...selectedPositionRanges.value.map((range) => range.max)));
const priorityMin = computed(() => (provider.value === 'firewalld' ? -32768 : positionalPriorityMin.value));
const priorityMax = computed(() => (provider.value === 'firewalld' ? 32767 : positionalPriorityMax.value));
const showingPreview = computed(() => previewVisible.value && previewRules.value.length > 0);

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

type RuleCheckDisplayStatus = 'creatable' | 'existing' | 'warning' | 'error';

const ruleCheckStatus = (result: Firewall.RuleCheckResult): RuleCheckDisplayStatus => {
    if (result.decision === 'blocked') return 'error';
    if (result.decision === 'no_change' || result.classification === 'exact_external') return 'existing';
    if (result.decision === 'confirmation_required') return 'warning';
    return 'creatable';
};

const ruleCheckDescription = (result: Firewall.RuleCheckResult) => {
    if (ruleCheckStatus(result) === 'creatable') return i18n.global.t('firewall.ruleCheckReadyHelper');
    if (result.classification === 'exact_external') return i18n.global.t('firewall.ruleCheckExternalExists');
    return planReasonMessage(result.reason);
};

const ruleCheckGroupDescription = (status: RuleCheckDisplayStatus) => {
    if (status === 'error') return i18n.global.t('firewall.ruleCheckBlockedHelper');
    if (status === 'warning') return i18n.global.t('firewall.ruleCheckWarningHelper');
    if (status === 'existing') return i18n.global.t('firewall.ruleCheckExistingHelper');
    return i18n.global.t('firewall.ruleCheckReadyHelper');
};

const ruleCheckCounts = computed(() => {
    const counts: Record<RuleCheckDisplayStatus, number> = {
        creatable: 0,
        existing: 0,
        warning: 0,
        error: 0,
    };
    for (const item of batchPlans.value) counts[ruleCheckStatus(item.plan)]++;
    return counts;
});

const ruleCheckGroupOrder: RuleCheckDisplayStatus[] = ['error', 'warning', 'creatable', 'existing'];
const ruleCheckGroups = computed(() =>
    ruleCheckGroupOrder
        .map((status) => ({
            status,
            label: i18n.global.t(`firewall.ruleCheckStatus_${status}`),
            items: batchPlans.value.filter((item) => ruleCheckStatus(item.plan) === status),
        }))
        .filter((group) => group.items.length > 0),
);

const previewRule = (item: BatchPlanItem) => item.plan.requestedRule || item.rule;
const previewRulePriority = (rule: Firewall.Rule) => rule.priority ?? rule.orderIndex;
const ruleCheckItemKey = (item: BatchPlanItem) =>
    [
        item.plan.checkFlag,
        previewRule(item).scope.family,
        previewRule(item).protocol,
        previewRule(item).sourceAddress,
        previewRule(item).destinationPort,
    ].join(':');

const hasBlockingRules = computed(() => ruleCheckCounts.value.error > 0);
const existingRuleItems = computed(() => batchPlans.value.filter((item) => ruleCheckStatus(item.plan) === 'existing'));
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
        (isWildcardAddress(item.family, item.address) ? [anywhereSourceValue] : splitTagValues([item.address])).flatMap(
            (address) => {
                const key = `${item.family}:${address}`;
                if (seen.has(key)) return [];
                seen.add(key);
                return [{ family: item.family, address }];
            },
        ),
    );
    const fallback = form.sourceAddresses[0] || { family: 'ipv4', address: anywhereSourceValue };
    form.sourceAddresses =
        mode.value === 'edit'
            ? [normalized[0] || { ...fallback, address: anywhereSourceValue }]
            : normalized.length > 0
              ? normalized
              : [{ ...fallback, address: anywhereSourceValue }];
};
const normalizeDestinationPorts = (values = form.destinationPorts) => {
    const splitPortList = provider.value === 'firewalld' || (provider.value === 'ufw' && form.protocol === 'tcp/udp');
    const splitPattern = splitPortList ? /[,，;；\s]+/ : /[;；\s]+/;
    const normalized = [
        ...new Set(
            values
                .map((value) => (splitPortList ? value : value.replace(/\s*[,，]\s*/g, ',')))
                .flatMap((value) => value.split(splitPattern))
                .map((value) => value.trim().replaceAll('，', ','))
                .filter(Boolean),
        ),
    ];
    if (mode.value === 'edit' && normalized.length > 1) {
        MsgError(i18n.global.t('commons.msg.notSupportOperation'));
        return false;
    }
    form.destinationPorts = mode.value === 'edit' ? [normalized[0] || ''] : normalized.length > 0 ? normalized : [''];
    return true;
};
const addSourceAddress = () => {
    const family = form.sourceAddresses.at(-1)?.family || 'ipv4';
    form.sourceAddresses.push({ family, address: anywhereSourceValue });
};
const addSourceAddressOnEnter = async (index: number) => {
    await nextTick();
    if (mode.value !== 'create' || !form.sourceAddresses[index]?.address.trim()) return;
    if (index < form.sourceAddresses.length - 1) {
        sourceAddressRefs.value[index + 1]?.focus();
        return;
    }
    addSourceAddress();
    await nextTick();
    sourceAddressRefs.value.at(-1)?.focus();
};
const removeSourceAddress = (index: number) => {
    form.sourceAddresses.splice(index, 1);
};
const addRuleRow = () => {
    form.destinationPorts.push('');
};
const addDestinationPortOnEnter = async (index: number) => {
    if (mode.value !== 'create' || !portProtocol.value || !form.destinationPorts[index]?.trim()) return;
    if (index < form.destinationPorts.length - 1) {
        destinationPortRefs.value[index + 1]?.focus();
        return;
    }
    addRuleRow();
    await nextTick();
    destinationPortRefs.value.at(-1)?.focus();
};
const removeRuleRow = (index: number) => {
    form.destinationPorts.splice(index, 1);
};

const resetForm = () => {
    form.protocol = 'tcp';
    form.sourceAddresses = [{ family: defaultFamily(), address: anywhereSourceValue }];
    form.sourcePort = '';
    form.destinationAddress = '';
    form.destinationPorts = [''];
    form.action = 'accept';
    form.priority = provider.value === 'firewalld' || mode.value === 'create' ? undefined : positionalPriorityMax.value;
    form.description = '';
    editingUUID.value = '';
    editingRule.value = undefined;
    resetBatch();
    formRef.value?.clearValidate();
};

const acceptParams = (
    value: Firewall.Provider,
    item?: Firewall.InventoryItem,
    ranges: Partial<Record<Firewall.Family, PriorityPositionRange>> = {},
    supportsExplicitPriority = true,
) => {
    provider.value = value;
    firewalldPrioritySupported.value = supportsExplicitPriority;
    positionRanges.value = ranges;
    mode.value = item?.desired?.uuid ? 'edit' : 'create';
    resetForm();
    if (mode.value === 'edit' && item?.desired?.uuid) {
        const rule = item.rule;
        const currentPosition = item.observed?.locator.position || rule.orderIndex;
        editingUUID.value = item.desired.uuid;
        editingRule.value = {
            ...rule,
            scope: { ...rule.scope },
            orderIndex: provider.value === 'firewalld' ? undefined : currentPosition,
        };
        form.protocol =
            provider.value === 'ufw' && rule.protocol === 'all' && rule.destinationPort ? 'tcp/udp' : rule.protocol;
        form.sourceAddresses = [
            {
                family: rule.scope.family,
                address: isWildcardAddress(rule.scope.family, rule.sourceAddress)
                    ? anywhereSourceValue
                    : rule.sourceAddress || anywhereSourceValue,
            },
        ];
        form.sourcePort = rule.sourcePort || '';
        form.destinationAddress = rule.destinationAddress || '';
        form.destinationPorts = [rule.destinationPort || ''];
        if (form.destinationPorts.length === 0) form.destinationPorts = [''];
        form.action = rule.action === 'reject' ? 'drop' : rule.action;
        form.priority = provider.value === 'firewalld' ? rule.priority : currentPosition || positionalPriorityMax.value;
        form.description = rule.description || '';
    }
    drawerVisible.value = true;
};

watch(priorityMax, (max) => {
    if (form.priority !== undefined && form.priority > max) form.priority = max;
});
watch(priorityMin, (min) => {
    if (form.priority !== undefined && form.priority < min) form.priority = min;
});
const handleClose = () => {
    drawerVisible.value = false;
    mode.value = 'create';
    resetBatch();
};

const resetBatch = () => {
    previewRules.value = [];
    previewVisible.value = false;
    checkCompleted.value = false;
    batchPlans.value = [];
};

watch(
    form,
    () => {
        if (!checkCompleted.value) return;
        checkCompleted.value = false;
        previewRules.value = [];
        batchPlans.value = [];
    },
    { deep: true },
);

const backToForm = () => resetBatch();

const changeProtocol = () => {
    if (!portProtocol.value) {
        form.sourcePort = '';
        form.destinationPorts = [''];
    }
};

const buildRule = (
    source: SourceAddressItem = form.sourceAddresses[0] || { family: 'ipv4', address: anywhereSourceValue },
    destinationPort = form.destinationPorts[0] || '',
    protocol = form.protocol,
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
        protocol,
        sourceAddress: isWildcardAddress(source.family, source.address) ? '' : source.address,
        sourcePort: form.sourcePort,
        destinationAddress: form.destinationAddress,
        destinationPort,
        action,
        priority: provider.value === 'firewalld' && firewalldPrioritySupported.value ? form.priority : undefined,
        orderIndex: provider.value === 'firewalld' || mode.value === 'create' ? undefined : form.priority,
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
    ['orderIndex', 'firewall.priority'],
    ['description', 'commons.table.description'],
];

const changedFieldLabels = (before: Firewall.Rule, after: Firewall.Rule) =>
    editableFieldLabels
        .filter(([field]) => JSON.stringify(before[field] ?? '') !== JSON.stringify(after[field] ?? ''))
        .map(([, label]) => i18n.global.t(label));

const availableResolutions = (result: Firewall.RuleCheckResult) =>
    (result.allowedActions || []).filter((item): item is Firewall.ApplicableCheckAction => item !== 'cancel');

const previewAddress = (rule: Firewall.Rule) => {
    if (rule.sourceAddress && !isWildcardAddress(rule.scope.family, rule.sourceAddress)) return rule.sourceAddress;
    return wildcardAddressLabel(rule.scope.family);
};
const previewProtocol = (rule: Firewall.Rule) =>
    rule.scope.provider === 'ufw' && rule.protocol === 'all' && rule.destinationPort
        ? 'TCP/UDP'
        : rule.protocol.toUpperCase();

const buildPreviewRules = () => {
    normalizeSourceAddresses();
    if (!normalizeDestinationPorts()) return false;
    const addresses =
        form.sourceAddresses.length > 0 ? form.sourceAddresses : [{ family: 'ipv4' as const, address: '' }];
    const ports = form.destinationPorts.length > 0 ? form.destinationPorts : [''];
    const orderOffsets = new Map<string, number>();
    const rules = addresses.flatMap((address) =>
        ports.flatMap((port) =>
            (form.protocol === 'tcp/udp'
                ? provider.value === 'ufw' && port && !port.includes(',')
                    ? ['all']
                    : ['tcp', 'udp']
                : [form.protocol]
            ).map((protocol) => {
                const rule = buildRule(address, port, protocol);
                if (provider.value === 'firewalld' || rule.orderIndex === undefined) return rule;
                const scopeKey = provider.value === 'ufw' ? 'ufw' : JSON.stringify(rule.scope);
                const offset = orderOffsets.get(scopeKey) || 0;
                orderOffsets.set(scopeKey, offset + 1);
                rule.orderIndex += offset;
                return rule;
            }),
        ),
    );
    if (rules.length > 256) {
        MsgError(i18n.global.t('firewall.batchRuleLimit', [256]));
        return false;
    }
    previewRules.value = rules;
    return true;
};

const prepareBatchPlans = async () => {
    checkCompleted.value = false;
    batchPlans.value = [];
    const results = (await checkFirewallRulesBatch({ rules: previewRules.value })).data.items || [];
    if (results.length !== previewRules.value.length) {
        MsgError(i18n.global.t('commons.msg.operationFailed'));
        return;
    }
    for (let ruleIndex = 0; ruleIndex < previewRules.value.length; ruleIndex++) {
        const rule = previewRules.value[ruleIndex];
        const result = results[ruleIndex];
        const available = availableResolutions(result);
        batchPlans.value.push({
            rule,
            plan: result,
            resolution:
                ruleCheckStatus(result) === 'creatable' || ruleCheckStatus(result) === 'warning'
                    ? available.find((action) => action !== 'select_adopt')
                    : undefined,
        });
    }
    previewRules.value = results.map((result) => result.requestedRule);
    checkCompleted.value = true;
    previewVisible.value = true;
};

const executeBatchPlans = async () => {
    if (hasBlockingRules.value) {
        MsgError(i18n.global.t('firewall.ruleCheckBlockedHelper'));
        return;
    }
    const items: Firewall.CreateRequest[] = [];
    for (const item of batchPlans.value) {
        if (ruleCheckStatus(item.plan) === 'existing') continue;
        const available = availableResolutions(item.plan);
        const selectedResolution = available.includes(item.resolution as Firewall.ApplicableCheckAction)
            ? item.resolution
            : undefined;
        if (!selectedResolution) {
            MsgError(i18n.global.t('commons.msg.operationFailed'));
            return;
        }
        items.push({
            checkFlag: item.plan.checkFlag,
            action: selectedResolution,
            rule: item.plan.requestedRule,
            sourceKind: 'user',
        });
    }

    if (items.length === 0) {
        MsgWarning(i18n.global.t('firewall.allRulesAlreadyExist', [existingRuleItems.value.length]));
        drawerVisible.value = false;
        return;
    }
    const result = (await createFirewallRulesBatch({ items })).data;
    if (result.succeeded > 0) emit('search');
    if (result.failed > 0 || (result.skipped || 0) > 0) {
        errDialogRef.value?.acceptParams(result);
        return;
    }
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    drawerVisible.value = false;
};

const prepareRulesFromForm = async () => {
    if (!formRef.value) return previewRules.value.length > 0;
    normalizeSourceAddresses();
    if (!normalizeDestinationPorts()) return false;
    if (
        splitTagValues(form.sourceAddresses.map((item) => item.address)).length === 0 &&
        !form.destinationAddress &&
        !form.sourcePort &&
        splitTagValues(form.destinationPorts).length === 0
    ) {
        MsgError(i18n.global.t('firewall.ruleTargetRequired'));
        return false;
    }
    const valid = await formRef.value.validate().catch(() => false);
    if (!valid) return false;
    return buildPreviewRules();
};

const checkRules = async () => {
    if (!(await prepareRulesFromForm())) return;
    if (mode.value === 'edit' && editingUUID.value) {
        const result = (await checkFirewallRule({ uuid: editingUUID.value, rule: previewRules.value[0] })).data;
        previewRules.value = [result.requestedRule];
        batchPlans.value = [{ rule: result.requestedRule, plan: result }];
        checkCompleted.value = true;
        previewVisible.value = true;
        return;
    }
    await prepareBatchPlans();
};

const executeEdit = async () => {
    if (!editingUUID.value || previewRules.value.length === 0) return;
    if (hasBlockingRules.value) {
        MsgError(i18n.global.t('firewall.ruleCheckBlockedHelper'));
        return;
    }
    const updatedRule = previewRules.value[0];
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
};

const submitCheckedRules = async () => {
    if (mode.value === 'edit') {
        await executeEdit();
        return;
    }
    await executeBatchPlans();
};

const onCheckOrSubmit = async () => {
    if (loading.value) return;
    loading.value = true;
    try {
        if (showingPreview.value) {
            await submitCheckedRules();
        } else {
            await checkRules();
        }
    } finally {
        loading.value = false;
    }
};

const emit = defineEmits<{ (event: 'search'): void }>();

defineExpose({ acceptParams });
</script>

<style lang="scss" scoped>
.ip-family-select {
    width: 100px;
    flex: none;
}

.source-address-row,
.destination-port-row {
    display: flex;
    align-items: center;
    width: 100%;
    gap: 8px;
}

.rule-check-alert {
    margin-bottom: 12px;
}

.source-address-select {
    flex: 1;
}

.destination-port-input {
    flex: 1;
}

.priority-range {
    margin-left: 12px;
    color: var(--el-text-color-secondary);
}

.rule-preview {
    margin-top: 4px;
    color: var(--el-text-color-primary);
}

.rule-preview-title {
    display: flex;
    align-items: center;
    margin-bottom: 12px;
    color: var(--el-text-color-primary);
    font-weight: 500;
}

.rule-check-groups {
    max-height: 460px;
    overflow-y: auto;
}

.rule-check-group + .rule-check-group {
    margin-top: 16px;
}

.rule-check-group-header {
    display: flex;
    align-items: center;
    min-height: 32px;
    padding: 0 4px;
    color: var(--el-text-color-primary);
    font-weight: 500;
}

.rule-check-group-title {
    display: flex;
    align-items: flex-start;
    min-width: 0;
}

.rule-check-group-status {
    display: flex;
    align-items: center;
    min-width: 0;
    flex-wrap: wrap;
    gap: 6px;
}

.rule-check-group-label {
    flex: none;
    font-size: 14px;
    font-weight: 500;
    color: var(--el-text-color-primary);
    line-height: 20px;
}

.rule-check-group-description {
    color: var(--el-text-color-secondary);
    font-size: 11px;
    font-weight: normal;
    line-height: 18px;
}

.rule-check-items {
    overflow: hidden;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 6px;
    background: var(--el-fill-color-light);
}

.rule-check-item {
    padding: 12px 14px;
    background: var(--el-fill-color-light);

    & + & {
        border-top: 1px solid var(--el-border-color-lighter);
    }
}

.rule-check-item-main {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
}

.rule-check-rule-summary {
    display: flex;
    align-items: center;
    min-width: 0;
    gap: 8px;
    color: var(--el-text-color-primary);
    font-size: 13px;
}

.rule-check-protocol {
    flex: none;
    font-weight: 500;
}

.rule-check-separator,
.rule-check-arrow {
    flex: none;
    color: var(--el-text-color-placeholder);
}

.rule-check-address {
    overflow-wrap: anywhere;
}

.rule-check-item-meta {
    display: flex;
    align-items: center;
    min-width: 0;
    margin-top: 7px;
    gap: 12px;
    color: var(--el-text-color-secondary);
    font-size: 12px;
}

.rule-check-action {
    display: inline-flex;
    align-items: center;
    flex: none;
    gap: 4px;
    color: var(--el-text-color-secondary);
    line-height: 18px;

    &.is-accept .rule-check-action-icon {
        color: var(--el-color-primary);
    }

    &.is-drop .rule-check-action-icon {
        color: var(--el-color-info);
    }
}

.rule-check-action-icon {
    font-size: 14px;
    line-height: 1;
}

.rule-check-description {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.rule-check-item-reason {
    &.is-warning {
        color: var(--el-color-warning);
    }

    &.is-error {
        color: var(--el-color-danger);
    }
}
</style>
