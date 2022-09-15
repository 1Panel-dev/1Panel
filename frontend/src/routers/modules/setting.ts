import { Layout } from '@/routers/constant';

const settingRouter = {
    sort: 3,
    path: '/settings',
    component: Layout,
    redirect: '/setting',
    meta: {
        title: 'menu.settings',
        icon: 'Setting',
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
