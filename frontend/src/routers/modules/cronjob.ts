import { Layout } from '@/routers/constant';

const cronRouter = {
    sort: 9,
    path: '/cronjobs',
    component: Layout,
    redirect: '/cronjobs',
    meta: {
        icon: 'p-plan',
        title: 'menu.cronjob',
    },
    children: [
        {
            path: '/cronjobs',
            name: 'Cronjob',
            component: () => import('@/views/cronjob/index.vue'),
            meta: {
                requiresAuth: false,
            },
        },
    ],
};

export default cronRouter;
