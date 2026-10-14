import { storeToRefs } from 'pinia';
import { nextTick, onActivated, onDeactivated, onScopeDispose, watch, type Ref } from 'vue';
import { panelCacheKey, readPanelCache, writePanelCache } from '@/utils/panel-cache';
import GlobalStore from '@/store/modules/global';

export type SearchFields = Record<string, Ref<string | number | undefined>>;
type SearchValues = Record<string, string | number | undefined>;

// Keep categories separate so changing search text does not renew pagination's lifetime.
const searchStorage = {
    save(key: string, values: SearchValues) {
        for (const category of ['search', 'pagination'] as const) {
            const entries = Object.entries(values).filter(
                ([name]) => (name === 'page' || name === 'pageSize') === (category === 'pagination'),
            );
            if (entries.length) {
                writePanelCache(
                    category,
                    key,
                    Object.fromEntries(entries.map(([name, value]) => [name, value ?? null])),
                );
            }
        }
    },
    read(key: string): SearchValues {
        const values = {
            ...readPanelCache<SearchValues>('search', key),
            ...readPanelCache<SearchValues>('pagination', key),
        };
        return Object.fromEntries(
            Object.entries(values)
                .filter(
                    ([, value]) =>
                        value === null ||
                        typeof value === 'string' ||
                        (typeof value === 'number' && Number.isFinite(value)),
                )
                .map(([name, value]) => [name, value === null ? undefined : value]),
        );
    },
};

export const bindSearchState = (key: () => string | undefined, fields: SearchFields, storage = searchStorage) => {
    const defaults = Object.fromEntries(Object.entries(fields).map(([name, field]) => [name, field.value]));
    const names = Object.keys(fields).sort((left, right) => Number(left === 'page') - Number(right === 'page'));
    let pending: { key: string; values: SearchValues } | undefined;
    let queued = false;
    const flush = () => {
        if (!pending) return;
        const snapshot = pending;
        pending = undefined;
        storage.save(snapshot.key, snapshot.values);
    };
    let restoring = false;
    let activeKey: string | undefined;
    let active = true;
    let resuming = false;
    let resumeVersion = 0;
    const pause = () => {
        flush();
        active = false;
        resuming = false;
        resumeVersion++;
    };
    onScopeDispose(pause);
    const restore = () => {
        // Commit the previous node/resource before restoring another one.
        flush();
        activeKey = key();
        if (!activeKey) return;
        const saved = storage.read(activeKey) || defaults;
        restoring = true;
        try {
            // Restore pagination last: search/filter watchers may reset it to page one.
            for (const name of names) {
                const value = Object.hasOwn(saved, name) ? saved[name] : defaults[name];
                const initial = defaults[name];
                const invalidPage =
                    (name === 'page' || name === 'pageSize') &&
                    (typeof value !== 'number' || !Number.isInteger(value) || value < 1);
                fields[name].value =
                    invalidPage || (initial !== undefined && value !== undefined && typeof value !== typeof initial)
                        ? initial
                        : value;
            }
        } finally {
            restoring = false;
        }
    };
    watch(
        key,
        () => {
            if (active || resuming) restore();
        },
        { immediate: true, flush: 'sync' },
    );
    watch(
        () => Object.fromEntries(Object.entries(fields).map(([name, field]) => [name, field.value])),
        (values, previous) => {
            if (restoring || !active || !activeKey) return;
            const changed = Object.entries(values).filter(([name, value]) => value !== previous[name]);
            const paginationChanged = changed.some(([name]) => name === 'page' || name === 'pageSize');
            const searchChanged = changed.some(([name]) => name !== 'page' && name !== 'pageSize');
            const snapshot = Object.fromEntries(
                Object.entries(values).filter(([name]) =>
                    name === 'page' || name === 'pageSize' ? paginationChanged : searchChanged,
                ),
            );
            pending = { key: activeKey, values: { ...pending?.values, ...snapshot } };
            if (queued) return;
            queued = true;
            queueMicrotask(() => {
                queued = false;
                flush();
            });
        },
        { flush: 'sync' },
    );
    return {
        restore,
        pause,
        resume: async () => {
            const version = ++resumeVersion;
            resuming = true;
            active = false;
            restore();
            const restoredKey = activeKey;
            const page = fields.page?.value;
            await nextTick();
            if (version !== resumeVersion) return false;
            // Search watchers may reset the page while restoring a changed query.
            if (activeKey === restoredKey && fields.page && page !== undefined) fields.page.value = page;
            active = true;
            resuming = false;
            return true;
        },
    };
};

interface SearchStateOptions {
    scope?: () => string | number | undefined;
    onRevisit: () => unknown;
}

/** Keep queries across page visits and reload data when a cached page becomes active. */
export const useSearchState = (page: string, fields: SearchFields, options: SearchStateOptions) => {
    const { cacheUserID } = storeToRefs(GlobalStore());
    const binding = bindSearchState(() => {
        const scope = options.scope?.();
        if (options.scope && scope === undefined) return undefined;
        // Track account and node changes, including initial auth synchronisation.
        if (!cacheUserID.value) return undefined;
        return panelCacheKey(JSON.stringify([page, scope ?? '']));
    }, fields);
    let deactivated = false;
    onDeactivated(() => {
        deactivated = true;
        binding.pause();
    });
    onActivated(async () => {
        if (!deactivated) return;
        deactivated = false;
        if (await binding.resume()) return options.onRevisit();
    });
};
