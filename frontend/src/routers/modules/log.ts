import { Layout } from '@/routers/constant';

const logsRouter = {
    sort: 7,
    path: '/logs',
    component: Layout,
    redirect: '/logs',
    meta: {
        title: 'menu.logs',
        icon: 'p-log',
    },
    children: [
        {
            path: '',
            name: 'OperationLog',
            component: () => import('@/views/log/operation/index.vue'),
            hidden: true,
            meta: {
                activeMenu: '/logs',
            },
        },
        {
            path: 'login',
            name: 'LoginLog',
            component: () => import('@/views/log/login/index.vue'),
            hidden: true,
            meta: {
                activeMenu: '/logs',
            },
        },
        {
            path: 'system',
            name: 'SystemLog',
            component: () => import('@/views/log/system/index.vue'),
            hidden: true,
            meta: {
                activeMenu: '/logs',
            },
        },
    ],
};

export default logsRouter;
