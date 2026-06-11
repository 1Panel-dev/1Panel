import { Setting } from '@/api/interface/setting';
import { listNodeOptions } from '@/api/modules/setting';
import { GlobalStore } from '@/store';

export const changeToLocal = async () => {
    const globalStore = GlobalStore();
    let nodes = await listNodes('all');
    if (nodes.length === 0) {
        setDefaultNodeInfo();
        return;
    }
    if (globalStore.isAdmin) {
        for (const item of nodes) {
            if (item.name === 'local') {
                globalStore.currentNode = 'local';
                globalStore.currentNodeAddr = item.addr;
                return;
            }
        }
    }
    globalStore.currentNode = nodes[0].name;
    globalStore.currentNodeAddr = nodes[0].addr;
};

export async function listNodes(type: string): Promise<Array<Setting.NodeItem>> {
    try {
        const res = await listNodeOptions(type);
        return res.data || [];
    } catch (error) {
        return [];
    }
}

export const setDefaultNodeInfo = () => {
    const globalStore = GlobalStore();
    globalStore.currentNode = 'local';
    globalStore.currentNodeAddr = '127.0.0.1';
};
