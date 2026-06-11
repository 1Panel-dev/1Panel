import { Comment, Fragment, Text, type VNode } from 'vue';
import type { PermissionBindingValue } from '@/utils/permission';

export interface FuTableColumnConfig {
    key: string;
    label: string;
    prop?: string;
    show: boolean;
    fixed?: boolean | string;
}

export interface FuTableOperationButton {
    label?: string | number;
    click?: (row: any) => void;
    disabled?: boolean | ((row: any) => boolean);
    permission?: true | PermissionBindingValue;
    nodeAdmin?: boolean;
    show?: boolean | ((row: any) => boolean);
    type?: string;
    icon?: any;
    divided?: boolean;
    [key: string]: any;
}

export const FU_TABLE_STORAGE_PREFIX = 'FU-T-';

export const flattenVNodes = (nodes: VNode[] = []): VNode[] => {
    const flattened: VNode[] = [];
    for (const node of nodes) {
        if (!node) {
            continue;
        }
        if (Array.isArray(node)) {
            flattened.push(...flattenVNodes(node));
            continue;
        }
        if (node.type === Fragment) {
            flattened.push(...flattenVNodes((node.children as VNode[]) || []));
            continue;
        }
        if (node.type === Comment || node.type === Text) {
            continue;
        }
        flattened.push(node);
    }
    return flattened;
};

export const getVNodeComponentName = (vnode: VNode) => {
    const type = vnode.type as any;
    return type?.name || type?.__name || type;
};

export const isElTableColumnVNode = (vnode: VNode) => {
    const type = vnode.type as any;
    return type === 'el-table-column' || type?.name === 'ElTableColumn' || type?.__name === 'ElTableColumn';
};

export const resolveMaybeFn = <T, R>(value: R | ((row: T) => R), row: T): R => {
    if (typeof value === 'function') {
        return (value as (row: T) => R)(row);
    }
    return value;
};

export const updateArrayInPlace = <T>(target: T[], next: T[]) => {
    if (target.length === next.length && target.every((item, index) => item === next[index])) {
        return;
    }
    target.splice(0, target.length, ...next);
};
