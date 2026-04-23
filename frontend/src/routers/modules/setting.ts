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
        adminOnly: true,
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
                        activeMenu: '/settings',
                        adminOnly: true,
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
                        activeMenu: '/settings',
                        adminOnly: true,
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
                        activeMenu: '/settings',
                        adminOnly: true,
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
                        activeMenu: '/settings',
                        adminOnly: true,
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
                        activeMenu: '/settings',
                        adminOnly: true,
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
                        activeMenu: '/settings',
                        adminOnly: true,
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
                        activeMenu: '/settings',
                        adminOnly: true,
                    },
                },
                {
                    path: 'expired',
                    name: 'Expired',
                    hidden: true,
                    component: () => import('@/views/setting/expired.vue'),
                    meta: {
                        activeMenu: '/settings',
                        ignoreTab: true,
                    },
                },
            ],
        },
    ],
};

export default settingRouter;
