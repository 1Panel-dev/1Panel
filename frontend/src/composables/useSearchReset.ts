import { storeToRefs } from 'pinia';
import { onActivated, onDeactivated } from 'vue';
import GlobalStore from '@/store/modules/global';
import { bindSearchState, type SearchFields } from '@/utils/search-reset';

interface SearchResetOptions {
    scope?: () => string | number | undefined;
    onRevisit: () => unknown;
}

/** Retain the existing entry point while keeping queries across all page visits. */
export const useSearchReset = (page: string, fields: SearchFields, options: SearchResetOptions) => {
    const { currentNode } = storeToRefs(GlobalStore());
    const restore = bindSearchState(() => {
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
        restore();
        return options.onRevisit();
    });
};
