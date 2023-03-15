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
            <fu-table v-bind="$attrs" @selection-change="handleSelectionChange">
                <slot></slot>
            </fu-table>
        </div>

        <div class="complex-table__pagination" v-if="$slots.pagination || paginationConfig">
            <slot name="pagination">
                <fu-table-pagination
                    v-model:current-page="paginationConfig.currentPage"
                    v-model:page-size="paginationConfig.pageSize"
                    v-bind="paginationConfig"
                    @change="search"
                />
            </slot>
        </div>
    </div>
</template>
<script setup lang="ts">
import { ref } from 'vue';
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
const condition = ref({});
function search(conditions: any, e: any) {
    if (conditions) {
        condition.value = conditions;
    }
    emit('search', condition.value, e);
}

function handleSelectionChange(row: any) {
    emit('update:selects', row);
}
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
