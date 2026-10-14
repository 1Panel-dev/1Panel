import { watch, type Ref, type WatchSource } from 'vue';

export type SearchFields = Record<string, Ref<string | number | undefined>>;

export const bindSearchReset = (sources: WatchSource[], fields: SearchFields) => {
    const defaults = Object.values(fields).map((field) => ({ field, value: field.value }));
    const reset = () => {
        for (const { field, value } of defaults) field.value = value;
    };
    watch(sources, reset, { flush: 'sync' });
    return reset;
};

type SearchValues = Record<string, string | number | undefined>;

/** A short-lived return state, discarded as soon as navigation leaves its list/edit flow. */
export const createSearchVisitCache = () => {
    const entries = new Map<string, { routes: string[]; values: SearchValues }>();
    let context: unknown;
    const syncContext = (next: unknown) => {
        if (context !== next) entries.clear();
        context = next;
    };
    return {
        navigate(route: string, nextContext: unknown) {
            syncContext(nextContext);
            for (const [key, entry] of entries) {
                if (!entry.routes.includes(route)) entries.delete(key);
            }
        },
        save(key: string, routes: string[], route: string, nextContext: unknown, values: SearchValues) {
            syncContext(nextContext);
            if (routes.includes(route)) entries.set(key, { routes, values: { ...values } });
        },
        read(key: string, nextContext: unknown) {
            syncContext(nextContext);
            const values = entries.get(key)?.values;
            return values ? { ...values } : undefined;
        },
    };
};
