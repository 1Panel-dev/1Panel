const DASHBOARD_CACHE_KEY = 'dashboardCache';

type CacheEntry = {
    value: any;
    expireAt: number;
};

const readCache = (): Record<string, CacheEntry> | null => {
    try {
        const cacheRaw = localStorage.getItem(DASHBOARD_CACHE_KEY);
        return cacheRaw ? JSON.parse(cacheRaw) : {};
    } catch {
        return null;
    }
};

export const getDashboardCache = (key: string) => {
    const cache = readCache();
    if (!cache) return null;
    const entry = cache[key];
    if (entry && entry.expireAt > Date.now()) {
        return entry.value;
    }
    return null;
};

export const setDashboardCache = (key: string, value: any, ttl: number) => {
    try {
        const cacheRaw = localStorage.getItem(DASHBOARD_CACHE_KEY);
        const cache = cacheRaw ? JSON.parse(cacheRaw) : {};
        cache[key] = {
            value,
            expireAt: Date.now() + ttl,
        };
        localStorage.setItem(DASHBOARD_CACHE_KEY, JSON.stringify(cache));
    } catch {
        localStorage.removeItem(DASHBOARD_CACHE_KEY);
    }
};

export const clearDashboardCache = () => {
    localStorage.removeItem(DASHBOARD_CACHE_KEY);
};

export const clearDashboardCacheByPrefix = (prefixes: string[]) => {
    try {
        const cacheRaw = localStorage.getItem(DASHBOARD_CACHE_KEY);
        if (!cacheRaw) return;
        const cache = JSON.parse(cacheRaw);
        Object.keys(cache).forEach((key: string) => {
            if (prefixes.some((prefix) => key.startsWith(prefix))) {
                delete cache[key];
            }
        });
        localStorage.setItem(DASHBOARD_CACHE_KEY, JSON.stringify(cache));
    } catch {
        clearDashboardCache();
    }
};

export { DASHBOARD_CACHE_KEY };
