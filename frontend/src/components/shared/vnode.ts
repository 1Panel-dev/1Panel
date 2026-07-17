import { Comment, Fragment, Text, type VNode } from 'vue';

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
