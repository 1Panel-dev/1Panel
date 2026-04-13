import type { RouteRecordRaw } from 'vue-router';

export function getXpackRoutes(): Record<string, RouteRecordRaw> {
    return import.meta.glob('@/xpack/routers/*.ts', { eager: true }) as Record<string, RouteRecordRaw>;
}
