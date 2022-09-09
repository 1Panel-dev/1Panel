import { Layout } from '@/routers/constant';

const planRouter = {
    sort: 6,
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
            meta: {},
        },
    ],
};

export default planRouter;
