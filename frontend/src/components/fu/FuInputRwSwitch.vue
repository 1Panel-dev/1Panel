<template>
    <div class="fu-input-rw-switch">
        <div
            v-if="!isWrite"
            class="fu-input-rw-switch__read"
            @click.stop="handleReadClick"
            @dblclick.stop="handleReadDblClick"
        >
            <span class="fu-input-rw-switch__text">{{ displayValue }}</span>
        </div>
        <el-input
            v-else
            ref="inputRef"
            v-bind="$attrs"
            :model-value="modelValue"
            @update:model-value="handleUpdate"
            @input="handleInput"
            @blur="handleBlur"
            @keyup.enter="handleEnter"
            @click.stop
        />
    </div>
</template>

<script setup lang="ts">
import { computed, nextTick, ref } from 'vue';

defineOptions({ name: 'FuInputRwSwitch', inheritAttrs: false });

const props = defineProps({
    modelValue: {
        type: [String, Number],
        default: '',
    },
    writeTrigger: {
        type: String,
        default: 'onClick',
    },
});

const emit = defineEmits(['update:modelValue', 'input', 'blur', 'enter']);

const inputRef = ref();
const isWrite = ref(false);
const permissionDisabled = ref(false);

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
    inputRef.value?.focus?.();
    inputRef.value?.select?.();
};

const closeWrite = () => {
    isWrite.value = false;
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

const handleUpdate = (value: string | number) => {
    emit('update:modelValue', value);
};

const handleInput = (value: string) => {
    emit('input', value);
};

const handleBlur = (event: FocusEvent) => {
    emit('blur', event);
    closeWrite();
};

const handleEnter = (event: KeyboardEvent) => {
    emit('enter', event);
    closeWrite();
};

defineExpose({
    setPermissionDisabled: (disabled: boolean) => {
        permissionDisabled.value = disabled;
        if (disabled) {
            isWrite.value = false;
        }
    },
    write: openWrite,
    read: closeWrite,
});
</script>
