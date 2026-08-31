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
                    <el-option label="ALL" value="all">
                        <div class="protocol-option">
                            <span>ALL</span>
                            <span class="protocol-option-description">{{ $t('firewall.allProtocolHelper') }}</span>
                        </div>
                    </el-option>
                </el-select>
            </el-form-item>
            <el-form-item label="IP" prop="sourceAddresses">
                <div v-for="(item, index) of form.sourceAddresses" :key="index" class="source-address-row mt-2">
                    <el-input
                        ref="sourceAddressRefs"
                        v-model.trim="item.address"
                        class="source-address-select"
                        clearable
                        :placeholder="$t('firewall.sourceAddressPlaceholder')"
                        @keyup.enter.prevent="addSourceAddressOnEnter(index)"
                    >
                        <template #append>
                            <el-button v-if="mode === 'create'" icon="Delete" @click="removeSourceAddress(index)" />
                        </template>
                    </el-input>
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
                        :placeholder="$t('firewall.destinationPortPlaceholder')"
                        @keyup.enter.prevent="addDestinationPortOnEnter(index)"
                    >
                        <template #append>
                            <el-button
                                v-if="mode === 'create'"
                                icon="Delete"
                                :disabled="!portProtocol"
                                @click="removeRuleRow(index)"
                            />
                        </template>
                    </el-input>
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

        <div v-if="showingPreview" v-loading="loading" class="rule-preview">
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
                                    v-if="ruleCheckStatus(item.plan) === 'error'"
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
                :disabled="loading || (showingPreview && hasBlockingRules)"
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
import { checkFirewallRules, createFirewallRules, updateFirewallRule } from '@/api/modules/firewall';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgError, MsgSuccess, MsgWarning } from '@/utils/message';
import { computed, nextTick, reactive, ref, watch } from 'vue';
import { ElMessageBox, type FormInstance, type FormRules } from 'element-plus';
import ErrDialog from './err-message.vue';
import {
    formatHostAddress,
    inferAddressFamily,
    isValidIPOrCIDR,
    isValidPortRange,
    normalizePortRange,
    splitTagValues,
} from '@/views/host/firewall/utils/validation';

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

const defaultFamily = (): Firewall.Family => 'ipv4';
type ValidationCallback = (error?: Error) => void;
const hasRuleTarget = () =>
    splitTagValues(form.sourceAddresses.map((item) => item.address)).length > 0 ||
    Boolean(form.destinationAddress) ||
    Boolean(form.sourcePort) ||
    splitTagValues(form.destinationPorts).length > 0;

const validateSourceAddresses = (_rule: unknown, value: SourceAddressItem[], callback: ValidationCallback) => {
    const addresses = splitTagValues((value || []).map((item) => item.address));
    if (addresses.some((address) => !isValidIPOrCIDR(address))) {
        callback(new Error(i18n.global.t('commons.rule.ip')));
        return;
    }
    if (mode.value === 'edit' && addresses.length > 1) {
        callback(
            new Error(`${i18n.global.t('firewall.sourceIP')}: ${i18n.global.t('commons.msg.notSupportOperation')}`),
        );
        return;
    }
    if (!hasRuleTarget()) {
        callback(new Error(i18n.global.t('firewall.ruleTargetRequired')));
        return;
    }
    callback();
};

const validateDestinationPorts = (_rule: unknown, value: string[], callback: ValidationCallback) => {
    const ports = splitTagValues(value || []);
    if (ports.some((port) => !isValidPortRange(port))) {
        callback(new Error(i18n.global.t('commons.rule.port')));
        return;
    }
    callback();
};

const rules = reactive<FormRules>({
    protocol: [Rules.requiredSelect],
    action: [Rules.requiredSelect],
    sourceAddresses: [{ validator: validateSourceAddresses, trigger: ['blur', 'change'] }],
    destinationPorts: [{ validator: validateDestinationPorts, trigger: ['blur', 'change'] }],
});

const portProtocol = computed(() => ['tcp', 'udp', 'tcp/udp'].includes(form.protocol));
const wildcardAddress = (family: Firewall.Family) => {
    if (family === 'ipv6') return '::/0';
    if (family === 'inet') return '0.0.0.0/0, ::/0';
    return '0.0.0.0/0';
};
const wildcardAddressLabel = (family: Firewall.Family) =>
    `${wildcardAddress(family)}（${i18n.global.t('firewall.anyWhere')}）`;
const isWildcardAddress = (_family: Firewall.Family, address?: string) => !address?.trim();
const priorityFieldLabel = computed(() => i18n.global.t('firewall.priority'));
const showPriorityField = computed(() => {
    if (provider.value === 'ufw') return mode.value === 'edit';
    if (provider.value !== 'firewalld') return true;
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
    'equivalent_managed_rule',
    'managed_rule_drifted',
    'opaque_rule_in_target_scope',
    'runtime_permanent_mismatch',
    'protected_rule',
]);

const planReasonMessage = (reason: string) =>
    i18n.global.t(`firewall.plan_${supportedPlanReasons.has(reason) ? reason : 'blocked'}`);

type RuleCheckDisplayStatus = 'creatable' | 'existing' | 'error';

const ruleCheckStatus = (result: Firewall.RuleCheckResult): RuleCheckDisplayStatus => {
    if (result.decision === 'blocked') return 'error';
    if (result.decision === 'no_change' || result.classification === 'exact_external') return 'existing';
    return 'creatable';
};

const ruleCheckDescription = (result: Firewall.RuleCheckResult) => {
    if (ruleCheckStatus(result) === 'creatable') return i18n.global.t('firewall.ruleCheckReadyHelper');
    if (result.classification === 'exact_external') return i18n.global.t('firewall.ruleCheckExternalExists');
    return planReasonMessage(result.reason);
};

const ruleCheckGroupDescription = (status: RuleCheckDisplayStatus) => {
    if (status === 'error') return i18n.global.t('firewall.ruleCheckBlockedHelper');
    if (status === 'existing') return i18n.global.t('firewall.ruleCheckExistingHelper');
    return i18n.global.t('firewall.ruleCheckReadyHelper');
};

const ruleCheckCounts = computed(() => {
    const counts: Record<RuleCheckDisplayStatus, number> = {
        creatable: 0,
        existing: 0,
        error: 0,
    };
    for (const item of batchPlans.value) counts[ruleCheckStatus(item.plan)]++;
    return counts;
});

const ruleCheckGroupOrder: RuleCheckDisplayStatus[] = ['error', 'creatable', 'existing'];
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
const normalizeSourceAddresses = () => {
    const seen = new Set<string>();
    let normalized = form.sourceAddresses.flatMap((item) => {
        const addresses = splitTagValues([item.address]);
        if (addresses.length === 0) return [{ family: item.family, address: '' }];
        return addresses.flatMap((address) => {
            const family = inferAddressFamily(address);
            const key = `${family}:${address}`;
            if (seen.has(key)) return [];
            seen.add(key);
            return [{ family, address }];
        });
    });
    if (normalized.some((item) => item.address)) {
        normalized = normalized.filter((item) => item.address);
    }
    if (mode.value === 'edit' && normalized.length > 1) {
        MsgError(`${i18n.global.t('firewall.sourceIP')}: ${i18n.global.t('commons.msg.notSupportOperation')}`);
        return false;
    }
    const fallback = form.sourceAddresses[0] || { family: 'ipv4', address: '' };
    form.sourceAddresses =
        mode.value === 'edit'
            ? [normalized[0] || { ...fallback, address: '' }]
            : normalized.length > 0
              ? normalized
              : [{ ...fallback, address: '' }];
    return true;
};
const normalizeDestinationPorts = (values = form.destinationPorts) => {
    const normalized = splitTagValues(values).map(normalizePortRange);
    if (provider.value === 'ufw' || provider.value === 'iptables') {
        form.destinationPorts = [normalized.join(',')];
        return true;
    }
    if (mode.value === 'edit' && normalized.length > 1) {
        MsgError(i18n.global.t('commons.msg.notSupportOperation'));
        return false;
    }
    form.destinationPorts = normalized.length > 0 ? normalized : [''];
    return true;
};
const addSourceAddress = () => {
    const family = form.sourceAddresses.at(-1)?.family || 'ipv4';
    form.sourceAddresses.push({ family, address: '' });
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
    form.sourceAddresses = [{ family: defaultFamily(), address: '' }];
    form.sourcePort = '';
    form.destinationAddress = '';
    form.destinationPorts = [''];
    form.action = 'accept';
    form.priority =
        provider.value === 'firewalld' || provider.value === 'ufw' ? undefined : positionalPriorityMax.value;
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
                address: formatHostAddress(rule.sourceAddress, rule.scope.family),
            },
        ];
        form.sourcePort = rule.sourcePort || '';
        form.destinationAddress = formatHostAddress(rule.destinationAddress, rule.scope.family);
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
    source: SourceAddressItem = form.sourceAddresses[0] || { family: 'ipv4', address: '' },
    destinationPort = form.destinationPorts[0] || '',
    protocol = form.protocol,
): Firewall.Rule => {
    const family = source.address
        ? source.family
        : protocol === 'icmpv6'
          ? 'ipv6'
          : protocol === 'icmp'
            ? 'ipv4'
            : provider.value === 'firewalld'
              ? 'inet'
              : source.family;
    const action =
        mode.value === 'edit' && editingRule.value?.action === 'reject' && form.action === 'drop'
            ? 'reject'
            : form.action;
    return {
        ...(editingRule.value || {}),
        scope:
            provider.value === 'iptables' || provider.value === 'nftables'
                ? {
                      provider: provider.value,
                      family,
                      table: 'filter',
                      chain: mode.value === 'edit' ? editingRule.value?.scope.chain || '1PANEL_BASIC' : '1PANEL_BASIC',
                      direction: 'input',
                  }
                : provider.value === 'firewalld'
                  ? {
                        provider: provider.value,
                        family,
                        zone: 'public',
                        direction: 'input',
                    }
                  : {
                        provider: provider.value,
                        family,
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
        orderIndex:
            provider.value === 'firewalld' || (provider.value === 'ufw' && mode.value === 'create')
                ? undefined
                : form.priority,
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
    if (rule.sourceAddress && !isWildcardAddress(rule.scope.family, rule.sourceAddress)) {
        return formatHostAddress(rule.sourceAddress, rule.scope.family);
    }
    return wildcardAddressLabel(rule.scope.family);
};
const previewProtocol = (rule: Firewall.Rule) =>
    rule.scope.provider === 'ufw' && rule.protocol === 'all' && rule.destinationPort
        ? 'TCP/UDP'
        : rule.protocol.toUpperCase();

const buildPreviewRules = () => {
    if (!normalizeSourceAddresses()) return false;
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
    if (mode.value === 'edit' && rules.length !== 1) {
        MsgError(i18n.global.t('commons.msg.notSupportOperation'));
        return false;
    }
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
    const results = (await checkFirewallRules({ items: previewRules.value.map((rule) => ({ rule })) })).data.items;
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
                ruleCheckStatus(result) === 'creatable'
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
    const items: Firewall.CreateItem[] = [];
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
    const result = (await createFirewallRules({ items })).data;
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
    const valid = await formRef.value.validate().catch(() => false);
    if (!valid) return false;
    if (!normalizeSourceAddresses()) return false;
    if (!normalizeDestinationPorts()) return false;
    return buildPreviewRules();
};

const checkRules = async () => {
    if (!(await prepareRulesFromForm())) return;
    if (mode.value === 'edit' && editingUUID.value) {
        const result = (
            await checkFirewallRules({
                items: [{ uuid: editingUUID.value, rule: previewRules.value[0] }],
            })
        ).data.items[0];
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
.protocol-option {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
}

.protocol-option-description {
    color: var(--el-text-color-secondary);
    font-size: 12px;
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
    &.is-error {
        color: var(--el-color-danger);
    }
}
</style>
