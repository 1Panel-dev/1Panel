import { Setting } from '@/api/interface/setting';
import { listNodeOptions, loadNodeByUser } from '@/api/modules/setting';
import { GlobalStore } from '@/store';

const getGlobalStore = () => GlobalStore();

export const changeToLocal = async () => {
    const globalStore = getGlobalStore();
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
    const globalStore = getGlobalStore();
    try {
        if (globalStore.isAdmin) {
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
    const globalStore = getGlobalStore();
    globalStore.currentNode = 'local';
    globalStore.currentNodeAddr = '127.0.0.1';
};
