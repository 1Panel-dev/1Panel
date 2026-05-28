import { getUserInfo } from '@/api/modules/auth';
import { getEnterpriseUserInfo } from '@/extensions/xpack';
import type { RouteMeta } from 'vue-router';
import { GlobalStore } from '@/store';

export type PermissionMetaValue = string | string[];

type RouteAccessMeta = {
    adminOnly?: boolean;
    protectedRoleOnly?: boolean;
    permission?: PermissionMetaValue;
};

type RouteAccessTarget = {
    matched: Array<{
        meta?: RouteMeta & RouteAccessMeta;
    }>;
};

export const syncAuthInfo = async (currentNode?: string) => {
    const globalStore = GlobalStore();
    const storeCurrentNode = globalStore.currentNode;
    if (!globalStore.isEnterprise) {
        const res = await getUserInfo();
        globalStore.setAuthInfo({
            isAdmin: res.data.role === 'ADMIN',
            permissions: res.data.permissions || [],
            masterOnlyPermissions: res.data.masterOnlyPermissions || [],
            nodeRoles: res.data.nodeRoles || [],
        });
        return res.data;
    }
    const res = await getEnterpriseUserInfo(currentNode ?? storeCurrentNode);
    globalStore.setAuthInfo({
        isAdmin: res.data.role === 'ADMIN',
        permissions: res.data.permissions || [],
        masterOnlyPermissions: res.data.masterOnlyPermissions || [],
        nodeRoles: res.data.nodeRoles || [],
    });
    return res.data;
};

export const hasPermission = (permission: string) => {
    return GlobalStore().hasPermission(permission);
};

export const hasPermissionMetaAccess = (permission?: PermissionMetaValue) => {
    if (!permission) {
        return true;
    }
    if (Array.isArray(permission)) {
        const permissions = permission.filter(Boolean);
        return permissions.length === 0 || permissions.some((item) => hasPermission(item));
    }
    return hasPermission(permission);
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
    return route.matched.every((record) => hasPermissionMetaAccess(record.meta?.permission));
};

export const hasRouteAccess = (route: RouteAccessTarget) => {
    return route.matched.every((record) => hasRouteRoleAccess(record.meta)) && hasRoutePermissionAccess(route);
};
