import { storeToRefs } from 'pinia';
import GlobalStore from '@/store/modules/global';
import { bindSearchPersistence } from '@/utils/search-persistence';

/** Restore the parent's search fields synchronously, before its initial request. */
export const useSearchPersistence = (
    page: string,
    fields: Parameters<typeof bindSearchPersistence>[1],
    scope?: () => string | number | undefined,
) => {
    const { currentNode } = storeToRefs(GlobalStore());
    return bindSearchPersistence(() => {
        const context = scope?.();
        if (scope && context === undefined) return undefined;
        return JSON.stringify([currentNode.value, page, context ?? '']);
    }, fields);
};
