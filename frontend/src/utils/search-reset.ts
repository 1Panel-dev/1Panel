import { watch, type Ref } from 'vue';

export type SearchFields = Record<string, Ref<string | number | undefined>>;
type SearchValues = Record<string, string | number | undefined>;

/** Shared by normal page visits and edit-return visits; navigation never evicts a query. */
export const createSearchVisitCache = (storage: () => Storage = () => localStorage) => {
    const entries = new Map<string, SearchValues>();
    return {
        save(key: string, values: SearchValues) {
            entries.set(key, { ...values });
            try {
                storage().setItem(
                    `1panel:search-state:v2:${key}`,
                    JSON.stringify(
                        Object.fromEntries(Object.entries(values).map(([name, value]) => [name, value ?? null])),
                    ),
                );
            } catch {
                // Keep in-memory state when storage is unavailable or full.
            }
        },
        read(key: string): SearchValues | undefined {
            const cached = entries.get(key);
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

const visitCache = createSearchVisitCache();

export const bindSearchState = (key: () => string | undefined, fields: SearchFields, cache = visitCache) => {
    const defaults = Object.fromEntries(Object.entries(fields).map(([name, field]) => [name, field.value]));
    let restoring = false;
    let activeKey: string | undefined;
    const restore = () => {
        activeKey = key();
        if (!activeKey) return;
        const saved = cache.read(activeKey) || defaults;
        restoring = true;
        try {
            // Restore pagination last: search/filter watchers may reset it to page one.
            const names = Object.keys(fields).sort((left, right) => Number(left === 'page') - Number(right === 'page'));
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
            if (!restoring && activeKey) cache.save(activeKey, values);
        },
        { immediate: true, flush: 'sync' },
    );
    return restore;
};
