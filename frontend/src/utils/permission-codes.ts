const masterOnlyPermissionCodes = new Set([
    'ai_proxy_view',
    'ai_proxy_manage',
    'ai_proxy_key_view',
    'ai_benchmark_view',
    'ai_benchmark_manage',
    'ai_skills_hub_view',
    'ai_skills_hub_manage',
    'xpack_ops_report_view',
    'xpack_ops_report_manage',
]);

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
