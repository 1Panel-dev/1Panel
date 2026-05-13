import router from '@/routers';
import { useGlobalStore } from '@/composables/useGlobalStore';

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

export const hasManagePermissionAccess = (value?: PermissionBindingValue) => {
    const { globalStore, isAdmin, isNodeAdmin } = useGlobalStore();
    if (isAdmin.value || isNodeAdmin.value) {
        return true;
    }
    const permissions = toPermissionList(value).map(toManagePermission).filter(Boolean);
    if (permissions.length === 0) {
        return false;
    }
    return permissions.every((permission) => globalStore.hasPermission(permission));
};

export const hasPermissionAccess = (value?: PermissionBindingValue) => {
    const { globalStore, isAdmin, isNodeAdmin } = useGlobalStore();

    if (isAdmin.value || isNodeAdmin.value) {
        return true;
    }
    const permissions = toPermissionList(value);
    if (permissions.length === 0) {
        return false;
    }
    return permissions.every((permission) => globalStore.hasPermission(permission));
};
