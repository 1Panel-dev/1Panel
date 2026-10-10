import { storeToRefs } from 'pinia';
import { onActivated, onDeactivated } from 'vue';
import GlobalStore from '@/store/modules/global';
import { bindSearchReset, type SearchFields } from '@/utils/search-reset';

interface SearchResetOptions {
    scope?: () => string | number | undefined;
    onRevisit: () => unknown;
}

export const useSearchReset = (fields: SearchFields, options: SearchResetOptions) => {
    const { currentNode } = storeToRefs(GlobalStore());
    const reset = bindSearchReset(options.scope ? [currentNode, options.scope] : [currentNode], fields);
    let deactivated = false;
    onDeactivated(() => {
        deactivated = true;
    });
    onActivated(() => {
        if (!deactivated) return;
        deactivated = false;
        reset();
        return options.onRevisit();
    });
};
