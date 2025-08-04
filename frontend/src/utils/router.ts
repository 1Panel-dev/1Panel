import router from '@/routers';
import { TabsStore } from '@/store';

export const routerToName = async (name: string) => {
    await router.push({ name: name });
    tabStoreMiddleWare();
};

export const routerToPath = async (path: string) => {
    await router.push({ path: path });
    tabStoreMiddleWare();
};

export const routerToFileWithPath = async (pathItem: string) => {
    await router.push({ name: 'File', query: { path: pathItem } });
    tabStoreMiddleWare();
};

export const routerToNameWithQuery = async (name: string, query: any) => {
    await router.push({ name: name, query: query });
    tabStoreMiddleWare();
};

export const routerToPathWithQuery = async (path: string, query: any) => {
    await router.push({ path: path, query: query });
    tabStoreMiddleWare();
};

export const routerToNameWithParams = async (name: string, params: any) => {
    await router.push({ name: name, params: params });
    tabStoreMiddleWare();
};

const tabStoreMiddleWare = () => {
    try {
        const tabsStore = TabsStore();
        let route = router.currentRoute;
        tabsStore.addTab(route.value);
        tabsStore.activeTabPath = route.value?.path;
    } catch (error) {}
};
