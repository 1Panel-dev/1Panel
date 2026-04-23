import type { RouteRecordRaw } from 'vue-router';

export function getXpackRoutes(): Record<string, RouteRecordRaw> {
    const xpackRoutes = import.meta.glob('@/xpack/routers/*.ts', { eager: true }) as Record<string, RouteRecordRaw>;
    const xpackEERoutes = import.meta.glob('@/xpack-ee/routers/*.ts', { eager: true }) as Record<
        string,
        RouteRecordRaw
    >;
    const routes = {
        ...xpackRoutes,
        ...xpackEERoutes,
    };
    const rawRoutes: RouteRecordRaw[] = Object.keys(routes)
        .map((key) => routes[key]['default'])
        .sort((r1, r2) => {
            r1.sort ??= Number.MAX_VALUE;
            r2.sort ??= Number.MAX_VALUE;
            return r1.sort - r2.sort;
        });

    const routeMap = new Map<string, RouteRecordRaw>();
    rawRoutes.forEach((route) => {
        if (route?.name) {
            routeMap.set(route.name as string, route);
        }
    });

    const mergedRouteNames = new Set<string>();
    rawRoutes.forEach((route) => {
        const parentRouteName = route.meta?.parentRouteName as string | undefined;
        if (!parentRouteName) {
            return;
        }
        const parentRoute = routeMap.get(parentRouteName);
        if (!parentRoute) {
            return;
        }
        const parentChildren = Array.isArray(parentRoute.children) ? parentRoute.children : [];
        const insertAfter = route.meta?.insertAfter as string | undefined;
        const insertIndex = insertAfter ? parentChildren.findIndex((item) => item.name === insertAfter) + 1 : -1;
        if (insertIndex > 0) {
            parentChildren.splice(insertIndex, 0, route);
        } else {
            parentChildren.push(route);
        }
        parentRoute.children = parentChildren;
        mergedRouteNames.add(route.name as string);
    });

    const mergedRoutes: Record<string, RouteRecordRaw> = {};
    Object.keys(routes).forEach((key) => {
        const route = routes[key]['default'];
        if (!mergedRouteNames.has(route?.name as string)) {
            mergedRoutes[key] = routes[key];
        }
    });
    return mergedRoutes;
}

export const getExtensionRoutes = getXpackRoutes;
