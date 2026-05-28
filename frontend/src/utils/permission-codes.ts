let masterOnlyPermissionCodes = new Set<string>();
let masterOnlyPermissionCodesKey = '';

export const setMasterOnlyPermissionCodes = (permissions?: string[]) => {
    const codes = (permissions || []).filter(Boolean);
    const key = codes.join('\n');
    if (key === masterOnlyPermissionCodesKey) {
        return;
    }
    masterOnlyPermissionCodesKey = key;
    masterOnlyPermissionCodes = new Set(codes);
};

export const toManageCode = (permission: string): string => {
    if (!permission) {
        return '';
    }
    return permission.endsWith('_view') ? permission.replace(/_view$/, '_manage') : '';
};

export const normalizeToManageCode = (permission: string): string => {
    if (!permission) {
        return '';
    }
    return toManageCode(permission) || permission;
};

export const isMasterOnlyPermissionCode = (permission: string): boolean => {
    return masterOnlyPermissionCodes.has(permission);
};
