<template>
    <el-dialog
        :title="title"
        v-model="dialogVisible"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :show-close="showClose"
        :width="size"
        :open="open"
        @opened="opened"
        :before-close="handleBeforeClose"
    >
        <slot name="header"></slot>
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
    size: {
        type: String,
        default: 'normal',
    },
    modelValue: {
        type: Boolean,
        default: false,
    },
});

const slots = useSlots();

const emit = defineEmits(['update:modelValue', 'close', 'open', 'opened']);

const size = computed(() => {
    switch (props.size) {
        case 'mini':
            return '20%';
        case 'small':
            return '30%';
        case 'normal':
            return '40%';
        case 'large':
            return '50%';
        case 'full':
            return '100%';
        case 'w-60':
            return '60%';
        case 'w-70':
            return '70%';
        case 'w-90':
            return '90%';
        default:
            return '50%';
    }
});

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
};
const open = () => {
    emit('open');
};
const opened = () => {
    emit('opened');
};
</script>
