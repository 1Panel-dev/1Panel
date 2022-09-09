import { Layout } from '@/routers/constant';

const containerRouter = {
    sort: 5,
    path: '/containers',
    component: Layout,
    redirect: '/containers',
    meta: {
        icon: 'p-docker',
        title: 'menu.container',
    },
    children: [
        {
            path: '/containers',
            name: 'Container',
            component: () => import('@/views/container/index.vue'),
            meta: {},
        },
    ],
};

export default containerRouter;
