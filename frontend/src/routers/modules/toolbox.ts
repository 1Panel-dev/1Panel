import { Layout } from '@/routers/constant';

const toolboxRouter = {
    sort: 8,
    path: '/toolbox',
    component: Layout,
    redirect: '/toolbox/supervisor',
    meta: {
        title: 'menu.toolbox',
        icon: 'p-toolbox',
    },
    children: [
        {
            path: '/toolbox',
            name: 'Toolbox',
            redirect: '/toolbox/device',
            component: () => import('@/views/toolbox/index.vue'),
            meta: {},
            children: [
                {
                    path: 'device',
                    name: 'Device',
                    component: () => import('@/views/toolbox/device/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/toolbox',
                        requiresAuth: false,
                    },
                },
                {
                    path: 'supervisor',
                    name: 'Supervisor',
                    component: () => import('@/views/toolbox/supervisor/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/toolbox',
                        requiresAuth: false,
                    },
                },
                {
                    path: 'clam',
                    name: 'Clam',
                    component: () => import('@/views/toolbox/clam/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/toolbox',
                        requiresAuth: false,
                    },
                },
                {
                    path: 'clam/setting',
                    name: 'Clam-Setting',
                    component: () => import('@/views/toolbox/clam/setting/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/toolbox',
                        requiresAuth: false,
                    },
                },
                {
                    path: 'ftp',
                    name: 'FTP',
                    component: () => import('@/views/toolbox/ftp/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/toolbox',
                        requiresAuth: false,
                    },
                },
                {
                    path: 'fail2ban',
                    name: 'Fail2ban',
                    component: () => import('@/views/toolbox/fail2ban/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/toolbox',
                        requiresAuth: false,
                    },
                },
                {
                    path: 'clean',
                    name: 'Clean',
                    component: () => import('@/views/toolbox/clean/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/toolbox',
                        requiresAuth: false,
                    },
                },
            ],
        },
    ],
};

export default toolboxRouter;
