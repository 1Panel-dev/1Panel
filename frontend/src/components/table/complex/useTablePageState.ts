import { watch } from 'vue';

type Pagination = { currentPage?: number; pageSize?: number } | undefined;

/** Refreshing data leaves page state alone; only user navigation or a new query clears selection. */
export const useTablePageState = (
    getPagination: () => Pagination,
    getContext: () => unknown,
    clearSelection: () => void,
) => {
    watch(() => [getPagination()?.currentPage, getPagination()?.pageSize], clearSelection, { flush: 'sync' });
    watch(
        () => JSON.stringify(getContext()),
        () => {
            clearSelection();
            const pagination = getPagination();
            if (pagination) pagination.currentPage = 1;
        },
        { flush: 'sync' },
    );
};
