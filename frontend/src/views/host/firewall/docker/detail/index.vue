<template>
    <DrawerPro
        v-model="drawerVisible"
        :header="$t('firewall.portDetails')"
        :resource="activeContainer?.name || ''"
        size="large"
    >
        <template #content>
            <div class="port-detail-toolbar">
                <el-checkbox
                    border
                    :model-value="allSelected"
                    :indeterminate="selectionIndeterminate"
                    @change="changeAllSelection"
                >
                    {{ $t('commons.button.selectAll') }}
                </el-checkbox>
                <el-button
                    type="primary"
                    :disabled="selectedEndpoints.length === 0"
                    @click="openPolicy(selectedEndpoints)"
                >
                    {{ $t('commons.button.set') }}
                </el-button>
                <el-button
                    v-permission
                    v-node-admin
                    :disabled="selectedPolicyUUIDs.length === 0"
                    @click="remove(selectedEndpoints, true)"
                >
                    {{ $t('commons.button.delete') }}
                </el-button>
                <el-select
                    v-model="familyFilter"
                    class="port-family-filter"
                    :placeholder="$t('firewall.addressFamily')"
                    @change="changeFamilyFilter"
                >
                    <el-option :label="$t('commons.table.all')" value="all" />
                    <el-option label="IPv4" value="ipv4" />
                    <el-option label="IPv6" value="ipv6" />
                </el-select>
            </div>
            <div class="port-card-grid">
                <el-card
                    v-for="group in filteredPortGroups"
                    :key="group.key"
                    class="port-detail-card"
                    :class="{ 'is-selected': selectedGroupKeys.includes(group.key) }"
                    shadow="never"
                    @click="toggleSelection(group.key)"
                >
                    <div class="port-card-header">
                        <div class="port-card-title">
                            <el-checkbox
                                :model-value="selectedGroupKeys.includes(group.key)"
                                @click.stop
                                @change="(checked: boolean) => changeSelection(group.key, checked)"
                            />
                            <el-tooltip :content="portMappingLabel(group)" placement="top" :show-after="400">
                                <span class="port-mapping">{{ portMappingLabel(group) }}</span>
                            </el-tooltip>
                        </div>
                        <div class="port-card-actions">
                            <el-button type="primary" link size="small" @click.stop="openPolicy(group.endpoints)">
                                {{ $t('commons.button.set') }}
                            </el-button>
                            <el-button
                                v-if="group.endpoint.policyUUID"
                                type="primary"
                                link
                                size="small"
                                @click.stop="remove(group.endpoints, false)"
                            >
                                {{ $t('commons.button.delete') }}
                            </el-button>
                        </div>
                    </div>
                    <div class="port-card-field">
                        <span class="port-card-label">{{ $t('firewall.protection') }}</span>
                        <el-tooltip :content="protectionSummary(group.endpoint)" placement="top" :show-after="400">
                            <span class="port-card-value">{{ protectionSummary(group.endpoint) }}</span>
                        </el-tooltip>
                    </div>
                    <div class="port-card-field">
                        <span class="port-card-label">{{ $t('commons.table.description') }}</span>
                        <el-tooltip :content="group.endpoint.description || '-'" placement="top" :show-after="400">
                            <span class="port-card-value">{{ group.endpoint.description || '-' }}</span>
                        </el-tooltip>
                    </div>
                </el-card>
            </div>
        </template>
    </DrawerPro>

    <DialogPro v-model="policyVisible" :title="$t('firewall.dockerGuardPolicy')" size="large">
        <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
            <el-alert
                v-if="hasMixedFamilies"
                class="mixed-family-alert"
                type="warning"
                :closable="false"
                show-icon
                :title="$t('firewall.dockerGuardMixedFamilyHelper')"
            />
            <el-alert
                v-if="hasInconsistentConfig"
                class="mixed-family-alert"
                type="warning"
                :closable="false"
                show-icon
                :title="$t('firewall.dockerGuardInconsistentConfigHelper')"
            />
            <el-form-item :label="$t('firewall.protectionMode')" prop="mode">
                <el-radio-group v-model="form.mode">
                    <el-radio value="deny_sources" :disabled="hasMixedFamilies">
                        {{ $t('firewall.denySources') }}
                    </el-radio>
                    <el-radio value="allow_sources" :disabled="hasMixedFamilies">
                        {{ $t('firewall.allowSources') }}
                    </el-radio>
                    <el-radio value="deny_all">{{ $t('firewall.denyAll') }}</el-radio>
                </el-radio-group>
            </el-form-item>
            <el-form-item v-if="form.mode && form.mode !== 'deny_all'" :label="sourcesLabel" prop="sources">
                <div v-for="(_, index) of form.sources" :key="index" class="source-address-row mt-2">
                    <el-input
                        ref="sourceAddressRefs"
                        v-model.trim="form.sources[index]"
                        class="source-address-input"
                        clearable
                        :placeholder="$t('firewall.sourceAddressPlaceholder')"
                        @keyup.enter.prevent="addSourceAddressOnEnter(index)"
                    >
                        <template #append>
                            <el-button icon="Delete" @click="removeSourceAddress(index)" />
                        </template>
                    </el-input>
                </div>
                <el-button class="mt-2" @click="addSourceAddress">
                    {{ $t('commons.button.add') }}
                </el-button>
            </el-form-item>
            <el-form-item :label="$t('commons.table.description')">
                <el-input v-model="form.description" type="textarea" :rows="3" />
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button @click="policyVisible = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button type="primary" @click="submitPolicy">{{ $t('commons.button.confirm') }}</el-button>
        </template>
    </DialogPro>
</template>

<script lang="ts" setup>
import { computed, nextTick, reactive, ref } from 'vue';
import { Firewall } from '@/api/interface/firewall';
import { deleteDockerPortGuardPolicies, upsertDockerPortGuardPolicies } from '@/api/modules/firewall';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { ElMessageBox, type FormInstance, type FormRules } from 'element-plus';
import { isValidDockerGuardSource } from '@/views/host/firewall/docker/model';
import { formatHostAddress, formatHostAddressList, splitTagValues } from '@/views/host/firewall/utils/validation';

const props = defineProps<{ containers: Firewall.DockerGuardContainer[] }>();
const emit = defineEmits<{ search: [] }>();

const drawerVisible = ref(false);
const policyVisible = ref(false);
const activeContainerKey = ref('');
const selectedGroupKeys = ref<string[]>([]);
const policyEndpoints = ref<Firewall.DockerGuardEndpoint[]>([]);
const familyFilter = ref<'all' | Firewall.DockerGuardEndpoint['family']>('all');
const formRef = ref<FormInstance>();
const sourceAddressRefs = ref<Array<{ focus: () => void }>>([]);
type PolicyMode = Firewall.DockerGuardPolicyBatch['mode'];
type PolicyForm = Omit<Firewall.DockerGuardPolicyBatch, 'endpoints' | 'mode'> & { mode: PolicyMode | '' };
const form = reactive<PolicyForm>({
    mode: 'deny_sources',
    sources: [''],
    description: '',
});

const activeContainer = computed(() => props.containers.find((row) => row.key === activeContainerKey.value));
const activePortGroups = computed(() => activeContainer.value?.portGroups || []);
const filteredPortGroups = computed(() =>
    familyFilter.value === 'all'
        ? activePortGroups.value
        : activePortGroups.value.filter((group) => group.endpoint.family === familyFilter.value),
);
const selectedEndpoints = computed(() =>
    filteredPortGroups.value
        .filter((group) => selectedGroupKeys.value.includes(group.key))
        .flatMap((group) => group.endpoints),
);
const allSelected = computed(
    () => filteredPortGroups.value.length > 0 && selectedGroupKeys.value.length === filteredPortGroups.value.length,
);
const selectionIndeterminate = computed(
    () => selectedGroupKeys.value.length > 0 && selectedGroupKeys.value.length < filteredPortGroups.value.length,
);
const hasMixedFamilies = computed(() => new Set(policyEndpoints.value.map((endpoint) => endpoint.family)).size > 1);
const normalizeSources = (sources: string[]) =>
    [...new Set(sources.map((source) => source.trim()).filter(Boolean))].sort();
const policyConfigKey = (endpoint: Firewall.DockerGuardEndpoint) =>
    JSON.stringify([endpoint.mode || '', normalizeSources(endpoint.sources || [])]);
const policyConfigConsistent = computed(() => {
    const first = policyEndpoints.value[0];
    return !first || policyEndpoints.value.every((endpoint) => policyConfigKey(endpoint) === policyConfigKey(first));
});
const descriptionConfigConsistent = computed(() => {
    const first = policyEndpoints.value[0];
    return (
        !first ||
        policyEndpoints.value.every(
            (endpoint) => (endpoint.description || '').trim() === (first.description || '').trim(),
        )
    );
});
const hasInconsistentConfig = computed(() => !policyConfigConsistent.value || !descriptionConfigConsistent.value);
const policyUUIDs = (endpoints: Firewall.DockerGuardEndpoint[]) => [
    ...new Set(endpoints.map((endpoint) => endpoint.policyUUID).filter(Boolean) as string[]),
];
const selectedPolicyUUIDs = computed(() => policyUUIDs(selectedEndpoints.value));
const sourcesLabel = computed(() =>
    form.mode === 'allow_sources' ? i18n.global.t('firewall.allowedSources') : i18n.global.t('firewall.deniedSources'),
);
type ValidationCallback = (error?: Error) => void;
const validateSources = (_rule: unknown, value: string[], callback: ValidationCallback) => {
    const sources = splitTagValues(value || []);
    if ((form.mode === 'deny_sources' || form.mode === 'allow_sources') && sources.length === 0) {
        callback(new Error(i18n.global.t('commons.rule.requiredInput')));
        return;
    }
    if (
        sources.some((source) =>
            policyEndpoints.value.some((endpoint) => !isValidDockerGuardSource(endpoint.family, source)),
        )
    ) {
        callback(new Error(i18n.global.t('commons.rule.ip')));
        return;
    }
    callback();
};
const rules = reactive<FormRules>({
    mode: [{ required: true, message: i18n.global.t('commons.rule.requiredSelect'), trigger: 'change' }],
    sources: [{ validator: validateSources, trigger: ['blur', 'change'] }],
});

const acceptParams = (container: Firewall.DockerGuardContainer) => {
    activeContainerKey.value = container.key;
    familyFilter.value = 'all';
    selectedGroupKeys.value = [];
    drawerVisible.value = true;
};
const changeFamilyFilter = () => {
    selectedGroupKeys.value = [];
};
const changeAllSelection = (checked: boolean) => {
    selectedGroupKeys.value = checked ? filteredPortGroups.value.map((group) => group.key) : [];
};
const changeSelection = (key: string, checked: boolean) => {
    selectedGroupKeys.value = checked
        ? [...selectedGroupKeys.value, key]
        : selectedGroupKeys.value.filter((item) => item !== key);
};
const toggleSelection = (key: string) => {
    changeSelection(key, !selectedGroupKeys.value.includes(key));
};
const openPolicy = (endpoints: Firewall.DockerGuardEndpoint[]) => {
    if (!endpoints.length) return;
    policyEndpoints.value = endpoints;
    const first = endpoints[0];
    form.mode = hasMixedFamilies.value ? 'deny_all' : policyConfigConsistent.value ? first.mode || 'deny_sources' : '';
    form.description = descriptionConfigConsistent.value ? (first.description || '').trim() : '';
    form.sources =
        policyConfigConsistent.value && first.sources?.length
            ? first.sources.map((source) => formatHostAddress(source, first.family))
            : [''];
    policyVisible.value = true;
    nextTick(() => formRef.value?.clearValidate());
};
const addSourceAddress = () => {
    form.sources.push('');
    nextTick(() => sourceAddressRefs.value.at(-1)?.focus());
};
const addSourceAddressOnEnter = (index: number) => {
    if (!form.sources[index]?.trim()) return;
    if (index < form.sources.length - 1) {
        sourceAddressRefs.value[index + 1]?.focus();
        return;
    }
    addSourceAddress();
};
const removeSourceAddress = (index: number) => {
    form.sources.splice(index, 1);
    if (form.sources.length === 0) form.sources.push('');
};
const submitPolicy = async () => {
    const valid = await formRef.value?.validate().catch(() => false);
    if (!valid || !form.mode) return;
    const sources = splitTagValues(form.sources);
    await upsertDockerPortGuardPolicies({
        endpoints: policyEndpoints.value.map(({ family, hostIP, hostPort, protocol }) => ({
            family,
            hostIP,
            hostPort,
            protocol,
        })),
        mode: form.mode,
        sources: form.mode === 'deny_all' ? [] : sources,
        description: form.description,
    });
    policyVisible.value = false;
    selectedGroupKeys.value = [];
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    emit('search');
};
const remove = async (endpoints: Firewall.DockerGuardEndpoint[], batch: boolean) => {
    const uuids = policyUUIDs(endpoints);
    if (!uuids.length) return;
    try {
        await ElMessageBox.confirm(
            i18n.global.t(
                batch ? 'firewall.clearDockerGuardPoliciesConfirm' : 'firewall.deleteDockerGuardPolicyConfirm',
            ),
            i18n.global.t('commons.button.delete'),
            { type: 'warning' },
        );
    } catch {
        return;
    }
    for (let offset = 0; offset < uuids.length; offset += 256) {
        await deleteDockerPortGuardPolicies({ uuids: uuids.slice(offset, offset + 256) });
    }
    selectedGroupKeys.value = [];
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    emit('search');
};
const portMappingLabel = (row: Firewall.DockerGuardPortGroup) => {
    const ports = row.endpoints.map((endpoint) => endpoint.containerPort).filter(Boolean) as number[];
    if (!ports.length) return row.label;
    const value = ports.length === 1 ? `${ports[0]}` : `${ports[0]}-${ports[ports.length - 1]}`;
    const publishedEndpoint = row.label.replace(`/${row.endpoint.protocol}`, '');
    return `${publishedEndpoint} → ${value}/${row.endpoint.protocol}`;
};
const protectionSummary = (row: Firewall.DockerGuardEndpoint) => {
    if (!row.policyUUID) return i18n.global.t('firewall.dockerGuardUnprotected');
    let summary = i18n.global.t('firewall.denyAll');
    if (row.mode === 'deny_sources') {
        summary = `${i18n.global.t('firewall.deny')}: ${formatHostAddressList(row.sources, row.family)}`;
    } else if (row.mode === 'allow_sources' && row.sources.length) {
        summary = `${i18n.global.t('firewall.allow')}: ${formatHostAddressList(row.sources, row.family)}`;
    }
    return row.effective ? summary : `${summary} · ${i18n.global.t('firewall.notEffective')}`;
};

defineExpose({ acceptParams, openPolicy });
</script>

<style lang="scss" scoped>
.port-detail-toolbar {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-bottom: 12px;
}

.port-family-filter {
    width: 130px;
    margin-left: auto;
}

.port-detail-toolbar :deep(.el-button + .el-button) {
    margin-left: 0;
}

.mixed-family-alert {
    margin-bottom: 16px;
}

.source-address-row {
    display: flex;
    align-items: center;
    width: 100%;
}

.source-address-input {
    flex: 1;
}

.port-card-grid {
    display: grid;
    grid-template-columns: minmax(0, 1fr);
    gap: 8px;
}

.port-detail-card {
    min-width: 0;
    cursor: pointer;
    font-size: 12px;
    transition: border-color 0.2s;
}

.port-detail-card.is-selected {
    border-color: var(--el-color-primary);
}

.port-detail-card :deep(.el-card__body) {
    padding: 8px 12px;
}

.port-card-header {
    display: flex;
    min-width: 0;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
}

.port-card-title {
    display: flex;
    min-width: 0;
    align-items: center;
    gap: 10px;
}

.port-mapping {
    min-width: 0;
    overflow: hidden;
    color: var(--el-text-color-primary);
    font-weight: 600;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.port-card-field {
    display: flex;
    min-width: 0;
    align-items: center;
    gap: 0;
    margin-top: 5px;
    padding-left: 28px;
}

.port-card-label {
    width: 44px;
    flex: none;
    color: var(--el-text-color-secondary);
}

.port-card-label::after {
    content: ':';
}

.port-card-value {
    min-width: 0;
    overflow: hidden;
    color: var(--el-text-color-regular);
    text-overflow: ellipsis;
    white-space: nowrap;
}

.port-card-actions {
    display: flex;
    flex: none;
    align-items: center;
}
</style>
