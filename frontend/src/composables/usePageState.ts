import { reactive, type UnwrapNestedRefs } from 'vue';

/** A new page visit starts from its defaults instead of restoring the previous page number. */
export const usePageState = <T extends object>(factory: () => T): UnwrapNestedRefs<T> => {
    return reactive(factory());
};
