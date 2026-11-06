import router from '@/routers';
import { TabsStore } from '@/store';
import { hasRouteAccess } from '@/utils/rbac';
import i18n from '@/lang';
import { MsgError } from '@/utils/message';
import type { RouteLocationRaw } from 'vue-router';
import { panelCacheKey, readPanelCache, removePanelCache } from '@/utils/panel-cache';

const pushWithAccessCheck = async (to: RouteLocationRaw) => {
    const resolvedRoute = router.resolve(to);
    if (!hasRouteAccess(resolvedRoute)) {
        MsgError(i18n.global.t('commons.res.forbidden'));
        return;
    }
    await router.push(to);
    tabStoreMiddleWare();
};

/** Explicit navigation always opens the requested page, independent of tab caching. */
export const openPage = pushWithAccessCheck;

/** Only module-menu entry points restore the last visited route within that module. */
export const openMenu = async (to: RouteLocationRaw) => {
    let entry = router.resolve(to);
    const visited = new Set<string>();
    // Menu roots often redirect and only their destination declares activeMenu.
    while (!visited.has(entry.fullPath)) {
        visited.add(entry.fullPath);
        const redirect = entry.matched[entry.matched.length - 1]?.redirect;
        if (!redirect) break;
        entry = router.resolve(typeof redirect === 'function' ? redirect(entry, router.currentRoute.value) : redirect);
    }
    const menu = entry.meta.activeMenu;
    if (menu) {
        const key = panelCacheKey(String(menu));
        const cached = readPanelCache<unknown>('tabs', key);
        if (typeof cached === 'string') {
            const destination = router.resolve(cached);
            if (destination.matched.length && destination.meta.activeMenu === menu && hasRouteAccess(destination)) {
                return openPage(cached);
            }
        }
        if (cached !== undefined) removePanelCache('tabs', key);
    }
    return openPage(to);
};

export const routerToName = async (name: string) => {
    await openPage({ name: name });
};

export const routerToPath = async (path: string) => {
    await openPage({ path: path });
};

export const routerToFileWithPath = async (pathItem: string) => {
    await openPage({ name: 'File', query: { path: pathItem, uncached: 'true' } });
};

export const routerToNameWithQuery = async (name: string, query: any) => {
    await openPage({ name: name, query: query });
};

export const routerToPathWithQuery = async (path: string, query: any) => {
    await openPage({ path: path, query: query });
};

export const jumpToPath = (router: any, path: string) => {
    routerToPathWithQuery(path, { uncached: 'true' });
};

export const routerToNameWithParams = async (name: string, params: any) => {
    await openPage({ name: name, params: params });
};

const tabStoreMiddleWare = () => {
    try {
        let route = router.currentRoute;
        if (route.value.meta.ignoreTab) {
            return;
        }
        const tabsStore = TabsStore();
        tabsStore.addTab(route.value);
        tabsStore.activeTabPath = route.value?.path;
    } catch (error) {}
};
