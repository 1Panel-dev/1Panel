import { Layout } from '@/routers/constant';

const cronRouter = {
    sort: 6,
    path: '/cronjobs',
    component: Layout,
    redirect: '/cronjobs',
    meta: {
        icon: 'p-plan',
        title: 'menu.cron',
    },
    children: [
        {
            path: '/cronjobs',
            name: 'Cronjob',
            component: () => import('@/views/cronjob/index.vue'),
            meta: {},
        },
    ],
};

export default cronRouter;
