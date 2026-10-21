import GlobalStore from '@/store/modules/global';

export type CacheCategory = 'tabs' | 'pagination' | 'refresh' | 'search';
export interface PanelCacheSettings {
    tabs: boolean;
    pagination: boolean;
    refresh: boolean;
    search: boolean;
    duration: '24h' | 'week' | 'month' | 'forever';
}
export const defaultPanelCache: PanelCacheSettings = {
    tabs: true,
    pagination: true,
    refresh: true,
    search: false,
    duration: '24h',
};
const categories: CacheCategory[] = ['tabs', 'pagination', 'refresh', 'search'];
const durations = { '24h': 86400000, week: 604800000, month: 2592000000, forever: Infinity };
const prefix = '1panel:page-cache:v1:';
interface CacheEntry {
    category: CacheCategory;
    updatedAt: number;
    value: unknown;
}

export const applyPanelCacheSettings = (raw: string) => {
    let value: Partial<PanelCacheSettings> = {};
    try {
        value = JSON.parse(raw) || {};
    } catch {
        // Older servers do not return this setting yet.
    }
    const settings = { ...defaultPanelCache };
    for (const category of categories) {
        if (typeof value[category] === 'boolean') settings[category] = value[category];
    }
    if (Object.hasOwn(durations, value.duration || '')) settings.duration = value.duration;
    const store = GlobalStore();
    const changed =
        settings.duration !== store.panelCache.duration ||
        categories.some((category) => settings[category] !== store.panelCache[category]);
    if (changed) store.panelCache = settings;
    // Repeated settings loads share the normal cleanup interval.
    prunePanelCache(changed);
};

export const panelCacheKey = (key: string): string => {
    const { cacheUserID, currentNode } = GlobalStore();
    return JSON.stringify([cacheUserID, currentNode, key]);
};
const storageKey = (category: CacheCategory, key: string) => prefix + JSON.stringify([category, key]);
const validEntry = (entry: CacheEntry, settings: PanelCacheSettings, now: number) => {
    return (
        entry &&
        categories.includes(entry.category) &&
        settings[entry.category] &&
        Number.isFinite(entry.updatedAt) &&
        entry.updatedAt <= now &&
        now - entry.updatedAt < durations[settings.duration]
    );
};

export const readPanelCache = <T>(category: CacheCategory, key: string): T | undefined => {
    const store = GlobalStore();
    if (!store.cacheUserID || !store.panelCache[category]) return;
    try {
        const target = storageKey(category, key);
        const raw = localStorage.getItem(target);
        if (!raw) return;
        const entry: CacheEntry = JSON.parse(raw);
        if (entry.category === category && validEntry(entry, store.panelCache, Date.now())) return entry.value as T;
        localStorage.removeItem(target);
    } catch {
        // Storage can be unavailable or contain invalid data; use page defaults.
    }
};

export const writePanelCache = (category: CacheCategory, key: string, value: unknown) => {
    const store = GlobalStore();
    if (!store.cacheUserID || !store.panelCache[category]) return;
    try {
        // Repeated visits and unchanged values must not extend the expiration time.
        const previous = readPanelCache(category, key);
        if (previous !== undefined && JSON.stringify(previous) === JSON.stringify(value)) return;
        localStorage.setItem(storageKey(category, key), JSON.stringify({ category, updatedAt: Date.now(), value }));
    } catch {
        // A full or disabled localStorage must not block page interactions.
    }
};

export const removePanelCache = (category: CacheCategory, key: string) => {
    try {
        localStorage.removeItem(storageKey(category, key));
    } catch {
        /* Storage is optional. */
    }
};

let lastPruned = 0;
export const prunePanelCache = (force = false) => {
    const now = Date.now();
    if (!force && now - lastPruned < 300000) return;
    lastPruned = now;
    const settings = GlobalStore().panelCache;
    try {
        for (let index = localStorage.length - 1; index >= 0; index--) {
            const key = localStorage.key(index);
            if (!key) continue;
            if (key.startsWith('1panel:search-state:v2:') || key.startsWith('cachedRoute')) {
                // Legacy entries have neither ownership nor timestamps and cannot be safely restored.
                localStorage.removeItem(key);
            } else if (key.startsWith(prefix)) {
                try {
                    if (validEntry(JSON.parse(localStorage.getItem(key)), settings, now)) continue;
                } catch {
                    /* Remove malformed entries too. */
                }
                localStorage.removeItem(key);
            }
        }
    } catch {
        /* Storage is optional. */
    }
};

export const readCachedPageSize = (key: string): number | undefined => {
    const value = readPanelCache<number>('pagination', panelCacheKey(`size:${key}`));
    return Number.isInteger(value) && value > 0 ? value : undefined;
};
