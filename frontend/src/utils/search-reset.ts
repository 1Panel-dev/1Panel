import { watch, type Ref } from 'vue';

type SearchFields = Record<string, Ref<string | number | undefined>>;

/** Search terms belong to the current tab visit, not browser storage or cached pages. */
export const bindSearchReset = (context: () => unknown, fields: SearchFields) => {
    const emptyValues = Object.fromEntries(
        Object.entries(fields).map(([name, field]) => [name, typeof field.value === 'string' ? '' : undefined]),
    );
    return watch(
        () => JSON.stringify(context()),
        () => {
            for (const [name, field] of Object.entries(fields)) field.value = emptyValues[name];
        },
        { immediate: true, flush: 'sync' },
    );
};
