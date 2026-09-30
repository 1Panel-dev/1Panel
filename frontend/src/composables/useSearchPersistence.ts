import { storeToRefs } from 'pinia';
import GlobalStore from '@/store/modules/global';
import router from '@/routers/router';
import { bindSearchReset } from '@/utils/search-reset';

/** Kept as a compatible entry point; searches now reset on tab, node or resource changes. */
export const useSearchPersistence = (
    page: string,
    fields: Parameters<typeof bindSearchReset>[1],
    scope?: () => string | number | undefined,
) => {
    const { currentNode } = storeToRefs(GlobalStore());
    return bindSearchReset(() => [router.currentRoute.value.path, currentNode.value, page, scope?.()], fields);
};
