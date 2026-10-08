import { storeToRefs } from 'pinia';
import { watch, type Ref } from 'vue';
import GlobalStore from '@/store/modules/global';

type SearchValue = string | number | undefined;
type SearchFields = Record<string, Ref<SearchValue>>;

export const useSearchPersistence = (page: string, fields: SearchFields, scope?: () => SearchValue) => {
    const { currentNode } = storeToRefs(GlobalStore());
    const defaults = Object.fromEntries(Object.entries(fields).map(([name, field]) => [name, field.value]));
    let initialized = false;
    let restoring = false;
    let activeKey: string | undefined;

    watch(
        () => {
            const context = scope?.();
            if (scope && context === undefined) return undefined;
            return `1panel:search:v1:${JSON.stringify([currentNode.value, page, context ?? ''])}`;
        },
        (key) => {
            activeKey = key;
            let saved: Record<string, unknown> = {};
            try {
                const parsed = activeKey ? JSON.parse(localStorage.getItem(activeKey) || '{}') : {};
                if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) saved = parsed;
            } catch {}
            restoring = true;
            for (const [name, field] of Object.entries(fields)) {
                const value = saved[name];
                const initial = defaults[name];
                let fallback = initial;
                if (initialized) {
                    fallback = typeof initial === 'string' ? '' : undefined;
                }
                const valid =
                    (typeof value === 'string' || (typeof value === 'number' && Number.isFinite(value))) &&
                    (fallback === undefined || typeof value === typeof fallback);
                if (value === null) {
                    field.value = typeof initial === 'string' ? '' : undefined;
                } else if (valid) {
                    field.value = value as SearchValue;
                } else {
                    field.value = fallback;
                }
            }
            restoring = false;
            initialized = true;
        },
        { immediate: true, flush: 'sync' },
    );

    watch(
        () => Object.fromEntries(Object.entries(fields).map(([name, field]) => [name, field.value ?? null])),
        (value) => {
            if (restoring || !activeKey) return;
            try {
                localStorage.setItem(activeKey, JSON.stringify(value));
            } catch {}
        },
        { flush: 'sync' },
    );
};
