import { Layout } from '@/routers/constant';

const logsRouter = {
    sort: 10,
    path: '/logs',
    component: Layout,
    redirect: '/logs/operation',
    meta: {
        title: 'menu.logs',
        icon: 'p-log',
    },
    children: [
        {
            path: '/logs',
            name: 'Log',
            redirect: '/logs/operation',
            component: () => import('@/views/log/index.vue'),
            meta: {},
            children: [
                {
                    path: 'operation',
                    name: 'OperationLog',
                    component: () => import('@/views/log/operation/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/logs',
                        requiresAuth: false,
                    },
                },
                {
                    path: 'login',
                    name: 'LoginLog',
                    component: () => import('@/views/log/login/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/logs',
                        requiresAuth: false,
                    },
                },
                {
                    path: 'website',
                    name: 'WebsiteLog',
                    component: () => import('@/views/log/website/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/logs',
                        requiresAuth: false,
                    },
                },
                {
                    path: 'system',
                    name: 'SystemLog',
                    component: () => import('@/views/log/system/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/logs',
                        requiresAuth: false,
                    },
                },
                {
                    path: 'ssh',
                    name: 'SSHLog2',
                    component: () => import('@/views/host/ssh/log/log.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/logs',
                        requiresAuth: false,
                    },
                },
            ],
        },
    ],
};

export default logsRouter;
