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
