<template>
    <router-view v-slot="{ Component, route }">
        <transition
            :appear="withTransition"
            :css="withTransition"
            :name="withTransition ? 'fade-transform' : undefined"
            :mode="withTransition ? 'out-in' : undefined"
        >
            <keep-alive :include="include">
                <component
                    :is="getRouteWrapper(route.path, route.name)"
                    :key="route.path"
                    :route-component="Component"
                    :route-props="routeProps"
                ></component>
            </keep-alive>
        </transition>
    </router-view>
</template>

<script setup lang="ts">
import {
    getTabCacheName,
    routeCacheContextKey,
    type KeepAlivePattern,
    type RouteCacheContext,
} from '@/utils/tab-cache';
import {
    computed,
    cloneVNode,
    defineComponent,
    inject,
    provide,
    type Component as VueComponent,
    type PropType,
    type VNode,
} from 'vue';
import type { RouteRecordName } from 'vue-router';

const props = withDefaults(
    defineProps<{
        keepAlive?: KeepAlivePattern | null;
        routeProps?: Record<string, unknown>;
        withTransition?: boolean;
    }>(),
    {
        withTransition: false,
    },
);

const inheritedCacheContext = inject(routeCacheContextKey, null);
const localCacheContext: RouteCacheContext = {
    include: computed<KeepAlivePattern>(() => {
        return props.keepAlive ?? [];
    }),
    tabCacheEnabled: computed(() => props.keepAlive !== undefined && props.keepAlive !== null),
};
const cacheContext = inheritedCacheContext || localCacheContext;

if (!inheritedCacheContext) {
    provide(routeCacheContextKey, cacheContext);
}

const include = cacheContext.include;
const routeWrappers = new Map<string, VueComponent>();
const getRouteWrapper = (path: string, routeName?: RouteRecordName | null) => {
    const wrapperName = cacheContext.tabCacheEnabled.value
        ? getTabCacheName(path)
        : String(routeName || getTabCacheName(path));
    let wrapper = routeWrappers.get(wrapperName);
    if (!wrapper) {
        wrapper = defineComponent({
            name: wrapperName,
            props: {
                routeComponent: {
                    type: Object as PropType<VNode>,
                    required: true,
                },
                routeProps: {
                    type: Object as PropType<Record<string, unknown>>,
                    required: false,
                },
            },
            setup(wrapperProps) {
                return () => {
                    return wrapperProps.routeProps
                        ? cloneVNode(wrapperProps.routeComponent, wrapperProps.routeProps)
                        : wrapperProps.routeComponent;
                };
            },
        });
        routeWrappers.set(wrapperName, wrapper);
    }
    return wrapper;
};
</script>
