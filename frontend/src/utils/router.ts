import router from '@/routers';
import { TabsStore } from '@/store';
import { hasRouteAccess } from '@/utils/rbac';
import i18n from '@/lang';
import { MsgError } from '@/utils/message';
import type { RouteLocationRaw } from 'vue-router';

const pushWithAccessCheck = async (to: RouteLocationRaw) => {
    const resolvedRoute = router.resolve(to);
    if (!hasRouteAccess(resolvedRoute)) {
        MsgError(i18n.global.t('commons.res.forbidden'));
        return;
    }
    await router.push(to);
    tabStoreMiddleWare();
};

export const routerToName = async (name: string) => {
    await pushWithAccessCheck({ name: name });
};

export const routerToPath = async (path: string) => {
    await pushWithAccessCheck({ path: path });
};

export const routerToFileWithPath = async (pathItem: string) => {
    await pushWithAccessCheck({ name: 'File', query: { path: pathItem, uncached: 'true' } });
};

export const routerToNameWithQuery = async (name: string, query: any) => {
    await pushWithAccessCheck({ name: name, query: query });
};

export const routerToPathWithQuery = async (path: string, query: any) => {
    await pushWithAccessCheck({ path: path, query: query });
};

export const jumpToPath = (router: any, path: string) => {
    routerToPathWithQuery(path, { uncached: 'true' });
};

export const routerToNameWithParams = async (name: string, params: any) => {
    await pushWithAccessCheck({ name: name, params: params });
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
