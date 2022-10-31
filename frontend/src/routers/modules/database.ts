import { Layout } from '@/routers/constant';

const databaseRouter = {
    sort: 4,
    path: '/databases',
    component: Layout,
    redirect: '/databases',
    meta: {
        icon: 'p-database',
        title: 'menu.database',
    },
    children: [
        {
            path: '',
            name: 'Mysql',
            component: () => import('@/views/database/mysql/index.vue'),
            hidden: true,
            meta: {
                activeMenu: '/databases',
            },
        },
        {
            path: 'redis',
            name: 'Redis',
            component: () => import('@/views/database/redis/index.vue'),
            hidden: true,
            meta: {
                activeMenu: '/databases',
            },
        },
    ],
};

export default databaseRouter;
