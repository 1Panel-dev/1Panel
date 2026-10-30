<template>
    <details open class="form-section">
        <summary class="form-section-header">
            <span>{{ $t('aiTools.model.gpuAcceleration') }}</span>
            <el-switch
                @click.stop
                @keydown.stop
                v-model="model.enabled"
                :disabled="loading"
                :aria-label="$t('aiTools.model.gpuAcceleration')"
                @change="changeEnabled"
            />
            <el-text v-if="!loading && !hasSupportedGPU" type="info" size="small">
                {{ $t(gpuDevices ? 'aiTools.model.gpuNotDetected' : 'aiTools.model.loadFailed') }}
            </el-text>
        </summary>
        <div v-if="model.enabled || model.error" v-show="!loading" class="form-section-body">
            <el-form-item
                prop="gpu.mode"
                :rules="gpuModeRules"
                :error="model.error"
                :label="model.enabled ? $t('aiTools.model.gpuMode') : undefined"
            >
                <el-radio-group v-if="model.enabled" v-model="model.mode" @change="changeMode">
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
                    <el-select
                        v-model="model.deviceIDs"
                        @change="model.error = undefined"
                        popper-class="model-gpu-device-select"
                        multiple
                        filterable
                        class="w-full"
                    >
                        <el-option
                            v-for="device in gpuDevices || []"
                            :key="device.id"
                            :value="device.id"
                            :label="device.name"
                            :disabled="device.type !== imageType || !device.available"
                        >
                            <div class="flex items-center justify-between gap-2">
                                <span class="truncate">{{ device.name }}</span>
                                <span class="flex shrink-0 items-center gap-1">
                                    <el-tag v-if="device.type !== imageType" size="small" type="warning">
                                        {{ $t('aiTools.model.gpuUnsupported') }}
                                    </el-tag>
                                    <el-tag v-else-if="!device.available" size="small" type="danger">
                                        {{ $t('aiTools.model.gpuMissing') }}
                                    </el-tag>
                                    <el-tag size="small" type="info">
                                        {{ gpuTypeLabels[device.type] || device.type.toUpperCase() }}
                                    </el-tag>
                                </span>
                            </div>
                        </el-option>
                        <el-option
                            v-for="id in missingGPUDevices"
                            :key="id"
                            :value="id"
                            :label="`${id} (${$t('aiTools.model.gpuMissing')})`"
                            disabled
                        />
                    </el-select>
                </el-form-item>
                <el-form-item v-if="imageType === 'nvidia'" label="Capabilities" prop="gpu.capabilities">
                    <el-checkbox-group v-model="model.capabilities">
                        <el-checkbox value="gpu" disabled>gpu</el-checkbox>
                        <el-checkbox value="utility">utility</el-checkbox>
                        <el-checkbox value="compute">compute</el-checkbox>
                    </el-checkbox-group>
                    <span class="input-help">{{ $t('aiTools.model.gpuCapabilitiesHelper') }}</span>
                </el-form-item>
            </template>
            <span v-if="selectingDevices && gpuWarnings.length" class="input-help">{{ gpuWarnings.join('; ') }}</span>
        </div>
    </details>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { AI } from '@/api/interface/ai';
import { loadModelGPUDevices } from '@/api/modules/ai';
import i18n from '@/lang';

const gpuTypeLabels: Record<string, string> = { nvidia: 'NVIDIA', intel: 'Intel', ascend: 'Ascend', amd: 'AMD' };
const props = withDefaults(defineProps<{ imageType?: string }>(), { imageType: 'nvidia' });
const model = defineModel<AI.ModelGPUConfig>({ required: true });
const loading = ref(false);
const gpuWarnings = ref<string[]>([]);
const gpuDevices = ref<AI.ModelGPUDevice[]>();
const hasSupportedGPU = computed(() =>
    Boolean(gpuDevices.value?.some((item) => item.type === props.imageType && item.available)),
);
const selectingDevices = computed(() => model.value.enabled && model.value.mode === 'devices');
const missingGPUDevices = computed(() =>
    model.value.deviceIDs.filter((id) => !(gpuDevices.value || []).some((item) => item.id === id)),
);
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
        validator: (_rule: unknown, value: string[], callback: (error?: Error) => void) => {
            if (selectingDevices.value && !value.length) {
                callback(new Error(i18n.global.t('aiTools.model.gpuSelectRequired')));
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
    if (!model.value.enabled) {
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
    if (create) {
        model.value.enabled = false;
        model.value.mode = 'all';
        model.value.count = undefined;
        model.value.deviceIDs = [];
        model.value.error = undefined;
    }
    try {
        const { data } = await loadModelGPUDevices();
        if (version !== requestVersion) return;
        gpuDevices.value = data.devices || [];
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
@use '@/views/ai/model/components/form-section.scss';
:global(.model-gpu-device-select .el-select-dropdown__item.is-selected::after) {
    display: none;
}
</style>
