const pageStateCache = new Map<string, object>();

export const getPageState = <T extends object>(key: string, factory: () => T): T => {
    const cached = pageStateCache.get(key) as T | undefined;
    if (cached) {
        return cached;
    }
    const state = factory();
    pageStateCache.set(key, state);
    return state;
};

export const clearPageStateCache = () => {
    pageStateCache.clear();
};
