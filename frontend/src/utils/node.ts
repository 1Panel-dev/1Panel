import { listNodeOptions } from '@/api/modules/setting';
import { GlobalStore } from '@/store';

const globalStore = GlobalStore();

export const changeToLocal = async () => {
    if (!globalStore.isMasterPro) {
        setDefaultNodeInfo();
        return;
    }
    await listNodeOptions('all')
        .then((res) => {
            if (!res) {
                setDefaultNodeInfo();
                return;
            }
            let nodes = res.data || [];
            if (nodes.length === 0) {
                setDefaultNodeInfo();
                return;
            }
            for (const item of nodes) {
                if (item.name === 'local') {
                    globalStore.currentNode = 'local';
                    globalStore.currentNodeAddr = item.addr;
                    return;
                }
            }
        })
        .catch(() => {
            setDefaultNodeInfo();
        });
};

export const setDefaultNodeInfo = () => {
    globalStore.currentNode = 'local';
    globalStore.currentNodeAddr = '127.0.0.1';
};
