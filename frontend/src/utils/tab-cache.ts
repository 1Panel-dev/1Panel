import type { ComputedRef, InjectionKey } from 'vue';

export type KeepAlivePattern = Array<string | RegExp>;

export interface RouteCacheContext {
    include: ComputedRef<KeepAlivePattern>;
    tabCacheEnabled: ComputedRef<boolean>;
}

export const routeCacheContextKey: InjectionKey<RouteCacheContext> = Symbol('route-cache-context');

export const getTabCacheName = (path: string): string => {
    const encodedPath = Array.from(path)
        .map((character) => character.codePointAt(0)?.toString(16))
        .join('_');

    return `TabCache_${encodedPath}`;
};
