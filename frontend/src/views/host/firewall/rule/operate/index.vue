<template>
    <DrawerPro
        v-model="drawerVisible"
        :header="$t(mode === 'edit' ? 'firewall.edit' : 'firewall.create')"
        size="large"
        :auto-close="!loading"
        @close="handleClose"
    >
        <el-form ref="formRef" v-loading="loading" label-position="top" :model="form" :rules="rules">
            <el-form-item :label="$t('firewall.action')" prop="action">
                <el-radio-group v-model="form.action" :disabled="descriptionOnly">
                    <el-radio-button value="accept">
                        {{ $t('firewall.accept') }}
                    </el-radio-button>
                    <el-radio-button value="drop">
                        {{ $t('firewall.drop') }}
                    </el-radio-button>
                </el-radio-group>
            </el-form-item>
            <el-form-item :label="$t('commons.table.protocol')" prop="protocol">
                <el-select v-model="form.protocol" :disabled="descriptionOnly" class="w-full" @change="changeProtocol">
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
                        :disabled="descriptionOnly"
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
                        :disabled="descriptionOnly || !portProtocol"
                        :placeholder="$t('firewall.destinationPortPlaceholder')"
                        @keyup.enter.prevent="addDestinationPortOnEnter(index)"
                    >
                        <template #append>
                            <el-button
                                v-if="mode === 'create'"
                                icon="Delete"
                                :disabled="descriptionOnly || !portProtocol"
                                @click="removeRuleRow(index)"
                            />
                        </template>
                    </el-input>
                </div>
                <el-button
                    v-if="mode === 'create'"
                    class="mt-2"
                    :disabled="descriptionOnly || !portProtocol"
                    @click="addRuleRow"
                >
                    {{ $t('commons.button.add') }}
                </el-button>
            </el-form-item>
            <el-form-item v-if="showPriorityField" :label="priorityFieldLabel">
                <el-input-number
                    v-model="form.priority"
                    :disabled="descriptionOnly"
                    :min="priorityMin"
                    :max="priorityMax"
                    controls-position="right"
                />
                <span v-if="!descriptionOnly" class="priority-range">{{ priorityMin }} ~ {{ priorityMax }}</span>
            </el-form-item>
            <el-form-item :label="$t('commons.table.description')">
                <el-input v-model.trim="form.description" clearable />
            </el-form-item>
        </el-form>

        <template #footer>
            <el-button :disabled="loading" @click="drawerVisible = false">
                {{ $t('commons.button.cancel') }}
            </el-button>
            <el-button type="primary" :disabled="loading" @click="onSubmit">
                {{ $t('commons.button.submit') }}
            </el-button>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { Firewall } from '@/api/interface/firewall';
import { createFirewallRules, updateFirewallRule } from '@/api/modules/firewall';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';
import { computed, nextTick, reactive, ref, watch } from 'vue';
import { ElMessageBox, type FormInstance, type FormRules } from 'element-plus';
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
const descriptionOnly = ref(false);
const editingRule = ref<Firewall.Rule>();
const originalFormRule = ref<Firewall.Rule>();
const drawerVisible = ref(false);
const loading = ref(false);
const formRef = ref<FormInstance>();
const sourceAddressRefs = ref<Array<{ focus: () => void }>>([]);
const destinationPortRefs = ref<Array<{ focus: () => void }>>([]);
const previewRules = ref<Firewall.Rule[]>([]);

const positionRanges = ref<Partial<Record<Firewall.Family, Firewall.PositionRange>>>({});
const firewalldPrioritySupported = ref(true);

interface SourceAddressItem {
    family: Firewall.Family;
    address: string;
}

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
const isWildcardAddress = (_family: Firewall.Family, address?: string) => !address?.trim();
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
const priorityMin = computed(() => Math.max(...selectedPositionRanges.value.map((range) => range.min)));
const priorityMax = computed(() => Math.min(...selectedPositionRanges.value.map((range) => range.max)));
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
    form.priority = undefined;
    form.description = '';
    editingUUID.value = '';
    descriptionOnly.value = false;
    editingRule.value = undefined;
    originalFormRule.value = undefined;
    resetBatch();
    formRef.value?.clearValidate();
};

const acceptParams = (
    value: Firewall.Provider,
    item?: Firewall.InventoryItem,
    ranges: Partial<Record<Firewall.Family, Firewall.PositionRange>> = {},
    supportsExplicitPriority = true,
    onlyDescription = false,
) => {
    provider.value = value;
    firewalldPrioritySupported.value = supportsExplicitPriority;
    positionRanges.value = ranges;
    mode.value = item?.desired?.uuid ? 'edit' : 'create';
    resetForm();
    descriptionOnly.value = mode.value === 'edit' && onlyDescription;
    if (mode.value === 'edit' && item?.desired?.uuid) {
        const rule = descriptionOnly.value ? item.desired.rule : item.rule;
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
        form.priority = provider.value === 'firewalld' ? rule.priority : currentPosition || priorityMax.value;
        form.description = rule.description || '';
        originalFormRule.value = buildRule();
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
};

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
            : mode.value === 'edit'
              ? editingRule.value?.scope.family || source.family
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
        protocol: mode.value === 'edit' && provider.value === 'ufw' && protocol === 'tcp/udp' ? 'all' : protocol,
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

const changedRuleFields = (before: Firewall.Rule, after: Firewall.Rule) =>
    editableFieldLabels
        .filter(([field]) => JSON.stringify(before[field] ?? '') !== JSON.stringify(after[field] ?? ''))
        .map(([field]) => field);

const isMetadataOnlyEdit = (before: Firewall.Rule, after: Firewall.Rule) =>
    changedRuleFields(before, after).every((field) => ['description', 'orderIndex', 'priority'].includes(field));

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
    previewRules.value = rules;
    return true;
};

const submitCreateTask = async () => {
    if (!(await prepareRulesFromForm())) return;
    const result = (
        await createFirewallRules({ items: previewRules.value.map((rule) => ({ rule, sourceKind: 'user' })) })
    ).data;
    if (!result.taskID || !result.queued) {
        MsgError(i18n.global.t('commons.msg.operationFailed'));
        return;
    }
    drawerVisible.value = false;
    emit('created', result.taskID);
};

const prepareRulesFromForm = async () => {
    if (!formRef.value) return previewRules.value.length > 0;
    const valid = await formRef.value.validate().catch(() => false);
    if (!valid) return false;
    if (!normalizeSourceAddresses()) return false;
    if (!normalizeDestinationPorts()) return false;
    return buildPreviewRules();
};

const executeEdit = async () => {
    if (!editingUUID.value || previewRules.value.length === 0) return;
    const updatedRule = previewRules.value[0];
    const before = originalFormRule.value || editingRule.value;
    const changed = before ? changedRuleFields(before, updatedRule) : [];
    if (changed.length === 0) {
        drawerVisible.value = false;
        return;
    }
    try {
        await ElMessageBox.confirm(i18n.global.t('firewall.editRuleConfirm'), i18n.global.t('firewall.edit'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'warning',
        });
    } catch {
        return;
    }
    let request: Firewall.UpdateRequest = { rule: updatedRule };
    if (before && isMetadataOnlyEdit(before, updatedRule)) {
        const description = changed.includes('description') ? updatedRule.description || '' : undefined;
        if (changed.includes('priority')) {
            request = { priority: updatedRule.priority!, description };
        } else if (changed.includes('orderIndex')) {
            request = { orderIndex: updatedRule.orderIndex!, description };
        } else {
            request = { description: description || '' };
        }
    }
    await updateFirewallRule(editingUUID.value, request);
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    emit('search');
    drawerVisible.value = false;
};

const onSubmit = async () => {
    if (loading.value) return;
    loading.value = true;
    try {
        if (mode.value === 'create') {
            await submitCreateTask();
        } else if (
            originalFormRule.value &&
            (descriptionOnly.value || isMetadataOnlyEdit(originalFormRule.value, buildRule()))
        ) {
            const rule = descriptionOnly.value
                ? { ...originalFormRule.value, description: form.description }
                : buildRule();
            const changed = changedRuleFields(originalFormRule.value, rule);
            if (changed.some((field) => field === 'priority' || field === 'orderIndex')) {
                const value = rule.priority ?? rule.orderIndex;
                if (
                    typeof value !== 'number' ||
                    !Number.isInteger(value) ||
                    value < priorityMin.value ||
                    value > priorityMax.value
                ) {
                    MsgError(i18n.global.t('commons.rule.numberRange', [priorityMin.value, priorityMax.value]));
                    return;
                }
            }
            previewRules.value = [rule];
            await executeEdit();
        } else if (await prepareRulesFromForm()) {
            await executeEdit();
        }
    } finally {
        loading.value = false;
    }
};

const emit = defineEmits<{
    (event: 'search'): void;
    (event: 'created', taskID: string): void;
}>();

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
</style>
