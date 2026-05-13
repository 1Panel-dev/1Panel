import { getUserInfo } from '@/api/modules/auth';
import { getEnterpriseUserInfo } from '@/extensions/xpack';
import { GlobalStore } from '@/store';
import type { RouteMeta } from 'vue-router';

type RouteAccessMeta = {
    adminOnly?: boolean;
    protectedRoleOnly?: boolean;
};

type RouteAccessTarget = {
    matched: Array<{
        meta?: RouteMeta & {
            permission?: string;
        };
    }>;
};

export const syncAuthInfo = async (currentNode?: string) => {
    const globalStore = GlobalStore();
    if (!globalStore.isEnterprise) {
        const res = await getUserInfo();
        return res.data;
    }
    const res = await getEnterpriseUserInfo(currentNode ?? globalStore.currentNode);
    globalStore.setAuthInfo({
        isAdmin: res.data.role === 'ADMIN',
        permissions: res.data.permissions || [],
        nodeRoles: res.data.nodeRoles || [],
    });
    return res.data;
};

export const hasPermission = (permission: string) => {
    return GlobalStore().hasPermission(permission);
};

export const hasRouteRoleAccess = (meta?: RouteMeta & RouteAccessMeta) => {
    const globalStore = GlobalStore();

    if (!meta) {
        return true;
    }
    if (globalStore.isAdmin) {
        return true;
    }
    if (meta.adminOnly && !globalStore.isAdmin) {
        return false;
    }
    if (meta.protectedRoleOnly) {
        return globalStore.isNodeAdmin;
    }
    return true;
};

export const hasRoutePermissionAccess = (route: RouteAccessTarget) => {
    const requiredPermissions = [
        ...new Set(
            route.matched
                .map((record) => record.meta?.permission)
                .filter((permission): permission is string => !!permission),
        ),
    ];
    return requiredPermissions.every((permission) => hasPermission(permission));
};

export const hasRouteAccess = (route: RouteAccessTarget) => {
    return route.matched.every((record) => hasRouteRoleAccess(record.meta)) && hasRoutePermissionAccess(route);
};
