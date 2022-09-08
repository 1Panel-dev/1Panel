import { Layout } from '@/routers/constant';

const monitorRouter = {
    sort: 2,
    path: '/monitors',
    component: Layout,
    redirect: '/monitor',
    meta: {
        title: 'menu.monitor',
        icon: 'monitor',
    },
    children: [
        {
            path: '/monitors/monitor',
            name: 'Monitor',
            component: () => import('@/views/monitor/index.vue'),
            meta: {
                requiresAuth: true,
                key: 'Monitor',
                title: 'menu.monitor',
                icon: 'Connection',
                activeMenu: '/monitors',
            },
        },
    ],
};

export default monitorRouter;
