<template>
    <DrawerPro
        v-model="visible"
        :header="$t(form.id ? 'commons.button.edit' : 'commons.button.create')"
        :resource="$t('aiTools.gpuGroup.groups')"
        size="large"
        :auto-close="false"
        confirm-before-close
        @before-close="close"
    >
        <el-form
            ref="formRef"
            :model="form"
            :rules="rules"
            label-position="top"
            v-loading="initializing || saving"
            :disabled="initializing || saving"
            @submit.prevent
        >
            <h3 class="group-form-heading">{{ $t('home.baseInfo') }}</h3>
            <el-form-item :label="$t('commons.table.name')" prop="name">
                <el-input v-model.trim="form.name" maxlength="128" clearable />
            </el-form-item>
            <el-form-item :label="$t('aiTools.gpuGroup.vendor')" prop="imageType">
                <el-select v-model="form.imageType" :disabled="!!imageType" @change="changeVendor">
                    <el-option
                        v-for="vendor in vendorOptions"
                        :key="vendor.value"
                        :value="vendor.value"
                        :label="`${vendor.label} (${vendor.count})`"
                        :disabled="vendor.count === 0"
                    />
                </el-select>
            </el-form-item>
            <GPUConfig ref="gpuRef" v-model="form.gpu" :image-type="form.imageType" group-editor>
                <template #device-meta="{ device }">
                    <span
                        v-if="groups.some((group) => group.id !== form.id && group.gpu.deviceIDs.includes(device.id))"
                        class="device-existing-groups"
                    >
                        {{ $t('aiTools.gpuGroup.groups') }} ·
                        {{
                            groups
                                .filter((group) => group.id !== form.id && group.gpu.deviceIDs.includes(device.id))
                                .map((group) => group.name)
                                .join(', ')
                        }}
                    </span>
                </template>
            </GPUConfig>
            <div v-if="form.id" class="input-help">{{ $t('aiTools.gpuGroup.snapshotHelp') }}</div>
            <el-form-item :label="$t('commons.table.description')" prop="description">
                <el-input v-model="form.description" type="textarea" :rows="2" maxlength="500" show-word-limit />
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button :disabled="saving" @click="visible = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button type="primary" :loading="saving" :disabled="initializing || gpuRef?.loading" @click="save">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DrawerPro>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue';
import { FormInstance } from 'element-plus';
import { AI } from '@/api/interface/ai';
import { saveGPUGroup } from '@/api/modules/ai';
import GPUConfig from '@/views/ai/components/gpu-config/index.vue';
import DrawerPro from '@/components/drawer-pro/index.vue';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { Rules } from '@/global/form-rules';
import { MsgSuccess } from '@/utils/message';
import i18n from '@/lang';

const props = defineProps<{
    group?: AI.GPUGroup;
    groups: AI.GPUGroup[];
    devices: AI.ModelGPUDevice[];
    imageType?: string;
}>();
const visible = defineModel<boolean>('visible', { required: true });
const emit = defineEmits<{ saved: [name: string] }>();
const { currentNode } = useGlobalStore();
const initializing = ref(true);
const vendorOptions = computed(() =>
    [
        { value: 'nvidia', label: 'NVIDIA' },
        { value: 'intel', label: 'Intel' },
        { value: 'ascend', label: 'Ascend' },
    ].map((vendor) => ({
        ...vendor,
        count: props.devices.filter((device) => device.type === vendor.value && device.available).length,
    })),
);
const saving = ref(false);
const formRef = ref<FormInstance>();
const gpuRef = ref<InstanceType<typeof GPUConfig>>();
const emptyGPU = (): AI.ModelGPUConfig => ({
    enabled: true,
    mode: 'devices',
    deviceIDs: [],
    capabilities: ['gpu', 'utility', 'compute'],
    numa: { mode: 'off', nodes: [] },
});
const form = ref(
    props.group
        ? {
              id: props.group.id,
              name: props.group.name,
              description: props.group.description || '',
              imageType: props.group.imageType,
              gpu: JSON.parse(JSON.stringify(props.group.gpu)) as AI.ModelGPUConfig,
          }
        : {
              id: 0,
              name: '',
              description: '',
              imageType: props.imageType || props.devices.find((device) => device.available)?.type || 'nvidia',
              gpu: emptyGPU(),
          },
);
const rules = {
    name: [
        Rules.requiredInput,
        {
            validator: (_rule: unknown, value: string, callback: (error?: Error) => void) => {
                callback(
                    props.groups.some((group) => group.id !== form.value.id && group.name === value.trim())
                        ? new Error(i18n.global.t('commons.rule.duplicate'))
                        : undefined,
                );
            },
            trigger: 'blur',
        },
    ],
    imageType: [Rules.requiredSelect],
};
const changeVendor = async () => {
    form.value.gpu = emptyGPU();
    await nextTick();
    formRef.value?.clearValidate();
    await gpuRef.value?.init(false);
};
const close = (done: () => void) => {
    if (!saving.value) done();
};
const save = async () => {
    if (initializing.value || saving.value || gpuRef.value?.loading) return;
    const node = currentNode.value;
    if (!(await formRef.value?.validate().catch(() => false))) return;
    if (node !== currentNode.value) return;
    saving.value = true;
    try {
        const name = form.value.name.trim();
        await saveGPUGroup(form.value);
        if (node !== currentNode.value) return;
        visible.value = false;
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        emit('saved', name);
    } finally {
        saving.value = false;
    }
};
onMounted(async () => {
    try {
        await nextTick();
        await gpuRef.value?.init(false);
    } finally {
        initializing.value = false;
    }
});
watch(currentNode, () => {
    visible.value = false;
});
</script>

<style lang="scss" scoped>
.group-form-heading {
    font-size: 14px;
    font-weight: 600;
    margin: 0 0 16px;
}
.device-existing-groups {
    width: 100%;
    font-size: 12px;
    line-height: 20px;
    color: var(--el-text-color-secondary);
}
</style>
