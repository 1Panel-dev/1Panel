<template>
    <div>
        <el-card class="cached-tabs">
            <el-radio-group :model-value="activeTab" @change="select">
                <el-radio-button v-for="item in items" :key="item.value" :value="item.value" size="large">
                    {{ item.label }}
                </el-radio-button>
            </el-radio-group>
        </el-card>
        <slot :active="activeTab" />
    </div>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router';
import { useCachedTab } from '@/composables/useCachedTab';

const props = withDefaults(
    defineProps<{
        items: { value: string; label: string }[];
        defaultValue: string;
        cacheKey?: string;
        requestedValue?: string;
    }>(),
    { cacheKey: 'default' },
);
const emit = defineEmits<{ activate: [value: string] }>();
const pagePath = useRoute().path;
const { activeTab, select } = useCachedTab({
    key: () => JSON.stringify([pagePath, props.cacheKey]),
    values: () => props.items.map((item) => item.value),
    defaultValue: () => props.defaultValue,
    requestedValue: () => props.requestedValue,
    onActivate: (value) => emit('activate', value),
});
</script>

<style scoped lang="scss">
.cached-tabs {
    --el-card-padding: 0;

    :deep(.el-radio-button__inner) {
        min-width: 100px;
        height: 100%;
        background-color: var(--panel-button-active) !important;
        box-shadow: none !important;
        outline: none !important;
        border: 2px solid transparent !important;
    }

    :deep(.el-radio-button__original-radio:checked + .el-radio-button__inner) {
        color: $primary-color;
        border-color: $primary-color !important;
        border-radius: 4px;
    }
}
</style>
