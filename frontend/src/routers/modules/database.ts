import { Layout } from '@/routers/constant';

const databaseRouter = {
    sort: 4,
    path: '/databases',
    component: Layout,
    redirect: '/databases/mysql',
    meta: {
        icon: 'p-database',
        title: 'menu.database',
    },
    children: [
        {
            path: '/databases',
            name: 'Databases',
            redirect: '/databases/mysql',
            component: () => import('@/views/database/index.vue'),
            meta: {},
            children: [
                {
                    path: 'mysql',
                    name: 'MySQL',
                    component: () => import('@/views/database/mysql/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/databases',
                        requiresAuth: false,
                    },
                },
                {
                    path: 'redis',
                    name: 'Redis',
                    component: () => import('@/views/database/redis/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/databases',
                        requiresAuth: false,
                    },
                },
            ],
        },
    ],
};

export default databaseRouter;
