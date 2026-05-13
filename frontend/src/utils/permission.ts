import router from '@/routers';
import { GlobalStore } from '@/store';

export type PermissionBindingValue = string | string[] | undefined;
export type PermissionMode = 'manage' | 'view';

const getRoutePermission = () => {
    const route = router.currentRoute.value;
    const metaPermission = route.meta?.permission;
    if (typeof metaPermission === 'string' && metaPermission) {
        return metaPermission;
    }

    for (const record of [...route.matched].reverse()) {
        const permission = record.meta?.permission;
        if (typeof permission === 'string' && permission) {
            return permission;
        }
    }

    return '';
};

export const toManagePermission = (permission: string) => {
    if (!permission) {
        return '';
    }
    return permission.endsWith('_view') ? permission.replace(/_view$/, '_manage') : permission;
};

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
    return routePermission ? [routePermission] : [];
};

const hasPermissionAccessByMode = (mode: PermissionMode, value?: PermissionBindingValue) => {
    const globalStore = GlobalStore();
    if (globalStore.isAdmin || globalStore.isNodeAdmin) {
        return true;
    }
    const permissions = toPermissionList(value);
    const normalizedPermissions = mode === 'manage' ? permissions.map(toManagePermission).filter(Boolean) : permissions;
    if (normalizedPermissions.length === 0) {
        return false;
    }
    return normalizedPermissions.every((permission) => globalStore.hasPermission(permission));
};

export const hasManagePermissionAccess = (value?: PermissionBindingValue) => {
    return hasPermissionAccessByMode('manage', value);
};

export const hasPermissionAccess = (value?: PermissionBindingValue) => {
    return hasPermissionAccessByMode('view', value);
};
