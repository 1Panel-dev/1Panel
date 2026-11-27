<template>
    <DialogPro
        :model-value="modelValue"
        :title="title"
        :show-close="!loading"
        @update:model-value="!loading && emit('update:modelValue', $event)"
    >
        <div class="network-change-content">
            <el-alert v-if="$slots.warning" type="warning" show-icon :closable="false">
                <template #title>
                    <slot name="warning" />
                </template>
            </el-alert>
            <el-descriptions :column="1" border class="network-change-details">
                <el-descriptions-item :label="$t('xpack.node.node')">{{ node }}</el-descriptions-item>
                <el-descriptions-item :label="$t('networkManage.interface')">{{ device }}</el-descriptions-item>
                <slot name="details" />
            </el-descriptions>
            <slot />
            <div class="network-change-confirmation">
                <label :for="confirmationId">{{ $t('networkManage.confirmInterface', [device]) }}</label>
                <el-input
                    :id="confirmationId"
                    v-model="confirmation"
                    :placeholder="device"
                    :disabled="loading"
                    autocomplete="off"
                    @keydown.enter.prevent.stop
                />
                <span class="input-help">{{ $t('networkManage.operationRisk') }}</span>
            </div>
        </div>
        <template #footer>
            <el-button :disabled="loading" @click="emit('update:modelValue', false)">
                {{ $t('commons.button.cancel') }}
            </el-button>
            <el-button
                v-permission
                :type="confirmType"
                :loading="loading"
                :disabled="disabled || loading || !device || confirmation !== device"
                @keydown.enter.prevent.stop
                @click="emit('confirm', confirmation)"
            >
                {{ confirmText || $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DialogPro>
</template>

<script setup lang="ts">
import { ref, useId, watch } from 'vue';

const props = withDefaults(
    defineProps<{
        modelValue: boolean;
        title: string;
        node: string;
        device: string;
        loading: boolean;
        disabled: boolean;
        confirmType?: 'primary' | 'danger';
        confirmText?: string;
    }>(),
    { confirmType: 'primary' },
);
const emit = defineEmits<{
    'update:modelValue': [value: boolean];
    confirm: [confirmation: string];
}>();
const confirmation = ref('');
const confirmationId = useId();
watch(
    () => [props.modelValue, props.device],
    () => (confirmation.value = ''),
);
</script>

<style scoped>
.network-change-content {
    display: grid;
    gap: 20px;
}
.network-change-content :deep(.el-alert) {
    align-items: flex-start;
}
.network-change-content :deep(.el-alert__title) {
    display: grid;
    gap: 4px;
    font-weight: normal;
    line-height: 1.6;
}
.network-change-details :deep(.el-descriptions__label) {
    width: 100px;
}
.network-change-details :deep(.el-descriptions__content) {
    overflow-wrap: anywhere;
}
.network-change-confirmation {
    display: grid;
    gap: 8px;
}
.network-change-content :deep(.el-form-item:last-child) {
    margin-bottom: 0;
}
</style>
