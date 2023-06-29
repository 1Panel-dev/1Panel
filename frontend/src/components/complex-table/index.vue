<template>
    <div class="complex-table">
        <div class="complex-table__header" v-if="$slots.header || header">
            <slot name="header">{{ header }}</slot>
        </div>
        <div v-if="$slots.toolbar && !searchConfig" style="margin-bottom: 10px">
            <slot name="toolbar"></slot>
        </div>

        <template v-if="searchConfig">
            <fu-filter-bar v-bind="searchConfig" @exec="search">
                <template #tl>
                    <slot name="toolbar"></slot>
                </template>
                <template #default>
                    <slot name="complex"></slot>
                </template>
                <template #buttons>
                    <slot name="buttons"></slot>
                </template>
            </fu-filter-bar>
        </template>

        <div class="complex-table__body">
            <fu-table v-bind="$attrs" ref="tableRef" @selection-change="handleSelectionChange">
                <slot></slot>
                <template #empty>
                    <slot name="empty"></slot>
                </template>
            </fu-table>
        </div>

        <div class="complex-table__pagination" v-if="$slots.pagination || paginationConfig">
            <slot name="pagination">
                <fu-table-pagination
                    v-model:current-page="paginationConfig.currentPage"
                    v-model:page-size="paginationConfig.pageSize"
                    v-bind="paginationConfig"
                    @change="search"
                    :small="mobile"
                    :layout="mobile ? 'total, prev, pager, next' : 'total, sizes, prev, pager, next, jumper'"
                />
            </slot>
        </div>
    </div>
</template>
<script setup lang="ts">
import { ref, computed } from 'vue';
import { GlobalStore } from '@/store';

defineOptions({ name: 'ComplexTable' });
defineProps({
    header: String,
    searchConfig: Object,
    paginationConfig: {
        type: Object,
        default: () => {},
    },
});
const emit = defineEmits(['search', 'update:selects']);

const globalStore = GlobalStore();

const mobile = computed(() => {
    return globalStore.isMobile();
});

const condition = ref({});
const tableRef = ref();
function search(conditions: any, e: any) {
    if (conditions) {
        condition.value = conditions;
    }
    emit('search', condition.value, e);
}

function handleSelectionChange(row: any) {
    emit('update:selects', row);
}

function sort(prop: string, order: string) {
    tableRef.value.refElTable.sort(prop, order);
}

function clearSelects() {
    tableRef.value.refElTable.clearSelection();
}
defineExpose({
    clearSelects,
    sort,
});
</script>

<style lang="scss">
@use '@/styles/mixins.scss' as *;

.complex-table {
    .complex-table__header {
        @include flex-row(flex-start, center);
        line-height: 60px;
        font-size: 18px;
    }

    .complex-table__toolbar {
        @include flex-row(space-between, center);

        .fu-search-bar {
            width: auto;
        }
    }
    .complex-table__pagination {
        margin-top: 20px;
        @include flex-row(flex-end);
    }
}
</style>
