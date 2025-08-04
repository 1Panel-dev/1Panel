import { Layout } from '@/routers/constant';

const logsRouter = {
    sort: 11,
    path: '/logs',
    name: 'Log-Menu',
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
                        parent: 'menu.logs',
                        title: 'logs.operation',
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
                        parent: 'menu.logs',
                        title: 'logs.login',
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
                        parent: 'menu.logs',
                        title: 'logs.websiteLog',
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
                        parent: 'menu.logs',
                        title: 'logs.system',
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
                        parent: 'menu.logs',
                        title: 'ssh.loginLogs',
                        activeMenu: '/logs',
                        requiresAuth: false,
                    },
                },
                {
                    path: 'task',
                    name: 'Task',
                    component: () => import('@/views/log/task/index.vue'),
                    hidden: true,
                    meta: {
                        parent: 'menu.logs',
                        title: 'logs.task',
                        activeMenu: '/logs',
                        requiresAuth: false,
                    },
                },
            ],
        },
    ],
};

export default logsRouter;
