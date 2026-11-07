import { nextTick, onActivated, onDeactivated, onScopeDispose, watch } from 'vue';
import GlobalStore from '@/store/modules/global';
import { panelCacheKey, readPanelCache, writePanelCache } from '@/utils/panel-cache';

interface Pagination {
    currentPage: number;
    pageSize: number;
    cacheSizeKey?: string;
}
interface Options {
    pagePath: string;
    pagination: () => Pagination | undefined;
    enabled: () => boolean | undefined;
    id: () => string | undefined;
    scope: () => string | number | undefined;
    search: () => void;
    refreshOnRestore?: () => boolean;
}

/** Restore during child setup, before the owning page's onMounted query. */
export const usePaginationCache = (options: Options) => {
    const store = GlobalStore();
    const defaults = new WeakMap<Pagination, { currentPage: number; pageSize: number }>();
    let active = true;
    let restoring = false;
    let activeKey: string | undefined;
    let queued = false;
    let disposed = false;
    let pending: { key: string; currentPage: number; pageSize: number } | undefined;
    let writeQueued = false;
    const flush = () => {
        if (!pending) return;
        const { key: target, currentPage, pageSize } = pending;
        pending = undefined;
        writePanelCache('pagination', target, { currentPage, pageSize });
    };
    const requestSearch = () => {
        if (queued || !key()) return;
        queued = true;
        void nextTick(() => {
            queued = false;
            if (active && !disposed && key() && options.refreshOnRestore?.() !== false) options.search();
        });
    };
    const key = () => {
        const pagination = options.pagination();
        if (options.enabled() === false || !pagination || !store.cacheUserID) return undefined;
        return panelCacheKey(
            JSON.stringify([
                'table-pagination',
                options.pagePath,
                options.id() ?? pagination.cacheSizeKey ?? 'default',
                options.scope() ?? '',
            ]),
        );
    };
    const positiveInteger = (value: unknown): value is number => Number.isInteger(value) && Number(value) > 0;
    const restore = () => {
        flush();
        activeKey = key();
        const pagination = options.pagination();
        if (!activeKey || !pagination) return false;
        if (!defaults.has(pagination)) {
            defaults.set(pagination, { currentPage: pagination.currentPage, pageSize: pagination.pageSize });
        }
        const initial = defaults.get(pagination)!;
        const saved = readPanelCache<Partial<Pagination>>('pagination', activeKey);
        const page = positiveInteger(saved?.currentPage) ? saved.currentPage : initial.currentPage;
        const size = positiveInteger(saved?.pageSize) ? saved.pageSize : initial.pageSize;
        const changed = pagination.currentPage !== page || pagination.pageSize !== size;
        restoring = true;
        try {
            pagination.pageSize = size;
            pagination.currentPage = page;
        } finally {
            restoring = false;
        }
        return changed;
    };
    // The initial restore deliberately does not emit: the parent owns its first query.
    restore();
    watch(
        () => [key(), options.pagination(), store.panelCache.pagination, store.panelCache.duration],
        () => {
            if (!active) return;
            const previousKey = activeKey;
            const changed = restore();
            if (activeKey && (changed || previousKey !== activeKey)) requestSearch();
        },
        { flush: 'sync' },
    );
    watch(
        () => [options.pagination()?.currentPage, options.pagination()?.pageSize],
        ([currentPage, pageSize]) => {
            // Never write old-page state under a newly selected node/account.
            if (!active || restoring || !activeKey || activeKey !== key()) return;
            if (!positiveInteger(currentPage) || !positiveInteger(pageSize)) return;
            pending = { key: activeKey, currentPage, pageSize };
            if (!writeQueued) {
                writeQueued = true;
                queueMicrotask(() => {
                    writeQueued = false;
                    flush();
                });
            }
        },
        { flush: 'sync' },
    );
    onDeactivated(() => {
        flush();
        active = false;
    });
    onActivated(() => {
        if (active) return;
        active = true;
        restore();
        requestSearch();
    });
    onScopeDispose(() => {
        flush();
        disposed = true;
        active = false;
    });
};
