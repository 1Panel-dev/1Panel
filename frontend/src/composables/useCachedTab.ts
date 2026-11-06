import { nextTick, onActivated, onDeactivated, onMounted, onScopeDispose, readonly, ref, watch } from 'vue';
import GlobalStore from '@/store/modules/global';
import { panelCacheKey, readPanelCache, writePanelCache } from '@/utils/panel-cache';

interface CachedTabOptions {
    key: () => string;
    values: () => string[];
    defaultValue: () => string;
    requestedValue?: () => string | undefined;
    onActivate: (value: string) => unknown;
}

export const useCachedTab = (options: CachedTabOptions) => {
    const store = GlobalStore();
    const activeTab = ref('');
    let active = true;
    let mounted = false;
    let version = 0;
    let restoredKey: string | undefined;
    const key = () => panelCacheKey(`inline-tabs:${options.key()}`);
    const valid = (value: unknown): value is string => typeof value === 'string' && options.values().includes(value);
    const restore = (useRequested = true) => {
        const requested = useRequested ? options.requestedValue?.() : undefined;
        const saved = readPanelCache<string>('tabs', key());
        restoredKey = key();
        activeTab.value = valid(requested)
            ? requested
            : valid(saved)
              ? saved
              : valid(options.defaultValue())
                ? options.defaultValue()
                : options.values()[0] || '';
    };
    const activate = async () => {
        const current = ++version;
        await nextTick();
        if (current !== version || !active || !mounted || !valid(activeTab.value)) return;
        options.onActivate(activeTab.value);
    };
    const select = (value: string) => {
        if (!active || !valid(value) || value === activeTab.value) return;
        activeTab.value = value;
        writePanelCache('tabs', key(), value);
        void activate();
    };
    watch(
        [
            key,
            () => store.panelCache.tabs,
            () => store.panelCache.duration,
            () => JSON.stringify(options.values()),
            options.defaultValue,
            () => options.requestedValue?.(),
        ],
        (values, previous) => {
            if (!active) return;
            // Reflecting the selected tab into the URL is not another selection.
            if (
                previous.length &&
                values.slice(0, 5).every((value, index) => value === previous[index]) &&
                values[5] === activeTab.value
            )
                return;
            // A URL left over from the previous node/account must not override its cache.
            restore(!previous.length || values[0] === previous[0]);
            if (mounted) void activate();
        },
        { immediate: true, flush: 'sync' },
    );
    onMounted(() => {
        mounted = true;
        void activate();
    });
    onDeactivated(() => {
        active = false;
        version++;
    });
    onActivated(() => {
        if (active) return;
        active = true;
        restore(restoredKey === key());
        void activate();
    });
    onScopeDispose(() => {
        active = false;
        version++;
    });
    return { activeTab: readonly(activeTab), select };
};
