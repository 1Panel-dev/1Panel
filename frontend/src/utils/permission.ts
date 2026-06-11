import router from '@/routers';
import { GlobalStore } from '@/store';
import { normalizeToManageCode } from '@/utils/permission-codes';

export type PermissionBindingValue = string | string[] | undefined;
export type PermissionMode = 'manage' | 'view';
export type PermissionAccessOptions = {
    nodeAdmin?: boolean;
};

const getRoutePermission = (): PermissionBindingValue => {
    const route = router.currentRoute.value;
    const metaPermission = route.meta?.permission;
    if (typeof metaPermission === 'string' && metaPermission) {
        return metaPermission;
    }
    if (Array.isArray(metaPermission)) {
        return metaPermission;
    }

    for (const record of [...route.matched].reverse()) {
        const permission = record.meta?.permission;
        if (typeof permission === 'string' && permission) {
            return permission;
        }
        if (Array.isArray(permission)) {
            return permission;
        }
    }

    return '';
};

export const toManagePermission = normalizeToManageCode;

export const toPermissionList = (value: PermissionBindingValue) => {
    if (Array.isArray(value)) {
        return value.filter(
            (permission): permission is string => typeof permission === 'string' && !!permission.trim(),
        );
    }
    if (typeof value === 'string' && value.trim()) {
        return [value];
    }
    const routePermission = getRoutePermission();
    if (Array.isArray(routePermission)) {
        return routePermission;
    }
    return routePermission ? [routePermission] : [];
};

const hasPermissionAccessByMode = (
    mode: PermissionMode,
    value?: PermissionBindingValue,
    options: PermissionAccessOptions = {},
) => {
    const globalStore = GlobalStore();
    if (options.nodeAdmin && globalStore.isNodeAdmin) {
        return true;
    }
    const permissions = toPermissionList(value);
    const normalizedPermissions = mode === 'manage' ? permissions.map(toManagePermission).filter(Boolean) : permissions;
    if (normalizedPermissions.length === 0) {
        return globalStore.isAdmin;
    }
    return normalizedPermissions.some((permission) => globalStore.hasPermission(permission));
};

export const hasManagePermissionAccess = (value?: PermissionBindingValue, options?: PermissionAccessOptions) => {
    return hasPermissionAccessByMode('manage', value, options);
};

export const hasPermissionAccess = (value?: PermissionBindingValue, options?: PermissionAccessOptions) => {
    return hasPermissionAccessByMode('view', value, options);
};
