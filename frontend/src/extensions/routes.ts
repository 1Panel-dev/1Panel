import type { RouteRecordRaw } from 'vue-router';

export function getXpackRoutes(): Record<string, RouteRecordRaw> {
    const xpackRoutes = import.meta.glob('@/xpack/routers/*.ts', { eager: true }) as Record<string, RouteRecordRaw>;
    const xpackEERoutes = import.meta.glob('@/xpack-ee/routers/*.ts', { eager: true }) as Record<
        string,
        RouteRecordRaw
    >;
    return {
        ...xpackRoutes,
        ...xpackEERoutes,
    };
}

export const getExtensionRoutes = getXpackRoutes;
