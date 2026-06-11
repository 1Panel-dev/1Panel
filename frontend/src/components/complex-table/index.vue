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
                @select="handleSelect"
                @selection-change="handleSelectionChange"
                :max-height="tableHeight"
                @row-contextmenu="handleRightClick"
                @row-click="handleRowClick"
                :tooltip-options="{
                    placement: 'bottom-start',
                }"
            >
                <slot></slot>
                <template #empty>
                    <slot name="empty"></slot>
                </template>
            </fu-table>
        </div>
        <div class="table-footer-container">
            <div class="footer-left" v-if="slots.footerLeft">
                <el-checkbox v-model="leftSelect" @change="toggleSelection"></el-checkbox>
                <div class="ml-4">
                    <slot name="footerLeft"></slot>
                </div>
            </div>

            <div
                ref="paginationRef"
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
                        :pager-count="responsivePagerCount"
                        :size="isMobile || paginationConfig.small ? 'small' : 'default'"
                        :layout="responsivePaginationLayout"
                    />
                </slot>
            </div>
        </div>
        <ul
            v-if="rightClick.visible"
            class="context-menu"
            ref="menuRef"
            :style="{ top: `${adjustedY}px`, left: `${adjustedX}px` }"
            @click.stop
        >
            <li
                v-for="(btn, index) in visibleRightButtons"
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
import { ref, computed, onMounted, useAttrs } from 'vue';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { hasManagePermissionAccess, hasPermissionAccess } from '@/utils/permission';
const slots = useSlots();
const attrs = useAttrs();
const { isMobile, openMenuTabs } = useGlobalStore();

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
const tableRef = ref();
const tableHeight = ref<number | string>('');
const menuRef = ref<HTMLElement | null>(null);
const paginationRef = ref<HTMLElement | null>(null);
const leftSelect = ref(false);
const paginationWidth = ref(0);
let paginationResizeObserver: ResizeObserver | null = null;
const shiftPressed = ref(false);
const lastSelectedRow = ref<any | null>(null);
const rangeBaseRows = ref<any[]>([]);
let isRangeSelecting = false;

const rightClick = ref({
    visible: false,
    left: 0,
    top: 0,
    currentRow: null,
});
const selectedRows = ref<any[]>([]);
const handleRightClick = (row, column, event) => {
    if (!tableRef.value) return;

    try {
        const selectionColumn = tableRef.value.refElTable.columns.find((col) => col.type === 'selection');
        const isSelectable = selectionColumn?.selectable ? selectionColumn.selectable(row) : true;
        if (!isSelectable) {
            if (!props.rightButtons) return;
            event.preventDefault();
            rightClick.value = {
                visible: true,
                left: event.clientX + 5,
                top: event.clientY,
                currentRow: row,
            };
            document.addEventListener('click', closeRightClick);
            return;
        }
    } catch {}

    if (!selectedRows.value.includes(row)) {
        clearSelects();
        tableRef.value.refElTable.toggleRowSelection(row);
    }
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
        let permissionDisabled = false;
        const permissionOptions = { nodeAdmin: btn.nodeAdmin === true };
        if (btn.permission === true) {
            permissionDisabled = !hasManagePermissionAccess(undefined, permissionOptions);
        } else if (btn.permission !== undefined) {
            permissionDisabled = !hasPermissionAccess(btn.permission, permissionOptions);
        }
        const buttonDisabled =
            typeof btn.disabled === 'function' ? btn.disabled(rightClick.value.currentRow) : btn.disabled;
        return permissionDisabled || buttonDisabled;
    };
});
const visibleRightButtons = computed(() => {
    if (!props.rightButtons) {
        return [];
    }
    return props.rightButtons.filter((btn: any) => {
        if (typeof btn.show === 'function') {
            return btn.show(rightClick.value.currentRow);
        }
        if (typeof btn.show === 'boolean') {
            return btn.show;
        }
        return true;
    });
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
    selectedRows.value = row;
    emit('update:selects', row);
    if (row.length > 0) {
        leftSelect.value = true;
    } else {
        leftSelect.value = false;
        if (!isRangeSelecting) {
            lastSelectedRow.value = null;
            rangeBaseRows.value = [];
        }
    }
}

const getTableData = () => {
    const data = attrs.data;
    return Array.isArray(data) ? data : [];
};

function isRowSelectable(row: any) {
    try {
        const selectionColumn = tableRef.value?.refElTable.columns.find((col) => col.type === 'selection');
        return typeof selectionColumn?.selectable === 'function' ? selectionColumn.selectable(row) : true;
    } catch {
        return true;
    }
}

function updateRowSelection(row: any, selected: boolean) {
    tableRef.value?.refElTable.toggleRowSelection(row, selected);
}

function syncSelection(targetRows: any[]) {
    const currentRows = selectedRows.value;
    const targetSet = new Set(targetRows);
    for (const row of currentRows) {
        if (!targetSet.has(row)) {
            updateRowSelection(row, false);
        }
    }
    for (const row of targetRows) {
        if (!currentRows.includes(row)) {
            updateRowSelection(row, true);
        }
    }
}

function applyRangeSelection(targetRow: any) {
    const table = tableRef.value?.refElTable;
    if (!table || !lastSelectedRow.value) {
        return false;
    }
    const tableData = getTableData();
    const startIndex = tableData.indexOf(lastSelectedRow.value);
    const endIndex = tableData.indexOf(targetRow);
    if (startIndex === -1 || endIndex === -1) {
        return false;
    }
    const [start, end] = [startIndex, endIndex].sort((a, b) => a - b);
    const nextRangeRows = tableData.slice(start, end + 1).filter((row) => isRowSelectable(row));
    const nextSelectionRows = [...rangeBaseRows.value];
    for (const row of nextRangeRows) {
        if (!nextSelectionRows.includes(row)) {
            nextSelectionRows.push(row);
        }
    }
    isRangeSelecting = true;
    try {
        syncSelection(nextSelectionRows);
    } finally {
        isRangeSelecting = false;
    }
    return true;
}

function handleSelect(selection: any[], row: any) {
    if (isRangeSelecting) {
        return;
    }
    if (shiftPressed.value && applyRangeSelection(row)) {
        clearTextSelection();
        return;
    }
    lastSelectedRow.value = row;
    rangeBaseRows.value = selection.filter((item) => item !== row);
    clearTextSelection();
}

function sort(prop: string, order: string) {
    tableRef.value.refElTable.sort(prop, order);
}

function clearSelects() {
    tableRef.value.refElTable.clearSelection();
    lastSelectedRow.value = null;
    rangeBaseRows.value = [];
}

function clearSort() {
    tableRef.value.refElTable.clearSort();
}

function clearTextSelection() {
    const selection = window.getSelection?.();
    if (selection && selection.rangeCount > 0) {
        selection.removeAllRanges();
    }
}

function hasActiveTextSelection() {
    const selection = window.getSelection?.();
    return !!selection && !selection.isCollapsed && selection.toString().trim().length > 0;
}

const updatePaginationWidth = () => {
    paginationWidth.value = paginationRef.value?.clientWidth || 0;
};

const responsivePaginationLayout = computed(() => {
    if (isMobile.value || props.paginationConfig?.small) {
        return 'total, prev, pager, next';
    }
    if (paginationWidth.value < 520) {
        return 'total, prev, pager, next';
    }
    return 'total, sizes, prev, pager, next, jumper';
});

const responsivePagerCount = computed(() => {
    if (isMobile.value || props.paginationConfig?.small || paginationWidth.value < 720) {
        return 5;
    }
    return 7;
});

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

function handleRowClick(row: any, column: any, event: any) {
    if (!tableRef.value) return;
    if (!isRowSelectable(row)) return;

    const target = event.target as HTMLElement;
    if (hasActiveTextSelection() && !event.shiftKey) {
        return;
    }

    if (target.closest('.el-checkbox')) return;
    if (
        target.closest('button') ||
        target.closest('a') ||
        target.closest('input') ||
        target.closest('textarea') ||
        target.closest('[contenteditable="true"]') ||
        target.closest('.el-input') ||
        target.closest('.el-textarea') ||
        target.closest('.el-input-number') ||
        target.closest('.el-date-editor') ||
        target.closest('.el-switch') ||
        target.closest('.el-select') ||
        target.closest('.table-link') ||
        target.closest('.cursor-pointer')
    ) {
        return;
    }
    if (event.shiftKey && applyRangeSelection(row)) {
        clearTextSelection();
        return;
    }
    const selected = !selectedRows.value.includes(row);
    tableRef.value.refElTable.toggleRowSelection(row);
    lastSelectedRow.value = row;
    rangeBaseRows.value = selected ? selectedRows.value : selectedRows.value.filter((item) => item !== row);
    clearTextSelection();
}

defineExpose({
    clearSelects,
    sort,
    clearSort,
    closeRightClick,
});

function calcHeight() {
    let heightDiff = props.heightDiff ?? 320;
    let tabHeight = openMenuTabs.value ? 48 : 0;

    if (props.height) {
        tableHeight.value = props.height - tabHeight;
    } else {
        tableHeight.value = window.innerHeight - heightDiff - tabHeight;
    }
}

const toggleSelection = () => {
    tableRef.value.refElTable.toggleAllSelection();
};

const handleKeyDown = (event: KeyboardEvent) => {
    if (event.key === 'Shift') {
        shiftPressed.value = true;
    }
};

const handleKeyUp = (event: KeyboardEvent) => {
    if (event.key === 'Shift') {
        shiftPressed.value = false;
    }
};

onMounted(() => {
    calcHeight();
    nextTick(() => {
        updatePaginationWidth();
        if (paginationRef.value) {
            paginationResizeObserver = new ResizeObserver(updatePaginationWidth);
            paginationResizeObserver.observe(paginationRef.value);
        }
    });
    window.addEventListener('resize', calcHeight);
    window.addEventListener('keydown', handleKeyDown);
    window.addEventListener('keyup', handleKeyUp);
    watch(
        () => [props.height, props.heightDiff],
        () => {
            calcHeight();
        },
    );
});

onBeforeUnmount(() => {
    window.removeEventListener('resize', calcHeight);
    window.removeEventListener('keydown', handleKeyDown);
    window.removeEventListener('keyup', handleKeyUp);
    paginationResizeObserver?.disconnect();
    paginationResizeObserver = null;
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
.table-footer-container {
    display: flex;
    align-items: center;
    justify-content: space-between;

    .footer-left {
        flex-shrink: 0;
        margin-right: 16px;
        margin-left: 12px;
        display: flex;

        .footer-left-button {
            margin-left: 17px;
            display: flex;
        }
    }
}

.complex-table__pagination {
    flex: 1;
    @include flex-row(flex-end);

    :deep(.el-pagination__sizes .el-select) {
        width: 128px;
        min-width: 128px;
    }

    :deep(.el-pagination--small .el-pagination__sizes .el-select) {
        width: 100px;
        min-width: 100px;
    }
}
</style>
