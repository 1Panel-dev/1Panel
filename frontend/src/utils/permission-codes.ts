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
