<template>
    <div class="fu-search-bar fu-table-column-select">
        <el-popover
            v-model:visible="visible"
            :trigger="trigger"
            :popper-class="resolvedPopperClass"
            placement="bottom-end"
            :width="popoverWidth"
        >
            <template #reference>
                <el-button class="fu-search-bar-button" :icon="iconComponent" @click.stop>
                    <span v-if="!onlyIcon">{{ resolvedTitle }}</span>
                </el-button>
            </template>
            <div class="fu-table-column-select__panel" @click.stop>
                <div class="fu-table-column-select__title">{{ resolvedTitle }}</div>
                <div class="fu-table-column-select__list">
                    <div v-for="column in columns" :key="column.key" class="fu-table-column-select__item">
                        <el-checkbox v-model="column.show" :disabled="isFixedColumn(column)">
                            {{ column.label }}
                        </el-checkbox>
                    </div>
                </div>
            </div>
        </el-popover>
    </div>
</template>

<script setup lang="ts">
import { Setting } from '@element-plus/icons-vue';
import { computed, ref, type PropType } from 'vue';
import { useI18n } from 'vue-i18n';

import type { FuTableColumnConfig } from './shared';

defineOptions({ name: 'FuTableColumnSelect' });

const { t } = useI18n();

const props = defineProps({
    columns: {
        type: Array as PropType<FuTableColumnConfig[]>,
        default: () => [],
    },
    title: {
        type: String,
        default: '',
    },
    icon: {
        type: [String, Object],
        default: undefined,
    },
    onlyIcon: {
        type: Boolean,
        default: false,
    },
    trigger: {
        type: String,
        default: 'click',
    },
    popperClass: {
        type: String,
        default: '',
    },
});

const visible = ref(false);
const resolvedTitle = computed(() => props.title || t('fu.table.custom_table_rows'));
const resolvedPopperClass = computed(() =>
    ['fu-table-column-select__popover', props.popperClass].filter(Boolean).join(' '),
);
const iconComponent = computed(() => props.icon || Setting);
const estimateTextWidth = (text: string) => {
    return Array.from(text).reduce((total, char) => total + (char.charCodeAt(0) > 255 ? 14 : 8), 0);
};

const popoverWidth = computed(() => {
    const texts = [resolvedTitle.value, ...props.columns.map((column) => String(column.label || ''))];
    const contentWidth = texts.reduce((maxWidth, text) => Math.max(maxWidth, estimateTextWidth(text)), 0);
    return Math.min(Math.max(contentWidth + 72, 180), 320);
});

const isFixedColumn = (column: FuTableColumnConfig) => {
    return Boolean(column.fixed);
};
</script>
