<template>
    <div class="source-options" role="group" :aria-label="$t('aiTools.model.modelSource')">
        <el-button
            v-for="source in sources"
            :key="source"
            class="source-option"
            :class="{ 'is-selected': modelValue === source }"
            :aria-pressed="modelValue === source"
            :disabled="disabled"
            @click="modelValue = source"
        >
            <svg-icon :iconName="source === 'modelscope' ? 'p-modelscope' : 'p-huggingface1'" class="source-icon" />
            <span class="source-name">{{ source === 'modelscope' ? 'ModelScope' : 'HuggingFace' }}</span>
        </el-button>
    </div>
</template>

<script setup lang="ts">
import type { ModelDownloader } from '@/api/interface/model-downloader';

defineProps<{ sources: ModelDownloader.Source[]; disabled?: boolean }>();
const modelValue = defineModel<ModelDownloader.Source>({ required: true });
</script>

<style scoped lang="scss">
.source-options {
    display: grid;
    grid-auto-flow: column;
    grid-auto-columns: minmax(0, 1fr);
    gap: 6px;
    width: 100%;
}

.source-option {
    min-width: 0;
    height: auto;
    margin: 0;
    padding: 10px 8px;
    border-radius: 6px;
    text-align: left;
    white-space: normal;

    :deep(> span) {
        width: 100%;
        display: flex;
        align-items: center;
        gap: 6px;
        min-width: 0;
    }

    &.is-selected {
        border-color: var(--el-color-primary);
        background: var(--el-color-primary-light-9);
    }
}

.source-icon {
    width: 18px;
    height: 18px;
    padding: 0;
    flex-shrink: 0;
}

.source-name {
    font-size: 12px;
    font-weight: 600;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
</style>
