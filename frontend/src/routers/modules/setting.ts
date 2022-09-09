import { Layout } from '@/routers/constant';

const settingRouter = {
    sort: 7,
    path: '/settings',
    component: Layout,
    redirect: '/setting',
    meta: {
        title: 'menu.settings',
        icon: 'p-config',
    },
    children: [
        {
            path: '/setting',
            name: 'Setting',
            component: () => import('@/views/setting/index.vue'),
            meta: {
                requiresAuth: true,
                key: 'Setting',
            },
        },
    ],
};

export default settingRouter;
