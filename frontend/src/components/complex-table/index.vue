<template>
    <div class="complex-table">
        <div class="complex-table__header" v-if="slots.header || header">
            <slot name="header">{{ header }}</slot>
        </div>
        <div v-if="slots.toolbar">
            <slot name="toolbar"></slot>
        </div>

        <div class="complex-table__body">
            <fu-table
                v-bind="$attrs"
                ref="tableRef"
                @selection-change="handleSelectionChange"
                :max-height="tableHeight"
            >
                <slot></slot>
                <template #empty>
                    <slot name="empty"></slot>
                </template>
            </fu-table>
        </div>

        <div
            class="complex-table__pagination flex items-center w-full sm:flex-row flex-col text-xs sm:text-sm"
            v-if="props.paginationConfig"
            :class="{ '!justify-between': slots.paginationLeft, '!justify-end': !slots.paginationLeft }"
        >
            <slot name="paginationLeft"></slot>
            <slot name="pagination">
                <el-pagination
                    v-model:current-page="paginationConfig.currentPage"
                    v-model:page-size="paginationConfig.pageSize"
                    :total="paginationConfig.total"
                    :page-sizes="[5, 10, 20, 50, 100, 200, 500]"
                    @size-change="sizeChange"
                    @current-change="currentChange"
                    :size="mobile || paginationConfig.small ? 'small' : 'default'"
                    :layout="
                        mobile || paginationConfig.small
                            ? 'total, prev, pager, next'
                            : 'total, sizes, prev, pager, next, jumper'
                    "
                />
            </slot>
        </div>
    </div>
</template>
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { GlobalStore } from '@/store';
const slots = useSlots();

defineOptions({ name: 'ComplexTable' });
const props = defineProps({
    header: String,
    paginationConfig: {
        type: Object,
        required: false,
    },
    heightDiff: {
        type: Number,
        default: 320,
    },
    height: {
        type: Number,
        default: 0,
    },
});
const emit = defineEmits(['search', 'update:selects', 'update:paginationConfig']);
const globalStore = GlobalStore();
const mobile = computed(() => {
    return globalStore.isMobile();
});
const tableRef = ref();
const tableHeight = ref(0);

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
    let heightDiff = 320;
    let tabHeight = 0;
    if (props.heightDiff) {
        heightDiff = props.heightDiff;
    }
    if (globalStore.openMenuTabs) {
        tabHeight = 48;
    }
    if (props.height) {
        tableHeight.value = props.height - tabHeight;
    } else {
        tableHeight.value = window.innerHeight - heightDiff - tabHeight;
    }

    window.onresize = () => {
        return (() => {
            if (props.height) {
                tableHeight.value = props.height - tabHeight;
            } else {
                tableHeight.value = window.innerHeight - heightDiff - tabHeight;
            }
        })();
    };
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

    .complex-table__body {
        margin-top: 10px;
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
