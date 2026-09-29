import { watch, type Ref } from 'vue';

type SearchValue = string | number | undefined;
type SearchFields = Record<string, Ref<SearchValue>>;

export const bindSearchPersistence = (key: () => string | undefined, fields: SearchFields) => {
    const defaults = Object.fromEntries(Object.entries(fields).map(([name, field]) => [name, field.value]));
    let initialized = false;
    let restoring = false;
    let activeKey: string | undefined;

    const stopRestore = watch(
        key,
        (value) => {
            activeKey = value ? `1panel:search:v1:${value}` : undefined;
            let saved: Record<string, unknown> = {};
            try {
                const parsed = activeKey ? JSON.parse(localStorage.getItem(activeKey) || '{}') : {};
                if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) saved = parsed;
            } catch {
                // Invalid or unavailable storage must not prevent a page from loading.
            }
            restoring = true;
            for (const [name, field] of Object.entries(fields)) {
                const value = saved[name];
                const initial = defaults[name];
                const fallback = initialized ? (typeof initial === 'string' ? '' : undefined) : initial;
                const valid =
                    (typeof value === 'string' || (typeof value === 'number' && Number.isFinite(value))) &&
                    (fallback === undefined || typeof value === typeof fallback);
                field.value =
                    value === null
                        ? typeof initial === 'string'
                            ? ''
                            : undefined
                        : valid
                          ? (value as SearchValue)
                          : fallback;
            }
            restoring = false;
            initialized = true;
        },
        { immediate: true, flush: 'sync' },
    );

    const stopSave = watch(
        () => Object.fromEntries(Object.entries(fields).map(([name, field]) => [name, field.value ?? null])),
        (value) => {
            if (restoring || !activeKey) return;
            try {
                localStorage.setItem(activeKey, JSON.stringify(value));
            } catch {
                // Search remains usable when browser storage is disabled or full.
            }
        },
        { flush: 'sync' },
    );

    return () => {
        stopRestore();
        stopSave();
    };
};
