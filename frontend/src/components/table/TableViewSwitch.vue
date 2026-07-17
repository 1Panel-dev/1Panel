<template>
    <el-radio-group v-model="currentViewMode">
        <el-tooltip :content="$t('commons.table.card')" placement="top">
            <el-radio-button value="card">
                <el-icon><Grid /></el-icon>
            </el-radio-button>
        </el-tooltip>
        <el-tooltip :content="$t('commons.table.table')" placement="top">
            <el-radio-button value="table">
                <el-icon><List /></el-icon>
            </el-radio-button>
        </el-tooltip>
    </el-radio-group>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';
import { Grid, List } from '@element-plus/icons-vue';

defineOptions({ name: 'TableViewSwitch' });

type ViewMode = 'table' | 'card';

const props = defineProps({
    modelValue: {
        type: String as PropType<ViewMode>,
        default: 'table',
    },
    storageKey: {
        type: String,
        required: true,
    },
});
const emit = defineEmits(['update:modelValue']);

const getStorageKey = () => `COMPLEX-T-V-${props.storageKey}`;
const getStoredViewMode = (): ViewMode | undefined => {
    if (typeof window === 'undefined') {
        return undefined;
    }
    const mode = localStorage.getItem(getStorageKey());
    return mode === 'card' || mode === 'table' ? mode : undefined;
};
const currentViewMode = ref<ViewMode>(getStoredViewMode() || props.modelValue);

watch(
    () => props.modelValue,
    (mode) => {
        if (mode !== currentViewMode.value) {
            currentViewMode.value = mode;
        }
    },
);
watch(
    currentViewMode,
    (mode) => {
        if (typeof window !== 'undefined') {
            localStorage.setItem(getStorageKey(), mode);
        }
        if (mode !== props.modelValue) {
            emit('update:modelValue', mode);
        }
    },
    { immediate: true },
);
</script>
