import { storeToRefs } from 'pinia';
import { onActivated, onDeactivated, watch } from 'vue';
import { useRouter, type Router } from 'vue-router';
import GlobalStore from '@/store/modules/global';
import { bindSearchReset, createSearchVisitCache, type SearchFields } from '@/utils/search-reset';

interface SearchResetOptions {
    scope?: () => string | number | undefined;
    onRevisit: () => unknown;
    preserveOnRoutes?: string[];
}

const visitCache = createSearchVisitCache();
const observedRouters = new WeakSet<Router>();

export const useSearchReset = (fields: SearchFields, options: SearchResetOptions) => {
    const { currentNode } = storeToRefs(GlobalStore());
    const reset = bindSearchReset(options.scope ? [currentNode, options.scope] : [currentNode], fields);
    let restoreVisit: (() => boolean) | undefined;
    if (options.preserveOnRoutes?.length) {
        const router = useRouter();
        const routes = options.preserveOnRoutes;
        const key = () => JSON.stringify([routes, options.scope?.()]);
        const context = () => currentNode.value;
        // This observer must outlive the list, which may unmount while the editor is open.
        if (!observedRouters.has(router)) {
            observedRouters.add(router);
            router.afterEach((to, _from, failure) => {
                if (!failure) visitCache.navigate(String(to.name || ''), context());
            });
        }
        visitCache.navigate(String(router.currentRoute.value.name || ''), context());
        restoreVisit = () => {
            const saved = visitCache.read(key(), context());
            if (!saved) return false;
            // Query watchers may reset pagination; restore the requested page last.
            for (const [name, field] of Object.entries(fields)) {
                if (name !== 'page') field.value = saved[name];
            }
            if (fields.page) fields.page.value = saved.page;
            return true;
        };
        restoreVisit();
        watch(
            () => Object.fromEntries(Object.entries(fields).map(([name, field]) => [name, field.value])),
            (values) => visitCache.save(key(), routes, String(router.currentRoute.value.name || ''), context(), values),
            { immediate: true, flush: 'sync' },
        );
    }
    let deactivated = false;
    onDeactivated(() => {
        deactivated = true;
    });
    onActivated(() => {
        if (!deactivated) return;
        deactivated = false;
        if (!restoreVisit?.()) reset();
        return options.onRevisit();
    });
};
