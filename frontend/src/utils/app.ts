import { jumpToPath } from './util';
import router from '@/routers';

export const jumpToInstall = (type: string, key: string) => {
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
