import { Layout } from '@/routers/constant';

const systemConfigRouter = {
    sort: 8,
    path: '/configs',
    component: Layout,
    redirect: '/configs',
    meta: {
        icon: 'p-config',
        title: 'menu.systemConfig',
    },
    children: [
        {
            path: '/configs',
            name: 'SystemConfig',
            component: () => import('@/views/system-config/index.vue'),
            meta: {},
        },
    ],
};

export default systemConfigRouter;
