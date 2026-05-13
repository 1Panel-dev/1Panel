<template>
    <div class="fu-select-rw-switch">
        <div
            v-if="!isWrite"
            class="fu-select-rw-switch__read"
            @click.stop="handleReadClick"
            @dblclick.stop="handleReadDblClick"
        >
            <slot name="read">
                <span class="fu-select-rw-switch__text">{{ displayValue }}</span>
            </slot>
        </div>
        <el-select
            v-else
            ref="selectRef"
            v-bind="$attrs"
            :model-value="modelValue"
            @update:model-value="handleUpdate"
            @change="handleChange"
            @blur="handleBlur"
            @visible-change="handleVisibleChange"
            @click.stop
        >
            <slot></slot>
        </el-select>
    </div>
</template>

<script setup lang="ts">
import { computed, nextTick, ref } from 'vue';

defineOptions({ name: 'FuSelectRwSwitch', inheritAttrs: false });

const props = defineProps({
    modelValue: {
        type: [String, Number, Boolean],
        default: '',
    },
    blurTime: {
        type: Number,
        default: 150,
    },
    writeTrigger: {
        type: String,
        default: 'onClick',
    },
});

const emit = defineEmits(['update:modelValue', 'input', 'blur', 'change']);

const selectRef = ref();
const isWrite = ref(false);
const permissionDisabled = ref(false);

defineExpose({
    setPermissionDisabled: (disabled: boolean) => {
        permissionDisabled.value = disabled;
        if (disabled) {
            isWrite.value = false;
        }
    },
});

const effectiveWriteTrigger = computed(() => {
    return permissionDisabled.value ? 'disabled' : props.writeTrigger;
});

const displayValue = computed(() => {
    return props.modelValue === '' || props.modelValue === undefined || props.modelValue === null
        ? '-'
        : String(props.modelValue);
});

const openWrite = async () => {
    isWrite.value = true;
    await nextTick();
    selectRef.value?.focus?.();
};

const closeWrite = () => {
    window.setTimeout(() => {
        isWrite.value = false;
    }, props.blurTime);
};

const handleReadClick = () => {
    if (effectiveWriteTrigger.value === 'onClick') {
        openWrite();
    }
};

const handleReadDblClick = () => {
    if (effectiveWriteTrigger.value === 'onDblclick') {
        openWrite();
    }
};

const handleUpdate = (value: string | number | boolean) => {
    emit('update:modelValue', value);
    emit('input', value);
};

const handleChange = (value: string | number | boolean) => {
    emit('change', value);
    closeWrite();
};

const handleBlur = (event: FocusEvent) => {
    emit('blur', event);
    closeWrite();
};

const handleVisibleChange = (visible: boolean) => {
    if (!visible) {
        closeWrite();
    }
};
</script>
