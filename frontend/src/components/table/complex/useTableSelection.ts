import { shallowRef, ref, toRaw, type Ref } from 'vue';

type TableRef = { refElTable?: any } | undefined;

export const useTableSelection = (
    tableRef: Ref<TableRef>,
    getTableData: () => any[],
    onSelectionChange: (rows: any[]) => void,
    isRowSelectable: (row: any) => boolean = () => true,
) => {
    const selectedRows = shallowRef<any[]>([]);
    const shiftPressed = ref(false);
    const lastSelectedRow = shallowRef<any | null>(null);
    const rangeBaseRows = shallowRef<any[]>([]);
    let isSyncingTableSelection = false;
    let skipNextSelectionChange = false;

    // Element Plus can return reactive proxies for rows supplied as plain objects.
    const sameRow = (left: any, right: any) => toRaw(left) === toRaw(right);
    const hasRow = (rows: any[], row: any) => rows.some((item) => sameRow(item, row));
    const isRowSelected = (row: any) => hasRow(selectedRows.value, row);
    const getTable = () => tableRef.value?.refElTable;
    const clearTextSelection = () => window.getSelection?.()?.removeAllRanges();
    const hasActiveTextSelection = () => {
        const selection = window.getSelection?.();
        return !!selection && !selection.isCollapsed && selection.toString().trim().length > 0;
    };
    const setSelectedRows = (rows: any[]) => {
        selectedRows.value = rows;
        onSelectionChange(rows);
    };
    const syncTableSelection = () => {
        const table = getTable();
        if (!table) {
            return;
        }
        isSyncingTableSelection = true;
        try {
            const tableData = getTableData();
            const nextRows = selectedRows.value.filter((row) => hasRow(tableData, row) && isRowSelectable(row));
            const currentRows = table.getSelectionRows();
            currentRows.filter((row) => !hasRow(nextRows, row)).forEach((row) => table.toggleRowSelection(row, false));
            nextRows.filter((row) => !hasRow(currentRows, row)).forEach((row) => table.toggleRowSelection(row, true));
        } finally {
            isSyncingTableSelection = false;
        }
    };
    const selectRow = (row: any, selected = !isRowSelected(row)) => {
        if (!isRowSelectable(row)) {
            return;
        }
        const nextRows = selected
            ? isRowSelected(row)
                ? selectedRows.value
                : [...selectedRows.value, row]
            : selectedRows.value.filter((item) => !sameRow(item, row));
        setSelectedRows(nextRows);
        syncTableSelection();
    };
    const applyRangeSelection = (targetRow: any) => {
        if (!lastSelectedRow.value) return false;
        const tableData = getTableData();
        const startIndex = tableData.findIndex((row) => sameRow(row, lastSelectedRow.value));
        const endIndex = tableData.findIndex((row) => sameRow(row, targetRow));
        if (startIndex === -1 || endIndex === -1) return false;

        const [start, end] = [startIndex, endIndex].sort((a, b) => a - b);
        const rangeRows = tableData.slice(start, end + 1).filter(isRowSelectable);
        const nextRows = [...rangeBaseRows.value];
        rangeRows.forEach((row) => !hasRow(nextRows, row) && nextRows.push(row));
        setSelectedRows(nextRows);
        syncTableSelection();
        return true;
    };
    const handleSelectionChange = (rows: any[]) => {
        if (isSyncingTableSelection) {
            return;
        }
        if (skipNextSelectionChange) {
            skipNextSelectionChange = false;
            syncTableSelection();
            return;
        }
        setSelectedRows(rows);
        if (rows.length === 0) {
            lastSelectedRow.value = null;
            rangeBaseRows.value = [];
        }
    };
    const handleSelect = (selection: any[], row: any) => {
        if (shiftPressed.value && applyRangeSelection(row)) {
            skipNextSelectionChange = true;
            clearTextSelection();
            return;
        }
        lastSelectedRow.value = row;
        rangeBaseRows.value = selection.filter((item) => !sameRow(item, row));
        clearTextSelection();
    };
    const clearSelects = () => {
        setSelectedRows([]);
        syncTableSelection();
        lastSelectedRow.value = null;
        rangeBaseRows.value = [];
    };
    const pruneSelection = () => {
        const nextRows = selectedRows.value.filter((row) => hasRow(getTableData(), row));
        if (nextRows.length !== selectedRows.value.length) {
            setSelectedRows(nextRows);
        }
        if (lastSelectedRow.value && !hasRow(nextRows, lastSelectedRow.value)) {
            lastSelectedRow.value = null;
            rangeBaseRows.value = [];
        }
        syncTableSelection();
    };
    const toggleSelection = () => {
        const selectableRows = getTableData().filter(isRowSelectable);
        const allSelected = selectableRows.length > 0 && selectableRows.every(isRowSelected);
        const nextRows = allSelected
            ? selectedRows.value.filter((row) => !hasRow(selectableRows, row))
            : [...selectedRows.value, ...selectableRows.filter((row) => !isRowSelected(row))];
        setSelectedRows(nextRows);
        syncTableSelection();
    };
    const handleRowClick = (row: any, _column: any, event: MouseEvent) => {
        if (!isRowSelectable(row) || (hasActiveTextSelection() && !event.shiftKey)) return;
        const target = event.target as HTMLElement;
        if (
            target.closest(
                '.el-checkbox, button, a, input, textarea, [contenteditable="true"], .el-input, .el-textarea, .el-input-number, .el-date-editor, .el-switch, .el-select, .table-link, .cursor-pointer',
            )
        )
            return;
        if (event.shiftKey && applyRangeSelection(row)) {
            clearTextSelection();
            return;
        }
        // Ordinary row clicks preserve existing selections; use the checkbox to deselect.
        if (!isRowSelected(row)) {
            selectRow(row, true);
        }
        lastSelectedRow.value = row;
        rangeBaseRows.value = selectedRows.value.filter((item) => !sameRow(item, row));
        clearTextSelection();
    };
    const handleKeyDown = (event: KeyboardEvent) => {
        if (event.key === 'Shift') shiftPressed.value = true;
    };
    const handleKeyUp = (event: KeyboardEvent) => {
        if (event.key === 'Shift') shiftPressed.value = false;
    };

    return {
        selectedRows,
        isRowSelected,
        clearSelects,
        pruneSelection,
        toggleSelection,
        selectRow,
        syncTableSelection,
        handleSelect,
        handleSelectionChange,
        handleRowClick,
        handleKeyDown,
        handleKeyUp,
    };
};
