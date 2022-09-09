import { Layout } from '@/routers/constant';

const databaseRouter = {
    sort: 4,
    path: '/database',
    component: Layout,
    redirect: '/database',
    meta: {
        icon: 'p-database',
        title: 'menu.database',
    },
    children: [
        {
            path: '/database',
            name: 'Database',
            component: () => import('@/views/database/index.vue'),
            meta: {},
        },
    ],
};

export default databaseRouter;
