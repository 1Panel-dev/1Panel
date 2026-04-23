import { Layout } from '@/routers/constant';

const toolboxRouter = {
    sort: 9,
    path: '/toolbox',
    name: 'Toolbox-Menu',
    component: Layout,
    redirect: '/toolbox/supervisor',
    meta: {
        title: 'menu.toolbox',
        icon: 'p-toolbox',
        permission: 'toolbox_view',
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
                        parent: 'menu.toolbox',
                        title: 'toolbox.device.toolbox',
                        activeMenu: '/toolbox',
                        permission: 'toolbox_view',
                    },
                },
                {
                    path: 'supervisor',
                    name: 'Supervisor',
                    component: () => import('@/views/toolbox/supervisor/index.vue'),
                    hidden: true,
                    meta: {
                        parent: 'menu.toolbox',
                        title: 'menu.supervisor',
                        activeMenu: '/toolbox',
                        permission: 'toolbox_view',
                    },
                },
                {
                    path: 'clam',
                    name: 'Clam',
                    component: () => import('@/views/toolbox/clam/index.vue'),
                    hidden: true,
                    meta: {
                        parent: 'menu.toolbox',
                        title: 'toolbox.clam.clam',
                        activeMenu: '/toolbox',
                        permission: 'toolbox_view',
                    },
                },
                {
                    path: 'clam/setting',
                    name: 'Clam-Setting',
                    component: () => import('@/views/toolbox/clam/setting/index.vue'),
                    hidden: true,
                    meta: {
                        ignoreTab: true,
                        activeMenu: '/toolbox',
                        permission: 'toolbox_view',
                    },
                },
                {
                    path: 'ftp',
                    name: 'FTP',
                    component: () => import('@/views/toolbox/ftp/index.vue'),
                    hidden: true,
                    meta: {
                        parent: 'menu.toolbox',
                        title: 'FTP',
                        activeMenu: '/toolbox',
                        permission: 'toolbox_view',
                    },
                },
                {
                    path: 'fail2ban',
                    name: 'Fail2ban',
                    component: () => import('@/views/toolbox/fail2ban/index.vue'),
                    hidden: true,
                    meta: {
                        parent: 'menu.toolbox',
                        title: 'Fail2Ban',
                        activeMenu: '/toolbox',
                        permission: 'toolbox_view',
                    },
                },
                {
                    path: 'clean',
                    name: 'Clean',
                    component: () => import('@/views/toolbox/clean/index.vue'),
                    hidden: true,
                    meta: {
                        parent: 'menu.toolbox',
                        title: 'setting.diskClean',
                        activeMenu: '/toolbox',
                        permission: 'toolbox_view',
                    },
                },
            ],
        },
    ],
};

export default toolboxRouter;
