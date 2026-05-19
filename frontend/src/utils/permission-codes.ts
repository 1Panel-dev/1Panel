/**
 * Returns the manage-flavoured code that corresponds to a `_view` permission.
 * Returns an empty string if the input is not a `_view` code (including empty
 * input), so callers can decide whether to fall back to the original code.
 */
export const toManageCode = (permission: string): string => {
    if (!permission) {
        return '';
    }
    return permission.endsWith('_view') ? permission.replace(/_view$/, '_manage') : '';
};

/**
 * Normalises a single permission code to its manage flavour. `_view` codes are
 * upgraded to `_manage`; manage/other codes are kept as-is. Empty input is
 * preserved so existing filters behave the same.
 */
export const normalizeToManageCode = (permission: string): string => {
    if (!permission) {
        return '';
    }
    return toManageCode(permission) || permission;
};
