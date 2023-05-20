import { Layout } from '@/routers/constant';

const settingRouter = {
    sort: 8,
    path: '/settings',
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
                        requiresAuth: true,
                        activeMenu: 'Setting',
                    },
                },
                {
                    path: 'backupaccount',
                    name: 'BackupAccount',
                    component: () => import('@/views/setting/backup-account/index.vue'),
                    hidden: true,
                    meta: {
                        requiresAuth: true,
                        activeMenu: 'Setting',
                    },
                },
                {
                    path: 'about',
                    name: 'About',
                    component: () => import('@/views/setting/about/index.vue'),
                    hidden: true,
                    meta: {
                        requiresAuth: true,
                        activeMenu: 'Setting',
                    },
                },
                {
                    path: 'safe',
                    name: 'Safe',
                    component: () => import('@/views/setting/safe/index.vue'),
                    hidden: true,
                    meta: {
                        requiresAuth: true,
                        activeMenu: 'Setting',
                    },
                },
                {
                    path: 'snapshot',
                    name: 'Snapshot',
                    hidden: true,
                    component: () => import('@/views/setting/snapshot/index.vue'),
                    meta: {
                        requiresAuth: true,
                        activeMenu: 'Setting',
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
