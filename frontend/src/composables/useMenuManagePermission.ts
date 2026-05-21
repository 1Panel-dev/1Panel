import { computed } from 'vue';
import { useRoute } from 'vue-router';
import { useGlobalStore } from '@/composables/useGlobalStore';
import {
    hasManagePermissionAccess,
    hasPermissionAccess,
    toManagePermission,
    type PermissionBindingValue,
} from '@/utils/permission';

const getRoutePermission = (route: ReturnType<typeof useRoute>): PermissionBindingValue => {
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

const toManagePermissionValue = (permission: PermissionBindingValue) => {
    if (Array.isArray(permission)) {
        return permission.map(toManagePermission).filter(Boolean);
    }
    return toManagePermission(permission || '');
};

export const useMenuManagePermission = (permission?: PermissionBindingValue) => {
    const route = useRoute();
    const { isAdmin, isNodeAdmin } = useGlobalStore();

    const sourcePermission = computed(() => {
        if (permission) {
            return permission;
        }
        return getRoutePermission(route);
    });
    const managePermission = computed(() => {
        return toManagePermissionValue(sourcePermission.value);
    });
    const hasAdminManagePermission = computed(() => isAdmin.value || isNodeAdmin.value);
    const hasPermission = computed(() => {
        return hasPermissionAccess(sourcePermission.value || []);
    });
    const hasManagePermission = computed(() => {
        return hasManagePermissionAccess(managePermission.value || []);
    });

    return {
        managePermission,
        hasAdminManagePermission,
        hasPermission,
        hasManagePermission,
    };
};

export const useCan = (permission?: PermissionBindingValue) => {
    return useMenuManagePermission(permission).hasPermission;
};
