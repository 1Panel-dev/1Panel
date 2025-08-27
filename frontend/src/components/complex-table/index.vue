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
                @row-click="handleRowClick"
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

        <ul
            v-if="rightClick.visible"
            class="context-menu"
            ref="menuRef"
            :style="{ top: `${adjustedY}px`, left: `${adjustedX}px` }"
            @click.stop
        >
            <li
                v-for="(btn, index) in rightButtons"
                :key="index"
                :class="[{ disabled: disabled(btn) }, { divided: btn.divided }]"
                @click="!disabled(btn) && rightButtonClick(btn)"
            >
                {{ btn.label }}
            </li>
        </ul>
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
const menuRef = ref<HTMLElement | null>(null);

const rightClick = ref({
    visible: false,
    left: 0,
    top: 0,
    currentRow: null,
});
const handleRightClick = (row, column, event) => {
    clearSelects();
    tableRef.value.refElTable.toggleRowSelection(row);
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
    clearSelects();
    document.removeEventListener('click', closeRightClick);
};
const disabled = computed(() => {
    return function (btn: any) {
        return typeof btn.disabled === 'function' ? btn.disabled(rightClick.value.currentRow) : btn.disabled;
    };
});
function rightButtonClick(btn: any) {
    closeRightClick();
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

const adjustedX = ref(rightClick.value.left);
const adjustedY = ref(rightClick.value.top);

watch(
    () => [rightClick.value.left, rightClick.value.top],
    async () => {
        await nextTick();
        if (!menuRef.value) return;

        const menuRect = menuRef.value.getBoundingClientRect();
        const windowWidth = window.innerWidth;
        const windowHeight = window.innerHeight;

        if (rightClick.value.left + menuRect.width > windowWidth) {
            adjustedX.value = windowWidth - menuRect.width - 4;
        } else {
            adjustedX.value = rightClick.value.left;
        }

        if (rightClick.value.top + menuRect.height > windowHeight) {
            adjustedY.value = windowHeight - menuRect.height - 4;
        } else {
            adjustedY.value = rightClick.value.top;
        }
    },
    { immediate: true },
);

function handleRowClick(row: any, column: any, event: MouseEvent) {
    if (!tableRef.value) return;
    const target = event.target as HTMLElement;

    if (target.closest('.el-checkbox')) return;
    if (
        target.closest('button') ||
        target.closest('a') ||
        target.closest('.el-switch') ||
        target.closest('.table-link') ||
        target.closest('.cursor-pointer')
    ) {
        return;
    }
    tableRef.value.refElTable.toggleRowSelection(row);
}

defineExpose({
    clearSelects,
    sort,
    clearSort,
    closeRightClick,
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

<style scoped lang="scss">
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
.context-menu {
    position: fixed;
    background: var(--panel-main-bg-color-9);
    border: 1px solid var(--el-border-color);
    border-radius: 4px;
    color: var(--el-color-primary);
    box-shadow: 0 2px 6px rgba(0, 0, 0, 0.15);
    list-style: none;
    font-size: 14px;
    padding: 4px 0;
    margin: 0;
    z-index: 9999;
    min-width: 120px;
}
.context-menu li {
    padding: 6px 12px;
    cursor: pointer;
}
.context-menu li:hover {
    background-color: var(--panel-menu-bg-color);
}
.context-menu li.disabled {
    color: var(--el-border-color);
    cursor: not-allowed;
}
.context-menu li.divided {
    border-top: 1px solid var(--el-border-color);
}
</style>
