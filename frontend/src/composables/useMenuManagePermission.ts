import { computed } from 'vue';
import { useRoute } from 'vue-router';
import { GlobalStore } from '@/store';

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

const toManagePermission = (permission: string) => {
    if (!permission) {
        return '';
    }
    return permission.endsWith('_view') ? permission.replace(/_view$/, '_manage') : permission;
};

export const useMenuManagePermission = (permission?: string) => {
    const route = useRoute();
    const globalStore = GlobalStore();

    const sourcePermission = computed(() => {
        if (permission) {
            return permission;
        }
        return getRoutePermission(route);
    });
    const managePermission = computed(() => {
        return toManagePermission(sourcePermission.value);
    });
    const hasAdminManagePermission = computed(() => globalStore.isAdmin || globalStore.isNodeAdmin);
    const hasPermission = computed(() => {
        if (hasAdminManagePermission.value) {
            return true;
        }
        if (!sourcePermission.value) {
            return false;
        }
        if (sourcePermission.value.endsWith('_view')) {
            return (
                globalStore.hasPermission(sourcePermission.value) ||
                globalStore.hasPermission(toManagePermission(sourcePermission.value))
            );
        }
        return globalStore.hasPermission(sourcePermission.value);
    });
    const hasManagePermission = computed(() => {
        if (hasAdminManagePermission.value) {
            return true;
        }
        if (!managePermission.value) {
            return false;
        }
        return globalStore.hasPermission(managePermission.value);
    });

    return {
        managePermission,
        hasAdminManagePermission,
        hasPermission,
        hasManagePermission,
    };
};
