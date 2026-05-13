import { Setting } from '@/api/interface/setting';
import { listNodeOptions, loadNodeByUser } from '@/api/modules/setting';
import { useGlobalStore } from '@/composables/useGlobalStore';

export const changeToLocal = async () => {
    const { currentNode, currentNodeAddr, isAdmin } = useGlobalStore();
    let nodes = await listNodes('all');
    if (nodes.length === 0) {
        setDefaultNodeInfo();
        return;
    }
    if (isAdmin.value) {
        for (const item of nodes) {
            if (item.name === 'local') {
                currentNode.value = 'local';
                currentNodeAddr.value = item.addr;
                return;
            }
        }
    }
    currentNode.value = nodes[0].name;
    currentNodeAddr.value = nodes[0].addr;
};

export async function listNodes(type: string): Promise<Array<Setting.NodeItem>> {
    const { isAdmin } = useGlobalStore();
    try {
        if (isAdmin.value) {
            const res = await listNodeOptions(type);
            return res.data || [];
        } else {
            const res = await loadNodeByUser();
            return res.data || [];
        }
    } catch (error) {
        return [];
    }
}

export const setDefaultNodeInfo = () => {
    const { currentNode, currentNodeAddr } = useGlobalStore();
    currentNode.value = 'local';
    currentNodeAddr.value = '127.0.0.1';
};
