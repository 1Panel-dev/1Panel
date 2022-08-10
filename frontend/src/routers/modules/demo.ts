import { Layout } from '@/routers/constant';

// demo
const demoRouter = {
    sort: 1,
    path: '/demos',
    component: Layout,
    redirect: '/demos/table',
    meta: {
        title: 'menu.demo',
    },
    children: [
        {
            path: '/demos/table',
            name: 'table',
            component: () => import('@/views/demos/table/index.vue'),
            meta: {
                keepAlive: true,
                requiresAuth: true,
                key: 'table',
            },
        },
    ],
};

export default demoRouter;
