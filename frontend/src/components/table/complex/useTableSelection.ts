import { ref, type Ref } from 'vue';

type TableRef = { refElTable?: any } | undefined;

export const useTableSelection = (
    tableRef: Ref<TableRef>,
    getTableData: () => any[],
    onSelectionChange: (rows: any[]) => void,
    isRowSelectable: (row: any) => boolean = () => true,
) => {
    const selectedRows = ref<any[]>([]);
    const shiftPressed = ref(false);
    const lastSelectedRow = ref<any | null>(null);
    const rangeBaseRows = ref<any[]>([]);
    let isSyncingTableSelection = false;

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
            table.clearSelection();
            selectedRows.value
                .filter((row) => getTableData().includes(row) && isRowSelectable(row))
                .forEach((row) => table.toggleRowSelection(row, true));
        } finally {
            isSyncingTableSelection = false;
        }
    };
    const selectRow = (row: any, selected = !selectedRows.value.includes(row)) => {
        if (!isRowSelectable(row)) {
            return;
        }
        const nextRows = selected
            ? selectedRows.value.includes(row)
                ? selectedRows.value
                : [...selectedRows.value, row]
            : selectedRows.value.filter((item) => item !== row);
        setSelectedRows(nextRows);
        syncTableSelection();
    };
    const applyRangeSelection = (targetRow: any) => {
        if (!lastSelectedRow.value) return false;
        const tableData = getTableData();
        const startIndex = tableData.indexOf(lastSelectedRow.value);
        const endIndex = tableData.indexOf(targetRow);
        if (startIndex === -1 || endIndex === -1) return false;

        const [start, end] = [startIndex, endIndex].sort((a, b) => a - b);
        const rangeRows = tableData.slice(start, end + 1).filter(isRowSelectable);
        const nextRows = [...rangeBaseRows.value];
        rangeRows.forEach((row) => !nextRows.includes(row) && nextRows.push(row));
        setSelectedRows(nextRows);
        syncTableSelection();
        return true;
    };
    const handleSelectionChange = (rows: any[]) => {
        if (isSyncingTableSelection) {
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
            clearTextSelection();
            return;
        }
        lastSelectedRow.value = row;
        rangeBaseRows.value = selection.filter((item) => item !== row);
        clearTextSelection();
    };
    const clearSelects = () => {
        setSelectedRows([]);
        syncTableSelection();
        lastSelectedRow.value = null;
        rangeBaseRows.value = [];
    };
    const pruneSelection = () => {
        const nextRows = selectedRows.value.filter((row) => getTableData().includes(row));
        if (nextRows.length !== selectedRows.value.length) {
            setSelectedRows(nextRows);
        }
        if (lastSelectedRow.value && !nextRows.includes(lastSelectedRow.value)) {
            lastSelectedRow.value = null;
            rangeBaseRows.value = [];
        }
        syncTableSelection();
    };
    const toggleSelection = () => {
        const selectableRows = getTableData().filter(isRowSelectable);
        const allSelected =
            selectableRows.length > 0 && selectableRows.every((row) => selectedRows.value.includes(row));
        const nextRows = allSelected
            ? selectedRows.value.filter((row) => !selectableRows.includes(row))
            : [...selectedRows.value, ...selectableRows.filter((row) => !selectedRows.value.includes(row))];
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
        const selected = !selectedRows.value.includes(row);
        selectRow(row, selected);
        lastSelectedRow.value = row;
        rangeBaseRows.value = selected ? selectedRows.value.filter((item) => item !== row) : selectedRows.value;
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
