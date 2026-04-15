<template>
    <div class="fu-search-bar fu-table-column-select">
        <el-popover
            v-model:visible="visible"
            :trigger="trigger"
            :popper-class="resolvedPopperClass"
            placement="bottom-end"
            :width="260"
        >
            <template #reference>
                <el-button class="fu-search-bar-button" :icon="iconComponent" @click.stop>
                    <span v-if="!onlyIcon">{{ resolvedTitle }}</span>
                </el-button>
            </template>
            <div class="fu-table-column-select__panel" @click.stop>
                <div class="fu-table-column-select__title">{{ resolvedTitle }}</div>
                <div ref="listRef" class="fu-table-column-select__list">
                    <div
                        v-for="column in columns"
                        :key="column.key"
                        class="fu-table-column-select__item"
                        :data-key="column.key"
                    >
                        <el-icon class="fu-table-column-select__drag"><Rank /></el-icon>
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
import Sortable from 'sortablejs';
import { Rank, Setting } from '@element-plus/icons-vue';
import { computed, nextTick, onBeforeUnmount, ref, watch, type PropType } from 'vue';
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
const listRef = ref<HTMLElement | null>(null);
let sortable: Sortable | undefined;

const resolvedTitle = computed(() => props.title || t('fu.table.custom_table_rows'));
const resolvedPopperClass = computed(() =>
    ['fu-table-column-select__popover', props.popperClass].filter(Boolean).join(' '),
);
const iconComponent = computed(() => props.icon || Setting);

const initSortable = async () => {
    await nextTick();
    sortable?.destroy();
    if (!listRef.value) {
        return;
    }
    sortable = Sortable.create(listRef.value, {
        animation: 150,
        handle: '.fu-table-column-select__drag',
        onEnd: ({ oldIndex, newIndex }) => {
            if (oldIndex === undefined || newIndex === undefined || oldIndex === newIndex) {
                return;
            }
            const nextColumns = [...props.columns];
            const [movedColumn] = nextColumns.splice(oldIndex, 1);
            if (!movedColumn) {
                return;
            }
            nextColumns.splice(newIndex, 0, movedColumn);
            props.columns.splice(0, props.columns.length, ...nextColumns);
        },
    });
};

const isFixedColumn = (column: FuTableColumnConfig) => {
    return Boolean(column.fixed);
};

watch(visible, (value) => {
    if (value) {
        initSortable();
        return;
    }
    sortable?.destroy();
    sortable = undefined;
});

onBeforeUnmount(() => {
    sortable?.destroy();
    sortable = undefined;
});
</script>
