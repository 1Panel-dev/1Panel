import { Layout } from '@/routers/constant';

const planRouter = {
    sort: 5,
    path: '/plans',
    component: Layout,
    redirect: '/plans',
    meta: {
        icon: 'p-plan',
        title: 'menu.plan',
    },
    children: [
        {
            path: '/plans',
            name: 'Plan',
            component: () => import('@/views/plan/index.vue'),
            meta: {
                keepAlive: true,
            },
        },
    ],
};

export default planRouter;
