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
                    {{ $t('commons.button.clean') }}
                </el-button>
            </div>
            <div class="port-card-grid">
                <el-card
                    v-for="group in activePortGroups"
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
                                {{ $t('commons.button.clean') }}
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

    <el-dialog v-model="policyVisible" :title="$t('firewall.dockerGuardPolicy')" width="520">
        <el-form label-position="top">
            <el-form-item :label="$t('firewall.protectionMode')">
                <el-radio-group v-model="form.mode">
                    <el-radio value="deny_sources">{{ $t('firewall.denySources') }}</el-radio>
                    <el-radio value="allow_sources">{{ $t('firewall.allowSources') }}</el-radio>
                    <el-radio value="deny_all">{{ $t('firewall.denyAll') }}</el-radio>
                </el-radio-group>
            </el-form-item>
            <el-form-item v-if="form.mode !== 'deny_all'" :label="sourcesLabel">
                <el-input v-model="sourcesText" type="textarea" :rows="4" :placeholder="$t('firewall.sourcesHelper')" />
            </el-form-item>
            <el-form-item :label="$t('commons.table.description')">
                <el-input v-model="form.description" type="textarea" :rows="3" />
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button @click="policyVisible = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button type="primary" @click="submitPolicy">{{ $t('commons.button.confirm') }}</el-button>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import { computed, reactive, ref } from 'vue';
import { Firewall } from '@/api/interface/firewall';
import { deleteDockerPortGuardPolicies, upsertDockerPortGuardPolicies } from '@/api/modules/firewall';
import i18n from '@/lang';
import { checkCidr, checkCidrV6, checkIpV6 } from '@/utils/validate';
import { MsgError, MsgSuccess } from '@/utils/message';
import { ElMessageBox } from 'element-plus';

const props = defineProps<{ containers: Firewall.DockerGuardContainer[] }>();
const emit = defineEmits<{ search: [] }>();

const drawerVisible = ref(false);
const policyVisible = ref(false);
const activeContainerKey = ref('');
const selectedGroupKeys = ref<string[]>([]);
const policyEndpoints = ref<Firewall.DockerGuardEndpoint[]>([]);
const sourcesText = ref('');
const form = reactive<Omit<Firewall.DockerGuardPolicyBatch, 'endpoints'>>({
    mode: 'deny_sources',
    sources: [],
    description: '',
});

const activeContainer = computed(() => props.containers.find((row) => row.key === activeContainerKey.value));
const activePortGroups = computed(() => activeContainer.value?.portGroups || []);
const selectedEndpoints = computed(() =>
    activePortGroups.value
        .filter((group) => selectedGroupKeys.value.includes(group.key))
        .flatMap((group) => group.endpoints),
);
const allSelected = computed(
    () => activePortGroups.value.length > 0 && selectedGroupKeys.value.length === activePortGroups.value.length,
);
const selectionIndeterminate = computed(
    () => selectedGroupKeys.value.length > 0 && selectedGroupKeys.value.length < activePortGroups.value.length,
);
const policyUUIDs = (endpoints: Firewall.DockerGuardEndpoint[]) => [
    ...new Set(endpoints.map((endpoint) => endpoint.policyUUID).filter(Boolean) as string[]),
];
const selectedPolicyUUIDs = computed(() => policyUUIDs(selectedEndpoints.value));
const sourcesLabel = computed(() =>
    form.mode === 'allow_sources' ? i18n.global.t('firewall.allowedSources') : i18n.global.t('firewall.deniedSources'),
);

const acceptParams = (container: Firewall.DockerGuardContainer) => {
    activeContainerKey.value = container.key;
    selectedGroupKeys.value = [];
    drawerVisible.value = true;
};
const changeAllSelection = (checked: boolean) => {
    selectedGroupKeys.value = checked ? activePortGroups.value.map((group) => group.key) : [];
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
    form.mode = first.mode || 'deny_sources';
    form.sources = first.sources || [];
    form.description = endpoints.every((endpoint) => endpoint.description === first.description)
        ? first.description || ''
        : '';
    sourcesText.value = form.sources.join('\n');
    policyVisible.value = true;
};
const validSource = (family: Firewall.DockerGuardEndpoint['family'], value: string) => {
    if (family === 'ipv6') return value.includes('/') ? !checkCidrV6(value) : !checkIpV6(value);
    return !checkCidr(value);
};
const submitPolicy = async () => {
    const sources = [...new Set(sourcesText.value.split(/[\s,]+/).filter(Boolean))];
    if (form.mode === 'deny_sources' && sources.length === 0) {
        MsgError(i18n.global.t('commons.rule.requiredInput'));
        return;
    }
    if (
        form.mode !== 'deny_all' &&
        sources.some((source) => policyEndpoints.value.some((endpoint) => !validSource(endpoint.family, source)))
    ) {
        MsgError(i18n.global.t('commons.rule.ip'));
        return;
    }
    form.sources = form.mode === 'deny_all' ? [] : sources;
    await upsertDockerPortGuardPolicies({
        endpoints: policyEndpoints.value.map(({ family, hostIP, hostPort, protocol }) => ({
            family,
            hostIP,
            hostPort,
            protocol,
        })),
        ...form,
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
            i18n.global.t('commons.button.clean'),
            { type: 'warning' },
        );
    } catch {
        return;
    }
    await deleteDockerPortGuardPolicies({ uuids });
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
        summary = `${i18n.global.t('firewall.deny')}: ${row.sources.join(', ')}`;
    } else if (row.mode === 'allow_sources' && row.sources.length) {
        summary = `${i18n.global.t('firewall.allow')}: ${row.sources.join(', ')}`;
    }
    return row.effective ? summary : `${summary} · ${i18n.global.t('firewall.notEffective')}`;
};

defineExpose({ acceptParams });
</script>

<style lang="scss" scoped>
.port-detail-toolbar {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-bottom: 12px;
}

.port-detail-toolbar :deep(.el-button + .el-button) {
    margin-left: 0;
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
