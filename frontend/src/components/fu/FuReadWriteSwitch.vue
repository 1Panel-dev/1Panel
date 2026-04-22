<template>
    <div class="fu-read-write-switch">
        <div
            v-if="!isWrite"
            class="fu-read-write-switch__read"
            @click.stop="handleReadClick"
            @dblclick.stop="handleReadDblClick"
        >
            <slot name="read" :write="openWrite">
                <span class="fu-read-write-switch__text">{{ displayValue }}</span>
            </slot>
        </div>
        <div v-else class="fu-read-write-switch__write" @click.stop>
            <slot :read="closeWrite"></slot>
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';

defineOptions({ name: 'FuReadWriteSwitch' });

const props = defineProps({
    modelValue: {
        type: [String, Number, Boolean, Object, Array],
        default: '',
    },
    data: {
        type: [String, Number, Boolean, Object, Array],
        default: '',
    },
    writeTrigger: {
        type: String,
        default: 'onClick',
    },
});

const emit = defineEmits(['update:modelValue', 'change']);

const isWrite = ref(false);

const displayValue = computed(() => {
    const value = props.modelValue !== '' && props.modelValue !== undefined ? props.modelValue : props.data;
    return value === '' || value === undefined || value === null ? '-' : String(value);
});

const openWrite = () => {
    isWrite.value = true;
};

const closeWrite = (value?: any) => {
    if (value !== undefined) {
        emit('update:modelValue', value);
        emit('change', value);
    }
    isWrite.value = false;
};

const handleReadClick = () => {
    if (props.writeTrigger === 'onClick') {
        openWrite();
    }
};

const handleReadDblClick = () => {
    if (props.writeTrigger === 'onDblclick') {
        openWrite();
    }
};

defineExpose({
    write: openWrite,
    read: closeWrite,
});
</script>
