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
                @row-contextmenu="handleRightClick"
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

        <div
            v-if="rightClick.visible"
            class="custom-context-menu"
            :style="{
                left: rightClick.left + 'px',
                top: rightClick.top + 'px',
            }"
        >
            <el-button
                class="no-border-button"
                v-for="(btn, i) in rightButtons"
                :disabled="disabled(btn)"
                @click="rightButtonClick(btn)"
                :key="i"
                :command="btn"
            >
                {{ btn.label }}
            </el-button>
        </div>
    </div>
</template>
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { GlobalStore } from '@/store';
const slots = useSlots();

defineOptions({ name: 'ComplexTable' });
export interface DropdownProps {
    disabled?: any;
    command?: string | number | object;
    label?: string | number;
    [k: string]: any;
}

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
    rightButtons: {
        type: Array as PropType<DropdownProps[]>,
    },
});
const emit = defineEmits(['search', 'update:selects', 'update:paginationConfig']);
const globalStore = GlobalStore();
const mobile = computed(() => {
    return globalStore.isMobile();
});
const tableRef = ref();
const tableHeight = ref(0);

const rightClick = ref({
    visible: false,
    left: 0,
    top: 0,
    currentRow: null,
});
const handleRightClick = (row, column, event) => {
    if (!props.rightButtons) {
        return;
    }
    event.preventDefault();
    rightClick.value = {
        visible: true,
        left: event.clientX + 5,
        top: event.clientY,
        currentRow: row,
    };
    document.addEventListener('click', closeRightClick);
};
const closeRightClick = () => {
    rightClick.value.visible = false;
    document.removeEventListener('click', closeRightClick);
};
const disabled = computed(() => {
    return function (btn: any) {
        return typeof btn.disabled === 'function' ? btn.disabled(rightClick.value.currentRow) : btn.disabled;
    };
});
function rightButtonClick(btn: any) {
    btn.click(rightClick.value.currentRow);
}

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
.custom-context-menu {
    position: fixed;
    z-index: 9999;
    border: 1px solid #dcdfe6;
    border-radius: 4px;
    padding: 0;
    width: 80px;
}

.no-border-button {
    border: 0;
    border-radius: 4;
    margin: 0 !important;
    width: 100%;
    text-align: left;
}
</style>
