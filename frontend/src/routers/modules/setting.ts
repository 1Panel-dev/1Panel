import { Layout } from '@/routers/constant';

const settingRouter = {
    sort: 12,
    path: '/settings',
    name: 'Setting-Menu',
    component: Layout,
    redirect: '/settings/panel',
    meta: {
        title: 'menu.settings',
        icon: 'p-config',
    },
    children: [
        {
            path: '/settings',
            name: 'Setting',
            redirect: '/settings/panel',
            component: () => import('@/views/setting/index.vue'),
            meta: {},
            children: [
                {
                    path: 'panel',
                    name: 'Panel',
                    component: () => import('@/views/setting/panel/index.vue'),
                    hidden: true,
                    meta: {
                        parent: 'menu.settings',
                        title: 'setting.panel',
                        requiresAuth: true,
                        activeMenu: '/settings',
                    },
                },
                {
                    path: 'alert',
                    name: 'Alert',
                    component: () => import('@/views/setting/alert/index.vue'),
                    hidden: true,
                    meta: {
                        parent: 'menu.settings',
                        title: 'xpack.alert.alertNotice',
                        requiresAuth: true,
                        activeMenu: '/settings',
                    },
                },
                {
                    path: 'backupaccount',
                    name: 'BackupAccount',
                    component: () => import('@/views/setting/backup-account/index.vue'),
                    hidden: true,
                    meta: {
                        parent: 'menu.settings',
                        title: 'setting.backupAccount',
                        requiresAuth: true,
                        activeMenu: '/settings',
                    },
                },
                {
                    path: 'license',
                    name: 'License',
                    component: () => import('@/views/setting/license/index.vue'),
                    hidden: true,
                    meta: {
                        parent: 'menu.settings',
                        title: 'setting.license',
                        requiresAuth: true,
                        activeMenu: '/settings',
                    },
                },
                {
                    path: 'about',
                    name: 'About',
                    component: () => import('@/views/setting/about/index.vue'),
                    hidden: true,
                    meta: {
                        parent: 'menu.settings',
                        title: 'setting.about',
                        requiresAuth: true,
                        activeMenu: '/settings',
                    },
                },
                {
                    path: 'safe',
                    name: 'Safe',
                    component: () => import('@/views/setting/safe/index.vue'),
                    hidden: true,
                    meta: {
                        parent: 'menu.settings',
                        title: 'setting.safe',
                        requiresAuth: true,
                        activeMenu: '/settings',
                    },
                },
                {
                    path: 'snapshot',
                    name: 'Snapshot',
                    hidden: true,
                    component: () => import('@/views/setting/snapshot/index.vue'),
                    meta: {
                        parent: 'menu.settings',
                        title: 'setting.snapshot',
                        requiresAuth: true,
                        activeMenu: '/settings',
                    },
                },
                {
                    path: 'expired',
                    name: 'Expired',
                    hidden: true,
                    component: () => import('@/views/setting/expired.vue'),
                    meta: {
                        requiresAuth: true,
                        activeMenu: 'Expired',
                    },
                },
            ],
        },
    ],
};

export default settingRouter;
