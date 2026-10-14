import { storeToRefs } from 'pinia';
import { onActivated, onDeactivated, onScopeDispose, watch, type Ref } from 'vue';
import GlobalStore from '@/store/modules/global';

export type SearchFields = Record<string, Ref<string | number | undefined>>;
type SearchValues = Record<string, string | number | undefined>;

/** Shared by normal page visits and edit-return visits; navigation never evicts a query. */
export const createSearchStateStorage = (storage: () => Storage = () => localStorage) => {
    const unsaved = new Map<string, SearchValues>();
    return {
        save(key: string, values: SearchValues) {
            try {
                const target = storage();
                const storageKey = `1panel:search-state:v2:${key}`;
                const serialized = JSON.stringify(
                    Object.fromEntries(Object.entries(values).map(([name, value]) => [name, value ?? null])),
                );
                if (target.getItem(storageKey) !== serialized) target.setItem(storageKey, serialized);
                unsaved.delete(key);
            } catch {
                // Only failed writes need an in-memory fallback.
                unsaved.set(key, { ...values });
            }
        },
        read(key: string): SearchValues | undefined {
            const cached = unsaved.get(key);
            if (cached) return { ...cached };
            try {
                const saved = JSON.parse(storage().getItem(`1panel:search-state:v2:${key}`) || 'null');
                if (!saved || typeof saved !== 'object' || Array.isArray(saved)) return undefined;
                return Object.fromEntries(
                    Object.entries(saved)
                        .filter(
                            ([, value]) =>
                                value === null ||
                                typeof value === 'string' ||
                                (typeof value === 'number' && Number.isFinite(value)),
                        )
                        .map(([name, value]) => [name, value === null ? undefined : value]),
                ) as SearchValues;
            } catch {
                return undefined;
            }
        },
    };
};

const searchStorage = createSearchStateStorage();

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
    onScopeDispose(flush);
    let restoring = false;
    let activeKey: string | undefined;
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
                fields[name].value =
                    initial !== undefined && value !== undefined && typeof value !== typeof initial ? initial : value;
            }
        } finally {
            restoring = false;
        }
    };
    watch(key, restore, { immediate: true, flush: 'sync' });
    watch(
        () => Object.fromEntries(Object.entries(fields).map(([name, field]) => [name, field.value])),
        (values) => {
            if (restoring || !activeKey) return;
            pending = { key: activeKey, values };
            if (queued) return;
            queued = true;
            queueMicrotask(() => {
                queued = false;
                flush();
            });
        },
        { flush: 'sync' },
    );
};

interface SearchStateOptions {
    scope?: () => string | number | undefined;
    onRevisit: () => unknown;
}

/** Keep queries across page visits and reload data when a cached page becomes active. */
export const useSearchState = (page: string, fields: SearchFields, options: SearchStateOptions) => {
    const { currentNode } = storeToRefs(GlobalStore());
    bindSearchState(() => {
        const scope = options.scope?.();
        if (options.scope && scope === undefined) return undefined;
        return JSON.stringify([currentNode.value, page, scope ?? '']);
    }, fields);
    let deactivated = false;
    onDeactivated(() => {
        deactivated = true;
    });
    onActivated(() => {
        if (!deactivated) return;
        deactivated = false;
        return options.onRevisit();
    });
};
