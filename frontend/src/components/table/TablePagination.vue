<template>
    <el-pagination
        v-bind="$attrs"
        class="fu-table-pagination"
        :current-page="currentPage"
        :page-size="pageSize"
        :page-sizes="pageSizes"
        :total="total"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
    />
</template>

<script setup lang="ts">
import { type PropType } from 'vue';

defineOptions({ name: 'FuTablePagination' });

const props = defineProps({
    currentPage: {
        type: Number,
        default: 1,
    },
    pageSize: {
        type: Number,
        default: 10,
    },
    pageSizes: {
        type: Array as PropType<number[]>,
        default: () => [10, 20, 50, 100],
    },
    total: {
        type: Number,
        default: 0,
    },
});

const emit = defineEmits(['update:pageSize', 'update:currentPage', 'size-change', 'current-change', 'change']);

const handleSizeChange = (pageSize: number) => {
    emit('update:pageSize', pageSize);
    emit('size-change', pageSize);
    emit('change', { currentPage: props.currentPage, pageSize });
};

const handleCurrentChange = (currentPage: number) => {
    emit('update:currentPage', currentPage);
    emit('current-change', currentPage);
    emit('change', { currentPage, pageSize: props.pageSize });
};
</script>
