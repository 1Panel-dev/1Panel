import { Layout } from '@/routers/constant';

const settingRouter = {
    sort: 8,
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
        {
            path: '/setting/backupaccount',
            name: 'BackupAccount',
            component: () => import('@/views/setting/backup-account/index.vue'),
            hidden: true,
            meta: {
                key: 'Setting',
            },
        },
        {
            path: '/setting/about',
            name: 'About',
            component: () => import('@/views/setting/about/index.vue'),
            hidden: true,
            meta: {
                key: 'Setting',
            },
        },
        {
            path: '/setting/monitor',
            name: 'Monitor',
            component: () => import('@/views/setting/monitor/index.vue'),
            hidden: true,
            meta: {
                key: 'Setting',
            },
        },
        {
            path: '/setting/panel',
            name: 'Panel',
            component: () => import('@/views/setting/panel/index.vue'),
            hidden: true,
            meta: {
                key: 'Setting',
            },
        },
        {
            path: '/setting/safe',
            name: 'Safe',
            component: () => import('@/views/setting/safe/index.vue'),
            hidden: true,
            meta: {
                key: 'Setting',
            },
        },
        {
            path: '/expired',
            name: 'Expired',
            hidden: true,
            component: () => import('@/views/setting/expired.vue'),
            meta: {
                requiresAuth: true,
                key: 'Expired',
            },
        },
    ],
};

export default settingRouter;
