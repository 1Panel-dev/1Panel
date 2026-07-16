import type { VNode } from 'vue';
import { hasManagePermissionAccess, hasPermissionAccess, type PermissionBindingValue } from '@/utils/permission';

export { flattenVNodes } from '@/components/shared/vnode';

export interface FuTableColumnConfig {
    key: string;
    label: string;
    prop?: string;
    show: boolean;
    fixed?: boolean | string;
}

export interface FuTableOperationButton<Row = any> {
    key?: string | number;
    label?: string | number;
    command?: string | number | object;
    click?: (row: Row) => void;
    disabled?: boolean | ((row: Row) => boolean);
    permission?: true | PermissionBindingValue;
    nodeAdmin?: boolean;
    show?: boolean | ((row: Row) => boolean);
    type?: string;
    icon?: any;
    divided?: boolean;
}

export const FU_TABLE_STORAGE_PREFIX = 'FU-T-';

export const isElTableColumnVNode = (vnode: VNode) => {
    const type = vnode.type as any;
    return type === 'el-table-column' || type?.name === 'ElTableColumn' || type?.__name === 'ElTableColumn';
};

export const getTableColumnKey = (vnode: VNode) => {
    const props = (vnode.props || {}) as Record<string, any>;
    return String(props.columnKey || props.prop || vnode.key || props.label || '');
};

export const resolveMaybeFn = <T, R>(value: R | ((row: T) => R), row: T): R => {
    if (typeof value === 'function') {
        return (value as (row: T) => R)(row);
    }
    return value;
};

export const isOperationVisible = <T>(button: FuTableOperationButton<T>, row: T) => {
    return resolveMaybeFn(button.show ?? true, row);
};

export const isOperationDisabled = <T>(button: FuTableOperationButton<T>, row: T) => {
    const permissionOptions = { nodeAdmin: button.nodeAdmin === true };
    const permissionDisabled =
        button.permission === true
            ? !hasManagePermissionAccess(undefined, permissionOptions)
            : button.permission !== undefined && !hasPermissionAccess(button.permission, permissionOptions);
    return permissionDisabled || Boolean(resolveMaybeFn(button.disabled ?? false, row));
};

export const updateArrayInPlace = <T>(target: T[], next: T[]) => {
    if (target.length === next.length && target.every((item, index) => item === next[index])) {
        return;
    }
    target.splice(0, target.length, ...next);
};
