<template>
    <div class="complex-table">
        <div class="complex-table__header" v-if="slots.header || header">
            <slot name="header">{{ header }}</slot>
        </div>
        <div v-if="slots.toolbar">
            <slot name="toolbar"></slot>
        </div>

        <div ref="complexTableRef" class="complex-table__body">
            <fu-table
                v-if="currentViewMode === 'table'"
                v-bind="$attrs"
                ref="tableRef"
                @select="handleSelect"
                @selection-change="handleSelectionChange"
                :max-height="tableHeight"
                @row-contextmenu="handleRightClick"
                @row-click="handleRowClick"
                :tooltip-options="resolvedTooltipOptions"
            >
                <slot></slot>
                <template #empty>
                    <slot name="empty"></slot>
                </template>
            </fu-table>
            <template v-if="currentViewMode === 'card'">
                <div v-if="tableData.length" class="complex-table__card-grid">
                    <el-card
                        v-for="(cardRow, cardIndex) in tableData"
                        :key="getRowKey(cardRow, cardIndex)"
                        class="complex-table__card"
                        @click="handleRowClick(cardRow, undefined, $event)"
                        @contextmenu="handleRightClick(cardRow, undefined, $event)"
                    >
                        <div
                            v-if="selectionColumn || cardColumns.name.length || cardColumns.status.length"
                            class="complex-table__card-header"
                        >
                            <el-checkbox
                                v-if="selectionColumn"
                                class="complex-table__card-selection"
                                :model-value="selectedRows.includes(cardRow)"
                                :disabled="!isRowSelectable(cardRow)"
                                @change="toggleCardSelection(cardRow, $event)"
                                @click.stop
                            />
                            <div class="complex-table__card-name">
                                <CardColumnValue
                                    v-for="nameColumn in cardColumns.name"
                                    :key="getColumnKey(nameColumn)"
                                    :column="nameColumn"
                                    :row="cardRow"
                                    :index="cardIndex"
                                />
                            </div>
                            <div class="complex-table__card-status">
                                <CardColumnValue
                                    v-for="statusColumn in cardColumns.status"
                                    :key="getColumnKey(statusColumn)"
                                    :column="statusColumn"
                                    :row="cardRow"
                                    :index="cardIndex"
                                />
                            </div>
                        </div>
                        <div
                            v-if="cardColumns.content.length"
                            class="complex-table__card-content"
                            :style="cardContentStyle"
                        >
                            <div
                                v-for="contentColumn in cardColumns.content"
                                :key="getColumnKey(contentColumn)"
                                class="complex-table__card-item"
                            >
                                <span class="complex-table__card-label">{{ getColumnLabel(contentColumn) }}</span>
                                <strong class="complex-table__card-value">
                                    <CardColumnValue :column="contentColumn" :row="cardRow" :index="cardIndex" />
                                </strong>
                            </div>
                        </div>
                        <div v-if="cardColumns['content-full'].length" class="complex-table__card-full-content">
                            <div
                                v-for="fullContentColumn in cardColumns['content-full']"
                                :key="getColumnKey(fullContentColumn)"
                                class="complex-table__card-full-content-item"
                            >
                                <span class="complex-table__card-label">{{ getColumnLabel(fullContentColumn) }}</span>
                                <div class="complex-table__card-full-content-value">
                                    <CardColumnValue :column="fullContentColumn" :row="cardRow" :index="cardIndex" />
                                </div>
                            </div>
                        </div>
                        <div v-if="cardColumns.description.length" class="complex-table__card-description">
                            <div
                                v-for="descriptionColumn in cardColumns.description"
                                :key="getColumnKey(descriptionColumn)"
                                class="complex-table__card-description-item"
                            >
                                <span>{{ getColumnLabel(descriptionColumn) }}</span>
                                <strong>
                                    <CardColumnValue :column="descriptionColumn" :row="cardRow" :index="cardIndex" />
                                </strong>
                            </div>
                        </div>
                        <div v-if="cardColumns.button.length" class="complex-table__card-buttons">
                            <CardColumnValue
                                v-for="buttonColumn in cardColumns.button"
                                :key="getColumnKey(buttonColumn)"
                                :column="buttonColumn"
                                :row="cardRow"
                                :index="cardIndex"
                            />
                        </div>
                    </el-card>
                </div>
                <slot v-else name="empty">
                    <el-empty />
                </slot>
            </template>
        </div>
        <div class="table-footer-container">
            <div class="footer-left" v-if="slots.footerLeft">
                <el-checkbox v-model="leftSelect" border @change="toggleSelection"></el-checkbox>
                <div>
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
                v-for="(btn, buttonIndex) in visibleRightButtons"
                :key="buttonIndex"
                :class="[{ disabled: isRightButtonDisabled(btn) }, { divided: btn.divided }]"
                @click="!isRightButtonDisabled(btn) && rightButtonClick(btn)"
            >
                {{ btn.label }}
            </li>
        </ul>
    </div>
</template>
<script setup lang="ts">
import { computed, onMounted, ref, useAttrs } from 'vue';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { flattenVNodes, type FuTableOperationButton } from '@/components/table/shared';
import { useCardColumns } from './useCardColumns';
import { useContextMenu } from './useContextMenu';
import { useResponsivePagination } from './useResponsivePagination';
import { useTableSelection } from './useTableSelection';
const slots = useSlots();
const attrs = useAttrs();
const { isMobile, openMenuTabs } = useGlobalStore();

defineOptions({ name: 'ComplexTable' });
export type DropdownProps = FuTableOperationButton;

type ViewMode = 'table' | 'card';

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
        type: Array as PropType<FuTableOperationButton[]>,
    },
    viewMode: {
        type: String as PropType<ViewMode>,
        default: undefined,
    },
    syncCardContentHeight: {
        type: Boolean,
        default: true,
    },
});
const emit = defineEmits(['search', 'update:selects', 'update:paginationConfig', 'update:viewMode']);
const tableRef = ref();
const complexTableRef = ref<HTMLElement>();
const tableHeight = ref<number | string>('');
const menuRef = ref<HTMLElement | null>(null);
const leftSelect = ref(false);
const cardContentMinHeight = ref(0);
let cardContentHeightFrame: number | undefined;

const currentViewMode = ref<ViewMode>(props.viewMode || 'table');
watch(
    () => props.viewMode,
    (mode) => {
        if (mode && mode !== currentViewMode.value) {
            currentViewMode.value = mode;
        }
    },
);
const columnNodes = computed(() => flattenVNodes(slots.default?.() || []));
const { CardColumnValue, cardColumns, getColumnKey, getColumnLabel } = useCardColumns(
    () => columnNodes.value,
    () => attrs.columns,
);
const selectionColumn = computed(() =>
    columnNodes.value.find((column) => (column.props as Record<string, any> | null)?.type === 'selection'),
);
const tableData = computed(() => {
    const data = attrs.data;
    return Array.isArray(data) ? data : [];
});
const cardContentStyle = computed(() => {
    if (!props.syncCardContentHeight || !cardContentMinHeight.value) {
        return undefined;
    }
    return { minHeight: `${cardContentMinHeight.value}px` };
});
const syncCardContentHeight = async () => {
    cardContentMinHeight.value = 0;
    await nextTick();
    if (!props.syncCardContentHeight || currentViewMode.value !== 'card') {
        return;
    }
    const contentElements = complexTableRef.value?.querySelectorAll<HTMLElement>('.complex-table__card-content');
    if (!contentElements?.length) {
        return;
    }
    cardContentMinHeight.value = Math.max(...Array.from(contentElements, (element) => element.offsetHeight));
};
const scheduleCardContentHeightSync = () => {
    if (cardContentHeightFrame !== undefined) {
        cancelAnimationFrame(cardContentHeightFrame);
    }
    cardContentHeightFrame = requestAnimationFrame(() => {
        cardContentHeightFrame = undefined;
        void syncCardContentHeight();
    });
};
const resolvedTooltipOptions = computed(() => ({
    placement: 'bottom-start',
    ...((attrs.tooltipOptions ?? attrs['tooltip-options']) as Record<string, unknown> | undefined),
}));
const getRowKey = (row: any, index: number) => {
    const rowKey = attrs.rowKey ?? attrs['row-key'];
    if (typeof rowKey === 'function') {
        return rowKey(row) ?? index;
    }
    if (typeof rowKey === 'string') {
        const value = rowKey.split('.').reduce((current, key) => current?.[key], row);
        if (value !== undefined && value !== null) {
            return value;
        }
    }
    return row?.id ?? row?.name ?? index;
};

const {
    contextMenu: rightClick,
    open: openContextMenu,
    close: closeRightClick,
    visibleButtons: visibleRightButtons,
    isDisabled: isRightButtonDisabled,
    click: rightButtonClick,
} = useContextMenu(() => props.rightButtons);
const getTableData = () => (Array.isArray(attrs.data) ? attrs.data : []);
const isRowSelectable = (row: any) => {
    const selectable = (selectionColumn.value?.props as Record<string, any> | null)?.selectable;
    return typeof selectable === 'function' ? selectable(row) : true;
};
const {
    selectedRows,
    clearSelects,
    pruneSelection,
    toggleSelection,
    selectRow,
    syncTableSelection,
    handleSelect,
    handleSelectionChange: syncSelectionChange,
    handleRowClick,
    handleKeyDown: handleSelectionKeyDown,
    handleKeyUp,
} = useTableSelection(tableRef, getTableData, (rows) => emit('update:selects', rows), isRowSelectable);
const toggleCardSelection = (row: any, selected: boolean) => {
    selectRow(row, selected);
};
const handleRightClick = (row, column, event) => {
    if (!props.rightButtons?.length) {
        return;
    }
    if (isRowSelectable(row) && !selectedRows.value.includes(row)) {
        clearSelects();
        selectRow(row, true);
    }
    openContextMenu(row, event);
};

function currentChange() {
    emit('search');
}

function sizeChange() {
    props.paginationConfig.currentPage = 1;
    localStorage.setItem(props.paginationConfig.cacheSizeKey, props.paginationConfig.pageSize);
    emit('search');
}

function handleSelectionChange(rows: any[]) {
    syncSelectionChange(rows);
}

watch(selectedRows, (rows) => {
    leftSelect.value = rows.length > 0;
});

watch(
    tableRef,
    async (table) => {
        if (!table || currentViewMode.value !== 'table') {
            return;
        }
        await nextTick();
        syncTableSelection();
    },
    { flush: 'post' },
);

function sort(prop: string, order: string) {
    tableRef.value?.refElTable?.sort(prop, order);
}

function clearSort() {
    tableRef.value?.refElTable?.clearSort();
}

const { paginationRef, responsivePaginationLayout, responsivePagerCount } = useResponsivePagination(
    () => props.paginationConfig,
    isMobile,
);

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

const handleKeyDown = (event: KeyboardEvent) => {
    handleSelectionKeyDown(event);
    if (event.key === 'Escape') {
        closeRightClick();
    }
};

onMounted(() => {
    if (currentViewMode.value !== props.viewMode) {
        emit('update:viewMode', currentViewMode.value);
    }
    scheduleCardContentHeightSync();
    calcHeight();
    window.addEventListener('resize', calcHeight);
    window.addEventListener('resize', scheduleCardContentHeightSync);
    window.addEventListener('keydown', handleKeyDown);
    window.addEventListener('keyup', handleKeyUp);
    watch(
        () => [props.height, props.heightDiff],
        () => {
            calcHeight();
        },
    );
});

watch([currentViewMode, tableData, () => props.syncCardContentHeight], scheduleCardContentHeightSync, {
    flush: 'post',
});

watch(tableData, pruneSelection);

onBeforeUnmount(() => {
    window.removeEventListener('resize', calcHeight);
    window.removeEventListener('resize', scheduleCardContentHeightSync);
    if (cardContentHeightFrame !== undefined) {
        cancelAnimationFrame(cardContentHeightFrame);
    }
    window.removeEventListener('keydown', handleKeyDown);
    window.removeEventListener('keyup', handleKeyUp);
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

    .complex-table__card-grid {
        display: grid;
        grid-template-columns: repeat(3, minmax(0, 1fr));
        gap: 12px;
    }

    .complex-table__card {
        position: relative;
        min-width: 0;
        --el-card-padding: 16px;

        :deep(.el-card__body) {
            display: flex;
            min-height: 248px;
            flex-direction: column;
            padding: 16px;
        }
    }

    .complex-table__card-header {
        display: flex;
        align-items: center;
        justify-content: flex-start;
        gap: 12px;
        min-height: 24px;
    }

    .complex-table__card-name {
        min-width: 0;
        font-size: 15px;
        font-weight: 600;
        line-height: 22px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: normal;
    }

    .complex-table__card-status {
        margin-left: auto;
        flex-shrink: 0;
    }

    .complex-table__card-content {
        display: grid;
        grid-template-columns: repeat(3, minmax(0, 1fr));
        gap: 8px;
        margin-top: 14px;
    }

    .complex-table__card-item {
        min-width: 0;
        padding: 10px;
        border-radius: 6px;
        background: var(--el-fill-color-extra-light);
    }

    .complex-table__card-label {
        display: block;
        margin-bottom: 4px;
        color: var(--el-text-color-secondary);
        font-size: 12px;
    }

    .complex-table__card-value {
        display: block;
        overflow: hidden;
        color: var(--el-text-color-primary);
        font-size: 13px;
        font-weight: 600;
        text-overflow: ellipsis;
        white-space: nowrap;
    }

    .complex-table__card-description {
        display: grid;
        gap: 8px;
        margin-top: 14px;
        margin-bottom: 14px;
    }

    .complex-table__card-full-content {
        display: grid;
        gap: 8px;
        margin-top: 14px;
    }

    .complex-table__card-full-content-item {
        min-width: 0;
        padding: 10px;
        border-radius: 6px;
        background: var(--el-fill-color-extra-light);
    }

    .complex-table__card-full-content-value {
        min-width: 0;
        color: var(--el-text-color-primary);
        font-size: 13px;
        line-height: 22px;

        :deep(.complex-table__card-column-value) {
            display: block;
        }

        :deep(.el-button) {
            margin: 0 6px 6px 0;
        }
    }

    .complex-table__card-description-item {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 12px;
        min-width: 0;

        > span {
            flex: none;
            color: var(--el-text-color-secondary);
            font-size: 12px;
        }

        > strong {
            min-width: 0;
            overflow: hidden;
            color: var(--el-text-color-primary);
            font-size: 13px;
            font-weight: 500;
            text-align: right;
            text-overflow: ellipsis;
            white-space: nowrap;
        }
    }

    .complex-table__card-buttons {
        display: flex;
        align-items: center;
        justify-content: flex-end;
        flex-wrap: wrap;
        gap: 4px 12px;
        align-items: center;
        margin-top: auto;
        padding-top: 14px;
        border-top: 1px solid var(--el-border-color-lighter);

        :deep(.el-button) {
            margin-left: 0;
        }
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

@media (max-width: 1200px) {
    .complex-table .complex-table__card-grid {
        grid-template-columns: repeat(2, minmax(0, 1fr));
    }
}

@media (max-width: 640px) {
    .complex-table .complex-table__card-grid,
    .complex-table .complex-table__card-content {
        grid-template-columns: 1fr;
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
        display: flex;
        align-items: center;

        > .el-checkbox.is-bordered {
            width: 32px;
            margin: 0;
            padding: 0;
            justify-content: center;
        }

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
