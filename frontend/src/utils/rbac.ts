import { getAuthInfo } from '@/api/modules/auth';
import { GlobalStore } from '@/store';

export const syncAuthInfo = async () => {
    const globalStore = GlobalStore();
    if (!globalStore.isXpackEE) {
        return;
    }
    const res = await getAuthInfo();
    globalStore.setAuthInfo({
        isAdmin: res.data.role === 'ADMIN',
        permissions: res.data.permissions || [],
        nodeScopes: res.data.nodeScopes || [],
        nodeRoles: res.data.nodeRoles || [],
    });
};

export const hasPermission = (permission: string) => {
    return GlobalStore().hasPermission(permission);
};
