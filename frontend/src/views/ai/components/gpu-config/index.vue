<template>
    <component
        :is="groupEditor ? 'div' : 'details'"
        ref="sectionRef"
        :open="groupEditor || !collapsed"
        class="form-section"
        :class="{ 'group-editor': groupEditor }"
    >
        <summary v-show="!groupEditor" class="form-section-header">
            <span>{{ $t('aiTools.model.gpuAcceleration') }}</span>
            <el-switch
                v-if="!groupEditor"
                @click.stop
                @keydown.stop
                v-model="model.enabled"
                :disabled="loading"
                :aria-label="$t('aiTools.model.gpuAcceleration')"
                @change="changeEnabled"
            />
            <el-text v-if="!loading && !hasSupportedGPU && !model.enabled" type="info" size="small">
                {{ $t(gpuDevices ? 'aiTools.model.gpuNotDetected' : 'aiTools.model.loadFailed') }}
            </el-text>
        </summary>
        <div v-if="model.enabled || model.error" v-show="!loading" class="form-section-body">
            <template v-if="!groupEditor && model.enabled">
                <el-form-item :label="$t('aiTools.gpuGroup.source')">
                    <el-radio-group v-model="selectionMode" @change="changeSource">
                        <el-radio-button value="group">{{ $t('aiTools.gpuGroup.groups') }}</el-radio-button>
                        <el-radio-button value="custom">{{ $t('aiTools.gpuGroup.custom') }}</el-radio-button>
                    </el-radio-group>
                </el-form-item>
                <el-form-item
                    v-if="selectionMode === 'group'"
                    :label="$t('aiTools.gpuGroup.groups')"
                    prop="gpu.group"
                    :rules="groupRules"
                >
                    <el-select :model-value="model.group?.id" class="w-full" @change="applyGroup">
                        <el-option
                            v-for="group in compatibleGroups"
                            :key="group.id"
                            :value="group.id"
                            :label="group.name"
                        />
                        <el-option
                            v-if="model.group && !compatibleGroups.some((group) => group.id === model.group?.id)"
                            :value="model.group.id"
                            :label="model.group.name"
                            disabled
                        />
                    </el-select>
                    <span class="input-help">{{ $t('aiTools.gpuGroup.snapshotHelp') }}</span>
                    <span v-if="groupsFailed" class="input-help">{{ $t('aiTools.model.loadFailed') }}</span>
                    <template v-if="model.group">
                        <div class="input-help">{{ groupSummary }}</div>
                        <el-alert
                            v-if="sourceChanged"
                            :closable="false"
                            type="warning"
                            :title="$t('aiTools.gpuGroup.changed')"
                        />
                        <el-button
                            v-if="sourceChanged && currentGroup"
                            link
                            type="primary"
                            @click="applyGroup(currentGroup.id)"
                        >
                            {{ $t('aiTools.gpuGroup.applyLatest') }}
                        </el-button>
                    </template>
                </el-form-item>
            </template>
            <div v-show="groupEditor || selectionMode === 'custom'">
                <el-form-item
                    v-show="!groupEditor"
                    prop="gpu.mode"
                    :rules="gpuModeRules"
                    :error="model.error"
                    :label="model.enabled ? $t('aiTools.model.gpuMode') : undefined"
                >
                    <el-radio-group v-if="model.enabled && !groupEditor" v-model="model.mode" @change="changeMode">
                        <el-radio-button value="all">
                            {{ $t('aiTools.model.gpuAll') }}
                        </el-radio-button>
                        <el-radio-button v-if="imageType === 'nvidia'" value="count">
                            {{ $t('aiTools.model.gpuCount') }}
                        </el-radio-button>
                        <el-radio-button value="devices">
                            {{ $t('aiTools.model.gpuSpecific') }}
                        </el-radio-button>
                    </el-radio-group>
                    <span v-if="model.enabled && model.mode === 'all'" class="input-help">
                        {{ $t('aiTools.model.gpuAllHelper') }}
                    </span>
                    <span v-if="model.enabled && model.mode === 'count'" class="input-help">
                        {{ $t('aiTools.model.gpuCountHelper') }}
                    </span>
                    <span v-if="model.enabled && model.mode === 'devices'" class="input-help">
                        {{ $t('aiTools.model.gpuSpecificHelper') }}
                    </span>
                </el-form-item>
                <template v-if="model.enabled">
                    <el-form-item
                        v-if="model.mode === 'count'"
                        :label="$t('aiTools.model.gpuCount')"
                        prop="gpu.count"
                        :rules="gpuCountRules"
                    >
                        <el-input-number
                            v-model="model.count"
                            :min="1"
                            :precision="0"
                            :step="1"
                            step-strictly
                            @change="model.error = undefined"
                        />
                    </el-form-item>
                    <el-form-item
                        v-if="model.mode === 'devices'"
                        :label="$t('aiTools.model.gpuDevices')"
                        prop="gpu.deviceIDs"
                        :rules="gpuDeviceRules"
                    >
                        <GPUTopology
                            v-model="model.deviceIDs"
                            :devices="(gpuDevices || []).filter((device) => device.type === imageType)"
                            :nodes="numaNodes"
                            :cards="groupEditor"
                            selectable
                            @update:model-value="model.error = undefined"
                        >
                            <template #device-meta="{ device }"><slot name="device-meta" :device="device" /></template>
                        </GPUTopology>
                        <span class="input-help w-full">
                            {{ $t('aiTools.gpuGroup.selectedGPU', { count: model.deviceIDs.length }) }}
                        </span>
                    </el-form-item>
                    <template v-if="model.numa">
                        <el-form-item
                            :label="$t('aiTools.gpuGroup.affinity')"
                            prop="gpu.numa.mode"
                            :rules="numaRules"
                            :show-message="model.numa.mode !== 'manual'"
                        >
                            <template #label>
                                <span class="gpu-capabilities-label">
                                    {{ $t('aiTools.gpuGroup.affinity') }}
                                    <el-tooltip placement="top">
                                        <template #content>
                                            <div class="gpu-capabilities-help">
                                                <div>{{ $t('aiTools.model.numaHelp') }}</div>
                                                <div class="mt-2">{{ $t('aiTools.model.numaHint') }}</div>
                                            </div>
                                        </template>
                                        <el-icon tabindex="0" :aria-label="$t('aiTools.model.numaHelp')">
                                            <InfoFilled />
                                        </el-icon>
                                    </el-tooltip>
                                </span>
                            </template>
                            <el-radio-group v-model="model.numa.mode">
                                <el-radio-button value="off">{{ $t('aiTools.model.numaOff') }}</el-radio-button>
                                <el-radio-button value="auto" :disabled="model.mode === 'count'">
                                    {{ $t('aiTools.gpuGroup.followGPU') }}
                                </el-radio-button>
                                <el-radio-button value="manual">{{ $t('aiTools.model.numaManual') }}</el-radio-button>
                            </el-radio-group>
                        </el-form-item>
                        <el-form-item
                            v-if="model.numa.mode === 'manual'"
                            :label="$t('aiTools.model.numaNodes')"
                            prop="gpu.numa.nodes"
                            :rules="numaRules"
                        >
                            <el-select v-model="model.numa.nodes" multiple class="w-full">
                                <el-option
                                    v-for="node in numaNodes"
                                    :key="node.id"
                                    :value="node.id"
                                    :label="`NUMA ${node.id} · CPU ${node.cpus || '—'}`"
                                    :disabled="!node.cpus || !node.memory"
                                />
                                <el-option
                                    v-for="id in missingNUMANodes"
                                    :key="id"
                                    :value="id"
                                    :label="`NUMA ${id} (${$t('aiTools.model.numaUnavailable')})`"
                                    disabled
                                />
                            </el-select>
                        </el-form-item>
                    </template>
                </template>
            </div>
            <p v-if="!groupEditor && model.enabled && model.numa && model.numa.mode !== 'off'" class="input-help">
                {{ $t('aiTools.model.numaRequirements') }}
            </p>
            <el-form-item
                v-if="!groupEditor && model.enabled && imageType === 'nvidia'"
                label="Capabilities"
                prop="gpu.capabilities"
            >
                <template #label>
                    <span class="gpu-capabilities-label">
                        Capabilities
                        <el-tooltip placement="top">
                            <template #content>
                                <div class="gpu-capabilities-help">
                                    {{ $t('aiTools.model.gpuCapabilitiesHelper') }}
                                </div>
                            </template>
                            <el-icon tabindex="0" :aria-label="$t('aiTools.model.gpuCapabilitiesHelper')">
                                <InfoFilled />
                            </el-icon>
                        </el-tooltip>
                    </span>
                </template>
                <el-checkbox-group v-model="model.capabilities">
                    <el-checkbox value="gpu" disabled>gpu</el-checkbox>
                    <el-checkbox value="utility">utility</el-checkbox>
                    <el-checkbox value="compute">compute</el-checkbox>
                </el-checkbox-group>
            </el-form-item>
            <span v-if="selectingDevices && gpuWarnings.length" class="input-help">{{ gpuWarnings.join('; ') }}</span>
        </div>
    </component>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { InfoFilled } from '@element-plus/icons-vue';
import GPUTopology from '@/views/ai/components/gpu-topology/index.vue';
import { AI } from '@/api/interface/ai';
import { loadModelGPUDevices, listGPUGroups } from '@/api/modules/ai';
import i18n from '@/lang';

const props = withDefaults(defineProps<{ imageType?: string; collapsed?: boolean; groupEditor?: boolean }>(), {
    imageType: 'nvidia',
    collapsed: false,
    groupEditor: false,
});
const model = defineModel<AI.ModelGPUConfig>({ required: true });
const sectionRef = ref<HTMLDetailsElement>();
const loading = ref(false);
const gpuWarnings = ref<string[]>([]);
const gpuDevices = ref<AI.ModelGPUDevice[]>();
const numaNodes = ref<AI.ModelNUMANode[]>([]);
const groups = ref<AI.GPUGroup[]>([]);
const groupsFailed = ref(false);
const selectionMode = ref<'group' | 'custom'>('custom');
const compatibleGroups = computed(() => groups.value.filter((group) => group.imageType === props.imageType));
const currentGroup = computed(() => compatibleGroups.value.find((group) => group.id === model.value.group?.id));
const sourceChanged = computed(
    () =>
        !groupsFailed.value &&
        model.value.group &&
        (!currentGroup.value || currentGroup.value.revision !== model.value.group.revision),
);
const groupSummary = computed(() => {
    const devices = model.value.deviceIDs
        .map((id) => gpuDevices.value?.find((device) => device.id === id)?.name || id)
        .join(', ');
    const config = model.value.numa;
    const nodes =
        config?.mode === 'auto'
            ? [...new Set(selectedNUMAGPUs.value.map((device) => device.numaNode).filter((id) => id !== undefined))]
            : config?.nodes || [];
    return (
        devices +
        ' · ' +
        (config?.mode && config.mode !== 'off'
            ? nodes
                  .map((id) => `NUMA ${id} · CPU ${numaNodes.value.find((node) => node.id === id)?.cpus || '—'}`)
                  .join(', ')
            : i18n.global.t('aiTools.model.numaOff'))
    );
});
const groupRules = [
    {
        validator: (_rule: unknown, _value: unknown, callback: (error?: Error) => void) => {
            let error = '';
            if (model.value.enabled && selectionMode.value === 'group') {
                if (!model.value.group) error = i18n.global.t('aiTools.gpuGroup.selectRequired');
                else if (numaError.value) error = numaError.value;
                else if (!gpuDevices.value) error = i18n.global.t('aiTools.model.loadFailed');
                else if (
                    model.value.deviceIDs.some(
                        (id) =>
                            !gpuDevices.value?.some(
                                (device) => device.id === id && device.available && device.type === props.imageType,
                            ),
                    )
                )
                    error = i18n.global.t('aiTools.model.gpuMissing');
                else if (model.value.error) error = model.value.error;
            }
            callback(error ? new Error(error) : undefined);
        },
        trigger: 'change',
    },
];
const changeSource = () => {
    if (selectionMode.value === 'custom') model.value.group = undefined;
};
const applyGroup = (id: number) => {
    const group = compatibleGroups.value.find((item) => item.id === id);
    if (!group) return;
    model.value = {
        ...JSON.parse(JSON.stringify(group.gpu)),
        capabilities: model.value.capabilities,
        group: { id: group.id, name: group.name, revision: group.revision },
    };
};
const selectedNUMAGPUs = computed(() =>
    (gpuDevices.value || []).filter(
        (device) =>
            device.type === props.imageType &&
            (model.value.mode === 'all' || model.value.deviceIDs.includes(device.id)),
    ),
);
const missingNUMANodes = computed(() =>
    (model.value.numa?.nodes || []).filter((id) => !numaNodes.value.some((node) => node.id === id)),
);
const numaError = computed(() => {
    const config = model.value.numa;
    if (!model.value.enabled || !config || config.mode === 'off') return '';
    let ids = config.nodes;
    if (config.mode === 'auto') {
        if (model.value.mode === 'count') return i18n.global.t('aiTools.model.numaCount');
        const devices = selectedNUMAGPUs.value;
        if (
            !devices.length ||
            (model.value.mode === 'devices' && devices.length !== model.value.deviceIDs.length) ||
            devices.some((device) => !device.available || device.numaNode === undefined)
        ) {
            return i18n.global.t('aiTools.model.numaUnknown');
        }
        ids = [...new Set(devices.map((device) => device.numaNode!))];
        if (ids.length !== 1) return i18n.global.t('aiTools.model.numaCross');
    }
    if (!ids.length) return i18n.global.t('aiTools.model.numaRequired');
    if (ids.some((id) => !numaNodes.value.some((node) => node.id === id && node.cpus && node.memory))) {
        return i18n.global.t('aiTools.model.numaUnavailable');
    }
    return '';
});
const numaRules = [
    {
        validator: (_rule: unknown, _value: unknown, callback: (error?: Error) => void) =>
            callback(numaError.value ? new Error(numaError.value) : undefined),
        trigger: 'change',
    },
];
const hasSupportedGPU = computed(() =>
    Boolean(gpuDevices.value?.some((item) => item.type === props.imageType && item.available)),
);
const selectingDevices = computed(() => model.value.enabled && model.value.mode === 'devices');
const gpuModeRules = [
    {
        validator: (_rule: unknown, _value: string, callback: (error?: Error) => void) => {
            if (model.value.error) {
                callback(new Error(model.value.error));
            } else if (model.value.enabled && !gpuDevices.value) {
                callback(new Error(i18n.global.t('aiTools.model.loadFailed')));
            } else if (model.value.enabled && !hasSupportedGPU.value) {
                callback(new Error(i18n.global.t('aiTools.model.gpuNotDetected')));
            } else if (
                model.value.enabled &&
                (!['all', 'count', 'devices'].includes(model.value.mode) ||
                    (model.value.mode === 'count' && props.imageType !== 'nvidia'))
            ) {
                callback(new Error(i18n.global.t('aiTools.model.gpuSelectRequired')));
            } else {
                callback();
            }
        },
        trigger: 'change',
    },
];
const gpuCountRules = [
    {
        validator: (_rule: unknown, value: number, callback: (error?: Error) => void) => {
            const availableCount =
                gpuDevices.value?.filter((item) => item.type === props.imageType && item.available).length || 0;
            if (model.value.enabled && model.value.mode === 'count' && (!Number.isInteger(value) || value < 1)) {
                callback(new Error(i18n.global.t('aiTools.model.gpuCountRequired')));
            } else if (
                model.value.enabled &&
                model.value.mode === 'count' &&
                availableCount > 0 &&
                value > availableCount
            ) {
                callback(new Error(i18n.global.t('commons.rule.numberRange', [1, availableCount])));
            } else {
                callback();
            }
        },
        trigger: 'change',
    },
];
const gpuDeviceRules = [
    {
        required: true,
        validator: (_rule: unknown, value: string[], callback: (error?: Error) => void) => {
            if (selectingDevices.value && !value.length) {
                callback(new Error(i18n.global.t('aiTools.model.gpuSelectRequired')));
            } else if (
                selectingDevices.value &&
                value.some(
                    (id) =>
                        !gpuDevices.value?.some(
                            (device) => device.id === id && device.type === props.imageType && device.available,
                        ),
                )
            ) {
                callback(new Error(i18n.global.t('aiTools.model.gpuMissing')));
            } else {
                callback();
            }
        },
        trigger: 'change',
    },
];

const changeMode = () => {
    model.value.error = undefined;
    model.value.deviceIDs = [];
    model.value.count = model.value.mode === 'count' ? 1 : undefined;
};

const changeEnabled = () => {
    model.value.error = undefined;
    if (model.value.enabled && sectionRef.value) {
        sectionRef.value.open = true;
    }
    if (!model.value.enabled) {
        model.value.group = undefined;
        selectionMode.value = 'custom';
        model.value.numa = { mode: 'off', nodes: [] };
        model.value.mode = 'all';
        model.value.count = undefined;
        model.value.deviceIDs = [];
    }
};

let requestVersion = 0;
const init = async (create: boolean) => {
    const version = ++requestVersion;
    loading.value = true;

    gpuDevices.value = undefined;
    gpuWarnings.value = [];
    numaNodes.value = [];
    model.value.numa ??= { mode: 'off', nodes: [] };
    if (create) {
        model.value.group = undefined;
        model.value.enabled = false;
        model.value.numa = { mode: 'off', nodes: [] };
        model.value.mode = 'all';
        model.value.count = undefined;
        model.value.deviceIDs = [];
        model.value.error = undefined;
    }
    selectionMode.value = model.value.group ? 'group' : 'custom';
    groups.value = [];
    groupsFailed.value = false;
    try {
        const [devicesResult, groupsResult] = await Promise.allSettled([
            loadModelGPUDevices(),
            props.groupEditor ? Promise.resolve({ data: [] as AI.GPUGroup[] }) : listGPUGroups(),
        ]);
        if (version !== requestVersion) return;
        if (groupsResult.status === 'fulfilled') groups.value = groupsResult.value.data || [];
        else groupsFailed.value = true;
        if (create && compatibleGroups.value.length) selectionMode.value = 'group';
        if (devicesResult.status === 'rejected') throw devicesResult.reason;
        const { data } = devicesResult.value;
        if (version !== requestVersion) return;
        gpuDevices.value = data.devices || [];
        if (props.imageType === 'nvidia' && model.value.mode === 'devices') {
            model.value.deviceIDs = [
                ...new Set(
                    model.value.deviceIDs.map(
                        (id) =>
                            gpuDevices.value.find(
                                (device) =>
                                    device.type === 'nvidia' &&
                                    (device.id === id ||
                                        (device.uuid && `nvidia:${device.uuid}` === id) ||
                                        (device.index !== undefined && `nvidia:${device.index}` === id)),
                            )?.id || id,
                    ),
                ),
            ];
        }
        numaNodes.value = data.numaNodes || [];
        gpuWarnings.value = data.warnings || [];
        if (create) model.value.enabled = hasSupportedGPU.value;
    } catch {
    } finally {
        if (version === requestVersion) {
            loading.value = false;
        }
    }
};

defineExpose({ init, loading });
</script>

<style lang="scss" scoped>
@use '@/components/form-section/index.scss';
.group-editor {
    border: 0;
    background: transparent;
}
.group-editor > .form-section-body {
    padding: 0;
    border: 0;
}
.group-editor .input-help {
    width: 100%;
}
:deep(.el-form-item__error) {
    position: static;
    width: 100%;
    padding-top: 4px;
    line-height: 1.5;
}
.gpu-capabilities-label {
    display: inline-flex;
    align-items: flex-start;
    gap: 4px;

    .el-icon {
        font-size: 12px;
        color: var(--el-text-color-secondary);
        cursor: help;
    }
}
.gpu-capabilities-help {
    max-width: 360px;
    white-space: pre-line;
}
</style>
