<template>
    <el-dialog
        :title="title"
        v-model="dialogVisible"
        :destroy-on-close="true"
        :close-on-click-modal="closeOnClickModal"
        :show-close="showClose"
        :width="width"
        :class="{ 'dialog-full-width': props.size === 'full', 'dialog-sized': !!props.size && props.size !== 'full' }"
        append-to-body
        draggable
        @open="open"
        @opened="opened"
        :before-close="handleBeforeClose"
    >
        <template #header v-if="slots.header">
            <slot name="header"></slot>
        </template>
        <div v-if="slots.content">
            <slot name="content"></slot>
        </div>
        <el-row v-else>
            <el-col :span="22" :offset="1">
                <slot></slot>
            </el-col>
        </el-row>

        <template #footer v-if="slots.footer">
            <slot name="footer"></slot>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import { computed, useSlots } from 'vue';
defineOptions({ name: 'DialogPro' });

const props = defineProps({
    title: String,
    showClose: {
        type: Boolean,
        default: true,
    },
    closeOnClickModal: {
        type: Boolean,
        default: false,
    },
    size: {
        type: String,
        default: '',
    },
    modelValue: {
        type: Boolean,
        default: false,
    },
});

// Explicit sizes preserve layouts such as imports; unspecified dialogs use the shared width.
const widths: Record<string, string> = {
    mini: '20%',
    small: '30%',
    normal: '40%',
    large: '50%',
    full: '100%',
    'w-60': '60%',
    'w-70': '70%',
    'w-90': '90%',
};
const width = computed(() => (props.size ? (widths[props.size] ?? '50%') : undefined));

const slots = useSlots();

const emit = defineEmits(['update:modelValue', 'close', 'open', 'opened']);

const dialogVisible = computed({
    get() {
        return props.modelValue;
    },
    set(value: boolean) {
        emit('update:modelValue', value);
    },
});

const handleBeforeClose = () => {
    emit('close');
    dialogVisible.value = false;
};
const open = () => {
    emit('open');
};
const opened = () => {
    emit('opened');
};
</script>
