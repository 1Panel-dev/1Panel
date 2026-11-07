<template>
    <div class="ai-model-list" :class="{ 'ai-model-list--single': visibleCount === 1 }">
        <span v-for="(model, index) in models.slice(0, visibleCount)" :key="index" class="model-preview">
            <el-tag size="small" :type="tagType" class="model-tag">
                <span :title="model">{{ model }}</span>
            </el-tag>
            <CopyButton v-if="copyable" :content="model" />
        </span>
        <el-popover
            v-if="models.length > visibleCount"
            trigger="click"
            placement="bottom-start"
            :width="360"
            :popper-options="{ modifiers: [{ name: 'preventOverflow', options: { padding: 16, altAxis: true } }] }"
            :popper-style="{ maxWidth: 'calc(100vw - 32px)', padding: '12px' }"
        >
            <template #reference>
                <el-button
                    size="small"
                    class="model-more"
                    :aria-label="$t('aiTools.model.model') + ' +' + (models.length - visibleCount)"
                    @click.stop
                >
                    +{{ models.length - visibleCount }}
                </el-button>
            </template>
            <div class="model-popover" @click.stop>
                <div class="model-popover__header">
                    <span>{{ $t('aiTools.model.model') }}</span>
                    <el-tag size="small" type="info" round>{{ models.length }}</el-tag>
                </div>
                <ul class="model-popover__list" tabindex="0" :aria-label="$t('aiTools.model.model')">
                    <li v-for="(model, index) in models" :key="index">
                        <span class="model-popover__name">{{ model }}</span>
                        <CopyButton v-if="copyable" :content="model" :aria-label="$t('commons.button.copy')" />
                    </li>
                </ul>
                <div v-if="$slots.footer" class="model-popover__footer"><slot name="footer" /></div>
            </div>
        </el-popover>
    </div>
</template>

<script setup lang="ts">
import CopyButton from '@/components/copy-button/index.vue';

withDefaults(
    defineProps<{
        models: string[];
        visibleCount?: number;
        copyable?: boolean;
        tagType?: 'primary' | 'success' | 'warning' | 'info' | 'danger';
    }>(),
    { visibleCount: 3, copyable: false, tagType: 'primary' },
);
</script>

<style scoped lang="scss">
.ai-model-list {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 6px;
    min-width: 0;
    max-width: 100%;

    &--single {
        flex-wrap: nowrap;
    }
}
.model-preview {
    display: inline-flex;
    align-items: center;
    min-width: 0;
    max-width: 100%;
}
.model-tag {
    min-width: 0;
    max-width: 100%;
    :deep(.el-tag__content) {
        min-width: 0;
        overflow: hidden;
        text-overflow: ellipsis;
    }
}
.model-more {
    flex-shrink: 0;
    margin: 0;
    padding: 0 8px;
    color: var(--el-color-primary);
    border-color: var(--el-color-primary-light-7);
    background: var(--el-color-primary-light-9);
}
.model-popover {
    &__header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
        padding-bottom: 10px;
        font-weight: 500;
        border-bottom: 1px solid var(--el-border-color-lighter);
    }
    &__list {
        max-height: min(320px, 50vh);
        overflow-y: auto;
        overscroll-behavior: contain;
        margin: 8px 0 0;
        padding: 0;
        list-style: none;
        li {
            display: flex;
            align-items: flex-start;
            gap: 8px;
            padding: 8px;
            line-height: 20px;
            border-radius: 4px;
            &:nth-child(odd) {
                background: var(--el-fill-color-light);
            }
        }
    }
    &__name {
        flex: 1;
        min-width: 0;
        white-space: normal;
        overflow-wrap: anywhere;
    }
    &__footer {
        margin-top: 8px;
        padding-top: 10px;
        border-top: 1px solid var(--el-border-color-lighter);
    }
    :deep(.el-button) {
        flex-shrink: 0;
    }
}
</style>
