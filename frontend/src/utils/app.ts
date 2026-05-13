import { jumpToPath } from './router';
import router from '@/routers';
import { GlobalStore } from '@/store';

export const jumpToInstall = (type: string, key: string) => {
    const globalStore = GlobalStore();
    switch (type) {
        case 'php':
        case 'node':
        case 'java':
        case 'go':
        case 'python':
        case 'dotnet':
            jumpToPath(router, '/websites/runtimes/' + type);
            return true;
    }
    switch (key) {
        case 'openclaw':
            jumpToPath(router, '/ai/agents/agent');
            return true;
        case 'copaw':
            router.push({
                path: '/ai/agents/agent',
                query: {
                    uncached: 'true',
                    open: 'create',
                    agentType: 'copaw',
                },
            });
            return true;
        case 'hermes-agent':
            router.push({
                path: '/ai/agents/agent',
                query: {
                    uncached: 'true',
                    open: 'create',
                    agentType: 'hermes-agent',
                },
            });
            return true;
        case 'vllm':
            if (globalStore.isProductPro) {
                router.push({
                    path: '/ai/model/local',
                    query: {
                        tab: 'vllm',
                        uncached: 'true',
                        open: 'create',
                    },
                });
                return true;
            }
            return false;
        case 'mysql-cluster':
            jumpToPath(router, '/xpack/cluster/mysql');
            return true;
        case 'redis-cluster':
            jumpToPath(router, '/xpack/cluster/redis');
            return true;
        case 'postgresql-cluster':
            jumpToPath(router, '/xpack/cluster/postgres');
            return true;
    }
    return false;
};
