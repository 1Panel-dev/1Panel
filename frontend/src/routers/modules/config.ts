import { Layout } from '@/routers/constant';

const systemConfigRouter = {
    sort: 7,
    path: '/config',
    component: Layout,
    redirect: '/config',
    meta: {
        icon: 'p-config',
        title: 'menu.systemConfig',
    },
    children: [
        {
            path: '/config',
            name: 'SystemConfig',
            component: () => import('@/views/system-config/index.vue'),
            meta: {
                hidden: true,
                keepAlive: true,
            },
        },
    ],
};

export default systemConfigRouter;
