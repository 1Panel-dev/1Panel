<template>
    <div class="complex-table">
        <div class="complex-table__header" v-if="$slots.header || header">
            <slot name="header">{{ header }}</slot>
        </div>
        <div v-if="$slots.toolbar" style="margin-bottom: 10px">
            <slot name="toolbar"></slot>
        </div>

        <div class="complex-table__body">
            <fu-table v-bind="$attrs" ref="tableRef" @selection-change="handleSelectionChange">
                <slot></slot>
                <template #empty>
                    <slot name="empty"></slot>
                </template>
            </fu-table>
        </div>

        <div class="complex-table__pagination" v-if="props.paginationConfig">
            <slot name="pagination">
                <el-pagination
                    v-model:current-page="paginationConfig.currentPage"
                    v-model:page-size="paginationConfig.pageSize"
                    :total="paginationConfig.total"
                    :page-sizes="[5, 10, 20, 50, 100]"
                    @size-change="sizeChange"
                    @current-change="currentChange"
                    :small="mobile"
                    :layout="mobile ? 'total, prev, pager, next' : 'total, sizes, prev, pager, next, jumper'"
                />
            </slot>
        </div>
    </div>
</template>
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { GlobalStore } from '@/store';

defineOptions({ name: 'ComplexTable' });
const props = defineProps({
    header: String,
    paginationConfig: {
        type: Object,
        required: false,
        default: () => {},
    },
});
const emit = defineEmits(['search', 'update:selects', 'update:paginationConfig']);

const globalStore = GlobalStore();

const mobile = computed(() => {
    return globalStore.isMobile();
});

const tableRef = ref();

function currentChange() {
    emit('search');
}

function sizeChange() {
    props.paginationConfig.currentPage = 1;
    localStorage.setItem(props.paginationConfig.cacheSizeKey, props.paginationConfig.pageSize);
    emit('search');
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

function clearSort() {
    tableRef.value.refElTable.clearSort();
}

defineExpose({
    clearSelects,
    sort,
    clearSort,
});

onMounted(() => {
    if (props.paginationConfig?.cacheSizeKey) {
        let itemSize = Number(localStorage.getItem(props.paginationConfig.cacheSizeKey));
        if (itemSize) {
            props.paginationConfig.pageSize = itemSize;
        }
    }
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
