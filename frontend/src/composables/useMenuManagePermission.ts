import { computed } from 'vue';
import { useRoute } from 'vue-router';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { hasManagePermissionAccess, hasPermissionAccess, toManagePermission } from '@/utils/permission';

const getRoutePermission = (route: ReturnType<typeof useRoute>) => {
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

export const useMenuManagePermission = (permission?: string) => {
    const route = useRoute();
    const { isAdmin, isNodeAdmin } = useGlobalStore();

    const sourcePermission = computed(() => {
        if (permission) {
            return permission;
        }
        return getRoutePermission(route);
    });
    const managePermission = computed(() => {
        return toManagePermission(sourcePermission.value);
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

export const useCan = (permission?: string) => {
    return useMenuManagePermission(permission).hasPermission;
};
